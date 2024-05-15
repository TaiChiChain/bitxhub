package saccount

import (
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/interfaces"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
)

const (
	ownerKey    = "owner_key"
	oldOwnerKey = "old_owner_key"
	guradianKey = "guardian_key"
	sessionKey  = "session_key"
	statusKey   = "status_key"

	callInnerMethodGas = 1000
	StatusUnlock       = 0
)

var (
	executeSig = crypto.Keccak256([]byte("execute(address,uint256,bytes)"))[:4]

	executeBatchSig = crypto.Keccak256([]byte("executeBatch(address[],bytes[])"))[:4]

	setGuardianSig = crypto.Keccak256([]byte("setGuardian(address)"))[:4]

	resetOwnerSig = crypto.Keccak256([]byte("resetOwner(address)"))[:4]

	setSessionSig = crypto.Keccak256([]byte("setSession(address,uint256,uint64,uint64)"))[:4]

	transferSig = crypto.Keccak256([]byte("transfer(address,uint256)"))[:4]

	executeMethod = abi.Arguments{
		{Name: "address", Type: common.AddressType},
		{Name: "value", Type: common.BigIntType},
		{Name: "callFunc", Type: common.BytesType},
	}

	executeBatchMethod = abi.Arguments{
		{Name: "dest", Type: common.AddressSliceType},
		{Name: "callFunc", Type: common.BytesSliceType},
	}

	LockedTime = 24 * time.Hour
)

// Session is temporary key to control the smart account
// Session has spending limit and valid time range
type Session struct {
	// used to check signature
	Addr ethcommon.Address

	// max limit for spending
	SpendingLimit *big.Int

	SpentAmount *big.Int

	// valid time range
	ValidUntil uint64

	ValidAfter uint64
}

var _ interfaces.IAccount = (*SmartAccount)(nil)

type SmartAccount struct {
	entryPoint interfaces.IEntryPoint
	logger     logrus.FieldLogger

	// Storage fields
	owner ethcommon.Address

	guardian    ethcommon.Address
	sessionSlot *common.VMSlot[Session]

	selfAddr *types.Address
	account  ledger.IAccount

	currentUser *ethcommon.Address
	currentLogs *[]common.Log
	stateLedger ledger.StateLedger
	currentEVM  *vm.EVM

	// remaining gas
	remainingGas *big.Int
}

func NewSmartAccount(logger logrus.FieldLogger, entryPoint interfaces.IEntryPoint) *SmartAccount {
	return &SmartAccount{
		logger:     logger,
		entryPoint: entryPoint,
	}
}

func (sa *SmartAccount) SetContext(context *common.VMContext) {
	sa.currentUser = context.CurrentUser
	sa.currentLogs = context.CurrentLogs
	sa.stateLedger = context.StateLedger
	sa.currentEVM = context.CurrentEVM
}

// InitializeOrLoad must be called after SetContext
// InitializeOrLoad can call anytimes, only initialize once
func (sa *SmartAccount) InitializeOrLoad(selfAddr, owner, guardian ethcommon.Address, gas *big.Int) {
	if sa.currentUser.Hex() != common.EntryPointContractAddr && sa.currentUser.Hex() != common.AccountFactoryContractAddr {
		return
	}

	sa.selfAddr = types.NewAddress(selfAddr.Bytes())
	sa.account = sa.stateLedger.GetOrCreateAccount(sa.selfAddr)
	sa.sessionSlot = common.NewVMSlotp[Session](sa.account, sessionKey)

	if isExist, ownerBytes := sa.account.GetState([]byte(ownerKey)); isExist {
		sa.owner = ethcommon.BytesToAddress(ownerBytes)
	} else if owner != (ethcommon.Address{}) {
		sa.account.SetState([]byte(ownerKey), owner.Bytes())
		sa.account.SetCodeAndHash(ethcommon.Hex2Bytes(common.EmptyContractBinCode))
		sa.owner = owner

		entryPointAddr := types.NewAddressByStr(common.EntryPointContractAddr)
		// emit AccountInitialized event
		sa.emitEvent(entryPointAddr.Bytes(), owner.Bytes())
	}

	// initialize guardian
	if err := sa.SetGuardian(guardian); err != nil {
		sa.logger.Warnf("initialize smart account failed: set guardian error: %v", err)
	}

	if gas == nil || gas.Sign() == 0 {
		sa.remainingGas = big.NewInt(MaxCallGasLimit)
	} else {
		sa.remainingGas = gas
	}
}

// nolint
func (sa *SmartAccount) getRemainingGas() *big.Int {
	return sa.remainingGas
}

