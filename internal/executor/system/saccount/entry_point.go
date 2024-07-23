package saccount

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"strings"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/interfaces"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/solidity/ientry_point"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/solidity/ientry_point_client"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/solidity/ownable"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

const (
	CreateAccountGas        = 10000
	RecoveryAccountOwnerGas = 10000
	MaxCallGasLimit         = 50000

	NoGasCallsKey = "no_gas_calls"
)

var EntryPointBuildConfig = &common.SystemContractBuildConfig[*EntryPoint]{
	Name:    "saccount_entry_point",
	Address: common.EntryPointContractAddr,
	AbiStr:  ientry_point_client.BindingContractMetaData.ABI,
	Constructor: func(systemContractBase common.SystemContractBase) *EntryPoint {
		return &EntryPoint{
			SystemContractBase: systemContractBase,
			Ownable:            Ownable{SystemContractBase: systemContractBase},
			NonceManager:       NewNonceManager(systemContractBase),
			StakeManager:       NewStakeManager(systemContractBase),
		}
	},
}

type MemoryUserOp struct {
	Sender               ethcommon.Address
	Nonce                *big.Int
	CallGasLimit         *big.Int
	VerificationGasLimit *big.Int
	PreVerificationGas   *big.Int
	Paymaster            ethcommon.Address
	MaxFeePerGas         *big.Int
	MaxPriorityFeePerGas *big.Int
}

type UserOpInfo struct {
	MUserOp        MemoryUserOp
	UserOpHash     [32]byte
	Prefund        *big.Int
	Context        []byte
	PreOpGas       *big.Int
	ValidationData *interfaces.Validation
}

var _ interfaces.IEntryPoint = (*EntryPoint)(nil)

type EntryPoint struct {
	common.SystemContractBase

	*NonceManager
	*StakeManager
	*common.ReentrancyGuard

	Ownable

	// NoGasCalls is a list of system contract call method id that are allowed to be called without gas cost.
	// EntryPoint will pay the gas for these calls.
	// Owner can update this list.
	NoGasCalls *common.VMSlot[[]byte]
}

func (ep *EntryPoint) GenesisInit(genesis *repo.GenesisConfig) error {
	ep.Init(ethcommon.HexToAddress(genesis.SmartAccountAdmin))
	ep.NoGasCalls = common.NewVMSlot[[]byte](ep.StateAccount, NoGasCallsKey)
	// initialize the list of no gas calls
	return ep.NoGasCalls.Put(setPasskeySig)
}

func (ep *EntryPoint) GetNoGasCalls() ([]byte, error) {
	return ep.NoGasCalls.MustGet()
}

func (ep *EntryPoint) SetNoGasCalls(calls []byte) error {
	if !ep.CheckOwner() {
		return ep.Revert(&ownable.ErrorOwnableUnauthorizedAccount{})
	}

	return ep.NoGasCalls.Put(calls)
}

func (ep *EntryPoint) SetContext(ctx *common.VMContext) {
	ep.SystemContractBase.SetContext(ctx)
	ep.NonceManager.SetContext(ctx)
	ep.StakeManager.SetContext(ctx)
	ep.ReentrancyGuard = common.NewReentrancyGuard()
	ep.Ownable.SetContext(ctx)

	ep.NoGasCalls = common.NewVMSlot[[]byte](ep.StateAccount, NoGasCallsKey)

	ep.Logger.Infof("EntryPoint SetContext")
}

func (ep *EntryPoint) compensate(beneficiary ethcommon.Address, amount *big.Int) error {
	if beneficiary == (ethcommon.Address{}) {
		return errors.New("AA90 invalid beneficiary")
	}

	// TODO: maybe no need to compensate?
	// collected used gas is not equals to bundler used gas
	// ep.addBalance(beneficiary, amount)
	return nil
}

// Execute a batch of UserOperations.
// @param ops the operations to execute
// @param beneficiary the address to receive the fees
// Attention: HandleOps must keep non reentrant
func (ep *EntryPoint) HandleOps(ops []interfaces.UserOperation, beneficiary ethcommon.Address) error {
	ep.Logger.Infof("entrypoint handleOps, ops: %+v", ops)
	if err := ep.Enter(); err != nil {
		return err
	}
	defer ep.Exit()

	opInfos := make([]UserOpInfo, len(ops))
	for i := 0; i < len(ops); i++ {
		validateData, paymasterValidationData, err := ep.validatePrepayment(big.NewInt(int64(i)), &ops[i], &opInfos[i])
		if err != nil {
			return err
		}
		err = ep.validateAccountAndPaymasterValidationData(big.NewInt(int64(i)), validateData, paymasterValidationData, &opInfos[i])
		if err != nil {
			return err
		}
	}

	ep.EmitEvent(&ientry_point.EventBeforeExecution{})

	collected := big.NewInt(0)
	for i := 0; i < len(ops); i++ {
		gasUsed, err := ep.executeUserOp(big.NewInt(int64(i)), &ops[i], &opInfos[i])
		if err != nil {
			return err
		}

		collected.Add(collected, gasUsed)
	}

	return ep.compensate(beneficiary, collected)
}

