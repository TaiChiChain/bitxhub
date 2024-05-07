package executor

import (
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
	syscommon "github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework/solidity/node_manager"
)

func (exec *BlockExecutor) updateEpochInfo(block *types.Block) {
	// check need turn into NewEpoch
	epochInfo := exec.chainState.EpochInfo
	if block.Header.Number == (epochInfo.StartBlock + epochInfo.EpochPeriod - 1) {
		nodeManagerContract := framework.NodeManagerBuildConfig.Build(syscommon.NewVMContextByExecutor(exec.ledger.StateLedger))
		votingPowers, err := nodeManagerContract.GetActiveValidatorVotingPowers()
		if err != nil {
			panic(err)
		}

		epochManagerContract := framework.EpochManagerBuildConfig.Build(syscommon.NewVMContextByExecutor(exec.ledger.StateLedger))
		newEpoch, err := epochManagerContract.TurnIntoNewEpoch()
		if err != nil {
			panic(err)
		}

		if err := exec.chainState.UpdateByEpochInfo(newEpoch, lo.SliceToMap(votingPowers, func(item node_manager.ConsensusVotingPower) (uint64, int64) {
			return item.NodeID, item.ConsensusVotingPower
		})); err != nil {
			panic(err)
		}

		exec.logger.WithFields(logrus.Fields{
			"height":                block.Header.Number,
			"new_epoch":             newEpoch.Epoch,
			"new_epoch_start_block": newEpoch.StartBlock,
		}).Info("Turn into new epoch")
	}
}

func (exec *BlockExecutor) updateMiningInfo(block *types.Block) {
	// calculate mining rewards and transfer the mining reward
	if err := exec.incentive.SetMiningRewards(block.Header.ProposerNodeID, exec.ledger.StateLedger,
		block.Header.GasFeeReward); err != nil {
		exec.logger.WithFields(logrus.Fields{
			"height": block.Height(),
			"err":    err.Error(),
		}).Errorf("set mining rewards error")
		// should panic the error, since there is no chance to get an error only if the logic is wrong
		panic(err)
	}
}