func (sa *SmartAccount) selfAddress() *types.Address {
	return sa.selfAddr
}

func (sa *SmartAccount) getOwner() ethcommon.Address {
	// if is lock status, return old owner
	isExist, statusBytes := sa.account.GetState([]byte(statusKey))
	if isExist && string(statusBytes) != fmt.Sprintf("%d", StatusUnlock) {
		// if lock time is expired, unlock owner
		lockTime, _ := strconv.ParseUint(string(statusBytes), 10, 64)
		if sa.currentEVM.Context.Time > lockTime {
			sa.account.SetState([]byte(statusKey), []byte(fmt.Sprintf("%d", StatusUnlock)))
			sa.logger.Infof("smart account unlock owner")
		} else {
			// get old owner and return
			isExist, oldOwnerBytes := sa.account.GetState([]byte(oldOwnerKey))
			if isExist {
				return ethcommon.BytesToAddress(oldOwnerBytes)
			}
		}
	}

	isExist, ownerBytes := sa.account.GetState([]byte(ownerKey))
	if isExist {
		sa.owner = ethcommon.BytesToAddress(ownerBytes)
	} else {
		// is no owner, belongs to self
		sa.owner = sa.selfAddr.ETHAddress()
	}

	return sa.owner
}

func (sa *SmartAccount) setOwner(owner ethcommon.Address) {
	if owner != (ethcommon.Address{}) {
		isExist, oldOwner := sa.account.GetState([]byte(ownerKey))
		if isExist {
			sa.account.SetState([]byte(oldOwnerKey), oldOwner)
		}

		sa.account.SetState([]byte(ownerKey), owner.Bytes())
		sa.owner = owner
	}
}

func (sa *SmartAccount) emitEvent(events ...[]byte) {
	var data []byte
	if len(events) > 4 {
		for _, ev := range events[4:] {
			data = append(data, ev...)
		}
		events = events[:4]
	}

	currentLog := common.Log{
		Address: sa.selfAddress(),
	}

	for _, ev := range events {
		currentLog.Topics = append(currentLog.Topics, types.NewHash(ev))
	}
	currentLog.Data = data
	currentLog.Removed = false

	*sa.currentLogs = append(*sa.currentLogs, currentLog)
}

// ValidateUserOp implements interfaces.IAccount.
// ValidateUserOp return SigValidationFailed when validate failed
// This allows making a "simulation call" without a valid signature
// Other failures (e.g. nonce mismatch, or invalid signature format) should still revert to signal failure.
func (sa *SmartAccount) ValidateUserOp(userOp interfaces.UserOperation, userOpHash []byte, missingAccountFunds *big.Int) (validationData *big.Int, err error) {
	validation, err := sa.validateUserOp(&userOp, userOpHash, missingAccountFunds)
	if validation != nil {
		validationData = interfaces.PackValidationData(validation)
	}
	return validationData, err
}

// nolint
func (sa *SmartAccount) validateUserOp(userOp *interfaces.UserOperation, userOpHash []byte, missingAccountFunds *big.Int) (*interfaces.Validation, error) {
	if sa.currentUser.Hex() != common.EntryPointContractAddr {
		return nil, common.NewRevertStringError("only entrypoint can call ValidateUserOp")
	}

	validationData := &interfaces.Validation{
		SigValidation: interfaces.SigValidationFailed,
	}
	// validate signature
	addr, err := recoveryAddrFromSignature(userOpHash, userOp.Signature)
	if err != nil {
		sa.logger.Warnf("validate user op failed: %v", err)
		return validationData, nil
	}
	owner := sa.getOwner()
	sa.logger.Debugf("validate user op, owner: %s, addr: %s, smart account addr: %s", owner.String(), addr.String(), sa.selfAddress().String())
	if addr != owner {
		session := sa.getSession()
		if session == nil {
			sa.logger.Errorf("userOp signature is not from owner, owner: %s, recovery owner: %s", owner.String(), addr.String())
			return validationData, nil
		}

		if session.Addr != addr {
			sa.logger.Errorf("userOp signature is not from session key, session key addr: %s, recovery addr: %s", session.Addr, addr.String())
			return validationData, nil
		}

		sa.logger.Infof("use session key to validate, session key addr: %s", session.Addr.String())

		// if use session key
		validationData.ValidAfter = session.ValidAfter
		validationData.ValidUntil = session.ValidUntil
		validationData.RemainingLimit = big.NewInt(0)
		if session.SpentAmount.Cmp(session.SpendingLimit) < 0 {
			validationData.RemainingLimit = new(big.Int).Sub(session.SpendingLimit, session.SpentAmount)
		}
	}
	validationData.SigValidation = interfaces.SigValidationSucceeded

	return validationData, nil
}