func (ep *EntryPoint) SimulateHandleOp(op interfaces.UserOperation, target ethcommon.Address, targetCallData []byte) error {
	ep.Logger.Infof("entrypoint simulateHandleOp, op %+v, callData: %x, paymasterAndData: %x, target: %s, targetCallData: %x\n", op, op.CallData, op.PaymasterAndData, target.Hex(), targetCallData)
	var opInfo UserOpInfo
	if err := ep.simulationOnlyValidations(&op); err != nil {
		return err
	}
	validation, pmValidationData, err := ep.validatePrepayment(big.NewInt(0), &op, &opInfo)
	if err != nil {
		return err
	}
	data := interfaces.IntersectTimeRange(validation, pmValidationData)

	targetSuccess := true
	var targetResult []byte
	paid, err := ep.executeUserOp(big.NewInt(0), &op, &opInfo)
	if err != nil {
		return err
	}

	if target != (ethcommon.Address{}) {
		result, _, err := ep.CrossCallEVMContract(op.CallGasLimit, target, targetCallData)
		if err != nil {
			targetSuccess = false
		}

		targetResult = result
	}

	return ep.Revert(&ientry_point.ErrorExecutionResult{
		PreOpGas:      opInfo.PreOpGas,
		Paid:          paid,
		ValidAfter:    big.NewInt(int64(data.ValidAfter)),
		ValidUntil:    big.NewInt(int64(data.ValidUntil)),
		TargetSuccess: targetSuccess,
		TargetResult:  targetResult,
	})
}

func (ep *EntryPoint) SimulateValidation(userOp interfaces.UserOperation) error {
	ep.Logger.Infof("entrypoint simulateValidation")
	var outOpInfo UserOpInfo
	if err := ep.simulationOnlyValidations(&userOp); err != nil {
		return err
	}
	validation, pmValidationData, err := ep.validatePrepayment(big.NewInt(0), &userOp, &outOpInfo)
	if err != nil {
		return err
	}
	data := interfaces.IntersectTimeRange(validation, pmValidationData)

	paymasterInfo := ep.getStakeInfo(outOpInfo.MUserOp.Paymaster)
	senderInfo := ep.getStakeInfo(outOpInfo.MUserOp.Sender)
	var factoryInfo *ientry_point.IStakeManagerStakeInfo
	var factory ethcommon.Address
	initCode := userOp.InitCode
	if len(initCode) >= 20 {
		factory = ethcommon.BytesToAddress(initCode[0:20])
	}
	factoryInfo = ep.getStakeInfo(factory)

	SigFailed := data.SigValidation == interfaces.SigValidationFailed
	returnInfo := &ientry_point.IEntryPointReturnInfo{
		PreOpGas:         outOpInfo.PreOpGas,
		Prefund:          outOpInfo.Prefund,
		SigFailed:        SigFailed,
		ValidAfter:       big.NewInt(int64(data.ValidAfter)),
		ValidUntil:       big.NewInt(int64(data.ValidUntil)),
		PaymasterContext: outOpInfo.Context,
	}

	return ep.Revert(&ientry_point.ErrorValidationResult{
		ReturnInfo:    *returnInfo,
		SenderInfo:    *senderInfo,
		FactoryInfo:   *factoryInfo,
		PaymasterInfo: *paymasterInfo,
	})
}

// validatePrepayment validate account and paymaster (if defined).
// also make sure total validation doesn't exceed verificationGasLimit
// this method is called off-chain (simulateValidation()) and on-chain (from handleOps)
// @param opIndex the index of this userOp into the "opInfos" array
// @param userOp the userOp to validate
func (ep *EntryPoint) validatePrepayment(opIndex *big.Int, userOp *interfaces.UserOperation, outOpInfo *UserOpInfo) (validationData *interfaces.Validation, paymasterValidationData *big.Int, err error) {
	mUserOp := &outOpInfo.MUserOp
	if err := copyUserOpToMemory(userOp, mUserOp); err != nil {
		return nil, nil, errors.New(err.Error())
	}
	outOpInfo.UserOpHash = ep.getUserOpHash(userOp)

	maxGasValues := new(big.Int).Or(mUserOp.PreVerificationGas, mUserOp.VerificationGasLimit).Or(mUserOp.CallGasLimit, userOp.MaxFeePerGas).Or(userOp.MaxPriorityFeePerGas, big.NewInt(0))
	// validate all numeric values in userOp are well below 128 bit, so they can safely be added
	// and multiplied without causing overflow
	if maxGasValues.BitLen() > 128 {
		return nil, nil, errors.New("AA94 gas values overflow")
	}

	requiredPreFund := getRequiredPrefund(mUserOp)
	gasUsedByValidateAccountPrepayment, validationData, err := ep.validateAccountPrepayment(opIndex, userOp, outOpInfo, requiredPreFund)
	if err != nil {
		return nil, nil, err
	}

	pass, err := ep.validateAndUpdateNonce(mUserOp.Sender, mUserOp.Nonce)
	if err != nil {
		return nil, nil, err
	}
	if !pass {
		return nil, nil, ep.Revert(&ientry_point.ErrorFailedOp{
			OpIndex: opIndex,
			Reason:  "AA25 invalid account nonce",
		})
	}

	var context []byte
	paymasterValidationData = big.NewInt(0)
	if mUserOp.Paymaster != (ethcommon.Address{}) {
		context, paymasterValidationData, err = ep.validatePaymasterPrepayment(opIndex, userOp, outOpInfo, requiredPreFund, gasUsedByValidateAccountPrepayment)
		if err != nil {
			return nil, nil, err
		}
	}

	gasUsed := gasUsedByValidateAccountPrepayment
	if userOp.VerificationGasLimit.Cmp(gasUsed) == -1 {
		return nil, nil, ep.Revert(&ientry_point.ErrorFailedOp{
			OpIndex: opIndex,
			Reason:  "AA40 over verificationGasLimit",
		})
	}

	outOpInfo.Prefund = requiredPreFund
	outOpInfo.Context = context
	outOpInfo.ValidationData = validationData

	outOpInfo.PreOpGas = new(big.Int).Add(gasUsed, userOp.PreVerificationGas)
	return validationData, paymasterValidationData, nil
}

