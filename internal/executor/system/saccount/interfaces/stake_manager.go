package interfaces

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

var (
	StakeInfoType, _ = abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{
			Name: "stake",
			Type: "uint256",
		},
		{
			Name: "unstakeDelaySec",
			Type: "uint256",
		},
	})
)

type DepositInfo struct {
	Deposit         *big.Int
	Staked          bool
	Stake           *big.Int
	UnstakeDelaySec uint32
	WithdrawTime    *big.Int
}

// API struct used by getStakeInfo and simulateValidation
type StakeInfo struct {
	Stake           *big.Int
	UnstakeDelaySec *big.Int
}

type IStakeManager interface {
	GetDepositInfo(account ethcommon.Address) *DepositInfo

	BalanceOf(account ethcommon.Address) *big.Int

	DepositTo(account ethcommon.Address)

	AddStake(unstakeDelaySec uint32)

	UnlockStake()

	WithdrawStake(withdrawAddress ethcommon.Address)

	WithdrawTo(withdrawAddress ethcommon.Address, withdrawAmount *big.Int)
}
