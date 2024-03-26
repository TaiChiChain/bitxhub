package axm

import (
	"context"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-ledger/internal/coreapi/api"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

type AxmAPI struct {
	ctx    context.Context
	cancel context.CancelFunc
	rep    *repo.Repo
	api    api.CoreAPI
	logger logrus.FieldLogger

	lock                  sync.Mutex
	epochCache            uint64
	incentiveAddressCache string
}

func NewAxmAPI(rep *repo.Repo, api api.CoreAPI, logger logrus.FieldLogger) *AxmAPI {
	ctx, cancel := context.WithCancel(context.Background())
	return &AxmAPI{ctx: ctx, cancel: cancel, rep: rep, api: api, logger: logger}
}

func (api *AxmAPI) GetIncentiveAddress() (res common.Address, err error) {
	if api.epochCache != api.rep.EpochInfo.Epoch {
		api.lock.Lock()
		defer api.lock.Unlock()
		if api.epochCache != api.rep.EpochInfo.Epoch {
			incentiveAddressNow, err := matchNodeIncentiveAddress(api.rep.P2PID, api.rep.EpochInfo)
			if err != nil {
				return common.Address{}, err
			}
			api.incentiveAddressCache = incentiveAddressNow
			api.epochCache = api.rep.EpochInfo.Epoch
		}
	}
	return common.HexToAddress(api.incentiveAddressCache), nil
}

func (api *AxmAPI) Status() any {
	syncStatus := make(map[string]string)
	err := api.api.Broker().ConsensusReady()
	if err != nil {
		syncStatus["status"] = "abnormal"
		syncStatus["error_msg"] = err.Error()
		return syncStatus
	}
	syncStatus["status"] = "normal"
	return syncStatus
}

func matchNodeIncentiveAddress(P2PID string, EpochInfo *rbft.EpochInfo) (res string, err error) {
	for _, set := range []*[]rbft.NodeInfo{&EpochInfo.ValidatorSet, &EpochInfo.CandidateSet, &EpochInfo.DataSyncerSet} {
		for _, nodeInfo := range *set {
			if P2PID == nodeInfo.P2PNodeID {
				return nodeInfo.AccountAddress, nil
			}
		}
	}
	return "", fmt.Errorf("unable to match node incentive address")
}

func (api *AxmAPI) SyncProgress() any {
	progress := api.api.Broker().GetSyncProgress()
	meta, err := api.api.Chain().Meta()
	var highestBlock uint64
	if err != nil {
		api.logger.Error(err)
		progress.HighestBlockHeight = 0
		progress.CatchUp = false
	} else {
		highestBlock = meta.Height
		progress.HighestBlockHeight = highestBlock
		if highestBlock >= progress.TargetHeight {
			progress.CatchUp = true
			progress.TargetHeight = 0
		}
	}

	return progress
}
