package finance

import (
	"math/big"
	"testing"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/token/axc"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	"github.com/stretchr/testify/require"
)

const mockAddr = "0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013"

func TestIncentive_Record(t *testing.T) {
	nvm := system.New()
	genesis := repo.DefaultGenesisConfig(false)
	ledger := axc.NewMockMinLedger(t)
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
			expectedRes: getRes(totalAmount, 5000, blocks),
		},
		{
			name:        "2nd year: 25%",
			blockNum:    genesis.Incentive.Mining.BlockNumToHalf + 1,
			expectedRes: getRes(totalAmount, 2500, blocks),
		},
		{
			name:        "3rd year: 12.5%",
			blockNum:    2*genesis.Incentive.Mining.BlockNumToHalf + 1,
			expectedRes: getRes(totalAmount, 1250, blocks),
		},
		{
			name:        "4th year: 6.25%",
			blockNum:    3*genesis.Incentive.Mining.BlockNumToHalf + 1,
			expectedRes: getRes(totalAmount, 625, blocks),
		},
		{
			name:        "5th year: 6.25%",
			blockNum:    4*genesis.Incentive.Mining.BlockNumToHalf + 1,
			expectedRes: getRes(totalAmount, 625, blocks),
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
			err = incentive.CalculateMiningRewards(addr, ledger, cases[tt].blockNum)
			require.Nil(t, err)

			afterBalance := incentive.axc.BalanceOf(addr)
			require.Equal(t, cases[tt].expectedRes.String(), new(big.Int).Sub(afterBalance, beforeBalance).String())
		})
	}
}

func getRes(totalAmount *big.Int, percentage int64, blocks *big.Int) *big.Int {
	totalReward := new(big.Int).Div(new(big.Int).Mul(totalAmount, big.NewInt(percentage)), big.NewInt(10000))
	res := new(big.Int).Div(totalReward, blocks)
	return res
}
