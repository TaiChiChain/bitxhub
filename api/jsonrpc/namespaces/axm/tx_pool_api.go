package axm

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-ledger/internal/coreapi/api"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

type TxPoolAPI struct {
	ctx    context.Context
	cancel context.CancelFunc
	rep    *repo.Repo
	api    api.CoreAPI
	logger logrus.FieldLogger
}

func NewTxPoolAPI(rep *repo.Repo, api api.CoreAPI, logger logrus.FieldLogger) *TxPoolAPI {
	ctx, cancel := context.WithCancel(context.Background())
	return &TxPoolAPI{ctx: ctx, cancel: cancel, rep: rep, api: api, logger: logger}
}

func (api *TxPoolAPI) TotalPendingTxCount() uint64 {
	return api.api.TxPool().GetTotalPendingTxCount()
}

func (api *TxPoolAPI) PendingTxCountFrom(addr common.Address) uint64 {
	return api.api.TxPool().GetPendingTxCountByAccount(addr.String())
}

func (api *TxPoolAPI) MetaFrom(addr common.Address, full *bool) any {
	f := false
	if full != nil {
		f = *full
	}
	return api.api.TxPool().GetAccountMeta(addr.String(), f)
}

func (api *TxPoolAPI) Meta(full *bool) any {
	f := false
	if full != nil {
		f = *full
	}
	return api.api.TxPool().GetMeta(f)
}
