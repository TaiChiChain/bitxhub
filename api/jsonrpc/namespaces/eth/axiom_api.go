package eth

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

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
	out := big.NewInt(int64(gasPrice))
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

// Syncing returns whether the current node is syncing with other peers. Returns false if not, or a struct
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

	accounts := api.rep.GenesisConfig.Accounts

	res := lo.Map(accounts, func(ac *repo.Account, index int) common.Address {
		return common.HexToAddress(ac.Address)
	})
	return res, nil
}
