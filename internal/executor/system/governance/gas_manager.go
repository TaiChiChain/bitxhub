package governance

import (
	"encoding/json"
	"errors"
	"fmt"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/base"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	vm "github.com/axiomesh/eth-kit/evm"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/samber/lo"
)

const (
	// GasProposalKey is key for GasProposal storage
	GasProposalKey = "gasProposalKey"

	GasParamsKey = "gasParamsKey"
)

var (
	ErrExistNotFinishedGasProposal     = errors.New("exist not gas finished proposal")
	ErrExistNotFinishedCouncilProposal = errors.New("exist not finished council proposal")
	ErrNotFoundGasProposal             = errors.New("gas proposal not found for the id")
	ErrUnKnownGasProposalArgs          = errors.New("unknown proposal args")
	ErrGasExtraArgs                    = errors.New("unmarshal gas extra arguments error")
	ErrGasArgsType                     = errors.New("gas arguments type error")
	ErrGasUpperOrLlower                = errors.New("gas upper or lower limit error")
	ErrRepeatedGasInfo                 = errors.New("repeated gas info")
)

type GasExtraArgs struct {
	MaxGasPrice        uint64
	MinGasPrice        uint64
	InitGasPrice       uint64
	GasChangeRateValue uint64
}

type GasProposalArgs struct {
	BaseProposalArgs
	GasExtraArgs
}

type GasProposal struct {
	BaseProposal
	GasExtraArgs
}

type GasVoteArgs struct {
	BaseVoteArgs
}

type GasManager struct {
	gov *Governance

	account                ledger.IAccount
	councilAccount         ledger.IAccount
	stateLedger            ledger.StateLedger
	currentLog             *common.Log
	proposalID             *ProposalID
	lastHeight             uint64
	notFinishedProposalMgr *NotFinishedProposalMgr
}

func NewGasManager(cfg *common.SystemContractConfig) *GasManager {
	gov, err := NewGov([]ProposalType{GasUpdate}, cfg.Logger)
	if err != nil {
		panic(err)
	}

	return &GasManager{
		gov: gov,
	}
}

func (gm *GasManager) Reset(lastHeight uint64, stateLedger ledger.StateLedger) {
	addr := types.NewAddressByStr(common.GasManagerContractAddr)
	gm.account = stateLedger.GetOrCreateAccount(addr)
	gm.stateLedger = stateLedger
	gm.currentLog = &common.Log{
		Address: addr,
	}
	gm.proposalID = NewProposalID(stateLedger)

	councilAddr := types.NewAddressByStr(common.CouncilManagerContractAddr)
	gm.councilAccount = stateLedger.GetOrCreateAccount(councilAddr)
	gm.notFinishedProposalMgr = NewNotFinishedProposalMgr(stateLedger)

	// check and update
	gm.checkAndUpdateState(lastHeight)
	gm.lastHeight = lastHeight
}

func (gm *GasManager) Run(msg *vm.Message) (*vm.ExecutionResult, error) {
	defer gm.gov.SaveLog(gm.stateLedger, gm.currentLog)

	// parse method and arguments from msg payload
	args, err := gm.gov.GetArgs(msg)
	if err != nil {
		return nil, err
	}

	var result *vm.ExecutionResult
	switch v := args.(type) {
	case *ProposalArgs:
		result, err = gm.propose(msg.From, v)
	case *VoteArgs:
		result, err = gm.vote(msg.From, v)
	case *GetProposalArgs:
		result, err = gm.getProposal(v.ProposalID)
	default:
		return nil, errors.New("unknown proposal args")
	}
	if result != nil {
		result.UsedGas = common.CalculateDynamicGas(msg.Data)
	}
	return result, err
}

func (gm *GasManager) propose(addr ethcommon.Address, args *ProposalArgs) (*vm.ExecutionResult, error) {
	// Check for outstanding council proposals and gas proposals
	if _, err := gm.checkNotFinishedProposal(); err != nil {
		return nil, err
	}

	result := &vm.ExecutionResult{}
	result.ReturnData, result.Err = gm.proposeGasUpdate(addr, args)

	return result, nil
}

