package interfaces

import "math/big"

type IValidator interface {
	// Validate signature
	Validate(userOp *UserOperation, userOpHash [32]byte) (*Validation, error)

	PostUserOp(validation *Validation, actualGasCost, totalValue *big.Int) error
}
