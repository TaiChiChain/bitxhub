package saccount

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-webauthn/webauthn/protocol/webauthncbor"
	"github.com/go-webauthn/webauthn/protocol/webauthncose"
	"github.com/pkg/errors"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/interfaces"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/solidity/smart_account"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/solidity/smart_account_client"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/solidity/smart_account_proxy"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

const (
	oldOwnerKey = "old_owner"
	guradianKey = "guardian"
	sessionKey  = "session"
	passkeyKey  = "passkey"
	statusKey   = "status"

	callInnerMethodGas = 1000
	StatusUnlock       = 0
)

var SmartAccountBuildConfig = &common.SystemContractBuildConfig[*SmartAccount]{
	Name:   "saccount_account",
	AbiStr: smart_account_client.BindingContractMetaData.ABI,
	Constructor: func(systemContractBase common.SystemContractBase) *SmartAccount {
		return &SmartAccount{
			SystemContractBase: systemContractBase,
		}
	},
}

var (
	executeSig = crypto.Keccak256([]byte("execute(address,uint256,bytes)"))[:4]

	executeBatchSig = crypto.Keccak256([]byte("executeBatch(address[],bytes[])"))[:4]

	setGuardianSig = crypto.Keccak256([]byte("setGuardian(address)"))[:4]

	resetOwnerSig = crypto.Keccak256([]byte("resetOwner(address)"))[:4]

	setSessionSig = crypto.Keccak256([]byte("setSession(address,uint256,uint64,uint64)"))[:4]

	transferSig = crypto.Keccak256([]byte("transfer(address,uint256)"))[:4]

	setPasskeySig = crypto.Keccak256([]byte("setPasskey(bytes,uint8)"))[:4]

	executeMethod = abi.Arguments{
		{Name: "address", Type: common.AddressType},
		{Name: "value", Type: common.BigIntType},
		{Name: "callFunc", Type: common.BytesType},
	}

	executeBatchMethod = abi.Arguments{
		{Name: "dest", Type: common.AddressSliceType},
		{Name: "callFunc", Type: common.BytesSliceType},
	}

	setPasskeyMethod = abi.Arguments{
		{Name: "publicKey", Type: common.BytesType},
		{Name: "algo", Type: common.UInt8Type},
	}

	LockedTime = time.Minute // 24 * time.Hour
)

type CallReturnResult struct {
	UsedGas        uint64
	TotalUsedValue *big.Int
	Result         []any
}

var _ interfaces.IAccount = (*SmartAccount)(nil)

type SmartAccount struct {
	common.SystemContractBase

	owner       *common.VMSlot[ethcommon.Address]
	oldOwner    *common.VMSlot[ethcommon.Address]
	guardian    *common.VMSlot[ethcommon.Address]
	sessionKeys *common.VMSlot[[]SessionKey]
	passkeys    *common.VMSlot[[]Passkey]
	status      *common.VMSlot[uint64]

	// remaining gas
	remainingGas *big.Int

	// total used value
	totalUsedValue *big.Int
}

func (sa *SmartAccount) GenesisInit(genesis *repo.GenesisConfig) error {
	return nil
}

func (sa *SmartAccount) SetContext(context *common.VMContext) {
	sa.SystemContractBase.SetContext(context)

	sa.owner = common.NewVMSlot[ethcommon.Address](sa.StateAccount, ownerKey)
	sa.oldOwner = common.NewVMSlot[ethcommon.Address](sa.StateAccount, oldOwnerKey)
	sa.guardian = common.NewVMSlot[ethcommon.Address](sa.StateAccount, guradianKey)
	sa.sessionKeys = common.NewVMSlot[[]SessionKey](sa.StateAccount, sessionKey)
	sa.passkeys = common.NewVMSlot[[]Passkey](sa.StateAccount, passkeyKey)
	sa.status = common.NewVMSlot[uint64](sa.StateAccount, statusKey)

	// reset remaining gas and total used value
	sa.remainingGas = big.NewInt(MaxCallGasLimit)
	sa.totalUsedValue = big.NewInt(0)
}

