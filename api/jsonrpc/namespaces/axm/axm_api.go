package axm

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-ledger/internal/coreapi/api"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

type AxmAPI struct {
	ctx    context.Context
	cancel context.CancelFunc
	rep    *repo.Repo
	api    api.CoreAPI
	logger logrus.FieldLogger
}

func NewAxmAPI(rep *repo.Repo, api api.CoreAPI, logger logrus.FieldLogger) *AxmAPI {
	ctx, cancel := context.WithCancel(context.Background())
	return &AxmAPI{ctx: ctx, cancel: cancel, rep: rep, api: api, logger: logger}
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