func (ep *EntryPoint) validateAccountAndPaymasterValidationData(opIndex *big.Int, validation *interfaces.Validation, paymasterValidationData *big.Int, opInfo *UserOpInfo) error {
	sigValidationResult := validation.SigValidation
	outOfTimeRange := ep.Ctx.CurrentEVM.Context.Time > validation.ValidUntil || ep.Ctx.CurrentEVM.Context.Time < validation.ValidAfter
	if validation.ValidUntil == 0 {
		validation.ValidUntil = interfaces.MaxUint48.Uint64()
		outOfTimeRange = false
	}
	if sigValidationResult == interfaces.SigValidationFailed {
		return ep.Revert(&ientry_point.ErrorFailedOp{
			OpIndex: opIndex,
			Reason:  "AA24 signature error",
		})
	}
	if outOfTimeRange {
		return ep.Revert(&ientry_point.ErrorFailedOp{
			OpIndex: opIndex,
			Reason:  "AA22 expired or not due",
		})
	}

	pmSigValidationResult, outOfTimeRange := ep.getValidationData(paymasterValidationData)
	if pmSigValidationResult == interfaces.SigValidationFailed {
		return ep.Revert(&ientry_point.ErrorFailedOp{
			OpIndex: opIndex,
			Reason:  "AA34 signature error",
		})
	}
	if outOfTimeRange {
		return ep.Revert(&ientry_point.ErrorFailedOp{
			OpIndex: opIndex,
			Reason:  "AA32 paymaster expired or not due",
		})
	}

	// when remain limit is not nil, it means the session key is used, then check
	if validation.RemainingLimit != nil && validation.RemainingLimit.Cmp(opInfo.PreOpGas) < 0 {
		return ep.Revert(&ientry_point.ErrorFailedOp{
			OpIndex: opIndex,
			Reason:  "account remain limit not enough",
		})
	}
	return nil
}

func (ep *EntryPoint) getValidationData(validationData *big.Int) (sigValidationResult uint, outOfTimeRange bool) {
	if validationData.Cmp(big.NewInt(interfaces.SigValidationSucceeded)) == 0 {
		return interfaces.SigValidationSucceeded, false
	}

	data := interfaces.ParseValidationData(validationData)
	outOfTimeRange = ep.Ctx.CurrentEVM.Context.Time > data.ValidUntil || ep.Ctx.CurrentEVM.Context.Time < data.ValidAfter
	if data.ValidUntil == interfaces.MaxUint48.Uint64() {
		outOfTimeRange = false
	}
	return data.SigValidation, outOfTimeRange
}

func (ep *EntryPoint) validateAccountPrepayment(opIndex *big.Int, op *interfaces.UserOperation, opInfo *UserOpInfo, requiredPrefund *big.Int) (gasUsedByValidateAccountPrepayment *big.Int, validationData *interfaces.Validation, err error) {
	mUserOp := &opInfo.MUserOp
	sender := mUserOp.Sender
	usedGas, err := ep.createSenderIfNeeded(opIndex, opInfo, op.InitCode)
	if err != nil {
		return big.NewInt(int64(usedGas)), nil, err
	}
	paymaster := mUserOp.Paymaster
	if paymaster == (ethcommon.Address{}) {
		if !ep.isNoGasUserOp(op) {
			// if not no gas user op call, need check pay prefund
			depositInfo := ep.GetDepositInfo(sender)
			if depositInfo.Deposit.Cmp(requiredPrefund) == -1 {
				return big.NewInt(int64(usedGas)), nil, ep.Revert(&ientry_point.ErrorFailedOp{
					OpIndex: opIndex,
					Reason:  "AA21 didn't pay prefund",
				})
			}
		}

		// if no paymaster set, just use sender balance
		// first, sub gas, after execute userOp, compensate left to sender
		if err := ep.subBalanceByUserOp(op, sender, requiredPrefund); err != nil {
			return big.NewInt(int64(usedGas)), nil, ep.Revert(&ientry_point.ErrorFailedOp{
				OpIndex: opIndex,
				Reason:  err.Error(),
			})
		}
	}

	// validate user operation
	sa := SmartAccountBuildConfig.BuildWithAddress(ep.CrossCallSystemContractContext(), sender).SetRemainingGas(mUserOp.VerificationGasLimit)
	validationData, err = sa.validateUserOp(op, opInfo.UserOpHash, big.NewInt(0))
	if err != nil {
		return big.NewInt(int64(usedGas)), nil, ep.Revert(&ientry_point.ErrorFailedOp{
			OpIndex: opIndex,
			Reason:  fmt.Sprintf("AA23 reverted: %s", err.Error()),
		})
	}

	return big.NewInt(int64(usedGas)), validationData, nil
}

