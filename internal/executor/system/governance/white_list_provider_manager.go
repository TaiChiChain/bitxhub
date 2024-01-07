package governance

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/samber/lo"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/access"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
)

var (
	ErrExistVotingProposal = errors.New("check finished council proposal fail: exist voting proposal of council elect and whitelist provider")
)

type WhiteListProviderManager struct {
	gov *Governance
}

// NewWhiteListProviderManager constructs a new NewWhiteListProviderManager
func NewWhiteListProviderManager(gov *Governance) *WhiteListProviderManager {
	return &WhiteListProviderManager{
		gov: gov,
	}
}

func (wlpm *WhiteListProviderManager) ProposeCheck(proposalType ProposalType, extra []byte) error {
	// Check if there are any finished council proposals and provider proposals
	if err := wlpm.checkFinishedProposal(); err != nil {
		return err
	}

	extraArgs, err := wlpm.getExtraArgs(extra)
	if err != nil {
		return err
	}

	if len(extraArgs.Providers) < 1 {
		return errors.New("empty providers")
	}

	// Check if the proposal providers have repeated addresses
	if len(lo.Uniq[string](lo.Map[access.WhiteListProvider, string](extraArgs.Providers, func(item access.WhiteListProvider, index int) string {
		return item.WhiteListProviderAddr
	}))) != len(extraArgs.Providers) {
		return errors.New("provider address repeated")
	}

	// Check if the providers already exist
	existProviders, err := access.GetProviders(wlpm.gov.stateLedger)
	if err != nil {
		return err
	}

	switch proposalType {
	case WhiteListProviderAdd:
		// Iterate through the args.Providers array and check if each provider already exists in existProviders
		for _, provider := range extraArgs.Providers {
			if common.IsInSlice[string](provider.WhiteListProviderAddr, lo.Map[access.WhiteListProvider, string](existProviders, func(item access.WhiteListProvider, index int) string {
				return item.WhiteListProviderAddr
			})) {
				return fmt.Errorf("provider already exists, %s", provider.WhiteListProviderAddr)
			}
		}
	case WhiteListProviderRemove:
		// Iterate through the args.Providers array and check all providers are in existProviders
		for _, provider := range extraArgs.Providers {
			if !common.IsInSlice[string](provider.WhiteListProviderAddr, lo.Map[access.WhiteListProvider, string](existProviders, func(item access.WhiteListProvider, index int) string {
				return item.WhiteListProviderAddr
			})) {
				return fmt.Errorf("provider does not exist, %s", provider.WhiteListProviderAddr)
			}
		}
	}
	return nil
}

// Execute add or remove providers in the WhiteListProviderManager.
func (wlpm *WhiteListProviderManager) Execute(proposal *Proposal) error {
	// if proposal is approved, update the node members
	if proposal.Status == Approved {
		modifyType := access.AddWhiteListProvider
		if proposal.Type == WhiteListProviderRemove {
			modifyType = access.RemoveWhiteListProvider
		}

		extraArgs, err := wlpm.getExtraArgs(proposal.Extra)
		if err != nil {
			return err
		}

		return access.AddAndRemoveProviders(wlpm.gov.stateLedger, modifyType, extraArgs.Providers)
	}
	return nil
}

func (wlpm *WhiteListProviderManager) getExtraArgs(extra []byte) (*access.WhiteListProviderArgs, error) {
	providerArgs := &access.WhiteListProviderArgs{}
	if err := json.Unmarshal(extra, providerArgs); err != nil {
		return nil, errors.New("unmarshal provider extra arguments error")
	}
	return providerArgs, nil
}

func (wlpm *WhiteListProviderManager) checkFinishedProposal() error {
	notFinishedProposals, err := wlpm.gov.notFinishedProposalMgr.GetProposals()
	if err != nil {
		return err
	}

	for _, notFinishedProposal := range notFinishedProposals {
		if notFinishedProposal.Type == CouncilElect || notFinishedProposal.Type == WhiteListProviderAdd || notFinishedProposal.Type == WhiteListProviderRemove {
			return ErrExistVotingProposal
		}
	}

	return nil
}