func (sa *SmartAccount) SetRemainingGas(gas *big.Int) *SmartAccount {
	if gas == nil || gas.Sign() == 0 {
		sa.remainingGas = big.NewInt(MaxCallGasLimit)
	} else {
		sa.remainingGas = gas
	}
	return sa
}

// Initialize must be called after SetContext
// Initialize can call anytimes, only initialize once
func (sa *SmartAccount) Initialize(owner, guardian ethcommon.Address) error {
	if sa.Ctx.From != ethcommon.HexToAddress(common.EntryPointContractAddr) && sa.Ctx.From != ethcommon.HexToAddress(common.AccountFactoryContractAddr) {
		return errors.Errorf("only entrypoint or account factory can call smart account init")
	}

	if sa.owner.Has() {
		return nil
	}
	if err := sa.owner.Put(owner); err != nil {
		return err
	}
	sa.StateAccount.SetCodeAndHash(ethcommon.Hex2Bytes(common.EmptyContractBinCode))
	// emit AccountInitialized event
	sa.EmitEvent(&smart_account.EventSmartAccountInitialized{
		EntryPoint: types.NewAddressByStr(common.EntryPointContractAddr).ETHAddress(),
		Owner:      owner,
	})

	// initialize guardian
	if err := sa.SetGuardian(guardian); err != nil {
		return errors.Errorf("initialize smart account failed: set guardian error: %v", err)
	}

	return nil
}

// nolint
func (sa *SmartAccount) getRemainingGas() *big.Int {
	return sa.remainingGas
}

func (sa *SmartAccount) getTotalUsedValue() *big.Int {
	return sa.totalUsedValue
}

func (sa *SmartAccount) GetOwner() (ethcommon.Address, error) {
	// if is lock status, return old owner
	isExist, lockTime, err := sa.status.Get()
	if err != nil {
		return ethcommon.Address{}, err
	}
	if isExist && lockTime != StatusUnlock {
		// if lock time is expired, unlock owner
		if sa.Ctx.CurrentEVM.Context.Time > lockTime {
			if err := sa.status.Put(StatusUnlock); err != nil {
				return ethcommon.Address{}, err
			}
			sa.Logger.Infof("smart account unlock owner")
		} else {
			// get old owner and return
			isExist, oldOwner, err := sa.oldOwner.Get()
			if err != nil {
				return ethcommon.Address{}, err
			}
			if isExist {
				return oldOwner, nil
			}
		}
	}

	isExist, owner, err := sa.owner.Get()
	if err != nil {
		return ethcommon.Address{}, err
	}
	if isExist {
		return owner, nil
	}
	// is no owner, belongs to self
	return sa.EthAddress, nil
}

func (sa *SmartAccount) setOwner(owner ethcommon.Address) error {
	if owner != (ethcommon.Address{}) {
		isExist, oldOwner, err := sa.owner.Get()
		if err != nil {
			return err
		}
		if isExist {
			// set old owner
			if err := sa.oldOwner.Put(oldOwner); err != nil {
				return err
			}
		}

		if err := sa.owner.Put(owner); err != nil {
			return err
		}
	}
	return nil
}

// GetGuardian return guardian address
func (sa *SmartAccount) GetGuardian() (ethcommon.Address, error) { return sa.guardian.MustGet() }

// GetStatus return status
func (sa *SmartAccount) GetStatus() (uint64, error) { return sa.status.MustGet() }

