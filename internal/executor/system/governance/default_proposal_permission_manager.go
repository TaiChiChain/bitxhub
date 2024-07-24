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
	return d.totalVotes(nil, true)
}

func (d *DefaultProposalPermissionManager) totalVotes(members []string, ignoreMembers bool) (uint64, error) {
	council, err := d.gov.council.MustGet()
	if err != nil {
		return 0, err
	}

	return lo.Sum(lo.Map(council.Members, func(item CouncilMember, index int) uint64 {
		if ignoreMembers {
			// count all council members's votes
			return item.Weight
		}

		if lo.Contains(members, item.Address) {
			return item.Weight
		}
		return 0
	})), nil
}

func (d *DefaultProposalPermissionManager) UpdateVoteStatus(proposal *Proposal) error {
	totalPassVotesWeight, err := d.totalVotes(proposal.PassVotes, false)
	if err != nil {
		return err
	}
	totalRejectVotesWeight, err := d.totalVotes(proposal.RejectVotes, false)
	if err != nil {
		return err
	}
	proposal.Status = CalcProposalStatus(proposal.Strategy, proposal.TotalVotes, totalPassVotesWeight, totalRejectVotesWeight)
	return nil
}

func (d *DefaultProposalPermissionManager) VotePermissionCheck(proposalType ProposalType, user ethcommon.Address) (has bool, err error) {
	return d.gov.isCouncilMember(user)
}