func (ep *EntryPoint) validatePaymasterPrepayment(opIndex *big.Int, op *interfaces.UserOperation, opInfo *UserOpInfo, requiredPrefund, gasUsedByValidateAccountPrepayment *big.Int) (context []byte, validationData *big.Int, err error) {
	mUserOp := opInfo.MUserOp
	verificationGasLimit := mUserOp.VerificationGasLimit
	if verificationGasLimit.Cmp(gasUsedByValidateAccountPrepayment) == -1 {
		return nil, nil, errors.New("AA41 too little verificationGas")
	}

	paymasterAddr := mUserOp.Paymaster
	paymaster, err := ep.getPaymaster(paymasterAddr)
	if err != nil {
		return nil, nil, err
	}
	paymasterInfo := ep.GetDepositInfo(paymasterAddr)

	if paymasterInfo.Deposit.Cmp(requiredPrefund) == -1 {
		return nil, nil, ep.Revert(&ientry_point.ErrorFailedOp{
			OpIndex: opIndex,
			Reason:  "AA31 paymaster deposit too low",
		})
	}
	// first, sub gas, after execute userOp, compensate left to paymaster
	if err := ep.subBalanceByUserOp(op, paymasterAddr, requiredPrefund); err != nil {
		return nil, nil, ep.Revert(&ientry_point.ErrorFailedOp{
			OpIndex: opIndex,
			Reason:  err.Error(),
		})
	}

	context, validationData, err = paymaster.ValidatePaymasterUserOp(*op, opInfo.UserOpHash, requiredPrefund)
	if err != nil {
		return nil, nil, ep.Revert(&ientry_point.ErrorFailedOp{
			OpIndex: opIndex,
			Reason:  fmt.Sprintf("AA33 reverted: %s", err.Error()),
		})
	}

	return context, validationData, nil
}

func (ep *EntryPoint) getPaymaster(paymasterAddr ethcommon.Address) (interfaces.IPaymaster, error) {
	if paymasterAddr == ethcommon.HexToAddress(common.VerifyingPaymasterContractAddr) {
		return VerifyingPaymasterBuildConfig.Build(ep.CrossCallSystemContractContext()), nil
	} else if paymasterAddr == ethcommon.HexToAddress(common.TokenPaymasterContractAddr) {
		return TokenPaymasterBuildConfig.Build(ep.CrossCallSystemContractContext()), nil
	} else {
		return nil, errors.New("paymaster not found")
	}
}

// executeUserOp execute a user op
// @param opIndex index into the opInfo array
// @param userOp the userOp to execute
// @param opInfo the opInfo filled by validatePrepayment for this userOp.
// @return actualGasCost the total amount this userOp paid.
func (ep *EntryPoint) executeUserOp(opIndex *big.Int, userOp *interfaces.UserOperation, opInfo *UserOpInfo) (actualGasCost *big.Int, err error) {
	return ep.innerHandleOp(opIndex, userOp, opInfo, opInfo.Context)
}

