package interfaces

import (
	"math/big"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
)

type PostOpMode uint

const (
	OpSucceeded PostOpMode = iota
	OpReverted
	PostOpReverted
)

// IPaymaster is the interface exposed by a paymaster contract, who agrees to pay the gas for user's operations.
// a paymaster must hold a stake to cover the required entrypoint stake and also the gas for the transaction.
type IPaymaster interface {
	common.SystemContract

	// ValidatePaymasterUserOp check if paymaster agrees to pay.
	// Must verify sender is the entryPoint.
	// Revert to reject this request.
	// Note that bundlers will reject this method if it changes the state, unless the paymaster is trusted (whitelisted)
	// The paymaster pre-pays using its deposit, and receive back a refund after the postOp method returns.
	ValidatePaymasterUserOp(userOp UserOperation, userOpHash [32]byte, maxCost *big.Int) (context []byte, validationData *big.Int, err error)

	PostOp(mode PostOpMode, context []byte, actualGasCost *big.Int) error
}
