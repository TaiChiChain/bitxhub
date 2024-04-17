package saccount

import (
	"errors"
	"fmt"
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/axiomesh/axiom-kit/intutil"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/interfaces"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/solidity/ientry_point"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

const (
	CreateAccountGas        = 10000
	RecoveryAccountOwnerGas = 10000
	MaxCallGasLimit         = 50000
)

var EntryPointBuildConfig = &common.SystemContractBuildConfig[*EntryPoint]{
	Name:    "saccount_entry_point",
	Address: common.EntryPointContractAddr,
	AbiStr:  ientry_point.BindingContractMetaData.ABI,
	Constructor: func(systemContractBase common.SystemContractBase) *EntryPoint {
		return &EntryPoint{
			SystemContractBase: systemContractBase,
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
	MUserOp       MemoryUserOp
	UserOpHash    []byte
	Prefund       *big.Int
	Context       []byte
	PreOpGas      *big.Int
	UseSessionKey bool
}

var _ interfaces.IEntryPoint = (*EntryPoint)(nil)

type EntryPoint struct {
	common.SystemContractBase

	*NonceManager
	*StakeManager
	*common.ReentrancyGuard
}

func (ep *EntryPoint) GenesisInit(genesis *repo.GenesisConfig) error {
	return nil
}

func (ep *EntryPoint) SetContext(ctx *common.VMContext) {
	ep.SystemContractBase.SetContext(ctx)
	ep.NonceManager.SetContext(ctx)
	ep.StakeManager.SetContext(ctx)
	ep.ReentrancyGuard = common.NewReentrancyGuard()

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

	ep.EmitEvent("BeforeExecution")

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

	return interfaces.ExecutionResult(opInfo.PreOpGas, paid, big.NewInt(int64(data.ValidAfter)), big.NewInt(int64(data.ValidUntil)), targetSuccess, targetResult)
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
	var factoryInfo *interfaces.StakeInfo
	var factory ethcommon.Address
	initCode := userOp.InitCode
	if len(initCode) >= 20 {
		factory = ethcommon.BytesToAddress(initCode[0:20])
	}
	factoryInfo = ep.getStakeInfo(factory)

	SigFailed := data.SigValidation == interfaces.SigValidationFailed
	returnInfo := &interfaces.ReturnInfo{
		PreOpGas:         outOpInfo.PreOpGas,
		Prefund:          outOpInfo.Prefund,
		SigFailed:        SigFailed,
		ValidAfter:       big.NewInt(int64(data.ValidAfter)),
		ValidUntil:       big.NewInt(int64(data.ValidUntil)),
		PaymasterContext: outOpInfo.Context,
	}

	return interfaces.ValidationResult(returnInfo, senderInfo, factoryInfo, paymasterInfo)
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
		return nil, nil, interfaces.FailedOp(opIndex, "AA25 invalid account nonce")
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
		return nil, nil, interfaces.FailedOp(opIndex, "AA40 over verificationGasLimit")
	}

	outOpInfo.Prefund = requiredPreFund
	outOpInfo.Context = context
	outOpInfo.UseSessionKey = false
	// if use session key, set UseSessionKey true
	if validationData != nil && validationData.RemainingLimit != nil {
		outOpInfo.UseSessionKey = true
	}
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
		return interfaces.FailedOp(opIndex, "AA24 signature error")
	}
	if outOfTimeRange {
		return interfaces.FailedOp(opIndex, "AA22 expired or not due")
	}

	pmSigValidationResult, outOfTimeRange := ep.getValidationData(paymasterValidationData)
	if pmSigValidationResult == interfaces.SigValidationFailed {
		return interfaces.FailedOp(opIndex, "AA34 signature error")
	}
	if outOfTimeRange {
		return interfaces.FailedOp(opIndex, "AA32 paymaster expired or not due")
	}

	// when remain limit is not nil, it means the session key is used, then check
	if validation.RemainingLimit != nil && validation.RemainingLimit.Cmp(opInfo.PreOpGas) < 0 {
		return interfaces.FailedOp(opIndex, "account remain limit not enough")
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
		depositInfo := ep.GetDepositInfo(sender)
		if depositInfo.Deposit.Cmp(requiredPrefund) == -1 {
			return big.NewInt(int64(usedGas)), nil, interfaces.FailedOp(opIndex, "AA21 didn't pay prefund")
		}

		// if no paymaster set, just use sender balance
		// first, sub gas, after execute userOp, compensate left to sender
		ep.subBalance(sender, requiredPrefund)
	}

	// validate user operation
	sa := SmartAccountBuildConfig.BuildWithAddress(ep.CrossCallSystemContractContext(), sender).SetRemainingGas(mUserOp.VerificationGasLimit)
	validationData, err = sa.validateUserOp(op, opInfo.UserOpHash, big.NewInt(0))
	if err != nil {
		return big.NewInt(int64(usedGas)), nil, interfaces.FailedOp(opIndex, fmt.Sprintf("AA23 reverted: %s", err.Error()))
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
		return nil, nil, interfaces.FailedOp(opIndex, "AA31 paymaster deposit too low")
	}
	// first, sub gas, after execute userOp, compensate left to paymaster
	ep.subBalance(paymasterAddr, requiredPrefund)

	context, validationData, err = paymaster.ValidatePaymasterUserOp(*op, opInfo.UserOpHash, requiredPrefund)
	if err != nil {
		return nil, nil, interfaces.FailedOp(opIndex, fmt.Sprintf("AA33 reverted: %s", err.Error()))
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
	return ep.innerHandleOp(opIndex, userOp.CallData, opInfo, opInfo.Context)
}

// innerHandleOp is inner function to handle a UserOperation.
func (ep *EntryPoint) innerHandleOp(opIndex *big.Int, callData []byte, opInfo *UserOpInfo, context []byte) (actualGasCost *big.Int, err error) {
	mUserOp := opInfo.MUserOp
	mode := interfaces.OpSucceeded
	usedGas := big.NewInt(0)
	totalValue := big.NewInt(0)
	if len(callData) > 0 {
		// check if call smart account method, directly call. otherwise, call evm
		if len(callData) < 4 {
			return nil, interfaces.FailedOp(opIndex, "callData length not enough")
		}

		sa := SmartAccountBuildConfig.BuildWithAddress(ep.CrossCallSystemContractContext(), mUserOp.Sender).SetRemainingGas(mUserOp.CallGasLimit)
		isInnerMethod, callGas, totalUsedValue, err := JudgeOrCallInnerMethod(callData, sa)
		if err != nil {
			ep.emitUserOperationRevertReason(opInfo.UserOpHash, mUserOp.Sender, mUserOp.Nonce, []byte(err.Error()))
			mode = interfaces.OpReverted
		}

		if totalUsedValue != nil {
			totalValue.Add(totalValue, totalUsedValue)
		}

		usedGas.Add(usedGas, big.NewInt(int64(callGas)))
		if !isInnerMethod {
			// call evm if not inner method
			_, callGas, err := ep.CrossCallEVMContract(mUserOp.CallGasLimit, mUserOp.Sender, callData)
			if err != nil {
				ep.EmitEvent("UserOperationRevertReason", opInfo.UserOpHash, mUserOp.Sender, mUserOp.Nonce, []byte(err.Error()))
				mode = interfaces.OpReverted
			}

			usedGas.Add(usedGas, big.NewInt(int64(callGas)))
		}
	}

	ep.Logger.Infof("handle op, opIndex: %s, mode: %s, usedGas: %s, preOpGas: %s", opIndex.Text(10), mode, usedGas.Text(10), opInfo.PreOpGas.Text(10))
	actualGas := new(big.Int).Add(usedGas, opInfo.PreOpGas)
	return ep.handlePostOp(opIndex, mode, opInfo, context, actualGas, totalValue)
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
func (ep *EntryPoint) handlePostOp(opIndex *big.Int, mode interfaces.PostOpMode, opInfo *UserOpInfo, context []byte, actualGas *big.Int, totalValue *big.Int) (*big.Int, error) {
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
				return nil, interfaces.FailedOp(opIndex, err.Error())
			}
			if err := paymaster.PostOp(mode, context, actualGasCost); err != nil {
				return nil, interfaces.FailedOp(opIndex, fmt.Sprintf("AA50 postOp reverted: %s", err.Error()))
			}
		}
	}

	sa := SmartAccountBuildConfig.BuildWithAddress(ep.CrossCallSystemContractContext(), mUserOp.Sender).SetRemainingGas(mUserOp.VerificationGasLimit)
	if err := sa.postUserOp(opInfo.UseSessionKey, actualGasCost, totalValue); err != nil {
		return nil, interfaces.FailedOp(opIndex, fmt.Sprintf("post user op reverted: %s", err.Error()))
	}

	if opInfo.Prefund.Cmp(actualGasCost) == -1 {
		ep.Logger.Errorf("prefund is below actual gas cost, prefund: %s, actual gas cost: %s, actual gas: %s, gas price: %s", opInfo.Prefund.Text(10), actualGasCost.Text(10), actualGas.Text(10), gasPrice.Text(10))
		return nil, interfaces.FailedOp(opIndex, "AA51 prefund below actualGasCost")
	}
	refund := new(big.Int).Sub(opInfo.Prefund, actualGasCost)
	ep.addBalance(refundAddress, refund)
	success := mode == interfaces.OpSucceeded

	ep.EmitEvent("UserOperationEvent", opInfo.UserOpHash, mUserOp.Sender, paymaster, mUserOp.Nonce, success, actualGasCost, actualGas)
	return actualGasCost, nil
}