// innerHandleOp is inner function to handle a UserOperation.
func (ep *EntryPoint) innerHandleOp(opIndex *big.Int, userOp *interfaces.UserOperation, opInfo *UserOpInfo, context []byte) (actualGasCost *big.Int, err error) {
	callData := userOp.CallData
	mUserOp := opInfo.MUserOp
	mode := interfaces.OpSucceeded
	usedGas := big.NewInt(0)
	totalValue := big.NewInt(0)
	if len(callData) > 0 {
		// check if call smart account method, directly call. otherwise, call evm
		if len(callData) < 4 {
			return nil, ep.Revert(&ientry_point.ErrorFailedOp{
				OpIndex: opIndex,
				Reason:  "callData length not enough",
			})
		}

		callReturnRes, err := ep.CallSmartAccount(mUserOp.Sender, callData, mUserOp.CallGasLimit)
		if err != nil {
			if !strings.Contains(err.Error(), "execute smart account callWithValue failed") {
				return nil, ep.Revert(&ientry_point.ErrorFailedOp{
					OpIndex: opIndex,
					Reason:  err.Error(),
				})
			}
			ep.EmitEvent(&ientry_point.EventUserOperationRevertReason{
				UserOpHash:   opInfo.UserOpHash,
				Sender:       mUserOp.Sender,
				Nonce:        mUserOp.Nonce,
				RevertReason: []byte(err.Error()),
			})
			mode = interfaces.OpReverted
		}

		if callReturnRes != nil {
			totalValue.Add(totalValue, callReturnRes.TotalUsedValue)
			usedGas.Add(usedGas, big.NewInt(int64(callReturnRes.UsedGas)))
		}
	}

	ep.Logger.Infof("handle op, opIndex: %s, mode: %v, usedGas: %s, preOpGas: %s", opIndex.Text(10), mode, usedGas.Text(10), opInfo.PreOpGas.Text(10))
	actualGas := new(big.Int).Add(usedGas, opInfo.PreOpGas)
	return ep.handlePostOp(opIndex, mode, userOp, opInfo, context, actualGas, totalValue)
}

// nolint
// handlePostOp process post-operation.
// called just after the callData is executed.
// if a paymaster is defined and its validation returned a non-empty context, its postOp is called.
// the excess amount is refunded to the account (or paymaster - if it was used in the request)
// @param opIndex index in the batch
// @param mode - whether is called from innerHandleOp, or outside (postOpReverted)
// @param opInfo userOp fields and info collected during validation
// @param context the context returned in validatePaymasterUserOp
func (ep *EntryPoint) handlePostOp(opIndex *big.Int, mode interfaces.PostOpMode, userOp *interfaces.UserOperation, opInfo *UserOpInfo, context []byte, actualGas *big.Int, totalValue *big.Int) (*big.Int, error) {
	var refundAddress ethcommon.Address
	mUserOp := opInfo.MUserOp
	gasPrice := getUserOpGasPrice(&mUserOp)

	paymaster := mUserOp.Paymaster
	actualGasCost := new(big.Int).Mul(actualGas, gasPrice)
	if paymaster == (ethcommon.Address{}) {
		refundAddress = mUserOp.Sender
	} else {
		refundAddress = paymaster
		if len(context) > 0 {
			paymaster, err := ep.getPaymaster(paymaster)
			if err != nil {
				return nil, ep.Revert(&ientry_point.ErrorFailedOp{
					OpIndex: opIndex,
					Reason:  err.Error(),
				})
			}
			if err := paymaster.PostOp(mode, context, actualGasCost); err != nil {
				return nil, ep.Revert(&ientry_point.ErrorFailedOp{
					OpIndex: opIndex,
					Reason:  fmt.Sprintf("AA50 postOp reverted: %s", err.Error()),
				})
			}
		}
	}

	sa := SmartAccountBuildConfig.BuildWithAddress(ep.CrossCallSystemContractContext(), mUserOp.Sender).SetRemainingGas(mUserOp.VerificationGasLimit)
	if err := sa.postUserOp(opInfo.ValidationData, actualGasCost, totalValue); err != nil {
		return nil, ep.Revert(&ientry_point.ErrorFailedOp{
			OpIndex: opIndex,
			Reason:  fmt.Sprintf("post user op reverted: %s", err.Error()),
		})
	}

	if opInfo.Prefund.Cmp(actualGasCost) == -1 {
		ep.Logger.Errorf("prefund is below actual gas cost, prefund: %s, actual gas cost: %s, actual gas: %s, gas price: %s", opInfo.Prefund.Text(10), actualGasCost.Text(10), actualGas.Text(10), gasPrice.Text(10))
		return nil, ep.Revert(&ientry_point.ErrorFailedOp{
			OpIndex: opIndex,
			Reason:  "AA51 prefund below actualGasCost",
		})
	}
	refund := new(big.Int).Sub(opInfo.Prefund, actualGasCost)
	if err := ep.addBalanceByUserOp(userOp, refundAddress, refund); err != nil {
		return nil, ep.Revert(&ientry_point.ErrorFailedOp{
			OpIndex: opIndex,
			Reason:  fmt.Sprintf("post user op reverted: %s", err.Error()),
		})
	}
	success := mode == interfaces.OpSucceeded

	ep.EmitEvent(&ientry_point.EventUserOperationEvent{
		UserOpHash:    opInfo.UserOpHash,
		Sender:        mUserOp.Sender,
		Paymaster:     paymaster,
		Nonce:         mUserOp.Nonce,
		Success:       success,
		ActualGasCost: actualGasCost,
		ActualGasUsed: actualGas,
	})
	return actualGasCost, nil
}

func (ep *EntryPoint) GetUserOpHash(userOp interfaces.UserOperation) ethcommon.Hash {
	return interfaces.GetUserOpHash(&userOp, ep.EthAddress, ep.Ctx.CurrentEVM.ChainConfig().ChainID)
}

