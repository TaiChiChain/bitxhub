package interfaces

import (
	"math/big"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
)

type UserOperation struct {
	Sender               ethcommon.Address `json:"sender"`
	Nonce                *big.Int          `json:"nonce"`
	InitCode             []byte            `json:"initCode"`
	CallData             []byte            `json:"callData"`
	CallGasLimit         *big.Int          `json:"callGasLimit"`
	VerificationGasLimit *big.Int          `json:"verificationGasLimit"`
	PreVerificationGas   *big.Int          `json:"preVerificationGas"`
	MaxFeePerGas         *big.Int          `json:"maxFeePerGas"`
	MaxPriorityFeePerGas *big.Int          `json:"maxPriorityFeePerGas"`
	PaymasterAndData     []byte            `json:"paymasterAndData"`
	Signature            []byte            `json:"signature"`
}

func PackForSignature(userOp *UserOperation) []byte {
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
		{Name: "paymasterAndData", Type: common.Bytes32Type},
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
		crypto.Keccak256Hash(userOp.PaymasterAndData),
	)

	return packed
}

// GetUserOpHash returns the hash of the userOp + entryPoint address + chainID.
func GetUserOpHash(userOp *UserOperation, entryPoint ethcommon.Address, chainID *big.Int) ethcommon.Hash {
	return crypto.Keccak256Hash(
		crypto.Keccak256(PackForSignature(userOp)),
		ethcommon.LeftPadBytes(entryPoint.Bytes(), 32),
		ethcommon.LeftPadBytes(chainID.Bytes(), 32),
	)
}

func GetGasPrice(userOp *UserOperation) *big.Int {
	return math.BigMin(userOp.MaxFeePerGas, userOp.MaxPriorityFeePerGas)
}
