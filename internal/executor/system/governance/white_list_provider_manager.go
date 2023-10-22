package governance

import (
	"encoding/json"
	"fmt"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/access"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	vm "github.com/axiomesh/eth-kit/evm"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"
)

const (
	WhiteListProviderProposalKey = "WhiteListProviderProposal"
)

var _ common.SystemContract = (*WhiteListProviderManager)(nil)

type WhiteListProviderProposalArgs struct {
	BaseProposalArgs
	access.WhiteListProviderArgs
}

type WhiteListProviderProposal struct {
	BaseProposal
	access.WhiteListProviderArgs
}

type WhiteListProviderVoteArgs struct {
	BaseVoteArgs
}

type WhiteListProviderManager struct {
	gov *Governance

	account        ledger.IAccount
	councilAccount ledger.IAccount
	stateLedger    ledger.StateLedger
	currentLog     *common.Log
	proposalID     *ProposalID
}

// Reset resets the state of the WhiteListProviderManager.
//
// stateLedger is the state ledger used to get or create an account.
func (ac *WhiteListProviderManager) Reset(stateLedger ledger.StateLedger) {
	addr := types.NewAddressByStr(common.WhiteListProviderManagerContractAddr)
	ac.account = stateLedger.GetOrCreateAccount(addr)
	ac.stateLedger = stateLedger
	ac.currentLog = &common.Log{
		Address: addr,
	}
	ac.proposalID = NewProposalID(stateLedger)

	councilAddr := types.NewAddressByStr(common.CouncilManagerContractAddr)
	ac.councilAccount = stateLedger.GetOrCreateAccount(councilAddr)
}

// NewWhiteListProviderManager constructs a new NewWhiteListProviderManager
func NewWhiteListProviderManager(cfg *common.SystemContractConfig) *WhiteListProviderManager {
	gov, err := NewGov([]ProposalType{WhiteListProviderAdd, WhiteListProviderRemove}, cfg.Logger)
	if err != nil {
		panic(err)
	}

	return &WhiteListProviderManager{
		gov: gov,
	}
}

func (ac *WhiteListProviderManager) Run(msg *vm.Message) (*vm.ExecutionResult, error) {
	defer ac.gov.SaveLog(ac.stateLedger, ac.currentLog)
	// parse method and arguments from msg payload
	args, err := ac.gov.GetArgs(msg)
	if err != nil {
		return nil, err
	}

	gasUse := common.CalculateDynamicGas(msg.Data)
	switch v := args.(type) {
	case *ProposalArgs:
		proposalArgs, err := ac.getProposalArgs(v)
		if err != nil {
			return nil, err
		}
		proposeRes, err := ac.propose(&msg.From, proposalArgs)
		// gas will not be used if err
		if proposeRes != nil {
			proposeRes.UsedGas = gasUse
		}
		return proposeRes, err
	case *VoteArgs:
		voteRes, err := ac.vote(&msg.From, &WhiteListProviderVoteArgs{BaseVoteArgs: v.BaseVoteArgs})
		// gas will not be used if err
		if voteRes != nil {
			voteRes.UsedGas = gasUse
		}
		return voteRes, err
	default:
		return nil, fmt.Errorf("unknown proposal args")
	}
}

func (ac *WhiteListProviderManager) EstimateGas(callArgs *types.CallArgs) (uint64, error) {
	_, err := ac.gov.GetArgs(&vm.Message{Data: *callArgs.Data})
	if err != nil {
		return 0, err
	}
	gas := common.CalculateDynamicGas(*callArgs.Data)
	return gas, nil
}

func (ac *WhiteListProviderManager) CheckAndUpdateState(u uint64, stateLedger ledger.StateLedger) {}

// vote processes a vote on a proposal.
//
// It takes a user address and vote arguments as parameters.
// It returns an execution result and an error.
func (ac *WhiteListProviderManager) vote(user *ethcommon.Address, voteArgs *WhiteListProviderVoteArgs) (*vm.ExecutionResult, error) {
	result := &vm.ExecutionResult{}

	// get proposal
	proposal, err := ac.loadProviderProposal(voteArgs.ProposalId)
	if err != nil {
		return nil, err
	}

	result.ReturnData, result.Err = ac.voteServicesAddRemove(user, proposal, voteArgs)
	if result.Err != nil {
		return nil, result.Err
	}
	return result, nil
}

