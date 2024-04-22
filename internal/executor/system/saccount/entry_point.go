package saccount

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/intutil"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/interfaces"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
)

const (
	CreateAccountGas        = 10000
	RecoveryAccountOwnerGas = 10000
	MaxCallGasLimit         = 50000
)

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
	*NonceManager
	*StakeManager
	*common.ReentrancyGuard

	logger   logrus.FieldLogger
	selfAddr *types.Address

	account       ledger.IAccount
	currentUser   *ethcommon.Address
	currentHeight uint64
	currentLogs   *[]common.Log
	stateLedger   ledger.StateLedger
	currentEVM    *vm.EVM
}

func NewEntryPoint(cfg *common.SystemContractConfig) *EntryPoint {
	ep := &EntryPoint{
		logger:          loggers.Logger(loggers.SystemContract),
		NonceManager:    NewNonceManager(),
		StakeManager:    NewStakeManager(),
		ReentrancyGuard: common.NewReentrancyGuard(),
		selfAddr:        types.NewAddressByStr(common.EntryPointContractAddr),
	}

	return ep
}

func (ep *EntryPoint) SetContext(context *common.VMContext) {
	ep.currentUser = context.CurrentUser
	ep.currentHeight = context.CurrentHeight
	ep.currentLogs = context.CurrentLogs
	ep.stateLedger = context.StateLedger
	ep.currentEVM = context.CurrentEVM

	addr := ep.selfAddress()
	ep.account = ep.stateLedger.GetOrCreateAccount(addr)

	ep.StakeManager.Init(ep.stateLedger)
	ep.NonceManager.Init(ep.account)

	ep.logger.Infof("EntryPoint SetContext")
}

func (ep *EntryPoint) selfAddress() *types.Address {
	return ep.selfAddr
}

// nolint
func (ep *EntryPoint) emitEvent(events ...[]byte) {
	var data []byte
	if len(events) > 4 {
		for _, ev := range events[4:] {
			data = append(data, ev...)
		}
		events = events[:4]
	}

	currentLog := common.Log{
		Address: ep.selfAddress(),
	}

	for _, ev := range events {
		currentLog.Topics = append(currentLog.Topics, types.NewHash(ev))
	}
	currentLog.Data = data
	currentLog.Removed = false

	*ep.currentLogs = append(*ep.currentLogs, currentLog)
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
	ep.logger.Infof("entrypoint handleOps, ops: %+v", ops)
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

	ep.emitBeforeExecution()

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
	ep.logger.Infof("entrypoint simulateHandleOp, op %+v, callData: %x, paymasterAndData: %x, target: %s, targetCallData: %x\n", op, op.CallData, op.PaymasterAndData, target.Hex(), targetCallData)
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
		result, _, err := call(ep.stateLedger, ep.currentEVM, op.CallGasLimit, ep.selfAddress(), &target, targetCallData)
		if err != nil {
			targetSuccess = false
		}

		targetResult = result
	}

	return interfaces.ExecutionResult(opInfo.PreOpGas, paid, big.NewInt(int64(data.ValidAfter)), big.NewInt(int64(data.ValidUntil)), targetSuccess, targetResult)
}

