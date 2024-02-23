package common

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
)

const (
	NOT_ENTERED = 1
	ENTERED     = 2
)

type ReentrancyGuard struct {
	status uint
}

func NewReentrancyGuard() *ReentrancyGuard {
	return &ReentrancyGuard{status: NOT_ENTERED}
}

func (rg *ReentrancyGuard) Enter() error {
	if rg.status == ENTERED {
		return ReentrancyGuardReentrantCall()
	}

	rg.status = ENTERED
	return nil
}

func (rg *ReentrancyGuard) Exit() {
	rg.status = NOT_ENTERED
}

func (rg *ReentrancyGuard) IsEntered() bool {
	return rg.status == ENTERED
}

func ReentrancyGuardReentrantCall() error {
	return NewRevertError("ReentrancyGuardReentrantCall", abi.Arguments{}, nil)
}
