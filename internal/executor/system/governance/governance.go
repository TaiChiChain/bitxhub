package governance

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	vm "github.com/axiomesh/eth-kit/evm"
)

var (
	ErrMethodName         = errors.New("no this method")
	ErrVoteResult         = errors.New("vote result is invalid")
	ErrProposalType       = errors.New("proposal type is invalid")
	ErrUser               = errors.New("user is invalid")
	ErrUseHasVoted        = errors.New("user has already voted")
	ErrTitle              = errors.New("title is invalid")
	ErrTooLongTitle       = errors.New("title is too long, max is 200 characters")
	ErrDesc               = errors.New("description is invalid")
	ErrTooLongDesc        = errors.New("description is too long, max is 10000 characters")
	ErrBlockNumber        = errors.New("block number is invalid")
	ErrProposalID         = errors.New("proposal id is invalid")
	ErrProposalFinished   = errors.New("proposal has already finished")
	ErrBlockNumberOutDate = errors.New("block number is out of date")
)

const jsondata = `
[
	{"type": "function", "name": "propose", "inputs": [{"name": "proposalType", "type": "uint8"}, {"name": "title", "type": "string"}, {"name": "desc", "type": "string"}, {"name": "blockNumber", "type": "uint64"}, {"name": "extra", "type": "bytes"}], "outputs": [{"name": "proposalId", "type": "uint64"}]},
	{"type": "function", "name": "vote", "inputs": [{"name": "proposalId", "type": "uint64"}, {"name": "voteResult", "type": "uint8"}, {"name": "extra", "type": "bytes"}]},
	{"type": "function", "name": "proposal", "stateMutability": "view", "inputs": [{"name": "proposalId", "type": "uint64"}], "outputs": [{"name": "proposal", "type": "bytes"}]}
]
`

const (
	ProposeMethod  = "propose"
	VoteMethod     = "vote"
	ProposalMethod = "proposal"
	MaxTitleLength = 200
	MaxDescLength  = 10000
)

var method2Sig = map[string]string{
	ProposeMethod:  "propose(uint8,string,string,uint64,bytes)",
	VoteMethod:     "vote(uint64,uint8,bytes)",
	ProposalMethod: "proposal(uint64)",
}

type ProposalType uint8

const (
	// CouncilElect is a proposal for elect the council
	CouncilElect ProposalType = iota

	// NodeUpgrade is a proposal for update or upgrade the node
	NodeUpgrade

	// NodeAdd is a proposal for adding a new node
	NodeAdd

	// NodeRemove is a proposal for removing a node
	NodeRemove

	WhiteListProviderAdd

	WhiteListProviderRemove
)

type VoteResult uint8

const (
	Pass VoteResult = iota
	Reject
)

type BaseProposalArgs struct {
	ProposalType uint8
	Title        string
	Desc         string
	BlockNumber  uint64
}

type ProposalArgs struct {
	BaseProposalArgs
	Extra []byte
}

type BaseVoteArgs struct {
	ProposalId uint64
	VoteResult uint8
}

type VoteArgs struct {
	BaseVoteArgs
	Extra []byte
}

type GetProposalArgs struct {
	ProposalID uint64
}

type Governance struct {
	proposalTypes []ProposalType
	logger        logrus.FieldLogger

	gabi       *abi.ABI
	method2Sig map[string][]byte
}

func NewGov(proposalTypes []ProposalType, logger logrus.FieldLogger) (*Governance, error) {
	gabi, err := GetABI()
	if err != nil {
		return nil, err
	}

	return &Governance{
		proposalTypes: proposalTypes,
		logger:        logger,
		gabi:          gabi,
		method2Sig:    initMethodSignature(),
	}, nil
}

// GetABI get system contract abi
func GetABI() (*abi.ABI, error) {
	gabi, err := abi.JSON(strings.NewReader(jsondata))
	if err != nil {
		return nil, err
	}
	return &gabi, nil
}