func (sa *SmartAccount) Execute(dest ethcommon.Address, value *big.Int, callFunc []byte) (*big.Int, *big.Int, error) {
	if sa.currentUser.Hex() != common.EntryPointContractAddr {
		return big.NewInt(0), big.NewInt(0), common.NewRevertStringError("only entrypoint can call smart account execute")
	}

	sa.logger.Infof("execute smart account, dest %s, value: %s, callFunc: %x, remainingGas: %s", dest.Hex(), value.Text(10), callFunc, sa.remainingGas.Text(10))

	// use left gas to call
	_, gasLeft, err := callWithValue(sa.stateLedger, sa.currentEVM, sa.remainingGas, value, sa.selfAddress(), &dest, callFunc)
	if err != nil {
		return big.NewInt(0), big.NewInt(0), common.NewRevertStringError(fmt.Sprintf("execute smart account callWithValue failed: %v", err))
	}
	gasCost := new(big.Int).Sub(sa.remainingGas, big.NewInt(int64(gasLeft)))
	sa.remainingGas = big.NewInt(int64(gasLeft))

	totalUsedValue, err := sa.getAxcOfTransferValue(dest, callFunc)
	if err != nil {
		return big.NewInt(0), big.NewInt(0), common.NewRevertStringError(fmt.Sprintf("execute smart account getAxcOfTransferValue failed: %v", err))
	}

	if value != nil {
		totalUsedValue.Add(totalUsedValue, value)
	}

	return gasCost, totalUsedValue, nil
}

func (sa *SmartAccount) ExecuteBatch(dest []ethcommon.Address, callFunc [][]byte) (*big.Int, *big.Int, error) {
	if sa.currentUser.Hex() != common.EntryPointContractAddr {
		return big.NewInt(0), big.NewInt(0), common.NewRevertStringError("only entrypoint can call smart account execute")
	}

	if len(dest) != len(callFunc) {
		return big.NewInt(0), big.NewInt(0), common.NewRevertStringError("dest and callFunc length mismatch")
	}

	gasCost := big.NewInt(0)
	totalUsedValue := big.NewInt(0)
	for i := 0; i < len(dest); i++ {
		usedGas, usedAxc, err := sa.Execute(dest[i], big.NewInt(0), callFunc[i])
		if err != nil {
			return nil, big.NewInt(0), err
		}
		gasCost.Add(gasCost, usedGas)
		totalUsedValue.Add(totalUsedValue, usedAxc)
	}

	return gasCost, totalUsedValue, nil
}

// SetGuardian set guardian for recover smart account's owner
func (sa *SmartAccount) SetGuardian(guardian ethcommon.Address) error {
	if sa.currentUser.Hex() != common.EntryPointContractAddr && sa.currentUser.Hex() != common.AccountFactoryContractAddr {
		return common.NewRevertStringError("only entrypoint or account factory can call smart account set guardian")
	}

	if guardian == (ethcommon.Address{}) {
		return nil
	}

	sa.account.SetState([]byte(guradianKey), guardian.Bytes())
	sa.guardian = guardian
	return nil
}

func (sa *SmartAccount) getGuardian() ethcommon.Address {
	if sa.guardian != (ethcommon.Address{}) {
		return sa.guardian
	}

	isExist, data := sa.account.GetState([]byte(guradianKey))
	if isExist {
		sa.guardian = ethcommon.BytesToAddress(data)
	}
	return sa.guardian
}

func (sa *SmartAccount) ValidateGuardianSignature(guardianSig []byte, userOpHash []byte) error {
	if sa.currentUser.Hex() != common.EntryPointContractAddr {
		return common.NewRevertStringError("only entrypoint can call ValidateUserOp")
	}

	guardian := sa.getGuardian()

	// validate signature
	addr, err := recoveryAddrFromSignature(userOpHash, guardianSig)
	if err != nil {
		return err
	}

	if addr != guardian {
		return common.NewRevertStringError("invalid guardian signature")
	}

	return nil
}

// ResetOwner recovery owner of smart account by reset owner
func (sa *SmartAccount) ResetOwner(owner ethcommon.Address) error {
	if sa.currentUser.Hex() != common.EntryPointContractAddr {
		return common.NewRevertStringError("only entrypoint can call ResetOwner")
	}

	sa.setOwner(owner)

	return nil
}

