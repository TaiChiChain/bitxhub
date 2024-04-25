package governance

import (
	"encoding/json"
	"errors"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

var (
	ErrChainParamInvalid                  = errors.New("chain param invalid")
	ErrRepeatedChainParam                 = errors.New("repeated chain param")
	ErrExistNotFinishedChainParamProposal = errors.New("propose chain param proposal failed, exist not finished proposal")
)

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
	DefaultProposalPermissionManager
}

func NewChainParamManager(gov *Governance) *ChainParamManager {
	return &ChainParamManager{
		gov:                              gov,
		DefaultProposalPermissionManager: NewDefaultProposalPermissionManager(gov),
	}
}

func (c *ChainParamManager) GenesisInit(genesis *repo.GenesisConfig) error {
	return nil
}

func (c *ChainParamManager) SetContext(ctx *common.VMContext) {}

func (c *ChainParamManager) ProposeArgsCheck(proposalType ProposalType, title, desc string, blockNumber uint64, extra []byte) error {
	_, err := c.getChainParamExtraArgs(extra)
	if err != nil {
		return err
	}

	return c.checkNotFinishedProposal()
}

func (c *ChainParamManager) VotePassExecute(proposal *Proposal) error {
	extraArgs, err := c.getChainParamExtraArgs(proposal.Extra)
	if err != nil {
		return err
	}

	epochManagerContract := framework.EpochManagerBuildConfig.Build(c.gov.CrossCallSystemContractContext())
	nextEpochInfo, err := epochManagerContract.NextEpoch()
	if err != nil {
		return err
	}

	nextEpochInfo.EpochPeriod = extraArgs.EpochPeriod
	nextEpochInfo.ConsensusParams.MaxValidatorNum = extraArgs.MaxValidatorNum
	nextEpochInfo.ConsensusParams.BlockMaxTxNum = extraArgs.BlockMaxTxNum
	nextEpochInfo.ConsensusParams.AbnormalNodeExcludeView = extraArgs.AbnormalNodeExcludeView
	nextEpochInfo.ConsensusParams.AgainProposeIntervalBlockInValidatorsNumPercentage = extraArgs.AgainProposeIntervalBlockInValidatorsNumPercentage
	nextEpochInfo.ConsensusParams.ContinuousNullRequestToleranceNumber = extraArgs.ContinuousNullRequestToleranceNumber
	nextEpochInfo.ConsensusParams.ReBroadcastToleranceNumber = extraArgs.RebroadcastToleranceNumber
	nextEpochInfo.ConsensusParams.EnableTimedGenEmptyBlock = extraArgs.EnableTimedGenEmptyBlock
	if err := epochManagerContract.UpdateNextEpoch(nextEpochInfo); err != nil {
		return err
	}
	return nil
}

func (c *ChainParamManager) getChainParamExtraArgs(extra []byte) (*ChainParamExtraArgs, error) {
	extraArgs := &ChainParamExtraArgs{}
	if err := json.Unmarshal(extra, extraArgs); err != nil {
		c.gov.Logger.Errorf("Unmarshal extra args failed: %v", err)
		return nil, ErrChainParamInvalid
	}

	// check param correct
	if extraArgs.EpochPeriod == 0 || extraArgs.MaxValidatorNum < 4 || extraArgs.BlockMaxTxNum == 0 ||
		extraArgs.AbnormalNodeExcludeView == 0 || extraArgs.AgainProposeIntervalBlockInValidatorsNumPercentage >= 100 ||
		extraArgs.ContinuousNullRequestToleranceNumber == 0 || extraArgs.RebroadcastToleranceNumber == 0 {
		return nil, ErrChainParamInvalid
	}

	// check whether chain param and GetNextEpochInfo are consistent
	epochManagerContract := framework.EpochManagerBuildConfig.Build(c.gov.CrossCallSystemContractContext())
	nextEpochInfo, err := epochManagerContract.NextEpoch()
	if err != nil {
		return nil, err
	}
	consensusParams := nextEpochInfo.ConsensusParams
	if nextEpochInfo.EpochPeriod == extraArgs.EpochPeriod && consensusParams.MaxValidatorNum == extraArgs.MaxValidatorNum &&
		consensusParams.BlockMaxTxNum == extraArgs.BlockMaxTxNum && consensusParams.AbnormalNodeExcludeView == extraArgs.AbnormalNodeExcludeView &&
		consensusParams.AgainProposeIntervalBlockInValidatorsNumPercentage == extraArgs.AgainProposeIntervalBlockInValidatorsNumPercentage &&
		consensusParams.ContinuousNullRequestToleranceNumber == extraArgs.ContinuousNullRequestToleranceNumber &&
		consensusParams.ReBroadcastToleranceNumber == extraArgs.RebroadcastToleranceNumber && consensusParams.EnableTimedGenEmptyBlock == extraArgs.EnableTimedGenEmptyBlock {
		return nil, ErrRepeatedChainParam
	}

	return extraArgs, nil
}

func (c *ChainParamManager) checkNotFinishedProposal() error {
	_, notFinishedProposals, err := c.gov.notFinishedProposals.Get()
	if err != nil {
		return err
	}

	if len(notFinishedProposals) > 0 {
		return ErrExistNotFinishedChainParamProposal
	}

	return nil
}
