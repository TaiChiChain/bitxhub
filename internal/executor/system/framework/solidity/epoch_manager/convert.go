package epoch_manager

import (
	"github.com/axiomesh/axiom-kit/types"
)

func (e *EpochInfo) ToTypesEpoch() *types.EpochInfo {
	return &types.EpochInfo{
		Epoch:       e.Epoch,
		EpochPeriod: e.EpochPeriod,
		StartBlock:  e.StartBlock,
		ConsensusParams: types.ConsensusParams{
			ProposerElectionType:          e.ConsensusParams.ProposerElectionType,
			CheckpointPeriod:              e.ConsensusParams.CheckpointPeriod,
			HighWatermarkCheckpointPeriod: e.ConsensusParams.HighWatermarkCheckpointPeriod,
			MaxValidatorNum:               e.ConsensusParams.MaxValidatorNum,
			MinValidatorNum:               e.ConsensusParams.MinValidatorNum,
			BlockMaxTxNum:                 e.ConsensusParams.BlockMaxTxNum,
			EnableTimedGenEmptyBlock:      e.ConsensusParams.EnableTimedGenEmptyBlock,
			NotActiveWeight:               e.ConsensusParams.NotActiveWeight,
			AbnormalNodeExcludeView:       e.ConsensusParams.AbnormalNodeExcludeView,
			AgainProposeIntervalBlockInValidatorsNumPercentage: e.ConsensusParams.AgainProposeIntervalBlockInValidatorsNumPercentage,
			ContinuousNullRequestToleranceNumber:               e.ConsensusParams.ContinuousNullRequestToleranceNumber,
			ReBroadcastToleranceNumber:                         e.ConsensusParams.ReBroadcastToleranceNumber,
		},
		FinanceParams: types.FinanceParams{
			GasLimit:    e.FinanceParams.GasLimit,
			MinGasPrice: types.CoinNumberByBigInt(e.FinanceParams.MinGasPrice),
		},
		StakeParams: types.StakeParams{
			StakeEnable:                      e.StakeParams.StakeEnable,
			MaxAddStakeRatio:                 e.StakeParams.MaxAddStakeRatio,
			MaxUnlockStakeRatio:              e.StakeParams.MaxUnlockStakeRatio,
			MaxUnlockingRecordNum:            e.StakeParams.MaxUnlockingRecordNum,
			UnlockPeriod:                     e.StakeParams.UnlockPeriod,
			MaxPendingInactiveValidatorRatio: e.StakeParams.MaxPendingInactiveValidatorRatio,
			MinDelegateStake:                 types.CoinNumberByBigInt(e.StakeParams.MinDelegateStake),
			MinValidatorStake:                types.CoinNumberByBigInt(e.StakeParams.MinValidatorStake),
			MaxValidatorStake:                types.CoinNumberByBigInt(e.StakeParams.MaxValidatorStake),
			EnablePartialUnlock:              e.StakeParams.EnablePartialUnlock,
		},
		MiscParams: types.MiscParams{
			TxMaxSize: e.MiscParams.TxMaxSize,
		},
	}
}

func FromTypesEpoch(e types.EpochInfo) EpochInfo {
	return EpochInfo{
		Epoch:       e.Epoch,
		EpochPeriod: e.EpochPeriod,
		StartBlock:  e.StartBlock,
		ConsensusParams: ConsensusParams{
			ProposerElectionType:          e.ConsensusParams.ProposerElectionType,
			CheckpointPeriod:              e.ConsensusParams.CheckpointPeriod,
			HighWatermarkCheckpointPeriod: e.ConsensusParams.HighWatermarkCheckpointPeriod,
			MaxValidatorNum:               e.ConsensusParams.MaxValidatorNum,
			MinValidatorNum:               e.ConsensusParams.MinValidatorNum,
			BlockMaxTxNum:                 e.ConsensusParams.BlockMaxTxNum,
			EnableTimedGenEmptyBlock:      e.ConsensusParams.EnableTimedGenEmptyBlock,
			NotActiveWeight:               e.ConsensusParams.NotActiveWeight,
			AbnormalNodeExcludeView:       e.ConsensusParams.AbnormalNodeExcludeView,
			AgainProposeIntervalBlockInValidatorsNumPercentage: e.ConsensusParams.AgainProposeIntervalBlockInValidatorsNumPercentage,
			ContinuousNullRequestToleranceNumber:               e.ConsensusParams.ContinuousNullRequestToleranceNumber,
			ReBroadcastToleranceNumber:                         e.ConsensusParams.ReBroadcastToleranceNumber,
		},
		FinanceParams: FinanceParams{
			GasLimit:    e.FinanceParams.GasLimit,
			MinGasPrice: e.FinanceParams.MinGasPrice.ToBigInt(),
		},
		StakeParams: StakeParams{
			StakeEnable:                      e.StakeParams.StakeEnable,
			MaxAddStakeRatio:                 e.StakeParams.MaxAddStakeRatio,
			MaxUnlockStakeRatio:              e.StakeParams.MaxUnlockStakeRatio,
			MaxUnlockingRecordNum:            e.StakeParams.MaxUnlockingRecordNum,
			UnlockPeriod:                     e.StakeParams.UnlockPeriod,
			MaxPendingInactiveValidatorRatio: e.StakeParams.MaxPendingInactiveValidatorRatio,
			MinDelegateStake:                 e.StakeParams.MinDelegateStake.ToBigInt(),
			MinValidatorStake:                e.StakeParams.MinValidatorStake.ToBigInt(),
			MaxValidatorStake:                e.StakeParams.MaxValidatorStake.ToBigInt(),
			EnablePartialUnlock:              e.StakeParams.EnablePartialUnlock,
		},
		MiscParams: MiscParams{
			TxMaxSize: e.MiscParams.TxMaxSize,
		},
	}
}
