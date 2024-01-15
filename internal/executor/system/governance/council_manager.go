package governance

import (
	"encoding/json"
	"errors"

	"github.com/samber/lo"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

var (
	ErrCouncilNumber            = errors.New("council members total count can't bigger than candidates count")
	ErrMinCouncilMembersCount   = errors.New("council members count can't less than 4")
	ErrRepeatedAddress          = errors.New("council member address repeated")
	ErrRepeatedName             = errors.New("council member name repeated")
	ErrNotFoundCouncilMember    = errors.New("council member is not found")
	ErrNotFoundCouncil          = errors.New("council is not found")
	ErrCouncilExtraArgs         = errors.New("unmarshal council extra arguments error")
	ErrNotFoundCouncilProposal  = errors.New("council proposal not found for the id")
	ErrExistNotFinishedProposal = errors.New("exist not finished proposal, must finished all proposal then propose council proposal")
	ErrDeadlineBlockNumber      = errors.New("can't vote, proposal is out of deadline block number")
)

const (
	// CouncilKey is key for council storage
	CouncilKey = "councilKey"

	// MinCouncilMembersCount is min council members count
	MinCouncilMembersCount = 4
)

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
}

func NewCouncilManager(gov *Governance) *CouncilManager {
	return &CouncilManager{
		gov: gov,
	}
}

func (cm *CouncilManager) ProposeCheck(proposalType ProposalType, extra []byte) error {
	var candidates Candidates
	if err := json.Unmarshal(extra, &candidates); err != nil {
		cm.gov.logger.Errorf("Unmarshal extra: %s to council member error: %s", extra, err)
		return err
	}

	for i, candidate := range candidates.Candidates {
		cm.gov.logger.Debugf("candidate %d: %+v", i, candidate)
	}

	// check proposal council member num
	if len(candidates.Candidates) < MinCouncilMembersCount {
		return ErrMinCouncilMembersCount
	}

	// check proposal candidates has repeated address
	if len(lo.Uniq[string](lo.Map[CouncilMember, string](candidates.Candidates, func(item CouncilMember, index int) string {
		return item.Address
	}))) != len(candidates.Candidates) {
		return ErrRepeatedAddress
	}

	if !checkAddr2Name(candidates.Candidates) {
		return ErrRepeatedName
	}

	if !cm.checkFinishedAllProposal() {
		return ErrExistNotFinishedProposal
	}

	return nil
}

func (cm *CouncilManager) Execute(proposal *Proposal) error {
	// if proposal is approved, update the council members
	if proposal.Status == Approved {
		var candidates Candidates
		if err := json.Unmarshal(proposal.Extra, &candidates); err != nil {
			cm.gov.logger.Errorf("Unmarshal extra: %s to council member error: %s", proposal.Extra, err)
			return err
		}

		council := &Council{
			Members: candidates.Candidates,
		}

		// save council
		cb, err := json.Marshal(council)
		if err != nil {
			return err
		}
		cm.gov.account.SetState([]byte(CouncilKey), cb)

		// set name when proposal approved
		setName(cm.gov.addr2NameSystem, council.Members)

		for i, member := range council.Members {
			cm.gov.logger.Debugf("after vote, now council member %d, %+v", i, member)
		}
	}

	return nil
}

func InitCouncilMembers(lg ledger.StateLedger, admins []*repo.Admin) error {
	addr2NameSystem := NewAddr2NameSystem(lg)

	council := &Council{}
	for _, admin := range admins {
		council.Members = append(council.Members, CouncilMember{
			Address: admin.Address,
			Weight:  admin.Weight,
			Name:    admin.Name,
		})

		// set name
		addr2NameSystem.SetName(admin.Address, admin.Name)
	}

	account := lg.GetOrCreateAccount(types.NewAddressByStr(common.GovernanceContractAddr))
	b, err := json.Marshal(council)
	if err != nil {
		return err
	}
	account.SetState([]byte(CouncilKey), b)
	return nil
}

func (cm *CouncilManager) checkFinishedAllProposal() bool {
	notFinishedProposals, err := cm.gov.notFinishedProposalMgr.GetProposals()
	if err != nil {
		cm.gov.logger.Errorf("get not finished proposals error: %s", err)
		return false
	}

	return len(notFinishedProposals) == 0
}

func CheckInCouncil(account ledger.IAccount, addr string) (bool, *Council) {
	// check council if is exist
	council := &Council{}
	council, isExistCouncil := GetCouncil(account)
	if !isExistCouncil {
		return false, nil
	}

	// check addr if is exist in council
	isExist := common.IsInSlice[string](addr, lo.Map[CouncilMember, string](council.Members, func(item CouncilMember, index int) string {
		return item.Address
	}))
	if !isExist {
		return false, nil
	}

	return true, council
}

func GetCouncil(account ledger.IAccount) (*Council, bool) {
	// Check if the council exists in the account's state
	isExist, data := account.GetState([]byte(CouncilKey))
	if !isExist {
		return nil, false
	}

	// Unmarshal the data into a Council object
	council := &Council{}
	if err := json.Unmarshal(data, council); err != nil {
		return nil, false
	}

	return council, true
}

func checkAddr2Name(members []CouncilMember) bool {
	// repeated name return false
	return len(lo.Uniq[string](lo.Map[CouncilMember, string](members, func(item CouncilMember, index int) string {
		return item.Name
	}))) == len(members)
}

func setName(addr2NameSystem *Addr2NameSystem, members []CouncilMember) {
	for _, member := range members {
		addr2NameSystem.SetName(member.Address, member.Name)
	}
}