// GetPasskeys return passkey list
func (sa *SmartAccount) GetPasskeys() ([]smart_account_proxy.PassKey, error) {
	passkeyList, err := sa.passkeys.MustGet()
	if err != nil {
		return nil, err
	}

	dstPasskeyList := make([]smart_account_proxy.PassKey, 0, len(passkeyList))
	for _, passkey := range passkeyList {
		var pk webauthncose.EC2PublicKeyData
		pk.KeyType = int64(webauthncose.EllipticKey)
		pk.Algorithm = int64(webauthncose.AlgES256)
		pk.Curve = int64(webauthncose.P256)
		pk.XCoord = passkey.PubKeyX.Bytes()
		pk.YCoord = passkey.PubKeyY.Bytes()

		pkBytes, err := webauthncbor.Marshal(pk)
		if err != nil {
			return nil, err
		}

		dstPasskeyList = append(dstPasskeyList, smart_account_proxy.PassKey{
			PublicKey: pkBytes,
			Algo:      passkey.Algo,
		})
	}

	return dstPasskeyList, nil
}

// GetSessions return session key list
func (sa *SmartAccount) GetSessions() ([]smart_account_proxy.SessionKey, error) {
	sessionKeyList, err := sa.sessionKeys.MustGet()
	if err != nil {
		return nil, err
	}

	dstSessionKeyList := make([]smart_account_proxy.SessionKey, 0, len(sessionKeyList))
	for _, sessionKey := range sessionKeyList {
		dstSessionKeyList = append(dstSessionKeyList, smart_account_proxy.SessionKey{
			Addr:          sessionKey.Addr,
			SpendingLimit: sessionKey.SpendingLimit,
			SpentAmount:   sessionKey.SpentAmount,
			ValidUntil:    sessionKey.ValidUntil,
			ValidAfter:    sessionKey.ValidAfter,
		})
	}

	return dstSessionKeyList, nil
}

// ValidateUserOp implements interfaces.IAccount.
// ValidateUserOp return SigValidationFailed when validate failed
// This allows making a "simulation call" without a valid signature
// Other failures (e.g. nonce mismatch, or invalid signature format) should still revert to signal failure.
func (sa *SmartAccount) ValidateUserOp(userOp interfaces.UserOperation, userOpHash [32]byte, missingAccountFunds *big.Int) (validationData *big.Int, err error) {
	validation, err := sa.validateUserOp(&userOp, userOpHash, missingAccountFunds)
	if validation != nil {
		validationData = interfaces.PackValidationData(validation)
	}
	return validationData, err
}

// nolint
func (sa *SmartAccount) validateUserOp(userOp *interfaces.UserOperation, userOpHash [32]byte, missingAccountFunds *big.Int) (*interfaces.Validation, error) {
	if sa.Ctx.From != ethcommon.HexToAddress(common.EntryPointContractAddr) {
		return nil, errors.New("only entrypoint can call ValidateUserOp")
	}

	validationData := &interfaces.Validation{
		SigValidation: interfaces.SigValidationFailed,
	}

	// first use validator to validate, if validator validate failed, use smart account to validate
	validatorList := sa.getAllValidators()
	for _, validator := range validatorList {
		sa.Logger.Infof("post user op, validator: %v", validator)
		validation, err := validator.Validate(userOp, userOpHash)
		if err != nil {
			// validator validate failed, maybe other validator or owner validate successfully
			sa.Logger.Warnf("smart account validator validate user op failed: %v", err)
		}
		if validation.SigValidation == interfaces.SigValidationSucceeded {
			return validation, nil
		}

		if validation.RecoveryAddr != (ethcommon.Address{}) {
			validationData.RecoveryAddr = validation.RecoveryAddr
		}
	}

	// validate owner signature
	addr := validationData.RecoveryAddr
	var err error
	if addr == (ethcommon.Address{}) {
		addr, err = recoveryAddrFromSignature(userOpHash, userOp.Signature)
		if err != nil {
			sa.Logger.Warnf("validate user op failed: %v", err)
			return validationData, nil
		}
	}

	owner, err := sa.GetOwner()
	if err != nil {
		sa.Logger.Warnf("get owner failed: %v", err)
		return validationData, nil
	}
	sa.Logger.Debugf("validate user op, owner: %s, addr: %s, smart account addr: %s", owner.String(), addr.String(), sa.EthAddress.String())

	if addr != owner {
		sa.Logger.Warnf("userOp signature is not from owner, owner: %s, recovery owner: %s", owner.String(), addr.String())
		return validationData, nil
	}
	validationData.SigValidation = interfaces.SigValidationSucceeded

	return validationData, nil
}