func (c *WhiteListProviderManager) loadProviderProposal(proposalID uint64) (*WhiteListProviderProposal, error) {
	isExist, data := c.account.GetState([]byte(fmt.Sprintf("%s%d", WhiteListProviderProposalKey, proposalID)))
	if !isExist {
		return nil, fmt.Errorf("provider proposal not found for the id")
	}

	proposal := &WhiteListProviderProposal{}
	if err := json.Unmarshal(data, proposal); err != nil {
		return nil, err
	}

	return proposal, nil
}

// voteServicesAddRemove is a function that allows a user to vote on adding or removing services in the WhiteListProviderManager.
//
// Parameters:
// - user: The address of the user who wants to vote.
// - proposal: The WhiteListProviderProposal that contains the details of the proposal being voted on.
// - voteArgs: The WhiteListProviderVoteArgs that contains the details of the user's vote.
//
// Returns:
// - []byte: The result of the vote.
// - error: An error if there was a problem with the vote.
func (ac *WhiteListProviderManager) voteServicesAddRemove(user *ethcommon.Address, proposal *WhiteListProviderProposal, voteArgs *WhiteListProviderVoteArgs) ([]byte, error) {
	// check user can vote
	isExist, _ := CheckInCouncil(ac.councilAccount, user.String())
	if !isExist {
		return nil, ErrNotFoundCouncilMember
	}

	res := VoteResult(voteArgs.VoteResult)
	proposalStatus, err := ac.gov.Vote(user, &proposal.BaseProposal, res)
	if err != nil {
		return nil, err
	}
	proposal.Status = proposalStatus

	b, err := ac.saveProviderProposal(proposal)
	if err != nil {
		return nil, err
	}

	// if proposal is approved, update the node members
	if proposal.Status == Approved {
		modifyType := access.AddWhiteListProvider
		if proposal.Type == WhiteListProviderRemove {
			modifyType = access.RemoveWhiteListProvider
		}
		err = access.AddAndRemoveProviders(ac.stateLedger, modifyType, proposal.Providers)
		if err != nil {
			return nil, err
		}
	}
	// record log
	ac.gov.RecordLog(ac.currentLog, VoteMethod, &proposal.BaseProposal, b)
	return b, nil
}

// propose is a method of the WhiteListProviderManager struct that handles the proposal of adding or removing providers.
// It takes in the address of the proposal and the proposal arguments and returns the execution result or an error.
func (ac *WhiteListProviderManager) propose(addr *ethcommon.Address, args *WhiteListProviderProposalArgs) (*vm.ExecutionResult, error) {
	result := &vm.ExecutionResult{}

	// Check if there are any finished council proposals and provider proposals
	if _, err := ac.checkFinishedProposal(); err != nil {
		return nil, err
	}

	if len(args.Providers) < 1 {
		return nil, fmt.Errorf("empty services")
	}

	// Check if the proposal services have repeated addresses
	if len(lo.Uniq[string](lo.Map[access.WhiteListProvider, string](args.Providers, func(item access.WhiteListProvider, index int) string {
		return item.WhiteListProviderAddr
	}))) != len(args.Providers) {
		return nil, fmt.Errorf("provider address repeated")
	}

	// Check if the services already exist
	existServices, err := access.GetProviders(ac.stateLedger)
	if err != nil {
		return nil, err
	}

	switch ProposalType(args.BaseProposalArgs.ProposalType) {
	case WhiteListProviderAdd:
		// Iterate through the args.Providers array and check if each service already exists in existServices
		for _, service := range args.Providers {
			if common.IsInSlice[string](service.WhiteListProviderAddr, lo.Map[access.WhiteListProvider, string](existServices, func(item access.WhiteListProvider, index int) string {
				return item.WhiteListProviderAddr
			})) {
				return nil, fmt.Errorf("provider already exists, %s", service.WhiteListProviderAddr)
			}
		}
	case WhiteListProviderRemove:
		// Iterate through the args.Providers array and check all services are in existServices
		for _, service := range args.Providers {
			if !common.IsInSlice[string](service.WhiteListProviderAddr, lo.Map[access.WhiteListProvider, string](existServices, func(item access.WhiteListProvider, index int) string {
				return item.WhiteListProviderAddr
			})) {
				return nil, fmt.Errorf("provider does not exist, %s", service.WhiteListProviderAddr)
			}
		}
	}

	// Propose adding or removing services and return the result
	result.ReturnData, result.Err = ac.proposeServicesAddRemove(addr, args)
	if result.Err != nil {
		return nil, result.Err
	}

	return result, nil
}

