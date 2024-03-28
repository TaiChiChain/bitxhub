package saccount

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/interfaces"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
)

const (
	VALID_TIMESTAMP_OFFSET = 20
	SIGNATURE_OFFSET       = 84

	verifyingPaymasterOwnerKey = "verifying_owner_key"
)

var _ interfaces.IPaymaster = (*VerifyingPaymaster)(nil)

/**
 * A sample paymaster that uses external service to decide whether to pay for the UserOp.
 * The paymaster trusts an external signer to sign the transaction.
 * The calling user must pass the UserOp to that external signer first, which performs
 * whatever off-chain verification before signing the UserOp.
 * Note that this signature is NOT a replacement for the account-specific signature:
 * - the paymaster checks a signature to agree to PAY for GAS.
 * - the account checks a signature to prove identity and account ownership.
 */
type VerifyingPaymaster struct {
	entryPoint interfaces.IEntryPoint
	selfAddr   *types.Address

	account ledger.IAccount
	// context fields
	currentUser *ethcommon.Address
	currentLogs *[]common.Log
	stateLedger ledger.StateLedger
	currentEVM  *vm.EVM
}

func NewVerifyingPaymaster(entryPoint interfaces.IEntryPoint) *VerifyingPaymaster {
	return &VerifyingPaymaster{entryPoint: entryPoint, selfAddr: types.NewAddressByStr(common.VerifyingPaymasterContractAddr)}
}

func InitializeVerifyingPaymaster(stateLedger ledger.StateLedger, owner ethcommon.Address) {
	account := stateLedger.GetOrCreateAccount(types.NewAddressByStr(common.VerifyingPaymasterContractAddr))
	account.SetState([]byte(verifyingPaymasterOwnerKey), owner.Bytes())
}

func (vp *VerifyingPaymaster) SetOwner(owner ethcommon.Address) {
	vp.account.SetState([]byte(verifyingPaymasterOwnerKey), owner.Bytes())
}

func (vp *VerifyingPaymaster) GetOwner() ethcommon.Address {
	isExist, ownerBytes := vp.account.GetState([]byte(verifyingPaymasterOwnerKey))
	if isExist {
		return ethcommon.BytesToAddress(ownerBytes)
	}
	return ethcommon.Address{}
}

func (vp *VerifyingPaymaster) SetContext(context *common.VMContext) {
	vp.currentUser = context.CurrentUser
	vp.currentLogs = context.CurrentLogs
	vp.stateLedger = context.StateLedger
	vp.currentEVM = context.CurrentEVM

	vp.account = vp.stateLedger.GetOrCreateAccount(vp.selfAddress())
}

func (vp *VerifyingPaymaster) selfAddress() *types.Address {
	return vp.selfAddr
}

// PostOp implements interfaces.IPaymaster.
func (vp *VerifyingPaymaster) PostOp(mode interfaces.PostOpMode, context []byte, actualGasCost *big.Int) error {
	if vp.currentUser.Hex() != common.EntryPointContractAddr {
		return common.NewRevertStringError("only entrypoint can call verifying paymaster post op")
	}
	return nil
}

// ValidatePaymasterUserOp implements interfaces.IPaymaster.
func (vp *VerifyingPaymaster) ValidatePaymasterUserOp(userOp interfaces.UserOperation, userOpHash []byte, maxCost *big.Int) (context []byte, validationData *big.Int, err error) {
	if vp.currentUser.Hex() != common.EntryPointContractAddr {
		return nil, nil, common.NewRevertStringError("only entrypoint can call validate paymaster user op")
	}

	context, validation, err := vp.validatePaymasterUserOp(userOp, userOpHash, maxCost)
	if validation != nil {
		validationData = interfaces.PackValidationData(validation)
	}

	return context, validationData, err
}

// nolint
func (vp *VerifyingPaymaster) validatePaymasterUserOp(userOp interfaces.UserOperation, userOpHash []byte, maxCost *big.Int) (context []byte, validationData *interfaces.Validation, err error) {
	validUntil, validAfter, signature, err := parsePaymasterAndData(userOp.PaymasterAndData)
	if err != nil {
		return nil, nil, common.NewRevertStringError(fmt.Sprintf("validate paymaster user op failed: %s", err.Error()))
	}

	if len(signature) != 65 {
		return nil, nil, common.NewRevertStringError("verifying paymaster: invalid signature length in paymasterAndData")
	}

	validationData = &interfaces.Validation{
		SigValidation: interfaces.SigValidationFailed,
	}
	// validate signature
	// paymaster validate hash is not the user op hash
	addr, err := recoveryAddrFromSignature(vp.getHash(userOp, validUntil, validAfter), signature)
	if err != nil {
		return []byte(""), validationData, common.NewRevertStringError("paymaster validate user op signature error")
	}

	if addr != vp.GetOwner() {
		return []byte(""), validationData, nil
	}
	validationData.SigValidation = interfaces.SigValidationSucceeded
	return []byte(""), validationData, nil
}

func (vp *VerifyingPaymaster) getHash(userOp interfaces.UserOperation, validUntil, validAfter *big.Int) []byte {
	return crypto.Keccak256(
		pack(userOp),
		ethcommon.LeftPadBytes(vp.currentEVM.ChainConfig().ChainID.Bytes(), 32),
		ethcommon.LeftPadBytes(vp.selfAddr.Bytes(), 32),
		ethcommon.LeftPadBytes(validUntil.Bytes(), 32),
		ethcommon.LeftPadBytes(validAfter.Bytes(), 32),
	)
}

func parsePaymasterAndData(paymasterAndData []byte) (validUntil, validAfter *big.Int, signature []byte, err error) {
	if len(paymasterAndData) < SIGNATURE_OFFSET {
		return nil, nil, nil, common.NewRevertStringError("parse paymasterAndData failed, length is too short")
	}

	validTimeData := paymasterAndData[VALID_TIMESTAMP_OFFSET:SIGNATURE_OFFSET]

	arg := abi.Arguments{
		{Name: "validUntil", Type: common.UInt48Type},
		{Name: "validAfter", Type: common.UInt48Type},
	}
	validTime, err := arg.Unpack(validTimeData)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(validTime) != 2 {
		return nil, nil, nil, common.NewRevertStringError("parse valid time failed from paymasterAndData")
	}
	validUntil = validTime[0].(*big.Int)
	validAfter = validTime[1].(*big.Int)

	signature = paymasterAndData[SIGNATURE_OFFSET:]
	return validUntil, validAfter, signature, nil
}

func pack(userOp interfaces.UserOperation) []byte {
	args := abi.Arguments{
		{Name: "sender", Type: common.AddressType},
		{Name: "nonce", Type: common.BigIntType},
		{Name: "initCode", Type: common.Bytes32Type},
		{Name: "callData", Type: common.Bytes32Type},
		{Name: "callGasLimit", Type: common.BigIntType},
		{Name: "verificationGasLimit", Type: common.BigIntType},
		{Name: "preVerificationGas", Type: common.BigIntType},
		{Name: "maxFeePerGas", Type: common.BigIntType},
		{Name: "maxPriorityFeePerGas", Type: common.BigIntType},
	}
	packed, _ := args.Pack(
		userOp.Sender,
		userOp.Nonce,
		crypto.Keccak256Hash(userOp.InitCode),
		crypto.Keccak256Hash(userOp.CallData),
		userOp.CallGasLimit,
		userOp.VerificationGasLimit,
		userOp.PreVerificationGas,
		userOp.MaxFeePerGas,
		userOp.MaxPriorityFeePerGas,
	)

	return packed
}
