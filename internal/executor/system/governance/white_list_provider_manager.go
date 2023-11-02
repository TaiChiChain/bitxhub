package governance

import (
	"encoding/json"
	"errors"
	"fmt"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/access"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	vm "github.com/axiomesh/eth-kit/evm"
)

var (
	ErrExistVotingProposal = errors.New("check finished council proposal fail: exist voting proposal of council elect and whitelist provider")
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

	account                ledger.IAccount
	councilAccount         ledger.IAccount
	stateLedger            ledger.StateLedger
	currentLog             *common.Log
	proposalID             *ProposalID
	lastHeight             uint64
	notFinishedProposalMgr *NotFinishedProposalMgr
}

// Reset resets the state of the WhiteListProviderManager.
//
// stateLedger is the state ledger used to get or create an account.
func (wlpm *WhiteListProviderManager) Reset(lastHeight uint64, stateLedger ledger.StateLedger) {
	addr := types.NewAddressByStr(common.WhiteListProviderManagerContractAddr)
	wlpm.account = stateLedger.GetOrCreateAccount(addr)
	wlpm.stateLedger = stateLedger
	wlpm.currentLog = &common.Log{
		Address: addr,
	}
	wlpm.proposalID = NewProposalID(stateLedger)

	councilAddr := types.NewAddressByStr(common.CouncilManagerContractAddr)
	wlpm.councilAccount = stateLedger.GetOrCreateAccount(councilAddr)
	wlpm.notFinishedProposalMgr = NewNotFinishedProposalMgr(stateLedger)

	// check and update
	wlpm.checkAndUpdateState(lastHeight)
	wlpm.lastHeight = lastHeight
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

func (wlpm *WhiteListProviderManager) Run(msg *vm.Message) (*vm.ExecutionResult, error) {
	defer wlpm.gov.SaveLog(wlpm.stateLedger, wlpm.currentLog)
	// parse method and arguments from msg payload
	args, err := wlpm.gov.GetArgs(msg)
	if err != nil {
		return nil, err
	}

	gasUse := common.CalculateDynamicGas(msg.Data)
	switch v := args.(type) {
	case *ProposalArgs:
		proposalArgs, err := wlpm.getProposalArgs(v)
		if err != nil {
			return nil, err
		}
		proposeRes, err := wlpm.propose(&msg.From, proposalArgs)
		// gas will not be used if err
		if proposeRes != nil {
			proposeRes.UsedGas = gasUse
		}
		return proposeRes, err
	case *VoteArgs:
		voteRes, err := wlpm.vote(&msg.From, &WhiteListProviderVoteArgs{BaseVoteArgs: v.BaseVoteArgs})
		// gas will not be used if err
		if voteRes != nil {
			voteRes.UsedGas = gasUse
		}
		return voteRes, err
	default:
		return nil, errors.New("unknown proposal args")
	}
}

func (wlpm *WhiteListProviderManager) EstimateGas(callArgs *types.CallArgs) (uint64, error) {
	_, err := wlpm.gov.GetArgs(&vm.Message{Data: *callArgs.Data})
	if err != nil {
		return 0, err
	}
	gas := common.CalculateDynamicGas(*callArgs.Data)
	return gas, nil
}

// vote processes a vote on a proposal.
//
// It takes a user address and vote arguments as parameters.
// It returns an execution result and an error.
func (wlpm *WhiteListProviderManager) vote(user *ethcommon.Address, voteArgs *WhiteListProviderVoteArgs) (*vm.ExecutionResult, error) {
	result := &vm.ExecutionResult{}

	// get proposal
	proposal, err := wlpm.loadProviderProposal(voteArgs.ProposalId)
	if err != nil {
		return nil, err
	}

	result.ReturnData, result.Err = wlpm.voteProviderAddRemove(user, proposal, voteArgs)
	if result.Err != nil {
		return nil, result.Err
	}
	return result, nil
}

func (wlpm *WhiteListProviderManager) loadProviderProposal(proposalID uint64) (*WhiteListProviderProposal, error) {
	return loadProviderProposal(wlpm.stateLedger, proposalID)
}

func (wlpm *WhiteListProviderManager) saveProviderProposal(proposal *WhiteListProviderProposal) ([]byte, error) {
	return saveProviderProposal(wlpm.stateLedger, proposal)
}

// voteProviderAddRemove is a function that allows a user to vote on adding or removing providers in the WhiteListProviderManager.
//
// Parameters:
// - user: The address of the user who wants to vote.
// - proposal: The WhiteListProviderProposal that contains the details of the proposal being voted on.
// - voteArgs: The WhiteListProviderVoteArgs that contains the details of the user's vote.
//
// Returns:
// - []byte: The result of the vote.
// - error: An error if there was a problem with the vote.
func (wlpm *WhiteListProviderManager) voteProviderAddRemove(user *ethcommon.Address, proposal *WhiteListProviderProposal, voteArgs *WhiteListProviderVoteArgs) ([]byte, error) {
	// check user can vote
	isExist, _ := CheckInCouncil(wlpm.councilAccount, user.String())
	if !isExist {
		return nil, ErrNotFoundCouncilMember
	}

	res := VoteResult(voteArgs.VoteResult)
	proposalStatus, err := wlpm.gov.Vote(user, &proposal.BaseProposal, res)
	if err != nil {
		return nil, err
	}
	proposal.Status = proposalStatus

	b, err := wlpm.saveProviderProposal(proposal)
	if err != nil {
		return nil, err
	}

	// update not finished proposal
	if proposal.Status == Approved || proposal.Status == Rejected {
		if err := wlpm.notFinishedProposalMgr.RemoveProposal(proposal.ID); err != nil {
			return nil, err
		}
	}

	// if proposal is approved, update the node members
	if proposal.Status == Approved {
		modifyType := access.AddWhiteListProvider
		if proposal.Type == WhiteListProviderRemove {
			modifyType = access.RemoveWhiteListProvider
		}
		err = access.AddAndRemoveProviders(wlpm.stateLedger, modifyType, proposal.Providers)
		if err != nil {
			return nil, err
		}
	}
	// record log
	wlpm.gov.RecordLog(wlpm.currentLog, VoteMethod, &proposal.BaseProposal, b)
	return b, nil
}

// propose is a method of the WhiteListProviderManager struct that handles the proposal of adding or removing providers.
// It takes in the address of the proposal and the proposal arguments and returns the execution result or an error.
func (wlpm *WhiteListProviderManager) propose(addr *ethcommon.Address, args *WhiteListProviderProposalArgs) (*vm.ExecutionResult, error) {
	result := &vm.ExecutionResult{}

	// Check if there are any finished council proposals and provider proposals
	if _, err := wlpm.checkFinishedProposal(); err != nil {
		return nil, err
	}

	if len(args.Providers) < 1 {
		return nil, errors.New("empty providers")
	}

	// Check if the proposal providers have repeated addresses
	if len(lo.Uniq[string](lo.Map[access.WhiteListProvider, string](args.Providers, func(item access.WhiteListProvider, index int) string {
		return item.WhiteListProviderAddr
	}))) != len(args.Providers) {
		return nil, errors.New("provider address repeated")
	}

	// Check if the providers already exist
	existProviders, err := access.GetProviders(wlpm.stateLedger)
	if err != nil {
		return nil, err
	}

	switch ProposalType(args.BaseProposalArgs.ProposalType) {
	case WhiteListProviderAdd:
		// Iterate through the args.Providers array and check if each provider already exists in existProviders
		for _, provider := range args.Providers {
			if common.IsInSlice[string](provider.WhiteListProviderAddr, lo.Map[access.WhiteListProvider, string](existProviders, func(item access.WhiteListProvider, index int) string {
				return item.WhiteListProviderAddr
			})) {
				return nil, fmt.Errorf("provider already exists, %s", provider.WhiteListProviderAddr)
			}
		}
	case WhiteListProviderRemove:
		// Iterate through the args.Providers array and check all providers are in existProviders
		for _, provider := range args.Providers {
			if !common.IsInSlice[string](provider.WhiteListProviderAddr, lo.Map[access.WhiteListProvider, string](existProviders, func(item access.WhiteListProvider, index int) string {
				return item.WhiteListProviderAddr
			})) {
				return nil, fmt.Errorf("provider does not exist, %s", provider.WhiteListProviderAddr)
			}
		}
	}

	// Propose adding or removing providers and return the result
	result.ReturnData, result.Err = wlpm.proposeProvidersAddRemove(addr, args)
	if result.Err != nil {
		return nil, result.Err
	}

	return result, nil
}

func (wlpm *WhiteListProviderManager) getProposalArgs(args *ProposalArgs) (*WhiteListProviderProposalArgs, error) {
	a := &WhiteListProviderProposalArgs{
		BaseProposalArgs: args.BaseProposalArgs,
	}
	providerArgs := &access.WhiteListProviderArgs{}
	if err := json.Unmarshal(args.Extra, providerArgs); err != nil {
		return nil, errors.New("unmarshal provider extra arguments error")
	}
	a.WhiteListProviderArgs = *providerArgs
	return a, nil
}

func (wlpm *WhiteListProviderManager) proposeProvidersAddRemove(addr *ethcommon.Address, args *WhiteListProviderProposalArgs) ([]byte, error) {
	baseProposal, err := wlpm.gov.Propose(addr, ProposalType(args.ProposalType), args.Title, args.Desc, args.BlockNumber, wlpm.lastHeight)
	if err != nil {
		return nil, err
	}
	isExist, council := CheckInCouncil(wlpm.councilAccount, addr.String())
	if !isExist {
		return nil, ErrNotFoundCouncilMember
	}
	proposal := &WhiteListProviderProposal{
		BaseProposal: *baseProposal,
	}
	id, err := wlpm.proposalID.GetAndAddID()
	if err != nil {
		return nil, err
	}
	proposal.ID = id
	proposal.Providers = args.Providers
	proposal.TotalVotes = lo.Sum[uint64](lo.Map[*CouncilMember, uint64](council.Members, func(item *CouncilMember, index int) uint64 {
		return item.Weight
	}))
	b, err := wlpm.saveProviderProposal(proposal)
	if err != nil {
		return nil, err
	}
	// propose generate not finished proposal
	if err = wlpm.notFinishedProposalMgr.SetProposal(&NotFinishedProposal{
		ID:                  proposal.ID,
		DeadlineBlockNumber: proposal.BlockNumber,
		ContractAddr:        common.WhiteListProviderManagerContractAddr,
	}); err != nil {
		return nil, err
	}
	// record log
	wlpm.gov.RecordLog(wlpm.currentLog, ProposeMethod, &proposal.BaseProposal, b)
	return b, nil
}

func (wlpm *WhiteListProviderManager) checkFinishedProposal() (bool, error) {
	notFinishedProposals, err := GetNotFinishedProposals(wlpm.stateLedger)
	if err != nil {
		return false, err
	}

	for _, notFinishedProposal := range notFinishedProposals {
		if notFinishedProposal.ContractAddr == common.CouncilManagerContractAddr || notFinishedProposal.ContractAddr == common.WhiteListProviderManagerContractAddr {
			return false, ErrExistVotingProposal
		}
	}

	return true, nil
}

func (wlpm *WhiteListProviderManager) checkAndUpdateState(lastHeight uint64) {
	if err := CheckAndUpdateState(lastHeight, wlpm.stateLedger); err != nil {
		wlpm.gov.logger.Errorf("check and update state error: %s", err)
	}
}

func loadProviderProposal(stateLedger ledger.StateLedger, proposalID uint64) (*WhiteListProviderProposal, error) {
	addr := types.NewAddressByStr(common.WhiteListProviderManagerContractAddr)
	account := stateLedger.GetOrCreateAccount(addr)

	isExist, data := account.GetState([]byte(fmt.Sprintf("%s%d", WhiteListProviderProposalKey, proposalID)))
	if !isExist {
		return nil, errors.New("provider proposal not found for the id")
	}

	proposal := &WhiteListProviderProposal{}
	if err := json.Unmarshal(data, proposal); err != nil {
		return nil, err
	}

	return proposal, nil
}

func saveProviderProposal(stateLedger ledger.StateLedger, proposal ProposalObject) ([]byte, error) {
	addr := types.NewAddressByStr(common.WhiteListProviderManagerContractAddr)
	account := stateLedger.GetOrCreateAccount(addr)

	b, err := json.Marshal(proposal)
	if err != nil {
		return nil, err
	}
	// save proposal
	account.SetState([]byte(fmt.Sprintf("%s%d", WhiteListProviderProposalKey, proposal.GetID())), b)
	return b, nil
}
