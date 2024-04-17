package governance

import (
	"encoding/json"
	"errors"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/base"
)

var (
	ErrExistNotFinishedGasProposal     = errors.New("exist not gas finished proposal")
	ErrExistNotFinishedCouncilProposal = errors.New("exist not finished council proposal")
	ErrNotFoundGasProposal             = errors.New("gas proposal not found for the id")
	ErrUnKnownGasProposalArgs          = errors.New("unknown proposal args")
	ErrGasExtraArgs                    = errors.New("unmarshal gas extra arguments error")
	ErrGasArgsType                     = errors.New("gas arguments type error")
	ErrGasUpperOrLlower                = errors.New("gas upper or lower limit error")
	ErrRepeatedGasInfo                 = errors.New("repeated gas info")
)

type GasExtraArgs struct {
	MinGasPrice uint64 // mol
}

type GasManager struct {
	gov *Governance
}

func NewGasManager(gov *Governance) *GasManager {
	return &GasManager{
		gov: gov,
	}
}

func (gm *GasManager) ProposeCheck(proposalType ProposalType, extra []byte) error {
	_, err := gm.getGasProposalExtraArgs(extra)
	if err != nil {
		return err
	}

	// Check for outstanding council proposals and gas proposals
	if _, err := gm.checkNotFinishedProposal(); err != nil {
		return err
	}

	return nil
}

func (gm *GasManager) Execute(proposal *Proposal) error {
	// if proposal is approved, update the EpochInfo gas
	if proposal.Status == Approved {
		extraArgs, err := gm.getGasProposalExtraArgs(proposal.Extra)
		if err != nil {
			return err
		}

		epochInfo, err := base.GetNextEpochInfo(gm.gov.stateLedger)
		if err != nil {
			return err
		}
		financeParams := epochInfo.FinanceParams
		financeParams.MinGasPrice = extraArgs.MinGasPrice
		financeParams.StartGasPriceAvailable = true
		epochInfo.FinanceParams = financeParams

		if err := base.SetNextEpochInfo(gm.gov.stateLedger, epochInfo); err != nil {
			return err
		}
	}

	return nil
}

func (gm *GasManager) checkNotFinishedProposal() (bool, error) {
	notFinishedProposals, err := gm.gov.notFinishedProposalMgr.GetProposals()
	if err != nil {
		return false, err
	}

	for _, notFinishedProposal := range notFinishedProposals {
		if notFinishedProposal.Type == CouncilElect {
			return false, ErrExistNotFinishedCouncilProposal
		}
		if notFinishedProposal.Type == GasUpdate {
			return false, ErrExistNotFinishedGasProposal
		}
	}
	return true, nil
}

func (gm *GasManager) getGasProposalExtraArgs(extra []byte) (*GasExtraArgs, error) {
	extraArgs := &GasExtraArgs{}
	if err := json.Unmarshal(extra, extraArgs); err != nil {
		gm.gov.logger.Errorf("Unmarshal extra args error: %s", err)
		return nil, ErrGasExtraArgs
	}

	// Check whether the gas proposal and GetNextEpochInfo are consistent
	epochInfo, err := base.GetNextEpochInfo(gm.gov.stateLedger)
	if err != nil {
		return nil, err
	}
	financeParams := epochInfo.FinanceParams
	if financeParams.MinGasPrice == extraArgs.MinGasPrice {
		return nil, ErrRepeatedGasInfo
	}

	return extraArgs, nil
}
