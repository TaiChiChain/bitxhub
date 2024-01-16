package finance

import (
	"math/big"
	"math/rand"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

const mockAddr = "0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013"

func TestNewIncentive(t *testing.T) {
	t.Parallel()
	nvm := system.New()
	genesis := repo.DefaultGenesisConfig(false)
	ledger := newMockMinLedger(t)
	err := system.InitGenesisData(genesis, ledger)
	require.Nil(t, err)

	t.Run("Should revert if no mining or userAcquisition part", func(t *testing.T) {
		genesis.Incentive.Distributions = []*repo.Distribution{
			{
				Name:         "Foundation",
				Addr:         "0x150776D3268c0eAEdAB7d880fd929fe1c5666666",
				Percentage:   0.15,
				InitEmission: 0,
				Locked:       true,
			},
		}
		_, err = NewIncentive(genesis, nvm)
		require.NotNil(t, err)
		require.EqualError(t, ErrNoIncentiveAddrFound, err.Error())
	})

	t.Run("Should create incentive success", func(t *testing.T) {
		genesis = repo.DefaultGenesisConfig(false)
		_, err = NewIncentive(genesis, nvm)
		require.Nil(t, err)
	})
}

func TestIncentive_CalculateMiningRewards(t *testing.T) {
	nvm := system.New()
	genesis := repo.DefaultGenesisConfig(false)
	ledger := newMockMinLedger(t)
	err := system.InitGenesisData(genesis, ledger)
	require.Nil(t, err)

	incentive, err := NewIncentive(genesis, nvm)
	require.Nil(t, err)

	axcAddr := types.NewAddressByStr(common.AXCContractAddr).ETHAddress()
	logs := make([]common.Log, 0)
	incentive.axc.SetContext(&common.VMContext{
		CurrentUser: &axcAddr,
		StateLedger: ledger,
		CurrentLogs: &logs,
	})

	type tests struct {
		name        string
		blockNum    uint64
		expectedRes *big.Int
	}
	totalAmount, _ := new(big.Int).SetString(genesis.Incentive.Mining.TotalAmount, 10)
	blocks := new(big.Int).SetUint64(genesis.Incentive.Mining.BlockNumToHalf)
	var cases = []tests{
		{
			name:        "1st year: 50%",
			blockNum:    1,
			expectedRes: getMiningRes(totalAmount, 5000, blocks),
		},
		{
			name:        "2nd year: 25%",
			blockNum:    genesis.Incentive.Mining.BlockNumToHalf + 1,
			expectedRes: getMiningRes(totalAmount, 2500, blocks),
		},
		{
			name:        "3rd year: 12.5%",
			blockNum:    2*genesis.Incentive.Mining.BlockNumToHalf + 1,
			expectedRes: getMiningRes(totalAmount, 1250, blocks),
		},
		{
			name:        "4th year: 6.25%",
			blockNum:    3*genesis.Incentive.Mining.BlockNumToHalf + 1,
			expectedRes: getMiningRes(totalAmount, 625, blocks),
		},
		{
			name:        "5th year: 6.25%",
			blockNum:    4*genesis.Incentive.Mining.BlockNumToHalf + 1,
			expectedRes: getMiningRes(totalAmount, 625, blocks),
		},
		{
			name:        "6th year: 0%",
			blockNum:    5*genesis.Incentive.Mining.BlockNumToHalf + 1,
			expectedRes: big.NewInt(0),
		},
	}
	for tt := range cases {
		t.Run(cases[tt].name, func(t *testing.T) {
			addr := types.NewAddressByStr(mockAddr).ETHAddress()
			beforeBalance := incentive.axc.BalanceOf(addr)
			err = incentive.SetMiningRewards(addr, ledger, cases[tt].blockNum)
			require.Nil(t, err)

			afterBalance := incentive.axc.BalanceOf(addr)
			require.Equal(t, cases[tt].expectedRes.String(), new(big.Int).Sub(afterBalance, beforeBalance).String())
		})
	}
}

func TestIncentive_CalculateUserAcquisitionReward(t *testing.T) {
	addrNum := 10
	nvm := system.New()
	genesis := repo.DefaultGenesisConfig(false)
	ledger := newMockMinLedger(t)
	err := system.InitGenesisData(genesis, ledger)
	require.Nil(t, err)

	incentive, err := NewIncentive(genesis, nvm)
	require.Nil(t, err)

	axcAddr := types.NewAddressByStr(common.AXCContractAddr).ETHAddress()
	logs := make([]common.Log, 0)
	incentive.axc.SetContext(&common.VMContext{
		CurrentUser: &axcAddr,
		StateLedger: ledger,
		CurrentLogs: &logs,
	})
	addrs := mockAddrs(addrNum)
	block := types.Block{
		Transactions: mockTxs(addrs),
	}

	err = incentive.SetUserAcquisitionReward(ledger, 1, &block)
	require.Nil(t, err)

	sum := big.NewInt(0)
	for i := 0; i < addrNum; i++ {
		balance := incentive.axc.BalanceOf(addrs[i].Addr.ETHAddress())
		sum = new(big.Int).Add(sum, balance)
	}
	require.Equal(t, genesis.Incentive.UserAcquisition.AvgBlockReward, sum.String())
}

func TestIncentive_Unlock(t *testing.T) {
	nvm := system.New()
	genesis := repo.DefaultGenesisConfig(false)
	ledger := newMockMinLedger(t)
	err := system.InitGenesisData(genesis, ledger)
	require.Nil(t, err)

	incentive, err := NewIncentive(genesis, nvm)
	require.Nil(t, err)

	incentive.Unlock(ethcommon.Address{}, nil)
}

func mockAddrs(num int) []*types.Signer {
	res := make([]*types.Signer, num)
	for i := 0; i < num; i++ {
		res[i], _ = types.GenerateSigner()
	}
	return res
}

func mockTxs(addrs []*types.Signer) []*types.Transaction {
	res := make([]*types.Transaction, len(addrs))
	for i := 0; i < len(addrs); i++ {
		index := rand.Intn(len(addrs))
		signer := addrs[index].Addr.ETHAddress()
		inner := &types.IncentiveTx{
			IncentiveAddress: &signer,
		}
		res[i] = &types.Transaction{
			Inner: inner,
			Time:  time.Now(),
		}
	}
	return res
}

func getMiningRes(totalAmount *big.Int, percentage int64, blocks *big.Int) *big.Int {
	totalReward := new(big.Int).Div(new(big.Int).Mul(totalAmount, big.NewInt(percentage)), big.NewInt(10000))
	res := new(big.Int).Div(totalReward, blocks)
	return res
}