func (ep *EntryPoint) SimulateValidation(userOp interfaces.UserOperation) error {
	ep.logger.Infof("entrypoint simulateValidation")
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
		return nil, nil, common.NewRevertStringError(err.Error())
	}
	outOpInfo.UserOpHash = ep.getUserOpHash(userOp)

	maxGasValues := new(big.Int).Or(mUserOp.PreVerificationGas, mUserOp.VerificationGasLimit).Or(mUserOp.CallGasLimit, userOp.MaxFeePerGas).Or(userOp.MaxPriorityFeePerGas, big.NewInt(0))
	// validate all numeric values in userOp are well below 128 bit, so they can safely be added
	// and multiplied without causing overflow
	if maxGasValues.BitLen() > 128 {
		return nil, nil, common.NewRevertStringError("AA94 gas values overflow")
	}

	requiredPreFund := getRequiredPrefund(mUserOp)
	gasUsedByValidateAccountPrepayment, validationData, err := ep.validateAccountPrepayment(opIndex, userOp, outOpInfo, requiredPreFund)
	if err != nil {
		return nil, nil, err
	}

	if !ep.validateAndUpdateNonce(mUserOp.Sender, mUserOp.Nonce) {
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
	outOfTimeRange := ep.currentEVM.Context.Time > validation.ValidUntil || ep.currentEVM.Context.Time < validation.ValidAfter
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
	outOfTimeRange = ep.currentEVM.Context.Time > data.ValidUntil || ep.currentEVM.Context.Time < data.ValidAfter
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
	entryPointAddr := ep.selfAddress().ETHAddress()
	sa := NewSmartAccount(ep.logger, ep)
	sa.SetContext(&common.VMContext{
		CurrentUser: &entryPointAddr,
		StateLedger: ep.stateLedger,
		CurrentLogs: ep.currentLogs,
		CurrentEVM:  ep.currentEVM,
	})
	sa.InitializeOrLoad(sender, ethcommon.Address{}, ethcommon.Address{}, mUserOp.VerificationGasLimit)
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
		return nil, nil, common.NewRevertStringError("AA41 too little verificationGas")
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

	// set paymaster context
	entrypointAddr := ep.selfAddress().ETHAddress()
	paymaster.SetContext(&common.VMContext{
		StateLedger: ep.stateLedger,
		CurrentLogs: ep.currentLogs,
		CurrentEVM:  ep.currentEVM,
		CurrentUser: &entrypointAddr,
	})

	context, validationData, err = paymaster.ValidatePaymasterUserOp(*op, opInfo.UserOpHash, requiredPrefund)
	if err != nil {
		return nil, nil, interfaces.FailedOp(opIndex, fmt.Sprintf("AA33 reverted: %s", err.Error()))
	}

	return context, validationData, nil
}

func (ep *EntryPoint) getPaymaster(paymasterAddr ethcommon.Address) (interfaces.IPaymaster, error) {
	if paymasterAddr == ethcommon.HexToAddress(common.VerifyingPaymasterContractAddr) {
		return NewVerifyingPaymaster(ep), nil
	} else if paymasterAddr == ethcommon.HexToAddress(common.TokenPaymasterContractAddr) {
		return NewTokenPaymaster(ep), nil
	} else {
		return nil, common.NewRevertStringError("paymaster not found")
	}
}

// executeUserOp execute a user op
// @param opIndex index into the opInfo array
// @param userOp the userOp to execute
// @param opInfo the opInfo filled by validatePrepayment for this userOp.
// @return actualGasCost the total amount this userOp paid.
func (ep *EntryPoint) executeUserOp(opIndex *big.Int, userOp *interfaces.UserOperation, opInfo *UserOpInfo) (actualGasCost *big.Int, err error) {
	context := opInfo.Context
	return ep.innerHandleOp(opIndex, userOp.CallData, opInfo, context)
}