func (ep *EntryPoint) GetUserOpHash(userOp interfaces.UserOperation) ethcommon.Hash {
	return interfaces.GetUserOpHash(&userOp, ep.EthAddress, ep.Ctx.CurrentEVM.ChainConfig().ChainID)
}

func (ep *EntryPoint) getUserOpHash(userOp *interfaces.UserOperation) []byte {
	return interfaces.GetUserOpHash(userOp, ep.EthAddress, ep.Ctx.CurrentEVM.ChainConfig().ChainID).Bytes()
}

func (ep *EntryPoint) createSenderIfNeeded(opIndex *big.Int, opInfo *UserOpInfo, initCode []byte) (uint64, error) {
	var usedGas uint64
	if len(initCode) > 0 {
		sender := opInfo.MUserOp.Sender
		account := ep.Ctx.StateLedger.GetOrCreateAccount(types.NewAddress(sender.Bytes()))
		if account.Code() != nil {
			return 0, interfaces.FailedOp(opIndex, "AA10 sender already constructed")
		}

		var sender1 ethcommon.Address
		sender1, usedGas, _ = ep.createSender(opInfo.MUserOp.VerificationGasLimit, initCode)
		if sender1 == (ethcommon.Address{}) {
			return usedGas, interfaces.FailedOp(opIndex, "AA13 initCode failed or OOG")
		}
		if sender1 != sender {
			return usedGas, interfaces.FailedOp(opIndex, "AA14 initCode must return sender")
		}
		if account.Code() == nil {
			return usedGas, interfaces.FailedOp(opIndex, "AA15 initCode must create sender")
		}
		factory := ethcommon.BytesToAddress(initCode[0:20])
		// post AccountDeployed event
		ep.EmitEvent("AccountDeployed", opInfo.UserOpHash, sender, factory, opInfo.MUserOp.Paymaster)
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
		sa, err := factory.CreateAccount(owner, salt, guardian)
		if err != nil {
			return ethcommon.Address{}, 0, err
		}
		return sa.EthAddress, CreateAccountGas, nil
	}

	return ethcommon.Address{}, 0, fmt.Errorf("only support factory address: %s", common.AccountFactoryContractAddr)
}

