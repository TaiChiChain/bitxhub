package governance

import (
	"encoding/json"
	"errors"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/base"
)

var (
	ErrChainParamInvalid                  = errors.New("chain param invalid")
	ErrRepeatedChainParam                 = errors.New("repeated chain param")
	ErrExistNotFinishedChainParamProposal = errors.New("propose chain param proposal failed, exist not finished proposal")
)

var _ IExecutor = (*ChainParamManager)(nil)

type ChainParamExtraArgs struct {
	EpochPeriod                                        uint64
	MaxValidatorNum                                    uint64
	BlockMaxTxNum                                      uint64
	EnableTimedGenEmptyBlock                           bool
	AbnormalNodeExcludeView                            uint64
	AgainProposeIntervalBlockInValidatorsNumPercentage uint64
	ContinuousNullRequestToleranceNumber               uint64
	RebroadcastToleranceNumber                         uint64
}

type ChainParamManager struct {
	gov *Governance
}

func NewChainParamManager(gov *Governance) *ChainParamManager {
	return &ChainParamManager{
		gov: gov,
	}
}

// ProposeCheck implements IExecutor.
func (c *ChainParamManager) ProposeCheck(proposalType ProposalType, extra []byte) error {
	_, err := c.getChainParamExtraArgs(extra)
	if err != nil {
		return err
	}

	return c.checkNotFinishedProposal()
}

// Execute implements IExecutor.
func (c *ChainParamManager) Execute(proposal *Proposal) error {
	if proposal.Status == Approved {
		extraArgs, err := c.getChainParamExtraArgs(proposal.Extra)
		if err != nil {
			return err
		}

		epochInfo, err := base.GetNextEpochInfo(c.gov.stateLedger)
		if err != nil {
			return err
		}

		epochInfo.EpochPeriod = extraArgs.EpochPeriod
		epochInfo.ConsensusParams.MaxValidatorNum = extraArgs.MaxValidatorNum
		epochInfo.ConsensusParams.BlockMaxTxNum = extraArgs.BlockMaxTxNum
		epochInfo.ConsensusParams.AbnormalNodeExcludeView = extraArgs.AbnormalNodeExcludeView
		epochInfo.ConsensusParams.AgainProposeIntervalBlockInValidatorsNumPercentage = extraArgs.AgainProposeIntervalBlockInValidatorsNumPercentage
		epochInfo.ConsensusParams.ContinuousNullRequestToleranceNumber = extraArgs.ContinuousNullRequestToleranceNumber
		epochInfo.ConsensusParams.ReBroadcastToleranceNumber = extraArgs.RebroadcastToleranceNumber
		epochInfo.ConsensusParams.EnableTimedGenEmptyBlock = extraArgs.EnableTimedGenEmptyBlock
		if err := base.SetNextEpochInfo(c.gov.stateLedger, epochInfo); err != nil {
			return err
		}
	}

	return nil
}

func (c *ChainParamManager) getChainParamExtraArgs(extra []byte) (*ChainParamExtraArgs, error) {
	extraArgs := &ChainParamExtraArgs{}
	if err := json.Unmarshal(extra, extraArgs); err != nil {
		c.gov.logger.Errorf("Unmarshal extra args failed: %v", err)
		return nil, ErrChainParamInvalid
	}

	// check param correct
	if extraArgs.EpochPeriod == 0 || extraArgs.MaxValidatorNum < 4 || extraArgs.BlockMaxTxNum == 0 ||
		extraArgs.AbnormalNodeExcludeView == 0 || extraArgs.AgainProposeIntervalBlockInValidatorsNumPercentage >= 100 ||
		extraArgs.ContinuousNullRequestToleranceNumber == 0 || extraArgs.RebroadcastToleranceNumber == 0 {
		return nil, ErrChainParamInvalid
	}

	// check whether chain param and GetNextEpochInfo are consistent
	epochInfo, err := base.GetNextEpochInfo(c.gov.stateLedger)
	if err != nil {
		return nil, err
	}
	consensusParams := epochInfo.ConsensusParams
	if epochInfo.EpochPeriod == extraArgs.EpochPeriod && consensusParams.MaxValidatorNum == extraArgs.MaxValidatorNum &&
		consensusParams.BlockMaxTxNum == extraArgs.BlockMaxTxNum && consensusParams.AbnormalNodeExcludeView == extraArgs.AbnormalNodeExcludeView &&
		consensusParams.AgainProposeIntervalBlockInValidatorsNumPercentage == extraArgs.AgainProposeIntervalBlockInValidatorsNumPercentage &&
		consensusParams.ContinuousNullRequestToleranceNumber == extraArgs.ContinuousNullRequestToleranceNumber &&
		consensusParams.ReBroadcastToleranceNumber == extraArgs.RebroadcastToleranceNumber && consensusParams.EnableTimedGenEmptyBlock == extraArgs.EnableTimedGenEmptyBlock {
		return nil, ErrRepeatedChainParam
	}

	return extraArgs, nil
}

func (c *ChainParamManager) checkNotFinishedProposal() error {
	notFinishedProposals, err := c.gov.notFinishedProposalMgr.GetProposals()
	if err != nil {
		return err
	}

	if len(notFinishedProposals) > 0 {
		return ErrExistNotFinishedChainParamProposal
	}

	return nil
}