// innerHandleOp is inner function to handle a UserOperation.
func (ep *EntryPoint) innerHandleOp(opIndex *big.Int, callData []byte, opInfo *UserOpInfo, context []byte) (actualGasCost *big.Int, err error) {
	mUserOp := opInfo.MUserOp
	mode := interfaces.OpSucceeded
	usedGas := big.NewInt(0)
	if len(callData) > 0 {
		// check if call smart account method, directly call. otherwise, call evm
		if len(callData) < 4 {
			return nil, interfaces.FailedOp(opIndex, "callData length not enough")
		}
		sa := NewSmartAccount(ep.logger, ep)
		entryPointAddr := ep.selfAddress().ETHAddress()
		sa.SetContext(&common.VMContext{
			CurrentUser: &entryPointAddr,
			StateLedger: ep.stateLedger,
			CurrentLogs: ep.currentLogs,
			CurrentEVM:  ep.currentEVM,
		})
		sa.InitializeOrLoad(mUserOp.Sender, ethcommon.Address{}, ethcommon.Address{}, mUserOp.CallGasLimit)

		isInnerMethod, callGas, err := JudgeOrCallInnerMethod(callData, sa)
		if err != nil {
			return nil, interfaces.FailedOp(opIndex, fmt.Sprintf("AA42 reverted: %s", err.Error()))
		}

		usedGas.Add(usedGas, big.NewInt(int64(callGas)))
		if !isInnerMethod {
			// call evm if not inner method
			_, callGas, err := call(ep.stateLedger, ep.currentEVM, mUserOp.CallGasLimit, ep.selfAddress(), &mUserOp.Sender, callData)
			if err != nil {
				ep.emitUserOperationRevertReason(opInfo.UserOpHash, mUserOp.Sender, mUserOp.Nonce, []byte(err.Error()))
				mode = interfaces.OpReverted
			}

			usedGas.Add(usedGas, big.NewInt(int64(callGas)))
		}
	}

	ep.logger.Infof("handle op, opIndex: %s, mode: %s, usedGas: %s, preOpGas: %s", opIndex.Text(10), mode, usedGas.Text(10), opInfo.PreOpGas.Text(10))
	actualGas := new(big.Int).Add(usedGas, opInfo.PreOpGas)
	return ep.handlePostOp(opIndex, mode, opInfo, context, actualGas)
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
func (ep *EntryPoint) handlePostOp(opIndex *big.Int, mode interfaces.PostOpMode, opInfo *UserOpInfo, context []byte, actualGas *big.Int) (*big.Int, error) {
	var refundAddress ethcommon.Address
	mUserOp := opInfo.MUserOp
	gasPrice := getUserOpGasPrice(&mUserOp)

	paymaster := mUserOp.Paymaster
	entrypointAddr := ep.selfAddress().ETHAddress()
	actualGasCost := new(big.Int).Mul(actualGas, gasPrice)
	if paymaster == (ethcommon.Address{}) {
		refundAddress = mUserOp.Sender
		sa := NewSmartAccount(ep.logger, ep)
		sa.SetContext(&common.VMContext{
			StateLedger: ep.stateLedger,
			CurrentLogs: ep.currentLogs,
			CurrentEVM:  ep.currentEVM,
			CurrentUser: &entrypointAddr,
		})
		sa.InitializeOrLoad(mUserOp.Sender, ethcommon.Address{}, ethcommon.Address{}, mUserOp.VerificationGasLimit)
		if err := sa.postUserOp(opInfo.UseSessionKey, actualGasCost); err != nil {
			return nil, interfaces.FailedOp(opIndex, fmt.Sprintf("post user op reverted: %s", err.Error()))
		}
	} else {
		refundAddress = paymaster
		if len(context) > 0 {
			paymaster, err := ep.getPaymaster(paymaster)
			if err != nil {
				return nil, interfaces.FailedOp(opIndex, err.Error())
			}
			// set paymaster context
			paymaster.SetContext(&common.VMContext{
				StateLedger: ep.stateLedger,
				CurrentLogs: ep.currentLogs,
				CurrentEVM:  ep.currentEVM,
				CurrentUser: &entrypointAddr,
			})
			if err := paymaster.PostOp(mode, context, actualGasCost); err != nil {
				return nil, interfaces.FailedOp(opIndex, fmt.Sprintf("AA50 postOp reverted: %s", err.Error()))
			}
		}
	}

	if opInfo.Prefund.Cmp(actualGasCost) == -1 {
		ep.logger.Errorf("prefund is below actual gas cost, prefund: %s, actual gas cost: %s, actual gas: %s, gas price: %s", opInfo.Prefund.Text(10), actualGasCost.Text(10), actualGas.Text(10), gasPrice.Text(10))
		return nil, interfaces.FailedOp(opIndex, "AA51 prefund below actualGasCost")
	}
	refund := new(big.Int).Sub(opInfo.Prefund, actualGasCost)
	ep.addBalance(refundAddress, refund)
	success := mode == interfaces.OpSucceeded

	ep.emitUserOperationEvent(opInfo.UserOpHash, mUserOp.Sender, paymaster, mUserOp.Nonce, success, actualGasCost, actualGas)
	return actualGasCost, nil
}

func (ep *EntryPoint) GetUserOpHash(userOp interfaces.UserOperation) ethcommon.Hash {
	return interfaces.GetUserOpHash(&userOp, ep.selfAddr.ETHAddress(), ep.currentEVM.ChainConfig().ChainID)
}

func (ep *EntryPoint) getUserOpHash(userOp *interfaces.UserOperation) []byte {
	return interfaces.GetUserOpHash(userOp, ep.selfAddr.ETHAddress(), ep.currentEVM.ChainConfig().ChainID).Bytes()
}

func (ep *EntryPoint) createSenderIfNeeded(opIndex *big.Int, opInfo *UserOpInfo, initCode []byte) (uint64, error) {
	var usedGas uint64
	if len(initCode) > 0 {
		sender := opInfo.MUserOp.Sender
		account := ep.stateLedger.GetOrCreateAccount(types.NewAddress(sender.Bytes()))
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
		ep.emitAccountDeployed(opInfo.UserOpHash, sender, factory, opInfo.MUserOp.Paymaster)
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

	ep.logger.Infof("create sender, init code: %x, factoryAddr: %s", initCode, factoryAddr.Hex())

	// if factoryAddr is system contract, directly call
	if factoryAddr.Hex() == common.AccountFactoryContractAddr {
		factory := NewSmartAccountFactory(&common.SystemContractConfig{
			Logger: ep.logger,
		}, ep)
		selfAddr := ep.selfAddress().ETHAddress()
		factory.SetContext(&common.VMContext{
			StateLedger:   ep.stateLedger,
			CurrentHeight: ep.currentHeight,
			CurrentLogs:   ep.currentLogs,
			CurrentUser:   &selfAddr,
			CurrentEVM:    ep.currentEVM,
		})
		// discard method signature
		callData := initCallData[4:]
		owner := ethcommon.BytesToAddress(callData[0:32])
		salt := new(big.Int).SetBytes(callData[32:64])
		// guardian
		var guardian ethcommon.Address
		if len(callData) >= 64+32 {
			guardian = ethcommon.BytesToAddress(callData[64 : 64+32])
		}
		sa := factory.CreateAccount(owner, salt, guardian)
		return sa.selfAddress().ETHAddress(), CreateAccountGas, nil
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
	ep.logger.Infof("entrypoint HandleAccountRecovery, ops: %+v", ops)
	if err := ep.Enter(); err != nil {
		return err
	}
	defer ep.Exit()

	opInfos := make([]UserOpInfo, len(ops))
	for i := 0; i < len(ops); i++ {
		// judge if account exist
		sender := types.NewAddress(ops[i].Sender.Bytes())
		account := ep.stateLedger.GetOrCreateAccount(sender)
		if code := account.Code(); len(code) == 0 {
			return common.NewRevertStringError("smart account is not initialized")
		}

		// validate guardian signature
		entryPointAddr := ep.selfAddress().ETHAddress()
		sa := NewSmartAccount(ep.logger, ep)
		sa.SetContext(&common.VMContext{
			CurrentUser: &entryPointAddr,
			StateLedger: ep.stateLedger,
			CurrentLogs: ep.currentLogs,
			CurrentEVM:  ep.currentEVM,
		})
		sa.InitializeOrLoad(sender.ETHAddress(), ethcommon.Address{}, ethcommon.Address{}, ops[i].VerificationGasLimit)
		err := sa.ValidateGuardianSignature(ops[i].Signature, ep.getUserOpHash(&ops[i]))
		if err != nil {
			return interfaces.FailedOp(big.NewInt(int64(i)), fmt.Sprintf("validate guardian signature reverted: %s", err.Error()))
		}

		_, _, err = ep.validatePrepayment(big.NewInt(int64(i)), &ops[i], &opInfos[i])
		if err != nil {
			return err
		}
	}

	ep.emitBeforeExecution()

	collected := big.NewInt(0)
	for i := 0; i < len(ops); i++ {
		op := ops[i]
		sender := types.NewAddress(op.Sender.Bytes())
		entryPointAddr := ep.selfAddress().ETHAddress()
		sa := NewSmartAccount(ep.logger, ep)
		sa.SetContext(&common.VMContext{
			CurrentUser: &entryPointAddr,
			StateLedger: ep.stateLedger,
			CurrentLogs: ep.currentLogs,
			CurrentEVM:  ep.currentEVM,
		})
		sa.InitializeOrLoad(sender.ETHAddress(), ethcommon.Address{}, ethcommon.Address{}, op.CallGasLimit)

		if len(op.CallData) < 20 {
			ep.emitUserOperationRevertReason(opInfos[i].UserOpHash, op.Sender, op.Nonce, []byte("callData is less than 20 bytes, should be contain address"))
		}

		// 20 bytes for owner address
		owner := ethcommon.BytesToAddress(op.CallData[:20])
		if err := sa.ResetAndLockOwner(owner); err != nil {
			return interfaces.FailedOp(big.NewInt(int64(i)), fmt.Sprintf("reset owner reverted: %s", err.Error()))
		}

		actualGas := big.NewInt(int64(RecoveryAccountOwnerGas))
		actualGasCost, err := ep.handlePostOp(big.NewInt(int64(i)), interfaces.OpSucceeded, &opInfos[i], []byte(""), actualGas)
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
	account := ep.stateLedger.GetOrCreateAccount(types.NewAddress(sender.Bytes()))
	if len(initCode) == 0 && account.Code() == nil {
		// it would revert anyway. but give a meaningful message
		return errors.New("AA20 account not deployed")
	}
	if len(paymasterAndData) >= 20 {
		paymaster := types.NewAddress(paymasterAndData[0:20])
		paymasterAccount := ep.stateLedger.GetOrCreateAccount(paymaster)
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
	info.Deposit = ep.stateLedger.GetBalance(types.NewAddress(account.Bytes()))
	return info
}

// BalanceOf overide StakeManager
func (ep *EntryPoint) BalanceOf(account ethcommon.Address) *big.Int {
	return ep.stateLedger.GetBalance(types.NewAddress(account.Bytes()))
}

func (ep *EntryPoint) addBalance(account ethcommon.Address, value *big.Int) {
	ep.stateLedger.AddBalance(types.NewAddress(account.Bytes()), value)
}

func (ep *EntryPoint) subBalance(account ethcommon.Address, value *big.Int) {
	ep.stateLedger.SubBalance(types.NewAddress(account.Bytes()), value)
}

func (ep *EntryPoint) emitUserOperationEvent(userOpHash []byte, sender, paymaster ethcommon.Address, nonce *big.Int, success bool, actualGasCost, actualGasUsed *big.Int) {
	userOpEvent := abi.NewEvent("UserOperationEvent", "UserOperationEvent", false, abi.Arguments{
		{Name: "userOpHash", Type: common.Bytes32Type, Indexed: true},
		{Name: "sender", Type: common.AddressType, Indexed: true},
		{Name: "paymaster", Type: common.AddressType, Indexed: true},
		{Name: "nonce", Type: common.BigIntType},
		{Name: "success", Type: common.BoolType},
		{Name: "actualGasCost", Type: common.BigIntType},
		{Name: "actualGasUsed", Type: common.BigIntType},
	})

	currentLog := common.Log{
		Address: ep.selfAddress(),
	}

	// log can only index 4 topics
	currentLog.Topics = append(currentLog.Topics, types.NewHash(userOpEvent.ID.Bytes()))
	currentLog.Topics = append(currentLog.Topics, types.NewHash(userOpHash))
	currentLog.Topics = append(currentLog.Topics, types.NewHash(sender.Bytes()))
	currentLog.Topics = append(currentLog.Topics, types.NewHash(paymaster.Bytes()))

	currentLog.Data = append(currentLog.Data, ethcommon.LeftPadBytes(nonce.Bytes(), 32)...)
	currentLog.Data = append(currentLog.Data, ethcommon.LeftPadBytes(common.Bool2Bytes(success), 32)...)
	currentLog.Data = append(currentLog.Data, ethcommon.LeftPadBytes(actualGasCost.Bytes(), 32)...)
	currentLog.Data = append(currentLog.Data, ethcommon.LeftPadBytes(actualGasUsed.Bytes(), 32)...)

	currentLog.Removed = false

	*ep.currentLogs = append(*ep.currentLogs, currentLog)
}

func (ep *EntryPoint) emitUserOperationRevertReason(userOpHash []byte, sender ethcommon.Address, nonce *big.Int, revertReason []byte) {
	userOpRevertEvent := abi.NewEvent("UserOperationRevertReason", "UserOperationRevertReason", false, abi.Arguments{
		{Name: "userOpHash", Type: common.Bytes32Type, Indexed: true},
		{Name: "sender", Type: common.AddressType, Indexed: true},
		{Name: "nonce", Type: common.BigIntType},
		{Name: "revertReason", Type: common.BytesType},
	})

	currentLog := common.Log{
		Address: ep.selfAddress(),
	}

	// log can only index 4 topics
	currentLog.Topics = append(currentLog.Topics, types.NewHash(userOpRevertEvent.ID.Bytes()))
	currentLog.Topics = append(currentLog.Topics, types.NewHash(userOpHash))
	currentLog.Topics = append(currentLog.Topics, types.NewHash(sender.Bytes()))

	currentLog.Data = append(currentLog.Data, ethcommon.LeftPadBytes(nonce.Bytes(), 32)...)
	currentLog.Data = append(currentLog.Data, ethcommon.LeftPadBytes(revertReason, 32)...)

	currentLog.Removed = false

	*ep.currentLogs = append(*ep.currentLogs, currentLog)
}

func (ep *EntryPoint) emitAccountDeployed(userOpHash []byte, sender, factory, paymaster ethcommon.Address) {
	accountDeployedEvent := abi.NewEvent("AccountDeployed", "AccountDeployed", false, abi.Arguments{
		{Name: "userOpHash", Type: common.Bytes32Type, Indexed: true},
		{Name: "sender", Type: common.AddressType, Indexed: true},
		{Name: "factory", Type: common.AddressType},
		{Name: "paymaster", Type: common.AddressType},
	})

	currentLog := common.Log{
		Address: ep.selfAddress(),
	}

	currentLog.Topics = append(currentLog.Topics, types.NewHash(accountDeployedEvent.ID.Bytes()))
	currentLog.Topics = append(currentLog.Topics, types.NewHash(userOpHash))
	currentLog.Topics = append(currentLog.Topics, types.NewHash(sender.Bytes()))

	currentLog.Data = append(currentLog.Data, ethcommon.LeftPadBytes(factory.Bytes(), 32)...)
	currentLog.Data = append(currentLog.Data, ethcommon.LeftPadBytes(paymaster.Bytes(), 32)...)

	currentLog.Removed = false
	*ep.currentLogs = append(*ep.currentLogs, currentLog)
}

func (ep *EntryPoint) emitBeforeExecution() {
	beforeExecutionEvent := abi.NewEvent("BeforeExecution", "BeforeExecution", false, abi.Arguments{})

	currentLog := common.Log{
		Address: ep.selfAddress(),
	}

	currentLog.Topics = append(currentLog.Topics, types.NewHash(beforeExecutionEvent.ID.Bytes()))

	currentLog.Removed = false
	*ep.currentLogs = append(*ep.currentLogs, currentLog)
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

// call return call result, left over gas and error
func call(stateLedger ledger.StateLedger, evm *vm.EVM, gas *big.Int, from *types.Address, to *ethcommon.Address, callData []byte) (returnData []byte, gasLeft uint64, err error) {
	return callWithValue(stateLedger, evm, gas, big.NewInt(0), from, to, callData)
}

// callWithValue return call result, left over gas and error
// nolint
func callWithValue(stateLedger ledger.StateLedger, evm *vm.EVM, gas, value *big.Int, from *types.Address, to *ethcommon.Address, callData []byte) (returnData []byte, gasLeft uint64, err error) {
	if gas == nil || gas.Sign() == 0 {
		gas = big.NewInt(MaxCallGasLimit)
	}
	v, err := intutil.BigIntToUint256(value)
	if err != nil {
		return nil, gas.Uint64(), err
	}
	result, gasLeft, err := evm.Call(vm.AccountRef(from.ETHAddress()), *to, callData, gas.Uint64(), v)
	if err == vm.ErrExecutionReverted {
		err = fmt.Errorf("%s, reason: %x", err.Error(), result)
	}
	return result, gasLeft, err
}

// Initialize in genesis
func Initialize(stateLedger ledger.StateLedger, admin string) error {
	if !ethcommon.IsHexAddress(admin) {
		return errors.New("invalid admin address")
	}
	owner := ethcommon.HexToAddress(admin)
	InitializeVerifyingPaymaster(stateLedger, owner)
	InitializeTokenPaymaster(stateLedger, owner)

	return nil
}