// ResetAndLockOwner reset owner and lock owner for some times
// lock when guardian try to recovery smart account's owner
func (sa *SmartAccount) ResetAndLockOwner(owner ethcommon.Address) error {
	if sa.currentUser.Hex() != common.EntryPointContractAddr {
		return common.NewRevertStringError("only entrypoint can call ResetAndLockOwner")
	}

	sa.setOwner(owner)

	// lock owner for some times, save lock timestamp
	timestamp := sa.currentEVM.Context.Time
	lockTimestamp := timestamp + uint64(LockedTime.Seconds())
	sa.account.SetState([]byte(statusKey), []byte(fmt.Sprintf("%d", lockTimestamp)))

	sa.emitUserLocked(sa.selfAddr.ETHAddress(), owner, new(big.Int).SetUint64(lockTimestamp))

	return nil
}

// SetSession set session key
func (sa *SmartAccount) SetSession(addr ethcommon.Address, spendingLimit *big.Int, validAfter, validUntil uint64) error {
	if sa.currentUser.Hex() != common.EntryPointContractAddr {
		return common.NewRevertStringError("only entrypoint can call SetSession")
	}
	if validAfter >= validUntil {
		return common.NewRevertStringError("session key validAfter must less than validUntil")
	}

	session := sa.getSession()
	if session == nil {
		session = &Session{
			Addr:          addr,
			SpendingLimit: spendingLimit,
			SpentAmount:   big.NewInt(0),
			ValidAfter:    validAfter,
			ValidUntil:    validUntil,
		}
	} else {
		session.Addr = addr
		session.SpendingLimit = spendingLimit
		session.SpentAmount = big.NewInt(0)
		session.ValidAfter = validAfter
		session.ValidUntil = validUntil
	}

	return sa.setSession(session)
}

func (sa *SmartAccount) setSession(session *Session) error {
	return sa.sessionSlot.Put(*session)
}

func (sa *SmartAccount) getSession() *Session {
	isExist, session, _ := sa.sessionSlot.Get()
	if isExist {
		return &session
	}

	return nil
}

// TODO: this function is too business, may be should be self defined by user
func (sa *SmartAccount) getAxcOfTransferValue(token ethcommon.Address, callFunc []byte) (*big.Int, error) {
	if string(callFunc[:4]) == string(transferSig) {
		value := new(big.Int).SetBytes(callFunc[4+32:])

		// transfer to axc value
		tokenPaymaster := NewTokenPaymaster(sa.entryPoint)
		tokenPaymaster.SetContext(&common.VMContext{
			StateLedger: sa.stateLedger,
			CurrentEVM:  sa.currentEVM,
			CurrentUser: sa.currentUser,
			CurrentLogs: sa.currentLogs,
		})
		//uint128
		axcValue, _ := new(big.Int).SetString("ffffffffffffffffffffffffffffffff", 16)
		tokenValue, err := tokenPaymaster.getTokenValueOfAxc(token, axcValue)
		if err != nil {
			return nil, err
		}

		if value.BitLen() > 128 {
			return nil, common.NewRevertStringError("transfer value exceeds 128 bits, would cause overflow")
		}

		// return axc value
		// value * axcValue / tokenValue
		return new(big.Int).Div(new(big.Int).Mul(value, axcValue), tokenValue), nil
	}
	return big.NewInt(0), nil
}

func (sa *SmartAccount) postUserOp(UseSessionKey bool, actualGasCost, totalValue *big.Int) error {
	if sa.currentUser.Hex() != common.EntryPointContractAddr {
		return common.NewRevertStringError("only entrypoint can call SetSession")
	}

	sa.logger.Infof("post user op, use session key: %v, actual gas cost: %s", UseSessionKey, actualGasCost.String())

	// is not use session key, no need to update spent amount
	if !UseSessionKey {
		return nil
	}

	session := sa.getSession()
	if session == nil {
		sa.logger.Infof("no session key, no need to update spent amount")
		// no session, no need to update spent amount
		return nil
	}

	spentAmout := new(big.Int).Add(session.SpentAmount, actualGasCost)
	spentAmout.Add(spentAmout, totalValue)
	sa.logger.Infof("update session key spent amount, spent amount: %s, spend limit: %s", spentAmout.String(), session.SpendingLimit.String())
	if spentAmout.Cmp(session.SpendingLimit) > 0 {
		return common.NewRevertStringError("spent amount exceeds session spending limit")
	}
	session.SpentAmount = spentAmout

	return sa.setSession(session)
}

