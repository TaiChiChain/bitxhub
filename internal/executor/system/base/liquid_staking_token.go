package base

import "math/big"

type UnlockingRecord struct {
	Amount          *big.Int
	UnlockTimestamp uint64
}

type LiquidStakingTokenInfo struct {
	ID          *big.Int
	Owner       string
	PoolID      uint64
	Principal   *big.Int
	Unlocked    *big.Int
	ActiveEpoch uint64
	// limit: MaxUnlockingRecords
	UnlockingRecords []UnlockingRecord
}

type LiquidStakingToken interface {
	// internal functions(called by System Contract)
	InternalMint(to string) (tokenID *big.Int, err error)
	// burned when Principal is zero and UnlockingRecords is empty
	InternalBurn(tokenID *big.Int) error

	// public functions
	Info(tokenID *big.Int) (info *LiquidStakingTokenInfo, err error)

	GetPendingReward(tokenID *big.Int) *big.Int
	GetUnlockingToken(tokenID *big.Int) *big.Int
	GetUnlockedToken(tokenID *big.Int) *big.Int
	GetLockedToken(tokenID *big.Int) *big.Int
	GetTotalToken(tokenID *big.Int) *big.Int

	BalanceOf(owner string) *big.Int

	OwnerOf(tokenID *big.Int) string

	SafeTransferFrom(from string, to string, tokenID *big.Int, data []byte) error

	SafeTransferFrom2(from string, to string, tokenID *big.Int) error

	TransferFrom(from string, to string, tokenID *big.Int) error

	Approve(to string, tokenID *big.Int) error

	SetApprovalForAll(operator string, approved bool) error

	GetApproved(tokenID *big.Int) (operator string)

	IsApprovedForAll(owner string, operator string) bool
}
