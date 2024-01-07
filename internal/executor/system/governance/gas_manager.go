package governance

import (
	"encoding/json"
	"errors"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/base"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/ethereum/go-ethereum/log"
)

const (
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

type GasManager struct {
	gov *Governance
}

func NewGasManager(gov *Governance) *GasManager {
	return &GasManager{
		gov: gov,
	}
}

func (gm *GasManager) ProposeCheck(proposalType ProposalType, extra []byte) error {
	_, err := gm.getGasProposalExtraArgs(extra)
	if err != nil {
		return err
	}

	// Check for outstanding council proposals and gas proposals
	if _, err := gm.checkNotFinishedProposal(); err != nil {
		return err
	}

	return nil
}

func (gm *GasManager) Execute(proposal *Proposal) error {
	// if proposal is approved, update the EpochInfo gas
	if proposal.Status == Approved {
		extraArgs, err := gm.getGasProposalExtraArgs(proposal.Extra)
		if err != nil {
			return err
		}

		epochInfo, err := base.GetNextEpochInfo(gm.gov.stateLedger)
		if err != nil {
			return err
		}
		financeParams := epochInfo.FinanceParams
		financeParams.MaxGasPrice = extraArgs.MaxGasPrice
		financeParams.MinGasPrice = extraArgs.MinGasPrice
		financeParams.StartGasPrice = extraArgs.InitGasPrice
		financeParams.GasChangeRateValue = extraArgs.GasChangeRateValue
		financeParams.StartGasPriceAvailable = true
		epochInfo.FinanceParams = financeParams

		if err := base.SetNextEpochInfo(gm.gov.stateLedger, epochInfo); err != nil {
			return err
		}

		cb, err := json.Marshal(extraArgs)
		if err != nil {
			return err
		}
		gm.gov.account.SetState([]byte(GasParamsKey), cb)
	}

	return nil
}

func (gm *GasManager) checkNotFinishedProposal() (bool, error) {
	notFinishedProposals, err := gm.gov.notFinishedProposalMgr.GetProposals()
	if err != nil {
		return false, err
	}

	for _, notFinishedProposal := range notFinishedProposals {
		if notFinishedProposal.Type == CouncilElect {
			return false, ErrExistNotFinishedCouncilProposal
		}
		if notFinishedProposal.Type == GasUpdate {
			return false, ErrExistNotFinishedGasProposal
		}
	}
	return true, nil
}

func (gm *GasManager) getGasProposalExtraArgs(extra []byte) (*GasExtraArgs, error) {
	extraArgs := &GasExtraArgs{}
	if err := json.Unmarshal(extra, extraArgs); err != nil {
		gm.gov.logger.Errorf("Unmarshal extra args error: %s", err)
		return nil, ErrGasExtraArgs
	}

	if !isPositiveInteger(extraArgs.GasChangeRateValue) || !isPositiveInteger(extraArgs.InitGasPrice) || !isPositiveInteger(extraArgs.MinGasPrice) || !isPositiveInteger(extraArgs.MaxGasPrice) {
		return nil, ErrGasArgsType
	}

	if extraArgs.InitGasPrice < extraArgs.MinGasPrice || extraArgs.InitGasPrice > extraArgs.MaxGasPrice || extraArgs.MinGasPrice > extraArgs.MaxGasPrice {
		return nil, ErrGasUpperOrLlower
	}

	// Check whether the gas proposal and GetNextEpochInfo are consistent
	epochInfo, err := base.GetNextEpochInfo(gm.gov.stateLedger)
	if err != nil {
		return nil, err
	}
	financeParams := epochInfo.FinanceParams
	if financeParams.GasChangeRateValue == extraArgs.GasChangeRateValue && financeParams.StartGasPrice == extraArgs.InitGasPrice && financeParams.MaxGasPrice == extraArgs.MaxGasPrice && financeParams.MinGasPrice == extraArgs.MinGasPrice {
		return nil, ErrRepeatedGasInfo
	}

	return extraArgs, nil
}

func isPositiveInteger(num uint64) bool {

	if num > 0 {
		log.Info("isPositiveInteger-Type is num:", num)
		return true
	}
	return false
}

func InitGasParam(lg ledger.StateLedger, epochInfo *rbft.EpochInfo) error {
	// read member config, write to ViewLedger
	gasParams := getGasInfo(epochInfo)
	g, err := json.Marshal(gasParams)
	if err != nil {
		return err
	}
	account := lg.GetOrCreateAccount(types.NewAddressByStr(common.GovernanceContractAddr))
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