func (sa *SmartAccount) Execute(dest ethcommon.Address, value *big.Int, callFunc []byte) error {
	if sa.Ctx.From != ethcommon.HexToAddress(common.EntryPointContractAddr) {
		return errors.New("only entrypoint can call smart account execute")
	}

	sa.Logger.Infof("execute smart account, dest %s, value: %s, callFunc: %x, remainingGas: %s", dest.Hex(), value.Text(10), callFunc, sa.remainingGas.Text(10))

	// support execute smart account method
	if dest == sa.EthAddress {
		// check if callFunc is execute or executeBatch
		methodId := callFunc[:4]
		if string(methodId) == string(sa.Abi.Methods["execute"].ID) || string(methodId) == string(sa.Abi.Methods["executeBatch"].ID) {
			return errors.New("smart account can not call execute or executeBatch recursively")
		}

		_, err := CallMethod(callFunc, sa)
		if err == nil {
			// call successfully
			return nil
		}

		if err != common.ErrMethodNotFound {
			return err
		}
		// if method is not system contract method, call solidity
	}

	// use left gas to call
	_, gasLeft, err := sa.CrossCallEVMContractWithValue(sa.remainingGas, value, dest, callFunc)
	if err != nil {
		return fmt.Errorf("execute smart account callWithValue failed: %v", err)
	}
	sa.remainingGas = big.NewInt(int64(gasLeft))

	totalUsedValue, err := sa.getAxcOfTransferValue(dest, callFunc)
	if err != nil {
		return fmt.Errorf("execute smart account getAxcOfTransferValue failed: %v", err)
	}

	if value != nil {
		totalUsedValue.Add(totalUsedValue, value)
	}
	sa.totalUsedValue.Add(sa.totalUsedValue, totalUsedValue)

	return nil
}