func (ep *EntryPoint) getUserOpHash(userOp *interfaces.UserOperation) [32]byte {
	return interfaces.GetUserOpHash(userOp, ep.EthAddress, ep.Ctx.CurrentEVM.ChainConfig().ChainID)
}

func (ep *EntryPoint) createSenderIfNeeded(opIndex *big.Int, opInfo *UserOpInfo, initCode []byte) (uint64, error) {
	var usedGas uint64
	if len(initCode) > 0 {
		sender := opInfo.MUserOp.Sender
		account := ep.Ctx.StateLedger.GetOrCreateAccount(types.NewAddress(sender.Bytes()))
		if account.Code() != nil {
			return 0, ep.Revert(&ientry_point.ErrorFailedOp{
				OpIndex: opIndex,
				Reason:  "AA10 sender already constructed",
			})
		}

		var sender1 ethcommon.Address
		sender1, usedGas, _ = ep.createSender(opInfo.MUserOp.VerificationGasLimit, initCode)
		if sender1 == (ethcommon.Address{}) {
			return usedGas, ep.Revert(&ientry_point.ErrorFailedOp{
				OpIndex: opIndex,
				Reason:  "AA13 initCode failed or OOG",
			})
		}
		if sender1 != sender {
			return usedGas, ep.Revert(&ientry_point.ErrorFailedOp{
				OpIndex: opIndex,
				Reason:  "AA14 initCode must return sender",
			})
		}
		if account.Code() == nil {
			return usedGas, ep.Revert(&ientry_point.ErrorFailedOp{
				OpIndex: opIndex,
				Reason:  "AA15 initCode must create sender",
			})
		}
		factory := ethcommon.BytesToAddress(initCode[0:20])
		// post AccountDeployed event
		ep.EmitEvent(&ientry_point.EventAccountDeployed{
			UserOpHash: opInfo.UserOpHash,
			Sender:     sender,
			Factory:    factory,
			Paymaster:  opInfo.MUserOp.Paymaster,
		})
	}

	return usedGas, nil
}

func (ep *EntryPoint) createSender(gas *big.Int, initCode []byte) (ethcommon.Address, uint64, error) {
	// initCode must contains factory address, method signature, owner address bytes(32), salt bytes(32)
	if len(initCode) < 20+4+32+32 {
		return ethcommon.Address{}, 0, fmt.Errorf("createSender, initCode length not enough: %d", len(initCode))
	}

	if gas.Cmp(big.NewInt(CreateAccountGas)) == -1 {
		return ethcommon.Address{}, 0, fmt.Errorf("createSender, gas not enough: %s, gas must bigger than %d", gas.Text(10), CreateAccountGas)
	}

	factoryAddr := ethcommon.BytesToAddress(initCode[0:20])
	// initCallData is method sig and packed arguments
	initCallData := initCode[20:]

	ep.Logger.Infof("create sender, init code: %x, factoryAddr: %s", initCode, factoryAddr.Hex())

	// if factoryAddr is system contract, directly call
	if factoryAddr.Hex() == common.AccountFactoryContractAddr {
		factory := SmartAccountFactoryBuildConfig.Build(ep.CrossCallSystemContractContext())
		// discard method signature
		callData := initCallData[4:]
		owner := ethcommon.BytesToAddress(callData[0:32])
		salt := new(big.Int).SetBytes(callData[32:64])
		// guardian
		var guardian ethcommon.Address
		if len(callData) >= 64+32 {
			guardian = ethcommon.BytesToAddress(callData[64 : 64+32])
		}
		accountAddr, err := factory.CreateAccount2(owner, salt, guardian)
		if err != nil {
			return ethcommon.Address{}, 0, err
		}
		return accountAddr, CreateAccountGas, nil
	}

	return ethcommon.Address{}, 0, fmt.Errorf("only support factory address: %s", common.AccountFactoryContractAddr)
}

func (ep *EntryPoint) GetSenderAddress(initCode []byte) error {
	sender, _, _ := ep.createSender(big.NewInt(300000), initCode)
	return ep.Revert(&ientry_point.ErrorSenderAddressResult{
		Sender: sender,
	})
}

