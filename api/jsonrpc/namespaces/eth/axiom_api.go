package eth

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/sirupsen/logrus"

	rpctypes "github.com/axiomesh/axiom-ledger/api/jsonrpc/types"
	"github.com/axiomesh/axiom-ledger/internal/coreapi/api"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

// AxiomAPI provides an API to get related info
type AxiomAPI struct {
	ctx    context.Context
	cancel context.CancelFunc
	rep    *repo.Repo
	api    api.CoreAPI
	logger logrus.FieldLogger
}

func NewAxiomAPI(rep *repo.Repo, api api.CoreAPI, logger logrus.FieldLogger) *AxiomAPI {
	ctx, cancel := context.WithCancel(context.Background())
	return &AxiomAPI{ctx: ctx, cancel: cancel, rep: rep, api: api, logger: logger}
}

// GasPrice returns the current gas price based on dynamic adjustment strategy.
func (api *AxiomAPI) GasPrice() *hexutil.Big {
	defer func(start time.Time) {
		invokeReadOnlyDuration.Observe(time.Since(start).Seconds())
		queryTotalCounter.Inc()
	}(time.Now())

	api.logger.Debug("eth_gasPrice")
	gasPrice, err := api.api.Gas().GetGasPrice()
	if err != nil {
		queryFailedCounter.Inc()
		api.logger.Errorf("get gas price err: %v", err)
	}
	var gasPremium int64
	if !api.rep.Config.JsonRPC.DisableGasPriceAPIPricePremium {
		gasPremium = int64(float64(gasPrice) * api.rep.Config.Genesis.EpochInfo.FinanceParams.GasPremiumRate)
	}
	out := big.NewInt(int64(gasPrice) + gasPremium)
	return (*hexutil.Big)(out)
}

// MaxPriorityFeePerGas returns a suggestion for a gas tip cap for dynamic transactions.
// todo Supplementary gas fee
func (api *AxiomAPI) MaxPriorityFeePerGas(ctx context.Context) (ret *hexutil.Big, err error) {
	defer func(start time.Time) {
		invokeReadOnlyDuration.Observe(time.Since(start).Seconds())
		queryTotalCounter.Inc()
		if err != nil {
			queryFailedCounter.Inc()
		}
	}(time.Now())

	api.logger.Debug("eth_maxPriorityFeePerGas")
	return (*hexutil.Big)(new(big.Int)), nil
}

type feeHistoryResult struct {
	OldestBlock  rpctypes.BlockNumber `json:"oldestBlock"`
	Reward       [][]*hexutil.Big     `json:"reward,omitempty"`
	BaseFee      []*hexutil.Big       `json:"baseFeePerGas,omitempty"`
	GasUsedRatio []float64            `json:"gasUsedRatio"`
}

// FeeHistory return feeHistory
// todo Supplementary feeHsitory
func (api *AxiomAPI) FeeHistory(blockCount rpctypes.DecimalOrHex, lastBlock rpctypes.BlockNumber, rewardPercentiles []float64) (ret *feeHistoryResult, err error) {
	defer func(start time.Time) {
		invokeReadOnlyDuration.Observe(time.Since(start).Seconds())
		queryTotalCounter.Inc()
		if err != nil {
			queryFailedCounter.Inc()
		}
	}(time.Now())

	api.logger.Debug("eth_feeHistory")
	return nil, ErrNotSupportApiError
}

// Syncing returns whether or not the current node is syncing with other peers. Returns false if not, or a struct
// outlining the state of the sync if it is.
func (api *AxiomAPI) Syncing() (ret any, err error) {
	defer func(start time.Time) {
		invokeReadOnlyDuration.Observe(time.Since(start).Seconds())
		queryTotalCounter.Inc()
		if err != nil {
			queryFailedCounter.Inc()
		}
	}(time.Now())

	api.logger.Debug("eth_syncing")

	// TODO
	// Supplementary data
	syncBlock := make(map[string]string)
	meta, err := api.api.Chain().Meta()
	if err != nil {
		return false, err
	}

	syncBlock["startingBlock"] = fmt.Sprintf("%d", hexutil.Uint64(1))
	syncBlock["highestBlock"] = fmt.Sprintf("%d", hexutil.Uint64(meta.Height))
	syncBlock["currentBlock"] = syncBlock["highestBlock"]
	return syncBlock, nil
}

func (api *AxiomAPI) Accounts() (ret []common.Address, err error) {
	defer func(start time.Time) {
		invokeReadOnlyDuration.Observe(time.Since(start).Seconds())
		queryTotalCounter.Inc()
		if err != nil {
			queryFailedCounter.Inc()
		}
	}(time.Now())

	accounts := api.rep.Config.Genesis.Accounts
	res := make([]common.Address, 0)
	for _, account := range accounts {
		res = append(res, common.HexToAddress(account))
	}
	return res, nil
}
