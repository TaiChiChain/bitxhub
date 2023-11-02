package governance

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"strings"
	"sync"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
)

type ProposalStatus uint8

const (
	Voting ProposalStatus = iota
	Approved
	Rejected
)

const (
	ProposalIDKey           = "proposalIDKey"
	NotFinishedProposalsKey = "notFinishedProposalKey"

	Addr2NameSystemAddrKey = "addrKey"
	Addr2NameSystemNameKey = "nameKey"
)

var (
	ErrNilProposalAccount = errors.New("ProposalID must be reset then use")
)

type BaseProposal struct {
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
}

func (baseProposal *BaseProposal) GetID() uint64 {
	return baseProposal.ID
}

func (baseProposal *BaseProposal) GetStatus() ProposalStatus {
	return baseProposal.Status
}

func (baseProposal *BaseProposal) SetStatus(status ProposalStatus) {
	baseProposal.Status = status
}

func (baseProposal *BaseProposal) GetBlockNumber() uint64 {
	return baseProposal.BlockNumber
}

type ProposalID struct {
	ID    uint64
	mutex sync.RWMutex

	account ledger.IAccount
}

// NewProposalID new proposal id from ledger
func NewProposalID(stateLedger ledger.StateLedger) *ProposalID {
	proposalID := &ProposalID{}
	// id is not initialized
	account := stateLedger.GetOrCreateAccount(types.NewAddressByStr(common.ProposalIDContractAddr))
	isExist, data := account.GetState([]byte(ProposalIDKey))
	if !isExist {
		proposalID.ID = 1
	} else {
		proposalID.ID = binary.BigEndian.Uint64(data)
	}

	proposalID.account = account

	return proposalID
}

func (pid *ProposalID) GetID() uint64 {
	pid.mutex.RLock()
	defer pid.mutex.RUnlock()

	return pid.ID
}

func (pid *ProposalID) GetAndAddID() (uint64, error) {
	pid.mutex.Lock()
	defer pid.mutex.Unlock()

	oldID := pid.ID
	pid.ID++

	if pid.account == nil {
		return 0, ErrNilProposalAccount
	}

	// persist id
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, pid.ID)
	pid.account.SetState([]byte(ProposalIDKey), data)

	return oldID, nil
}

type Addr2NameSystem struct {
	account ledger.IAccount
}

func NewAddr2NameSystem(stateLedger ledger.StateLedger) *Addr2NameSystem {
	addr2NameSystem := &Addr2NameSystem{}

	addr2NameSystem.account = stateLedger.GetOrCreateAccount(types.NewAddressByStr(common.Addr2NameContractAddr))

	return addr2NameSystem
}

// SetName set address to new name
func (ans *Addr2NameSystem) SetName(addr, name string) {
	ak := addrKey(addr)
	nk := nameKey(name)

	ans.account.SetState(ak, []byte(name))
	ans.account.SetState(nk, []byte(addr))
}

func (ans *Addr2NameSystem) GetName(addr string) (bool, string) {
	isExist, name := ans.account.GetState(addrKey(addr))
	return isExist, string(name)
}

func (ans *Addr2NameSystem) GetAddr(name string) (bool, string) {
	isExist, addr := ans.account.GetState(nameKey(name))
	return isExist, string(addr)
}

func addrKey(addr string) []byte {
	return []byte(strings.Join([]string{Addr2NameSystemAddrKey, addr}, "-"))
}

func nameKey(name string) []byte {
	return []byte(strings.Join([]string{Addr2NameSystemNameKey, name}, "-"))
}

type NotFinishedProposal struct {
	ID                  uint64
	DeadlineBlockNumber uint64
	ContractAddr        string
}

type NotFinishedProposalMgr struct {
	account ledger.IAccount
}

func NewNotFinishedProposalMgr(stateLedger ledger.StateLedger) *NotFinishedProposalMgr {
	notFinishedProposalMgr := &NotFinishedProposalMgr{}

	notFinishedProposalMgr.account = stateLedger.GetOrCreateAccount(types.NewAddressByStr(common.NotFinishedProposalContractAddr))

	return notFinishedProposalMgr
}

