package eth

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-ledger/internal/coreapi/api"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

// EthereumAPI provides an API to get related info
type EthereumAPI struct {
	ctx    context.Context
	cancel context.CancelFunc
	rep    *repo.Repo
	api    api.CoreAPI
	logger logrus.FieldLogger
}

func NewEthereumAPI(rep *repo.Repo, api api.CoreAPI, logger logrus.FieldLogger) *EthereumAPI {
	ctx, cancel := context.WithCancel(context.Background())
	return &EthereumAPI{ctx: ctx, cancel: cancel, rep: rep, api: api, logger: logger}
}

// Syncing returns whether the current node is syncing with other peers. Returns false if not, or a struct
// outlining the state of the sync if it is.
func (api *EthereumAPI) Syncing() (ret any, err error) {
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

func (api *EthereumAPI) Accounts() (ret []common.Address, err error) {
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
