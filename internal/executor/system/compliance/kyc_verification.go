package compliance

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	gov "github.com/axiomesh/axiom-ledger/internal/executor/system/governance"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	vm "github.com/axiomesh/eth-kit/evm"
)

const (
	KycProposalGas uint64 = 30000
	KycVoteGas     uint64 = 30000
	KycInfosKey           = "kycinfosKey"
	KycServicesKey        = "kycservicesKey"
	KycProposalKey        = "KycProposalKey"
)

var _ common.SystemContract = (*KycVerification)(nil)

var (
	ErrCheckKycInfoFail      = errors.New("check kyc info fail")
	ErrCheckKycServiceFail   = errors.New("check kyc service fail")
	ErrKycServiceArgs        = errors.New("unmarshal node extra arguments error")
	ErrNotFoundCouncilMember = errors.New("council member is not found")
	ErrNotFoundKycProposal   = errors.New("node proposal not found for the id")
)

type KycInfo struct {
	User    ethcommon.Address
	KycAddr ethcommon.Address
	KycFlag uint8
	Expires int64
}

type KycService struct {
	KycAddr ethcommon.Address
}

type kycServiceArgs struct {
	services []*KycService
}

type KycProposalArgs struct {
	gov.BaseProposalArgs
	kycServiceArgs
}

type KycVerification struct {
	gov *gov.Governance

	proposalID     *gov.ProposalID
	stateLedger    ledger.StateLedger
	account        ledger.IAccount
	councilAccount ledger.IAccount
	currentLog     *common.Log
}

// NewKycVerification constructs a new KycVerification
func NewKycVerification(cfg *common.SystemContractConfig) common.SystemContract {
	gov, err := gov.NewGov([]gov.ProposalType{gov.KycServiceAdd, gov.KycServiceRemove}, cfg.Logger)
	if err != nil {
		panic(err)
	}

	return &KycVerification{
		gov: gov,
	}
}

func (c *KycVerification) Reset(stateLedger ledger.StateLedger) {
	addr := types.NewAddressByStr(common.KycContractAddr)
	c.account = stateLedger.GetOrCreateAccount(addr)
	c.stateLedger = stateLedger
	c.currentLog = &common.Log{
		Address: addr,
	}
	c.proposalID = gov.NewProposalID(stateLedger)

	councilAddr := types.NewAddressByStr(common.CouncilManagerContractAddr)
	c.councilAccount = stateLedger.GetOrCreateAccount(councilAddr)
}

func (c *KycVerification) EstimateGas(callArgs *types.CallArgs) (uint64, error) {
	args, err := c.gov.GetArgs(&vm.Message{Data: *callArgs.Data})
	if err != nil {
		return 0, err
	}

	var gas uint64
	switch args.(type) {
	case *gov.ProposalArgs:
		gas = KycProposalGas
	case *gov.VoteArgs:
		gas = KycVoteGas
	default:
		return 0, errors.New("unknown proposal args")
	}

	return gas, nil
}

func (c *KycVerification) CheckAndUpdateState(lastHeight uint64, stateLedger ledger.StateLedger) {
	c.Reset(stateLedger)
}

func (c *KycVerification) Run(msg *vm.Message) (*vm.ExecutionResult, error) {
	defer c.gov.SaveLog(c.stateLedger, c.currentLog)

	// parse method and arguments from msg payload
	args, err := c.gov.GetArgs(msg)
	if err != nil {
		return nil, err
	}

	var result *vm.ExecutionResult
	switch v := args.(type) {
	case *gov.ProposalArgs:
		result, err = c.propose(msg.From, v)
	case *gov.VoteArgs:
		result, err = c.vote(msg.From, v)
	default:
		return nil, errors.New("unknown proposal args")
	}

	return result, err
}

func (c *KycVerification) Submit(From *ethcommon.Address, kycInfo *KycInfo) (*vm.ExecutionResult, error) {
	success, _ := CheckInServices(c.account, From.String())
	if !success {
		return nil, ErrCheckKycServiceFail
	}

	b, err := c.saveKycInfo(kycInfo)
	if err != nil {
		return nil, err
	}

	return &vm.ExecutionResult{
		UsedGas:    KycProposalGas,
		ReturnData: b,
		Err:        nil,
	}, nil
}