func initMethodSignature() map[string][]byte {
	m2sig := make(map[string][]byte)
	for methodName, methodSig := range method2Sig {
		m2sig[methodName] = crypto.Keccak256([]byte(methodSig))
	}
	return m2sig
}

// GetMethodName quickly returns the name of a method.
// This is a quick way to get the name of a method.
// The method name is the first 4 bytes of the keccak256 hash of the method signature.
// If the method name is not found, the empty string is returned.
func (g *Governance) GetMethodName(data []byte) (string, error) {
	for methodName, methodSig := range g.method2Sig {
		id := methodSig[:4]
		g.logger.Debugf("method id: %v, get method id: %v", id, data[:4])
		if bytes.Equal(id, data[:4]) {
			return methodName, nil
		}
	}

	return "", ErrMethodName
}

// ParseArgs parse the arguments to specified interface by method name
func (g *Governance) ParseArgs(msg *vm.Message, methodName string, ret any) error {
	if len(msg.Data) < 4 {
		return fmt.Errorf("msg data length is not improperly formatted: %q - Bytes: %+v", msg.Data, msg.Data)
	}

	// discard method id
	data := msg.Data[4:]

	var args abi.Arguments
	if method, ok := g.gabi.Methods[methodName]; ok {
		if len(data)%32 != 0 {
			return fmt.Errorf("gabi: improperly formatted output: %q - Bytes: %+v", data, data)
		}
		args = method.Inputs
	}

	if args == nil {
		return fmt.Errorf("gabi: could not locate named method: %s", methodName)
	}

	unpacked, err := args.Unpack(data)
	if err != nil {
		return err
	}
	return args.Copy(ret, unpacked)
}

// GetArgs get system contract arguments from a message
func (g *Governance) GetArgs(msg *vm.Message) (any, error) {
	data := msg.Data
	if data == nil {
		return nil, vm.ErrExecutionReverted
	}

	method, err := g.GetMethodName(data)
	if err != nil {
		return nil, err
	}

	switch method {
	case ProposeMethod:
		proposalArgs := &ProposalArgs{}
		if err := g.ParseArgs(msg, ProposeMethod, proposalArgs); err != nil {
			return nil, err
		}
		return proposalArgs, nil
	case VoteMethod:
		voteArgs := &VoteArgs{}
		if err := g.ParseArgs(msg, VoteMethod, voteArgs); err != nil {
			return nil, err
		}
		return voteArgs, nil
	case ProposalMethod:
		getProposalArgs := &GetProposalArgs{}
		if err := g.ParseArgs(msg, ProposalMethod, getProposalArgs); err != nil {
			return nil, err
		}
		return getProposalArgs, nil
	default:
		return nil, ErrMethodName
	}
}

// PackOutputArgs pack the output arguments by method name
func (g *Governance) PackOutputArgs(methodName string, outputArgs ...any) ([]byte, error) {
	var args abi.Arguments
	if method, ok := g.gabi.Methods[methodName]; ok {
		args = method.Outputs
	}

	if args == nil {
		return nil, fmt.Errorf("gabi: could not locate named method: %s", methodName)
	}

	return args.Pack(outputArgs...)
}

// UnpackOutputArgs unpack the output arguments by method name
func (g *Governance) UnpackOutputArgs(methodName string, packed []byte) ([]any, error) {
	var args abi.Arguments
	if method, ok := g.gabi.Methods[methodName]; ok {
		args = method.Outputs
	}

	if args == nil {
		return nil, fmt.Errorf("gabi: could not locate named method: %s", methodName)
	}

	return args.Unpack(packed)
}