func (gm *GasManager) proposeGasUpdate(addr ethcommon.Address, args *ProposalArgs) ([]byte, error) {
	gasArgs, err := gm.getGasProposalArgs(args)
	if err != nil {
		return nil, err
	}

	isExist, council := CheckInCouncil(gm.councilAccount, addr.String())
	if !isExist {
		return nil, ErrNotFoundCouncilMember
	}

	baseProposal, err := gm.gov.Propose(&addr, ProposalType(gasArgs.ProposalType), gasArgs.Title, gasArgs.Desc, gasArgs.BlockNumber, gm.lastHeight, false)
	if err != nil {
		return nil, err
	}

	// set proposal id
	proposal := &GasProposal{
		BaseProposal: *baseProposal,
	}

	id, err := gm.proposalID.GetAndAddID()
	if err != nil {
		return nil, err
	}
	proposal.ID = id
	proposal.GasExtraArgs = gasArgs.GasExtraArgs

	proposal.TotalVotes = lo.Sum[uint64](lo.Map[*CouncilMember, uint64](council.Members, func(item *CouncilMember, index int) uint64 {
		return item.Weight
	}))

	b, err := gm.saveGasProposal(proposal)
	if err != nil {
		return nil, err
	}

	// propose generate not finished proposal
	if err = gm.notFinishedProposalMgr.SetProposal(&NotFinishedProposal{
		ID:                  proposal.ID,
		DeadlineBlockNumber: proposal.BlockNumber,
		ContractAddr:        common.GasManagerContractAddr,
	}); err != nil {
		return nil, err
	}

	returnData, err := gm.gov.PackOutputArgs(ProposeMethod, id)
	if err != nil {
		return nil, err
	}

	// record log
	gm.gov.RecordLog(gm.currentLog, ProposeMethod, &proposal.BaseProposal, b)

	return returnData, nil
}

// Vote a proposal, return vote status
func (gm *GasManager) vote(user ethcommon.Address, voteArgs *VoteArgs) (*vm.ExecutionResult, error) {
	result := &vm.ExecutionResult{}

	// get proposal
	proposal, err := gm.loadGasProposal(voteArgs.ProposalId)
	if err != nil {
		return nil, err
	}

	// check user can vote
	isExist, _ := CheckInCouncil(gm.councilAccount, user.String())
	if !isExist {
		return nil, ErrNotFoundCouncilMember
	}

	res := VoteResult(voteArgs.VoteResult)
	proposalStatus, err := gm.gov.Vote(&user, &proposal.BaseProposal, res)
	if err != nil {
		return nil, err
	}
	proposal.Status = proposalStatus

	b, err := gm.saveGasProposal(proposal)
	if err != nil {
		return nil, err
	}

	// update not finished proposal
	if proposal.Status == Approved || proposal.Status == Rejected {
		if err := gm.notFinishedProposalMgr.RemoveProposal(proposal.ID); err != nil {
			return nil, err
		}
	}

	// if proposal is approved, update the EpochInfo gas
	if proposal.Status == Approved {
		epochInfo, err := base.GetNextEpochInfo(gm.stateLedger)
		if err != nil {
			return nil, err
		}
		financeParams := epochInfo.FinanceParams
		financeParams.MaxGasPrice = proposal.MaxGasPrice
		financeParams.MinGasPrice = proposal.MinGasPrice
		financeParams.StartGasPrice = proposal.InitGasPrice
		financeParams.GasChangeRateValue = proposal.GasChangeRateValue
		financeParams.StartGasPriceAvailable = true
		epochInfo.FinanceParams = financeParams

		if err := base.SetNextEpochInfo(gm.stateLedger, epochInfo); err != nil {
			return nil, err
		}

		cb, err := json.Marshal(proposal.GasExtraArgs)
		if err != nil {
			return nil, err
		}
		gm.account.SetState([]byte(GasParamsKey), cb)
	}

	// record log
	gm.gov.RecordLog(gm.currentLog, VoteMethod, &proposal.BaseProposal, b)

	return result, nil
}

// getProposal view proposal details
func (gm *GasManager) getProposal(proposalID uint64) (*vm.ExecutionResult, error) {
	result := &vm.ExecutionResult{}

	isExist, b := gm.account.GetState([]byte(fmt.Sprintf("%s%d", GasProposalKey, proposalID)))
	if isExist {
		packed, err := gm.gov.PackOutputArgs(ProposalMethod, b)
		if err != nil {
			return nil, err
		}
		result.ReturnData = packed
		return result, nil
	}
	return nil, ErrNotFoundGasProposal
}

func (gm *GasManager) checkNotFinishedProposal() (bool, error) {
	notFinishedProposals, err := GetNotFinishedProposals(gm.stateLedger)

	if err != nil {
		return false, err
	}

	for _, notFinishedProposal := range notFinishedProposals {
		if notFinishedProposal.ContractAddr == common.CouncilManagerContractAddr {
			return false, ErrExistNotFinishedCouncilProposal
		}
		if notFinishedProposal.ContractAddr == common.GasManagerContractAddr {
			return false, ErrExistNotFinishedGasProposal
		}
	}
	return true, nil
}