func Verify(lg ledger.StateLedger, needApprove ethcommon.Address) (bool, error) {
	account := lg.GetOrCreateAccount(types.NewAddressByStr(common.KycContractAddr))
	bytes := account.GetCommittedState([]byte(KycInfosKey))
	if len(bytes) == 0 {
		return false, ErrCheckKycInfoFail
	}
	var infos []*KycInfo
	if err := json.Unmarshal(bytes, &infos); err != nil {
		return false, err
	}

	addrToInfoMap := lo.Associate(infos, func(info *KycInfo) (ethcommon.Address, *KycInfo) {
		return info.User, info
	})

	info := addrToInfoMap[needApprove]

	if info == nil {
		return false, nil
	}

	// if expires == -1 means long-term validity
	if info.Expires == -1 {
		return true, nil
	}

	if time.Now().Unix() > info.Expires || info.KycFlag != 1 {
		return false, ErrCheckKycInfoFail
	}

	return true, nil
}

func (c *KycVerification) saveKycInfo(info *KycInfo) ([]byte, error) {
	success, bytes := c.account.GetState([]byte(KycInfosKey))
	if !success {
		return nil, errors.New("get kyc infos error")
	}

	var infos []*KycInfo
	if err := json.Unmarshal(bytes, &infos); err != nil {
		return nil, err
	}
	infos = append(infos, info)
	print(infos)

	b, err := json.Marshal(infos)
	if err != nil {
		return nil, err
	}

	c.account.SetState([]byte(KycInfosKey), b)

	return b, nil
}

func CheckInServices(account ledger.IAccount, addr string) (bool, *kycServiceArgs) {
	// check council if is exist
	isExist, data := account.GetState([]byte(KycServicesKey))
	if !isExist {
		return false, nil
	}
	kycServices := &kycServiceArgs{}
	if err := json.Unmarshal(data, kycServices); err != nil {
		return false, nil
	}

	// check addr if is exist in council
	isExist = common.IsInSlice[string](addr, lo.Map[*KycService, string](kycServices.services, func(item *KycService, index int) string {
		return item.KycAddr.String()
	}))
	if !isExist {
		return false, nil
	}

	return true, kycServices
}

func (c *KycVerification) propose(addr ethcommon.Address, args *gov.ProposalArgs) (*vm.ExecutionResult, error) {
	result := &vm.ExecutionResult{
		UsedGas: KycProposalGas,
	}

	serviceArgs, err := c.getKycProposalArgs(args)
	if err != nil {
		return nil, err
	}

	result.ReturnData, result.Err = c.proposeServicesAddRemove(addr, serviceArgs)

	return result, nil
}

type KycVoteArgs struct {
	gov.BaseVoteArgs
}

func (c *KycVerification) vote(user ethcommon.Address, voteArgs *gov.VoteArgs) (*vm.ExecutionResult, error) {
	result := &vm.ExecutionResult{UsedGas: KycVoteGas}

	// get proposal
	proposal, err := c.loadKycProposal(voteArgs.ProposalId)
	if err != nil {
		return nil, err
	}

	result.ReturnData, result.Err = c.voteKycAddRemove(user, proposal, &KycVoteArgs{BaseVoteArgs: voteArgs.BaseVoteArgs})
	return result, nil
}

func (c *KycVerification) getKycProposalArgs(args *gov.ProposalArgs) (*KycProposalArgs, error) {
	kycArgs := &KycProposalArgs{
		BaseProposalArgs: args.BaseProposalArgs,
	}

	serviceArgs := &kycServiceArgs{}
	if err := json.Unmarshal(args.Extra, serviceArgs); err != nil {
		return nil, ErrKycServiceArgs
	}

	kycArgs.kycServiceArgs = *serviceArgs
	return kycArgs, nil
}

type KycProposal struct {
	gov.BaseProposal
	kycServiceArgs
}

func (c *KycVerification) proposeServicesAddRemove(addr ethcommon.Address, args *KycProposalArgs) ([]byte, error) {
	baseProposal, err := c.gov.Propose(&addr, gov.ProposalType(args.ProposalType), args.Title, args.Desc, args.BlockNumber)
	if err != nil {
		return nil, err
	}

	isExist, council := gov.CheckInCouncil(c.councilAccount, addr.String())
	if !isExist {
		return nil, ErrNotFoundCouncilMember
	}

	proposal := &KycProposal{
		BaseProposal: *baseProposal,
	}

	id, err := c.proposalID.GetAndAddID()
	if err != nil {
		return nil, err
	}
	proposal.ID = id
	proposal.services = args.services
	proposal.TotalVotes = lo.Sum[uint64](lo.Map[*gov.CouncilMember, uint64](council.Members, func(item *gov.CouncilMember, index int) uint64 {
		return item.Weight
	}))

	b, err := c.saveKycProposal(proposal)
	if err != nil {
		return nil, err
	}

	// record log
	c.gov.RecordLog(c.currentLog, gov.ProposeMethod, &proposal.BaseProposal, b)

	return b, nil
}