func (g *Governance) checkBeforePropose(user *ethcommon.Address, proposalType ProposalType, title, desc string, blockNumber uint64, lastHeight uint64) (bool, error) {
	if user == nil {
		return false, ErrUser
	}

	isVaildProposalType := common.IsInSlice[ProposalType](proposalType, g.proposalTypes)
	if !isVaildProposalType {
		return false, ErrProposalType
	}

	if title == "" || len(title) > MaxTitleLength {
		if title == "" {
			return false, ErrTitle
		}
		return false, ErrTooLongTitle
	}

	if desc == "" || len(desc) > MaxDescLength {
		if desc == "" {
			return false, ErrDesc
		}
		return false, ErrTooLongDesc
	}

	if blockNumber == 0 {
		return false, ErrBlockNumber
	}

	// check out of date block number
	if blockNumber <= lastHeight {
		return false, ErrBlockNumberOutDate
	}

	return true, nil
}

func (g *Governance) Propose(user *ethcommon.Address, proposalType ProposalType, title, desc string, blockNumber uint64, lastHeight uint64, nodeAddRole bool) (*BaseProposal, error) {
	_, err := g.checkBeforePropose(user, proposalType, title, desc, blockNumber, lastHeight)
	if err != nil {
		return nil, err
	}

	proposal := &BaseProposal{
		Type:        proposalType,
		Strategy:    NowProposalStrategy,
		Proposer:    user.String(),
		Title:       title,
		Desc:        desc,
		BlockNumber: blockNumber,
		Status:      Voting,
	}

	// proposer vote pass by default
	if !nodeAddRole {
		proposal.PassVotes = []string{user.String()}
	}

	return proposal, nil
}

func (g *Governance) checkBeforeVote(user *ethcommon.Address, proposal *BaseProposal, voteResult VoteResult) (bool, error) {
	if user == nil {
		return false, ErrUser
	}

	if proposal.ID == 0 {
		return false, ErrProposalID
	}

	if voteResult != Pass && voteResult != Reject {
		return false, ErrVoteResult
	}

	// check if user has voted
	if common.IsInSlice[string](user.String(), proposal.PassVotes) || common.IsInSlice[string](user.String(), proposal.RejectVotes) {
		return false, ErrUseHasVoted
	}

	// check proposal status
	if proposal.Status == Approved || proposal.Status == Rejected {
		return false, ErrProposalFinished
	}

	return true, nil
}

// Vote a proposal, return vote status
func (g *Governance) Vote(user *ethcommon.Address, proposal *BaseProposal, voteResult VoteResult) (ProposalStatus, error) {
	if _, err := g.checkBeforeVote(user, proposal, voteResult); err != nil {
		return Voting, err
	}

	switch voteResult {
	case Pass:
		proposal.PassVotes = append(proposal.PassVotes, user.String())
	case Reject:
		proposal.RejectVotes = append(proposal.RejectVotes, user.String())
	}

	return CalcProposalStatus(proposal.Strategy, proposal.TotalVotes, uint64(len(proposal.PassVotes)), uint64(len(proposal.RejectVotes))), nil
}

// RecordLog record execution log for governance
func (g *Governance) RecordLog(currentLog *common.Log, method string, proposal *BaseProposal, data []byte) {
	// set method signature, proposal id, proposal type, proposer as log topic for index
	idhash := make([]byte, 8)
	binary.BigEndian.PutUint64(idhash, proposal.ID)
	typeHash := make([]byte, 2)
	binary.BigEndian.PutUint16(typeHash, uint16(proposal.Type))
	currentLog.Topics = append(currentLog.Topics, types.NewHash(g.method2Sig[method]),
		types.NewHash(idhash), types.NewHash(typeHash), types.NewHash([]byte(proposal.Proposer)))
	currentLog.Data = data
	currentLog.Removed = false
}

// SaveLog save log
func (g *Governance) SaveLog(stateLedger ledger.StateLedger, currentLog *common.Log) {
	if currentLog.Data != nil {
		stateLedger.AddLog(&types.EvmLog{
			Address: currentLog.Address,
			Topics:  currentLog.Topics,
			Data:    currentLog.Data,
			Removed: currentLog.Removed,
		})
	}
}