func (gm *GasManager) getGasProposalArgs(args *ProposalArgs) (*GasProposalArgs, error) {
	gasArgs := &GasProposalArgs{
		BaseProposalArgs: args.BaseProposalArgs,
	}

	extraArgs := &GasExtraArgs{}
	if err := json.Unmarshal(args.Extra, extraArgs); err != nil {
		return nil, ErrGasExtraArgs
	}

	if !isPositiveInteger(extraArgs.GasChangeRateValue) || !isPositiveInteger(extraArgs.InitGasPrice) || !isPositiveInteger(extraArgs.MinGasPrice) || !isPositiveInteger(extraArgs.MaxGasPrice) {
		return nil, ErrGasArgsType
	}

	if extraArgs.InitGasPrice < extraArgs.MinGasPrice || extraArgs.InitGasPrice > extraArgs.MaxGasPrice || extraArgs.MinGasPrice > extraArgs.MaxGasPrice {
		return nil, ErrGasUpperOrLlower
	}

	// Check whether the gas proposal and GetNextEpochInfo are consistent
	epochInfo, err := base.GetNextEpochInfo(gm.stateLedger)
	if err != nil {
		return nil, err
	}
	financeParams := epochInfo.FinanceParams
	if financeParams.GasChangeRateValue == extraArgs.GasChangeRateValue && financeParams.StartGasPrice == extraArgs.InitGasPrice && financeParams.MaxGasPrice == extraArgs.MaxGasPrice && financeParams.MinGasPrice == extraArgs.MinGasPrice {
		return nil, ErrRepeatedGasInfo
	}

	gasArgs.GasExtraArgs = *extraArgs
	return gasArgs, nil
}

func isPositiveInteger(num uint64) bool {

	if num > 0 {
		log.Info("isPositiveInteger-Type is num:", num)
		return true
	}
	return false
}

func InitGasMembers(lg ledger.StateLedger, epochInfo *rbft.EpochInfo) error {
	// read member config, write to ViewLedger
	gasParams := getGasInfo(epochInfo)
	g, err := json.Marshal(gasParams)
	if err != nil {
		return err
	}
	account := lg.GetOrCreateAccount(types.NewAddressByStr(common.GasManagerContractAddr))
	account.SetState([]byte(GasParamsKey), g)
	return nil
}

func getGasInfo(epochInfo *rbft.EpochInfo) GasExtraArgs {
	var result GasExtraArgs
	financeParams := epochInfo.FinanceParams
	result.InitGasPrice = financeParams.StartGasPrice
	result.MinGasPrice = financeParams.MinGasPrice
	result.MaxGasPrice = financeParams.MaxGasPrice
	result.GasChangeRateValue = financeParams.GasChangeRateValue

	return result
}

func (gm *GasManager) checkAndUpdateState(lastHeight uint64) {
	if err := CheckAndUpdateState(lastHeight, gm.stateLedger); err != nil {
		gm.gov.logger.Errorf("check and update state error: %s", err)
	}
}

func (gm *GasManager) EstimateGas(callArgs *types.CallArgs) (uint64, error) {
	_, err := gm.gov.GetArgs(&vm.Message{Data: *callArgs.Data})
	if err != nil {
		return 0, err
	}

	return common.CalculateDynamicGas(*callArgs.Data), nil
}

func (gm *GasManager) saveGasProposal(proposal *GasProposal) ([]byte, error) {
	return saveGasProposal(gm.stateLedger, proposal)
}

func (gm *GasManager) loadGasProposal(proposalID uint64) (*GasProposal, error) {
	return loadGasProposal(gm.stateLedger, proposalID)
}

func saveGasProposal(stateLedger ledger.StateLedger, proposal ProposalObject) ([]byte, error) {
	addr := types.NewAddressByStr(common.GasManagerContractAddr)
	account := stateLedger.GetOrCreateAccount(addr)

	b, err := json.Marshal(proposal)
	if err != nil {
		return nil, err
	}
	// save proposal
	account.SetState([]byte(fmt.Sprintf("%s%d", GasProposalKey, proposal.GetID())), b)

	return b, nil
}

func loadGasProposal(stateLedger ledger.StateLedger, proposalID uint64) (*GasProposal, error) {
	addr := types.NewAddressByStr(common.GasManagerContractAddr)
	account := stateLedger.GetOrCreateAccount(addr)

	isExist, data := account.GetState([]byte(fmt.Sprintf("%s%d", GasProposalKey, proposalID)))
	if !isExist {
		return nil, ErrNotFoundGasProposal
	}

	proposal := &GasProposal{}
	if err := json.Unmarshal(data, proposal); err != nil {
		return nil, err
	}

	return proposal, nil
}
