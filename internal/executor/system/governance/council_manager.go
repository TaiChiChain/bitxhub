package governance

import (
	"encoding/json"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

var (
	ErrMinCouncilMembersCount   = errors.New("council members count can't less than 4")
	ErrRepeatedAddress          = errors.New("council member address repeated")
	ErrRepeatedName             = errors.New("council member name repeated")
	ErrExistNotFinishedProposal = errors.New("exist not finished proposal, must finished all proposal then propose council proposal")
)

const (
	// councilStorageKey is key for council storage
	councilStorageKey = "council"

	// MinCouncilMembersCount is min council members count
	MinCouncilMembersCount = 4
)

var _ ProposalHandler = (*CouncilManager)(nil)

// Council is storage of council
type Council struct {
	Members []CouncilMember
}

type CouncilMember struct {
	Address string
	Weight  uint64
	Name    string
}

type Candidates struct {
	Candidates []CouncilMember
}

type CouncilManager struct {
	gov *Governance
	DefaultProposalPermissionManager
}

func NewCouncilManager(gov *Governance) *CouncilManager {
	return &CouncilManager{
		gov:                              gov,
		DefaultProposalPermissionManager: NewDefaultProposalPermissionManager(gov),
	}
}

func (cm *CouncilManager) GenesisInit(genesis *repo.GenesisConfig) error {
	council := &Council{}
	for _, admin := range genesis.CouncilMembers {
		if !ethcommon.IsHexAddress(admin.Address) {
			return errors.Errorf("invalid council member address: %s", admin.Address)
		}

		council.Members = append(council.Members, CouncilMember{
			Address: admin.Address,
			Weight:  admin.Weight,
			Name:    admin.Name,
		})
	}

	if err := cm.gov.council.Put(*council); err != nil {
		return err
	}

	return nil
}

func (cm *CouncilManager) SetContext(ctx *common.VMContext) {}

func (cm *CouncilManager) ProposeArgsCheck(proposalType ProposalType, title, desc string, blockNumber uint64, extra []byte) error {
	var candidates Candidates
	if err := json.Unmarshal(extra, &candidates); err != nil {
		return errors.Wrapf(err, "unmarshal extra: %s to council member error", extra)
	}

	// check proposal council member num
	if len(candidates.Candidates) < MinCouncilMembersCount {
		return ErrMinCouncilMembersCount
	}

	// check proposal candidates has repeated address
	if len(lo.UniqBy(candidates.Candidates, func(item CouncilMember) string {
		return item.Address
	})) != len(candidates.Candidates) {
		return ErrRepeatedAddress
	}

	if len(lo.UniqBy(candidates.Candidates, func(item CouncilMember) string {
		return item.Name
	})) != len(candidates.Candidates) {
		return ErrRepeatedName
	}

	finished, err := cm.gov.checkFinishedAllProposal()
	if err != nil {
		return err
	}
	if !finished {
		return ErrExistNotFinishedProposal
	}

	return nil
}

func (cm *CouncilManager) VotePassExecute(proposal *Proposal) error {
	var candidates Candidates
	if err := json.Unmarshal(proposal.Extra, &candidates); err != nil {
		return errors.Wrapf(err, "unmarshal extra: %s to council member error", proposal.Extra)
	}

	council := &Council{
		Members: candidates.Candidates,
	}
	if err := cm.gov.council.Put(*council); err != nil {
		return err
	}
	return nil
}