func (ac *WhiteListProviderManager) getProposalArgs(args *ProposalArgs) (*WhiteListProviderProposalArgs, error) {
	a := &WhiteListProviderProposalArgs{
		BaseProposalArgs: args.BaseProposalArgs,
	}
	serviceArgs := &access.WhiteListProviderArgs{}
	if err := json.Unmarshal(args.Extra, serviceArgs); err != nil {
		return nil, fmt.Errorf("unmarshal services extra arguments error")
	}
	a.WhiteListProviderArgs = *serviceArgs
	return a, nil
}

func (ac *WhiteListProviderManager) proposeServicesAddRemove(addr *ethcommon.Address, args *WhiteListProviderProposalArgs) ([]byte, error) {
	baseProposal, err := ac.gov.Propose(addr, ProposalType(args.ProposalType), args.Title, args.Desc, args.BlockNumber)
	if err != nil {
		return nil, err
	}
	isExist, council := CheckInCouncil(ac.councilAccount, addr.String())
	if !isExist {
		return nil, ErrNotFoundCouncilMember
	}
	proposal := &WhiteListProviderProposal{
		BaseProposal: *baseProposal,
	}
	id, err := ac.proposalID.GetAndAddID()
	if err != nil {
		return nil, err
	}
	proposal.ID = id
	proposal.Providers = args.Providers
	proposal.TotalVotes = lo.Sum[uint64](lo.Map[*CouncilMember, uint64](council.Members, func(item *CouncilMember, index int) uint64 {
		return item.Weight
	}))
	b, err := ac.saveProviderProposal(proposal)
	if err != nil {
		return nil, err
	}
	// record log
	ac.gov.RecordLog(ac.currentLog, ProposeMethod, &proposal.BaseProposal, b)
	return b, nil
}

func (ac *WhiteListProviderManager) saveProviderProposal(proposal *WhiteListProviderProposal) ([]byte, error) {
	b, err := json.Marshal(proposal)
	if err != nil {
		return nil, err
	}
	// save proposal
	ac.account.SetState([]byte(fmt.Sprintf("%s%d", WhiteListProviderProposalKey, proposal.ID)), b)
	return b, nil
}

func (ac *WhiteListProviderManager) checkFinishedProposal() (bool, error) {
	if isExist, data := ac.councilAccount.Query(CouncilProposalKey); isExist {
		for _, proposalData := range data {
			proposal := &CouncilProposal{}
			if err := json.Unmarshal(proposalData, proposal); err != nil {
				return false, fmt.Errorf("check finished council proposal fail: json.Unmarshal fail")
			}

			if proposal.Status == Voting {
				return false, fmt.Errorf("check finished council proposal fail: exist voting proposal")
			}
		}
	}

	if isExist, data := ac.account.Query(WhiteListProviderProposalKey); isExist {
		for _, proposalData := range data {
			proposal := &WhiteListProviderProposal{}
			if err := json.Unmarshal(proposalData, proposal); err != nil {
				return false, fmt.Errorf("check finished provider proposal fail: json.Unmarshal fail")
			}

			if proposal.Status == Voting {
				return false, fmt.Errorf("check finished provider proposal fail: exist voting proposal")
			}
		}
	}

	return true, nil
}
