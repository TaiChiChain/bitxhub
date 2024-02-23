package saccount

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/interfaces"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
)

const (
	ownerKey    = "owner_key"
	guradianKey = "guardian_key"
	sessionKey  = "session_key"
)

// Session is temporary key to control the smart account
// Session has spending limit and valid time range
type Session struct {
	// used to check signature
	Addr ethcommon.Address

	// max limit for spending
	SpendingLimit *big.Int
	SpentAmount   *big.Int

	// valid time range
	ValidUntil uint64
	ValidAfter uint64
}

var _ interfaces.IAccount = (*SmartAccount)(nil)

type SmartAccount struct {
	entryPoint interfaces.IEntryPoint
	logger     logrus.FieldLogger

	// Storage fields
	owner    ethcommon.Address
	guardian ethcommon.Address
	session  *Session

	selfAddr *types.Address
	account  ledger.IAccount

	currentUser *ethcommon.Address
	currentLogs *[]common.Log
	stateLedger ledger.StateLedger
	currentEVM  *vm.EVM
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
func (sa *SmartAccount) InitializeOrLoad(selfAddr, owner, guardian ethcommon.Address) {
	sa.selfAddr = types.NewAddress(selfAddr.Bytes())
	sa.account = sa.stateLedger.GetOrCreateAccount(sa.selfAddr)

	if isExist, ownerBytes := sa.account.GetState([]byte(ownerKey)); isExist {
		sa.owner = ethcommon.BytesToAddress(ownerBytes)
	} else if owner != (ethcommon.Address{}) {
		sa.account.SetState([]byte(ownerKey), owner.Bytes())
		sa.account.SetCodeAndHash(selfAddr.Bytes())
		sa.owner = owner

		entryPointAddr := types.NewAddressByStr(common.EntryPointContractAddr)
		// emit AccountInitialized event
		sa.emitEvent(entryPointAddr.Bytes(), owner.Bytes())
	}

	// initialize guardian
	if err := sa.SetGuardian(guardian); err != nil {
		sa.logger.Warnf("initialize smart account failed: set guardian error: %v", err)
	}
}

func (sa *SmartAccount) selfAddress() *types.Address {
	return sa.selfAddr
}

func (sa *SmartAccount) getOwner() ethcommon.Address {
	if sa.owner == (ethcommon.Address{}) {
		isExist, ownerBytes := sa.account.GetState([]byte(ownerKey))
		if isExist {
			sa.owner = ethcommon.BytesToAddress(ownerBytes)
		} else {
			// is no owner, belongs to self
			sa.owner = sa.selfAddr.ETHAddress()
		}
	}
	return sa.owner
}

func (sa *SmartAccount) setOwner(owner ethcommon.Address) {
	if owner != (ethcommon.Address{}) {
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
	if addr != sa.getOwner() {
		session := sa.getSession()
		if session == nil {
			return validationData, nil
		}

		if session.Addr != addr {
			return validationData, nil
		}

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

func (sa *SmartAccount) Execute(dest ethcommon.Address, value *big.Int, callFunc []byte) error {
	if sa.currentUser.Hex() != common.EntryPointContractAddr {
		return common.NewRevertStringError("only entrypoint can call smart account execute")
	}

	result, err := callWithValue(sa.stateLedger, sa.currentEVM, nil, value, sa.selfAddress(), &dest, callFunc)
	if err != nil {
		return common.NewRevertStringError(fmt.Sprintf("execute smart account callWithValue failed: %v", err))
	}

	if result != nil && result.Err != nil {
		return common.NewRevertStringError(fmt.Sprintf("execute smart account callWithValue failed: %v", result.Err))
	}

	return nil
}

func (sa *SmartAccount) ExecuteBatch(dest []ethcommon.Address, callFunc [][]byte) error {
	if sa.currentUser.Hex() != common.EntryPointContractAddr {
		return common.NewRevertStringError("only entrypoint can call smart account execute")
	}

	if len(dest) != len(callFunc) {
		return common.NewRevertStringError("dest and callFunc length mismatch")
	}

	for i := 0; i < len(dest); i++ {
		result, err := callWithValue(sa.stateLedger, sa.currentEVM, nil, big.NewInt(0), sa.selfAddress(), &dest[i], callFunc[i])
		if err != nil {
			return common.NewRevertStringError(fmt.Sprintf("execute batch smart account callWithValue failed: %v, index: %d", err, i))
		}

		if result != nil && result.Err != nil {
			return common.NewRevertStringError(fmt.Sprintf("execute batch smart account callWithValue failed: %v, index: %d", result.Err, i))
		}
	}

	return nil
}

// SetGuardian set guardian for recover smart account's owner
func (sa *SmartAccount) SetGuardian(guardian ethcommon.Address) error {
	if sa.currentUser.Hex() != common.EntryPointContractAddr {
		return common.NewRevertStringError("only entrypoint can call smart account set guardian")
	}

	if sa.guardian == (ethcommon.Address{}) {
		return common.NewRevertStringError("can't set guardian empty address")
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
		return common.NewRevertStringError("only entrypoint can call ValidateUserOp")
	}

	sa.setOwner(owner)

	return nil
}

// SetSession set session key
func (sa *SmartAccount) SetSession(session Session) error {
	if sa.currentUser.Hex() != common.EntryPointContractAddr {
		return common.NewRevertStringError("only entrypoint can call SetSession")
	}
	blockTime := sa.currentEVM.Context.Time
	if blockTime < session.ValidAfter || blockTime > session.ValidUntil {
		return common.NewRevertStringError("session valid time is out of date")
	}

	data, err := rlp.EncodeToBytes(session)
	if err != nil {
		return common.NewRevertStringError("rlp encode session error")
	}
	sa.account.SetState([]byte(sessionKey), data)

	sa.session = &session
	return nil
}

func (sa *SmartAccount) getSession() *Session {
	if sa.session != nil {
		return sa.session
	}

	var session Session
	isExist, data := sa.account.GetState([]byte(sessionKey))
	if isExist {
		if err := rlp.DecodeBytes(data, &session); err != nil {
			return nil
		}
		sa.session = &session
	}

	return sa.session
}

func recoveryAddrFromSignature(hash, signature []byte) (ethcommon.Address, error) {
	if len(signature) != 65 {
		return ethcommon.Address{}, common.NewRevertStringError("invalid signature length")
	}

	ethHash := accounts.TextHash(hash)
	// ethers js return r|s|v, v only 1 byte
	// golang return rid, v = rid +27
	if signature[64] > 27 {
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
