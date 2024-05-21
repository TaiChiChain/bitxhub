package governance

import (
	"encoding/json"
	"fmt"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/access"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/access/solidity/whitelist"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

var (
	ErrExistVotingProposal = errors.New("check finished council proposal fail: exist voting proposal of council elect and whitelist provider")
)

var _ ProposalHandler = (*WhitelistProviderManager)(nil)

type WhitelistProviderInfo struct {
	Addr string
}

type WhitelistProviderArgs struct {
	Providers []WhitelistProviderInfo
}

type WhitelistProviderManager struct {
	gov *Governance
	DefaultProposalPermissionManager
}

func NewWhiteListProviderManager(gov *Governance) *WhitelistProviderManager {
	return &WhitelistProviderManager{
		gov:                              gov,
		DefaultProposalPermissionManager: NewDefaultProposalPermissionManager(gov),
	}
}

func (m *WhitelistProviderManager) GenesisInit(genesis *repo.GenesisConfig) error {
	return nil
}

func (m *WhitelistProviderManager) SetContext(ctx *common.VMContext) {}

func (m *WhitelistProviderManager) ProposeArgsCheck(proposalType ProposalType, title, desc string, blockNumber uint64, extra []byte) error {
	// Check if there are any finished council proposals and provider proposals
	if err := m.checkFinishedProposal(); err != nil {
		return err
	}

	extraArgs, err := m.getExtraArgs(extra)
	if err != nil {
		return err
	}

	if len(extraArgs.Providers) == 0 {
		return errors.New("empty providers")
	}
	var providersAddrs []ethcommon.Address
	for _, provider := range extraArgs.Providers {
		if !ethcommon.IsHexAddress(provider.Addr) {
			return errors.Errorf("invalid provider address, %s", provider.Addr)
		}
		providersAddrs = append(providersAddrs, ethcommon.HexToAddress(provider.Addr))
	}

	// Check if the proposal providers have repeated addresses
	if len(lo.Uniq[string](lo.Map(extraArgs.Providers, func(item WhitelistProviderInfo, index int) string {
		return item.Addr
	}))) != len(extraArgs.Providers) {
		return errors.New("provider address repeated")
	}

	whitelistContract := access.WhitelistBuildConfig.Build(m.gov.CrossCallSystemContractContext())

	switch proposalType {
	case WhitelistProviderAdd:
		for _, provider := range providersAddrs {
			if whitelistContract.ExistProvider(provider) {
				return errors.Errorf("provider already exists, %v", provider)
			}
		}
	case WhitelistProviderRemove:
		for _, provider := range providersAddrs {
			if !whitelistContract.ExistProvider(provider) {
				return fmt.Errorf("provider does not exist, %v", provider)
			}
		}
	default:
		return errors.New("invalid proposal type")
	}
	return nil
}

func (m *WhitelistProviderManager) VotePassExecute(proposal *Proposal) error {
	extraArgs, err := m.getExtraArgs(proposal.Extra)
	if err != nil {
		return err
	}
	whitelistContract := access.WhitelistBuildConfig.Build(m.gov.CrossCallSystemContractContext())
	return whitelistContract.UpdateProviders(proposal.Type == WhitelistProviderAdd, lo.Map(extraArgs.Providers, func(item WhitelistProviderInfo, index int) whitelist.ProviderInfo {
		return whitelist.ProviderInfo{
			Addr: ethcommon.HexToAddress(item.Addr),
		}
	}))
}

func (m *WhitelistProviderManager) getExtraArgs(extra []byte) (*WhitelistProviderArgs, error) {
	providerArgs := &WhitelistProviderArgs{}
	if err := json.Unmarshal(extra, providerArgs); err != nil {
		return nil, errors.New("unmarshal provider extra arguments error")
	}
	return providerArgs, nil
}

func (m *WhitelistProviderManager) checkFinishedProposal() error {
	_, notFinishedProposals, err := m.gov.notFinishedProposals.Get()
	if err != nil {
		return err
	}

	for _, notFinishedProposal := range notFinishedProposals {
		if notFinishedProposal.Type == CouncilElect || notFinishedProposal.Type == WhitelistProviderAdd || notFinishedProposal.Type == WhitelistProviderRemove {
			return ErrExistVotingProposal
		}
	}

	return nil
}
