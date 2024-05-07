package finance

import (
	"math/big"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

type Incentive struct {
	repo                *repo.GenesisConfig
	logger              logrus.FieldLogger
	miningAddr          ethcommon.Address
	userAcquisitionAddr ethcommon.Address

	totalAmount *big.Int

	// todo: change the remaining in memory to the receiver address in the future, in case restart vanish the data
	remaining *big.Int

	blockNumHalf uint64
	rules        []MiningRules
}

func NewIncentive(config *repo.GenesisConfig) (*Incentive, error) {
	logger := loggers.Logger(loggers.Finance)

	return &Incentive{
		repo:   config,
		logger: logger,
	}, nil
}

func (in *Incentive) SetMiningRewards(nodeId uint64, ledger ledger.StateLedger, gasFee *big.Int) error {
	stakingManager := framework.StakingManagerBuildConfig.Build(common.NewVMContextByExecutor(ledger))
	reward, err := stakingManager.InternalRecordReward(nodeId, gasFee)
	if err != nil {
		return err
	}
	in.logger.Infof("SetMiningRewards to node %d, reward: %d", nodeId, reward.Uint64())
	return nil
}

func (in *Incentive) calculateMiningRewards(currentBlock uint64) *big.Int {
	return big.NewInt(0)
}

func (in *Incentive) Unlock(_ ethcommon.Address, _ *big.Int) {
	// todo: implement it
}

func (in *Incentive) SetUserAcquisitionReward(ledger ledger.StateLedger, blockHeight uint64, block *types.Block) error {
	if blockHeight > in.repo.Incentive.Referral.BlockToNone {
		return nil
	}
	return in.setUserAcquisitionReward(block)
}

func (in *Incentive) setUserAcquisitionReward(block *types.Block) error {
	// axc := token.AXCBuildConfig.Build(common.NewVMContextByExecutor(ledger))
	return nil
}