// HandleAccountRecovery handle account recovery
// Attention: HandleAccountRecovery must keep non reentrant
func (ep *EntryPoint) HandleAccountRecovery(ops []interfaces.UserOperation, beneficiary ethcommon.Address) error {
	ep.Logger.Infof("entrypoint HandleAccountRecovery, ops: %+v", ops)
	if err := ep.Enter(); err != nil {
		return err
	}
	defer ep.Exit()

	opInfos := make([]UserOpInfo, len(ops))
	for i := 0; i < len(ops); i++ {
		// judge if account exist
		sender := types.NewAddress(ops[i].Sender.Bytes())
		account := ep.Ctx.StateLedger.GetOrCreateAccount(sender)
		if code := account.Code(); len(code) == 0 {
			return errors.New("smart account is not initialized")
		}

		// validate guardian signature
		sa := SmartAccountBuildConfig.BuildWithAddress(ep.CrossCallSystemContractContext(), ops[i].Sender).SetRemainingGas(ops[i].VerificationGasLimit)
		err := sa.ValidateGuardianSignature(ops[i].Signature, ep.getUserOpHash(&ops[i]))
		if err != nil {
			return ep.Revert(&ientry_point.ErrorFailedOp{
				OpIndex: big.NewInt(int64(i)),
				Reason:  fmt.Sprintf("validate guardian signature reverted: %s", err.Error()),
			})
		}

		_, _, err = ep.validatePrepayment(big.NewInt(int64(i)), &ops[i], &opInfos[i])
		if err != nil {
			return err
		}
	}

	ep.EmitEvent(&ientry_point.EventBeforeExecution{})

	collected := big.NewInt(0)
	for i := 0; i < len(ops); i++ {
		op := ops[i]
		sa := SmartAccountBuildConfig.BuildWithAddress(ep.CrossCallSystemContractContext(), op.Sender).SetRemainingGas(op.CallGasLimit)
		if len(op.CallData) < 20 {
			return ep.Revert(&ientry_point.ErrorFailedOp{
				OpIndex: big.NewInt(int64(i)),
				Reason:  "callData is less than 20 bytes, should be contain address",
			})
		}

		// 20 bytes for owner address
		owner := ethcommon.BytesToAddress(op.CallData[:20])
		if err := sa.ResetAndLockOwner(owner); err != nil {
			return ep.Revert(&ientry_point.ErrorFailedOp{
				OpIndex: big.NewInt(int64(i)),
				Reason:  fmt.Sprintf("reset owner reverted: %s", err.Error()),
			})
		}

		actualGas := big.NewInt(int64(RecoveryAccountOwnerGas))
		actualGasCost, err := ep.handlePostOp(big.NewInt(int64(i)), interfaces.OpSucceeded, &op, &opInfos[i], []byte(""), actualGas, big.NewInt(0))
		if err != nil {
			return err
		}

		collected.Add(collected, actualGasCost)
	}

	return ep.compensate(beneficiary, collected)
}

func (ep *EntryPoint) simulationOnlyValidations(userOp *interfaces.UserOperation) error {
	if err := ep.validateSenderAndPaymaster(userOp.InitCode, userOp.Sender, userOp.PaymasterAndData); err != nil {
		return ep.Revert(&ientry_point.ErrorFailedOp{
			OpIndex: big.NewInt(int64(0)),
			Reason:  err.Error(),
		})
	}

	return nil
}

func (ep *EntryPoint) validateSenderAndPaymaster(initCode []byte, sender ethcommon.Address, paymasterAndData []byte) error {
	account := ep.Ctx.StateLedger.GetOrCreateAccount(types.NewAddress(sender.Bytes()))
	if len(initCode) == 0 && account.Code() == nil {
		// it would revert anyway. but give a meaningful message
		return errors.New("AA20 account not deployed")
	}
	if len(paymasterAndData) >= 20 {
		paymaster := types.NewAddress(paymasterAndData[0:20])
		paymasterAccount := ep.Ctx.StateLedger.GetOrCreateAccount(paymaster)
		if paymasterAccount.Code() == nil {
			// it would revert anyway. but give a meaningful message
			return errors.New("AA30 paymaster not deployed")
		}
	}
	return nil
}

func (ep *EntryPoint) WithdrawTo(withdrawAddress ethcommon.Address, withdrawAmount *big.Int) error {
	if !ep.CheckOwner() {
		return ep.Revert(&ownable.ErrorOwnableUnauthorizedAccount{})
	}

	depositInfo := ep.GetDepositInfo(ep.EthAddress)
	if depositInfo.Deposit.Cmp(withdrawAmount) < 0 {
		return common.NewRevertStringError("Withdraw amount too large")
	}
	_, _, err := ep.CrossCallEVMContractWithValue(big.NewInt(MaxCallGasLimit), withdrawAmount, withdrawAddress, nil)
	return err
}

// GetDepositInfo overide StakeManager
func (ep *EntryPoint) GetDepositInfo(account ethcommon.Address) *interfaces.DepositInfo {
	info := ep.StakeManager.GetDepositInfo(account)
	info.Deposit = ep.Ctx.StateLedger.GetBalance(types.NewAddress(account.Bytes()))
	return info
}

// BalanceOf overide StakeManager
func (ep *EntryPoint) BalanceOf(account ethcommon.Address) *big.Int {
	return ep.Ctx.StateLedger.GetBalance(types.NewAddress(account.Bytes()))
}

// if userOp call method id in NoGasCalls, will use EntryPoint to pay
func (ep *EntryPoint) addBalanceByUserOp(userOp *interfaces.UserOperation, account ethcommon.Address, value *big.Int) error {
	if ep.isNoGasUserOp(userOp) {
		ep.StateAccount.AddBalance(value)
		return nil
	}

	return ep.addBalance(account, value)
}

