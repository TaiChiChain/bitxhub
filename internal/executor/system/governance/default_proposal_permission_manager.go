package governance

import (
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"
)

var _ ProposalPermissionManager = (*DefaultProposalPermissionManager)(nil)

// DefaultProposalPermissionManager implements ProposalPermissionManager
// only governance council member has permission
type DefaultProposalPermissionManager struct {
	gov *Governance
}

func NewDefaultProposalPermissionManager(gov *Governance) DefaultProposalPermissionManager {
	return DefaultProposalPermissionManager{
		gov: gov,
	}
}

func (d *DefaultProposalPermissionManager) ProposePermissionCheck(_ ProposalType, user ethcommon.Address) (has bool, err error) {
	return d.gov.isCouncilMember(user)
}

func (d *DefaultProposalPermissionManager) TotalVotes(_ ProposalType) (uint64, error) {
	council, err := d.gov.council.MustGet()
	if err != nil {
		return 0, err
	}

	return lo.Sum(lo.Map(council.Members, func(item CouncilMember, index int) uint64 {
		return item.Weight
	})), nil
}

func (d *DefaultProposalPermissionManager) UpdateVoteStatus(proposal *Proposal) error {
	proposal.Status = CalcProposalStatus(proposal.Strategy, proposal.TotalVotes, uint64(len(proposal.PassVotes)), uint64(len(proposal.RejectVotes)))
	return nil
}

func (d *DefaultProposalPermissionManager) VotePermissionCheck(proposalType ProposalType, user ethcommon.Address) (has bool, err error) {
	return d.gov.isCouncilMember(user)
}