func (sa *SmartAccount) emitUserLocked(sender, newOwner ethcommon.Address, lockedTime *big.Int) {
	userLockedEvent := abi.NewEvent("UserLocked", "UserLocked", false, abi.Arguments{
		{Name: "sender", Type: common.AddressType, Indexed: true},
		{Name: "newOwner", Type: common.AddressType, Indexed: true},
		{Name: "lockedTime", Type: common.BigIntType},
	})

	currentLog := common.Log{
		Address: sa.selfAddress(),
	}

	currentLog.Topics = append(currentLog.Topics, types.NewHash(userLockedEvent.ID.Bytes()))
	currentLog.Topics = append(currentLog.Topics, types.NewHash(sender.Bytes()))
	currentLog.Topics = append(currentLog.Topics, types.NewHash(newOwner.Bytes()))

	currentLog.Data = append(currentLog.Data, ethcommon.LeftPadBytes(lockedTime.Bytes(), 32)...)
	currentLog.Removed = false

	*sa.currentLogs = append(*sa.currentLogs, currentLog)
}

func recoveryAddrFromSignature(hash, signature []byte) (ethcommon.Address, error) {
	if len(signature) != 65 {
		return ethcommon.Address{}, common.NewRevertStringError("invalid signature length")
	}

	ethHash := accounts.TextHash(hash)
	// ethers js return r|s|v, v only 1 byte
	// golang return rid, v = rid +27
	if signature[64] >= 27 {
		signature[64] -= 27
	}

	recoveredPub, err := crypto.Ecrecover(ethHash, signature)
	if err != nil {
		return ethcommon.Address{}, err
	}
	pubKey, _ := crypto.UnmarshalPubkey(recoveredPub)
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	return recoveredAddr, nil
}

// TODO: use abi to find method
// JudgeOrCallInnerMethod judge if call data is inner method, if yes, call it, return true.
func JudgeOrCallInnerMethod(callData []byte, sa *SmartAccount) (bool, uint64, *big.Int, error) {
	if len(callData) < 4 {
		return false, 0, nil, common.NewRevertStringError("callData length is too short")
	}

	var err error
	usedGas := big.NewInt(0)
	var gas *big.Int
	var totalUsedValue *big.Int
	var value *big.Int
	methodSig := callData[:4]
	switch string(methodSig) {
	case string(executeSig):
		if len(callData) < 36 {
			return false, 0, nil, common.NewRevertStringError("call smart account execute, callData length is too short")
		}
		var res []any
		res, err = executeMethod.Unpack(callData[4:])
		if err != nil {
			return false, 0, nil, common.NewRevertStringError(fmt.Sprintf("call smart account execute, unpack error: %v", err))
		}
		if len(res) != 3 {
			return false, 0, nil, common.NewRevertStringError("call smart account execute error, unpack result length is not 3")
		}
		dest := res[0].(ethcommon.Address)
		value = res[1].(*big.Int)
		callFunc := res[2].([]byte)
		gas, totalUsedValue, err = sa.Execute(dest, value, callFunc)
	case string(executeBatchSig):
		var res []any
		res, err = executeBatchMethod.Unpack(callData[4:])
		if err != nil {
			return false, 0, nil, common.NewRevertStringError(fmt.Sprintf("call smart account execute batch, unpack error: %v", err))
		}
		if len(res) != 2 {
			return false, 0, nil, common.NewRevertStringError("call smart account execute batch error, unpack result length is not 2")
		}
		dest := res[0].([]ethcommon.Address)
		callFunc := res[1].([][]byte)
		gas, totalUsedValue, err = sa.ExecuteBatch(dest, callFunc)
	case string(setGuardianSig):
		if len(callData) < 36 {
			return false, 0, nil, common.NewRevertStringError("call smart account set guardian, callData length is too short")
		}
		err = sa.SetGuardian(ethcommon.BytesToAddress(callData[4:36]))
	case string(resetOwnerSig):
		if len(callData) < 36 {
			return false, 0, nil, common.NewRevertStringError("call smart account reset owner, callData length is too short")
		}
		err = sa.ResetOwner(ethcommon.BytesToAddress(callData[4:36]))
	case string(setSessionSig):
		if len(callData) < 132 {
			return false, 0, nil, common.NewRevertStringError("call smart account set session, callData length is too short")
		}
		err = sa.SetSession(ethcommon.BytesToAddress(callData[4:36]),
			new(big.Int).SetBytes(callData[36:68]),
			new(big.Int).SetBytes(callData[68:100]).Uint64(),
			new(big.Int).SetBytes(callData[100:132]).Uint64(),
		)
	default:
		return false, 0, nil, nil
	}

	if gas == nil {
		gas = big.NewInt(callInnerMethodGas)
	}

	usedGas.Add(usedGas, gas)

	return true, usedGas.Uint64(), totalUsedValue, err
}