func (nfpm *NotFinishedProposalMgr) SetProposal(proposal *NotFinishedProposal) error {
	proposals, err := nfpm.GetProposals()
	if err != nil {
		return err
	}

	proposals[proposal.ID] = *proposal
	data, err := json.Marshal(proposals)
	if err != nil {
		return err
	}

	nfpm.account.SetState(notFinishedProposalsKey(), data)
	return nil
}

func (nfpm *NotFinishedProposalMgr) RemoveProposal(id uint64) error {
	proposals, err := nfpm.GetProposals()
	if err != nil {
		return err
	}

	delete(proposals, id)
	data, err := json.Marshal(proposals)
	if err != nil {
		return err
	}

	nfpm.account.SetState(notFinishedProposalsKey(), data)
	return nil
}

func (nfpm *NotFinishedProposalMgr) GetProposals() (map[uint64]NotFinishedProposal, error) {
	isExist, data := nfpm.account.GetState(notFinishedProposalsKey())
	proposals := make(map[uint64]NotFinishedProposal, 0)
	if isExist {
		if err := json.Unmarshal(data, &proposals); err != nil {
			return nil, err
		}
	}
	return proposals, nil
}

func notFinishedProposalsKey() []byte {
	return []byte(NotFinishedProposalsKey)
}

func loadProposal(stateLedger ledger.StateLedger, nfp *NotFinishedProposal) (ProposalObject, error) {
	switch nfp.ContractAddr {
	case common.CouncilManagerContractAddr:
		return loadCouncilProposal(stateLedger, nfp.ID)
	case common.NodeManagerContractAddr:
		return loadNodeProposal(stateLedger, nfp.ID)
	case common.WhiteListProviderManagerContractAddr:
		return loadProviderProposal(stateLedger, nfp.ID)
	default:
		return nil, errors.New("no this contract")
	}
}

func saveProposal(stateLedger ledger.StateLedger, nfp *NotFinishedProposal, proposal ProposalObject) error {
	switch nfp.ContractAddr {
	case common.CouncilManagerContractAddr:
		_, err := saveCouncilProposal(stateLedger, proposal)
		return err
	case common.NodeManagerContractAddr:
		_, err := saveNodeProposal(stateLedger, proposal)
		return err
	case common.WhiteListProviderManagerContractAddr:
		_, err := saveProviderProposal(stateLedger, proposal)
		return err
	default:
		return errors.New("no this contract")
	}
}

type ProposalObject interface {
	GetID() uint64
	GetStatus() ProposalStatus
	SetStatus(status ProposalStatus)
	GetBlockNumber() uint64
}

func CheckAndUpdateState(lastHeight uint64, stateLedger ledger.StateLedger) error {
	notFinishedProposalMgr := NewNotFinishedProposalMgr(stateLedger)
	notFinishedProposals, err := notFinishedProposalMgr.GetProposals()
	if err != nil {
		return err
	}

	for _, notFinishedProposal := range notFinishedProposals {
		if notFinishedProposal.DeadlineBlockNumber <= lastHeight {
			// update original proposal status
			proposal, err := loadProposal(stateLedger, &notFinishedProposal)
			if err != nil {
				return err
			}

			if proposal.GetStatus() == Approved || proposal.GetStatus() == Rejected {
				// proposal is finnished, no need update
				continue
			}

			// means proposal is out of deadline,status change to rejected
			proposal.SetStatus(Rejected)

			err = saveProposal(stateLedger, &notFinishedProposal, proposal)
			if err != nil {
				return err
			}

			// remove proposal from not finished proposals
			if err = notFinishedProposalMgr.RemoveProposal(notFinishedProposal.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func GetNotFinishedProposals(stateLedger ledger.StateLedger) (map[uint64]NotFinishedProposal, error) {
	notFinishedProposalMgr := NewNotFinishedProposalMgr(stateLedger)
	notFinishedProposals, err := notFinishedProposalMgr.GetProposals()
	if err != nil {
		return nil, err
	}

	return notFinishedProposals, nil
}
