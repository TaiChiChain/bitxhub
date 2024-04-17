package governance

import (
	"github.com/samber/lo"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
)

type ProposalStatus uint8

const (
	Voting ProposalStatus = iota
	Approved
	Rejected
)

const (
	nextProposalIDStorageKey       = "nextProposalID"
	proposalsStorageKey            = "proposals"
	notFinishedProposalsStorageKey = "notFinishedProposals"
)

type Proposal struct {
	ID          uint64
	Type        ProposalType
	Strategy    ProposalStrategy
	Proposer    string
	Title       string
	Desc        string
	BlockNumber uint64

	// totalVotes is total votes for this proposal
	// attention: some users may not vote for this proposal
	TotalVotes uint64

	// passVotes record user address for passed vote
	PassVotes []string

	RejectVotes []string
	Status      ProposalStatus

	// Extra information for some special proposal
	Extra []byte

	// CreatedBlockNumber is block number when the proposal has be created
	CreatedBlockNumber uint64

	// EffectiveBlockNumber is block number when the proposal has be take effect
	EffectiveBlockNumber uint64

	ExecuteSuccess   bool
	ExecuteFailedMsg string
}

type NotFinishedProposal struct {
	ID                  uint64
	DeadlineBlockNumber uint64
	Type                ProposalType
}

func AddNotFinishedProposal(notFinishedProposals *common.VMSlot[[]NotFinishedProposal], proposal *NotFinishedProposal) error {
	exist, proposals, err := notFinishedProposals.Get()
	if err != nil {
		return err
	}
	if !exist {
		proposals = make([]NotFinishedProposal, 0)
	}

	proposals = lo.UniqBy[NotFinishedProposal](append(proposals, *proposal), func(item NotFinishedProposal) uint64 {
		return item.ID
	})

	if err := notFinishedProposals.Put(proposals); err != nil {
		return err
	}
	return nil
}

func RemoveNotFinishedProposal(notFinishedProposals *common.VMSlot[[]NotFinishedProposal], id uint64) error {
	exist, proposals, err := notFinishedProposals.Get()
	if err != nil {
		return err
	}
	if !exist {
		return ErrNotFoundProposal
	}

	newProposals := lo.Filter[NotFinishedProposal](proposals, func(item NotFinishedProposal, index int) bool {
		return item.ID != id
	})

	if len(newProposals) == len(proposals) {
		return ErrNotFoundProposal
	}

	if err := notFinishedProposals.Put(proposals); err != nil {
		return err
	}
	return nil
}