func (sa *SmartAccount) ExecuteBatch(dest []ethcommon.Address, callFunc [][]byte) error {
	if sa.Ctx.From != ethcommon.HexToAddress(common.EntryPointContractAddr) {
		return errors.New("only entrypoint can call smart account execute")
	}

	if len(dest) != len(callFunc) {
		return errors.New("dest and callFunc length mismatch")
	}

	for i := 0; i < len(dest); i++ {
		err := sa.Execute(dest[i], big.NewInt(0), callFunc[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (sa *SmartAccount) ExecuteBatch0(dest []ethcommon.Address, value []*big.Int, callFunc [][]byte) error {
	if sa.Ctx.From != ethcommon.HexToAddress(common.EntryPointContractAddr) {
		return errors.New("only entrypoint can call smart account execute")
	}

	if len(dest) != len(callFunc) || len(dest) != len(value) {
		return errors.New("dest, value and callFunc length mismatch")
	}

	for i := 0; i < len(dest); i++ {
		err := sa.Execute(dest[i], value[i], callFunc[i])
		if err != nil {
			return err
		}
	}

	return nil
}

// SetGuardian set guardian for recover smart account's owner
func (sa *SmartAccount) SetGuardian(guardian ethcommon.Address) error {
	if sa.Ctx.From != ethcommon.HexToAddress(common.EntryPointContractAddr) && sa.Ctx.From != ethcommon.HexToAddress(common.AccountFactoryContractAddr) {
		return errors.New("only entrypoint or smart account factory can call smart account setGuardian")
	}

	if guardian == (ethcommon.Address{}) {
		return nil
	}
	return sa.guardian.Put(guardian)
}

func (sa *SmartAccount) ValidateGuardianSignature(guardianSig []byte, userOpHash [32]byte) error {
	if sa.Ctx.From != ethcommon.HexToAddress(common.EntryPointContractAddr) {
		return errors.New("only entrypoint can call ValidateUserOp")
	}

	_, guardian, err := sa.guardian.Get()
	if err != nil {
		return err
	}

	// validate signature
	addr, err := recoveryAddrFromSignature(userOpHash, guardianSig)
	if err != nil {
		return err
	}

	if addr != guardian {
		return errors.New("invalid guardian signature")
	}

	return nil
}

// ResetOwner recovery owner of smart account by reset owner
func (sa *SmartAccount) ResetOwner(owner ethcommon.Address) error {
	if sa.Ctx.From != ethcommon.HexToAddress(common.EntryPointContractAddr) {
		return errors.New("only entrypoint can call ResetOwner")
	}
	if err := sa.setOwner(owner); err != nil {
		return err
	}

	return nil
}

// ResetAndLockOwner reset owner and lock owner for some times
// lock when guardian try to recovery smart account's owner
func (sa *SmartAccount) ResetAndLockOwner(owner ethcommon.Address) error {
	if sa.Ctx.From != ethcommon.HexToAddress(common.EntryPointContractAddr) {
		return errors.New("only entrypoint can call ResetAndLockOwner")
	}

	if err := sa.setOwner(owner); err != nil {
		return err
	}

	// lock owner for some times, save lock timestamp
	lockTimestamp := sa.Ctx.CurrentEVM.Context.Time + uint64(LockedTime.Seconds())
	if err := sa.status.Put(lockTimestamp); err != nil {
		return err
	}

	sa.EmitEvent(&smart_account.EventUserLocked{
		Sender:     sa.EthAddress,
		NewOwner:   owner,
		LockedTime: new(big.Int).SetUint64(lockTimestamp),
	})

	return nil
}

// SetSession set session key
func (sa *SmartAccount) SetSession(addr ethcommon.Address, spendingLimit *big.Int, validAfter, validUntil uint64) error {
	if sa.Ctx.From != ethcommon.HexToAddress(common.EntryPointContractAddr) {
		return errors.New("only entrypoint can call SetSession")
	}
	if validAfter >= validUntil {
		return errors.New("session key validAfter must less than validUntil")
	}

	newSession := SessionKey{
		Addr:          addr,
		SpendingLimit: spendingLimit,
		SpentAmount:   big.NewInt(0),
		ValidAfter:    validAfter,
		ValidUntil:    validUntil,
	}
	// reset session keys
	return sa.sessionKeys.Put([]SessionKey{newSession})
}

// SetPasskey set passkey
func (sa *SmartAccount) SetPasskey(publicKey []byte, algo uint8) error {
	if sa.Ctx.From != ethcommon.HexToAddress(common.EntryPointContractAddr) {
		return errors.New("only entrypoint can call SetPasskey")
	}

	var pk webauthncose.EC2PublicKeyData
	if err := webauthncbor.Unmarshal(publicKey, &pk); err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}

	newPasskey := Passkey{
		PubKeyX: new(big.Int).SetBytes(pk.XCoord),
		PubKeyY: new(big.Int).SetBytes(pk.YCoord),
		Algo:    algo,
	}
	// reset passkeys
	sa.Logger.Infof("sender %s set passkey successfully, public key: %x", sa.EthAddress.Hex(), publicKey)
	return sa.passkeys.Put([]Passkey{newPasskey})
}

func (sa *SmartAccount) getAllValidators() []interfaces.IValidator {
	var validatorList []interfaces.IValidator
	isExist, sessionKeys, _ := sa.sessionKeys.Get()
	if isExist {
		for _, sessionKey := range sessionKeys {
			validatorList = append(validatorList, &sessionKey)
		}
	}

	isExist, passkeys, _ := sa.passkeys.Get()
	if isExist {
		for _, passkey := range passkeys {
			validatorList = append(validatorList, &passkey)
		}
	}

	return validatorList
}

func (sa *SmartAccount) updateAllValidators(validatorList []interfaces.IValidator) error {
	sessionKeys := make([]SessionKey, 0)
	passkeys := make([]Passkey, 0)
	for _, validator := range validatorList {
		switch v := validator.(type) {
		case *SessionKey:
			sessionKeys = append(sessionKeys, *v)
		case *Passkey:
			passkeys = append(passkeys, *v)
		}
	}

	if err := sa.sessionKeys.Put(sessionKeys); err != nil {
		return err
	}
	if err := sa.passkeys.Put(passkeys); err != nil {
		return err
	}
	return nil
}

// TODO: this function is too business, may be should be self defined by user
func (sa *SmartAccount) getAxcOfTransferValue(token ethcommon.Address, callFunc []byte) (*big.Int, error) {
	if string(callFunc[:4]) == string(transferSig) {
		value := new(big.Int).SetBytes(callFunc[4+32:])

		// transfer to axc value
		tokenPaymaster := TokenPaymasterBuildConfig.Build(sa.CrossCallSystemContractContext())

		// uint128
		axcValue, _ := new(big.Int).SetString("ffffffffffffffffffffffffffffffff", 16)
		tokenValue, err := tokenPaymaster.getTokenValueOfAxc(token, axcValue)
		if err != nil {
			return nil, err
		}

		if value.BitLen() > 128 {
			return nil, errors.New("transfer value exceeds 128 bits, would cause overflow")
		}

		// return axc value
		// value * axcValue / tokenValue
		return new(big.Int).Div(new(big.Int).Mul(value, axcValue), tokenValue), nil
	}
	return big.NewInt(0), nil
}

func (sa *SmartAccount) postUserOp(validateData *interfaces.Validation, actualGasCost, totalValue *big.Int) error {
	if sa.Ctx.From != ethcommon.HexToAddress(common.EntryPointContractAddr) {
		return errors.New("only entrypoint can call postUserOp")
	}

	sa.Logger.Infof("post user op, validate data: %v, actual gas cost: %s", validateData, actualGasCost.String())

	// call all validators PostUserOp
	validatorList := sa.getAllValidators()
	for _, validator := range validatorList {
		sa.Logger.Infof("post user op, validator: %v", validator)
		if err := validator.PostUserOp(validateData, actualGasCost, totalValue); err != nil {
			sa.Logger.Errorf("post user op failed, validator: %v, error: %v", validator, err)
			return err
		}
	}
	// update validators
	return sa.updateAllValidators(validatorList)
}

func recoveryAddrFromSignature(hash [32]byte, signature []byte) (ethcommon.Address, error) {
	if len(signature) != 65 {
		return ethcommon.Address{}, errors.New("invalid signature length")
	}

	ethHash := accounts.TextHash(hash[:])
	// ethers js return r|s|v, v only 1 byte
	// golang return rid, v = rid +27
	sig := make([]byte, len(signature))
	copy(sig, signature)
	if sig[64] >= 27 {
		sig[64] -= 27
	}

	recoveredPub, err := crypto.Ecrecover(ethHash, sig)
	if err != nil {
		return ethcommon.Address{}, err
	}
	pubKey, _ := crypto.UnmarshalPubkey(recoveredPub)
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	return recoveredAddr, nil
}

func CallMethod(callData []byte, sa *SmartAccount) (res CallReturnResult, returnErr error) {
	defer common.Recovery(sa.Logger)

	remainingGas := new(big.Int).SetBytes(sa.getRemainingGas().Bytes())
	totalUsedValue := new(big.Int).SetBytes(sa.getTotalUsedValue().Bytes())

	execRes, err := common.CallSystemContract(sa.Logger, sa, sa.Address.String(), sa.Abi, callData)
	if err != nil {
		return res, err
	}

	res.Result = execRes
	res.UsedGas = new(big.Int).Sub(remainingGas, sa.getRemainingGas()).Uint64()
	res.TotalUsedValue = new(big.Int).Sub(sa.getTotalUsedValue(), totalUsedValue)

	return res, nil
}