func (ep *EntryPoint) subBalanceByUserOp(userOp *interfaces.UserOperation, account ethcommon.Address, value *big.Int) error {
	if ep.isNoGasUserOp(userOp) {
		depositInfo := ep.GetDepositInfo(ep.EthAddress)
		if depositInfo.Deposit.Cmp(value) < 0 {
			return errors.New("EntryPoint account balance too low")
		}

		ep.StateAccount.SubBalance(value)
		return nil
	}

	return ep.subBalance(account, value)
}

func (ep *EntryPoint) isNoGasUserOp(userOp *interfaces.UserOperation) bool {
	isExist, noGasCallsBytes, _ := ep.NoGasCalls.Get()
	if isExist {
		// parse method id from userOp's InitCode and CallData
		var initMethodId, callMethodId []byte
		if len(userOp.InitCode) > 20+4 {
			initMethodId = userOp.InitCode[20 : 20+4]
		}

		if len(userOp.CallData) > 4 {
			callMethodId = userOp.CallData[:4]
		}

		for start := 0; start+4 <= len(noGasCallsBytes); start = start + 4 {
			if bytes.Equal(noGasCallsBytes[start:start+4], initMethodId) || bytes.Equal(noGasCallsBytes[start:start+4], callMethodId) {
				return true
			}
		}
	}
	return false
}

func (ep *EntryPoint) addBalance(account ethcommon.Address, value *big.Int) error {
	depositInfo := ep.GetDepositInfo(ep.EthAddress)
	if depositInfo.Deposit.Cmp(value) < 0 {
		return errors.New("EntryPoint account balance too low")
	}

	ep.StateAccount.SubBalance(value)
	ep.Ctx.StateLedger.AddBalance(types.NewAddress(account.Bytes()), value)
	return nil
}

func (ep *EntryPoint) subBalance(account ethcommon.Address, value *big.Int) error {
	depositInfo := ep.GetDepositInfo(account)
	if depositInfo.Deposit.Cmp(value) < 0 {
		return fmt.Errorf("account %s balance too low", account.Hex())
	}

	ep.Ctx.StateLedger.SubBalance(types.NewAddress(account.Bytes()), value)
	ep.StateAccount.AddBalance(value)
	return nil
}

func (ep *EntryPoint) CallSmartAccount(account ethcommon.Address, callData []byte, callGasLimit *big.Int) (*CallReturnResult, error) {
	sa := SmartAccountBuildConfig.BuildWithAddress(ep.CrossCallSystemContractContext(), account).SetRemainingGas(callGasLimit)

	callReturnRes, err := CallMethod(callData, sa)
	if err != nil {
		return nil, err
	}

	return &callReturnRes, nil
}

// TODO: use evm context gas price?
func getUserOpGasPrice(mUserOp *MemoryUserOp) *big.Int {
	// no basefee, direct return MaxFeePerGas
	// return math.BigMin(mUserOp.MaxFeePerGas, mUserOp.MaxPriorityFeePerGas)
	return mUserOp.MaxFeePerGas
}

func copyUserOpToMemory(userOp *interfaces.UserOperation, mUserOp *MemoryUserOp) error {
	mUserOp.Sender = ethcommon.BytesToAddress(userOp.Sender.Bytes())
	mUserOp.Nonce = new(big.Int).SetBytes(userOp.Nonce.Bytes())
	mUserOp.CallGasLimit = new(big.Int).SetBytes(userOp.CallGasLimit.Bytes())
	mUserOp.VerificationGasLimit = new(big.Int).SetBytes(userOp.VerificationGasLimit.Bytes())
	mUserOp.PreVerificationGas = new(big.Int).SetBytes(userOp.PreVerificationGas.Bytes())
	mUserOp.MaxFeePerGas = new(big.Int).SetBytes(userOp.MaxFeePerGas.Bytes())
	mUserOp.MaxPriorityFeePerGas = new(big.Int).SetBytes(userOp.MaxPriorityFeePerGas.Bytes())
	paymasterAndData := userOp.PaymasterAndData
	if len(paymasterAndData) > 0 {
		if len(paymasterAndData) < 20 {
			return errors.New("AA93 invalid paymasterAndData")
		}
		mUserOp.Paymaster = ethcommon.BytesToAddress(paymasterAndData[:20])
	} else {
		mUserOp.Paymaster = ethcommon.Address{}
	}
	return nil
}

func getRequiredPrefund(mUserOp *MemoryUserOp) *big.Int {
	// when using a Paymaster, the verificationGasLimit is used also to as a limit for the postOp call.
	// our security model might call postOp eventually twice
	mul := big.NewInt(1)
	if mUserOp.Paymaster != (ethcommon.Address{}) {
		mul = big.NewInt(3)
	}
	// requiredGas = mUserOp.callGasLimit + mUserOp.verificationGasLimit * mul + mUserOp.preVerificationGas
	requiredGas := new(big.Int).Add(mUserOp.CallGasLimit, new(big.Int).Mul(mUserOp.VerificationGasLimit, mul))
	requiredGas.Add(requiredGas, mUserOp.PreVerificationGas)
	requiredPrefund := new(big.Int).Mul(requiredGas, mUserOp.MaxFeePerGas)

	return requiredPrefund
}