func (ep *EntryPoint) GetSenderAddress(initCode []byte) error {
	sender, _, _ := ep.createSender(big.NewInt(300000), initCode)
	return interfaces.SenderAddressResult(sender)
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
			return interfaces.FailedOp(big.NewInt(int64(i)), fmt.Sprintf("validate guardian signature reverted: %s", err.Error()))
		}

		_, _, err = ep.validatePrepayment(big.NewInt(int64(i)), &ops[i], &opInfos[i])
		if err != nil {
			return err
		}
	}

	ep.EmitEvent("BeforeExecution")

	collected := big.NewInt(0)
	for i := 0; i < len(ops); i++ {
		op := ops[i]
		sa := SmartAccountBuildConfig.BuildWithAddress(ep.CrossCallSystemContractContext(), op.Sender).SetRemainingGas(op.CallGasLimit)
		if len(op.CallData) < 20 {
			return interfaces.FailedOp(big.NewInt(int64(i)), "callData is less than 20 bytes, should be contain address")
		}

		// 20 bytes for owner address
		owner := ethcommon.BytesToAddress(op.CallData[:20])
		if err := sa.ResetAndLockOwner(owner); err != nil {
			return interfaces.FailedOp(big.NewInt(int64(i)), fmt.Sprintf("reset owner reverted: %s", err.Error()))
		}

		actualGas := big.NewInt(int64(RecoveryAccountOwnerGas))
		actualGasCost, err := ep.handlePostOp(big.NewInt(int64(i)), interfaces.OpSucceeded, &opInfos[i], []byte(""), actualGas, big.NewInt(0))
		if err != nil {
			return err
		}

		collected.Add(collected, actualGasCost)
	}

	return ep.compensate(beneficiary, collected)
}

func (ep *EntryPoint) simulationOnlyValidations(userOp *interfaces.UserOperation) error {
	if err := ep.validateSenderAndPaymaster(userOp.InitCode, userOp.Sender, userOp.PaymasterAndData); err != nil {
		return interfaces.FailedOp(big.NewInt(0), err.Error())
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

func (ep *EntryPoint) addBalance(account ethcommon.Address, value *big.Int) {
	ep.Ctx.StateLedger.AddBalance(types.NewAddress(account.Bytes()), value)
}

func (ep *EntryPoint) subBalance(account ethcommon.Address, value *big.Int) {
	ep.Ctx.StateLedger.SubBalance(types.NewAddress(account.Bytes()), value)
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