func (c *KycVerification) saveKycProposal(proposal *KycProposal) ([]byte, error) {
	b, err := json.Marshal(proposal)
	if err != nil {
		return nil, err
	}
	// save proposal
	c.account.SetState([]byte(fmt.Sprintf("%s%d", KycProposalKey, proposal.ID)), b)

	return b, nil
}

func (c *KycVerification) loadKycProposal(proposalID uint64) (*KycProposal, error) {
	isExist, data := c.account.GetState([]byte(fmt.Sprintf("%s%d", KycProposalKey, proposalID)))
	if !isExist {
		return nil, ErrNotFoundKycProposal
	}

	proposal := &KycProposal{}
	if err := json.Unmarshal(data, proposal); err != nil {
		return nil, err
	}

	return proposal, nil
}

func (c *KycVerification) voteKycAddRemove(user ethcommon.Address, proposal *KycProposal, voteArgs *KycVoteArgs) ([]byte, error) {
	// check user can vote
	isExist, _ := gov.CheckInCouncil(c.councilAccount, user.String())
	if !isExist {
		return nil, ErrNotFoundCouncilMember
	}

	res := gov.VoteResult(voteArgs.VoteResult)
	proposalStatus, err := c.gov.Vote(&user, &proposal.BaseProposal, res)
	if err != nil {
		return nil, err
	}
	proposal.Status = proposalStatus

	b, err := c.saveKycProposal(proposal)
	if err != nil {
		return nil, err
	}

	// if proposal is approved, update the node members
	if proposal.Status == gov.Approved {
		services, err := c.getKycServices()
		if err != nil {
			return nil, err
		}

		if proposal.Type == gov.KycServiceAdd {
			services = append(services, proposal.services...)
		}

		if proposal.Type == gov.KycServiceRemove {
			addrToServiceMap := lo.Associate(proposal.services, func(service *KycService) (ethcommon.Address, *KycService) {
				return service.KycAddr, service
			})

			filteredMembers := lo.Reject(services, func(service *KycService, _ int) bool {
				_, exists := addrToServiceMap[service.KycAddr]
				return exists
			})

			services = filteredMembers
		}

		cb, err := json.Marshal(services)
		if err != nil {
			return nil, err
		}
		c.account.SetState([]byte(KycServicesKey), cb)
	}

	// record log
	c.gov.RecordLog(c.currentLog, gov.VoteMethod, &proposal.BaseProposal, b)

	return b, nil
}

func (c *KycVerification) getKycServices() ([]*KycService, error) {
	success, data := c.account.GetState([]byte(KycServicesKey))
	if success {
		var services []*KycService
		if err := json.Unmarshal(data, &services); err != nil {
			return nil, err
		}
		return services, nil
	}
	return nil, errors.New("kyc services should register first")
}

func InitKycServicesAndInfos(lg ledger.StateLedger, admins []string, accounts []string) error {
	account := lg.GetOrCreateAccount(types.NewAddressByStr(common.KycContractAddr))

	allAddresses := append(admins, accounts...)

	var kycInfos []*KycInfo
	var kycServices []*KycService

	for _, addrStr := range allAddresses {
		addr := ethcommon.HexToAddress(addrStr)

		info := &KycInfo{
			User:    addr,
			KycAddr: addr,
			KycFlag: 1,
			Expires: -1,
		}
		kycInfos = append(kycInfos, info)

		service := &KycService{
			KycAddr: addr,
		}
		kycServices = append(kycServices, service)
	}

	kycInfosBytes, err := json.Marshal(kycInfos)
	if err != nil {
		return err
	}
	account.SetState([]byte(KycInfosKey), kycInfosBytes)

	kycServicesBytes, err := json.Marshal(kycServices)
	if err != nil {
		return err
	}
	account.SetState([]byte(KycServicesKey), kycServicesBytes)

	return nil
}
