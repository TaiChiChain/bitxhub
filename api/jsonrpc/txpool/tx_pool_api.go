package txpool

import (
	"context"
	"fmt"

	"github.com/axiomesh/axiom-kit/txpool"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/api/jsonrpc/namespaces/eth"
	rpctypes "github.com/axiomesh/axiom-ledger/api/jsonrpc/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-ledger/internal/coreapi/api"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

const (
	pendingType = "pending"
	queuedType  = "queued"

	unit = "wei"
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

// convert account's txMeta to <nonce:RPCTransaction>
// pending: All currently processable transactions, store in txpool with priorityIndex field
// queued: Queued but non-processable transactions, store in txpool with parkingLotIndex field
func (api *TxPoolAPI) formatTxMeta(meta *txpool.AccountMeta[types.Transaction, *types.Transaction]) (map[string]*rpctypes.RPCTransaction, map[string]*rpctypes.RPCTransaction) {
	if len(meta.Txs) == 0 {
		return make(map[string]*rpctypes.RPCTransaction), make(map[string]*rpctypes.RPCTransaction)
	}
	txs := lo.Map(meta.Txs, func(tx *txpool.TxInfo[types.Transaction, *types.Transaction], index int) *types.Transaction {
		return tx.Tx
	})

	pending := make(map[string]*rpctypes.RPCTransaction)
	queued := make(map[string]*rpctypes.RPCTransaction)
	lo.ForEach(txs, func(tx *types.Transaction, _ int) {
		nonce := tx.GetNonce()
		if nonce > meta.PendingNonce {
			queued[fmt.Sprintf("%d", nonce)] = eth.NewRPCTransaction(tx, common.Hash{}, 0, 0)
		} else {
			pending[fmt.Sprintf("%d", nonce)] = eth.NewRPCTransaction(tx, common.Hash{}, 0, 0)
		}
	})

	return pending, queued
}

// Content returns the transactions contained within the transaction pool.
// - pending/queued: The transactions status in the pool.
//   - account: The address of the account.
//   - nonce: The nonce of the transaction.
//   - tx: The transaction content.
func (api *TxPoolAPI) Content() any {
	return api.getContent()
}

func (api *TxPoolAPI) SimpleContent() any {
	return api.getSimpleContent()
}

func (api *TxPoolAPI) getContent() ContentResponse {
	data := api.api.TxPool().GetMeta(true)

	txpoolMeta, ok := data.(*txpool.Meta[types.Transaction, *types.Transaction])
	if !ok {
		api.logger.Errorf("failed to get txpool meta")
		return ContentResponse{}
	}

	pendingTxsM := make(map[string]map[string]*rpctypes.RPCTransaction)
	queuedTxsM := make(map[string]map[string]*rpctypes.RPCTransaction)
	for account, accMeta := range txpoolMeta.Accounts {
		pending, queue := api.formatTxMeta(accMeta)
		if len(pending) != 0 {
			pendingTxsM[account] = pending
		}
		if len(queue) != 0 {
			queuedTxsM[account] = queue
		}
	}
	resp := ContentResponse{
		Pending:         pendingTxsM,
		Queued:          queuedTxsM,
		TxCountLimit:    txpoolMeta.TxCountLimit,
		TxCount:         txpoolMeta.TxCount,
		ReadyTxCount:    txpoolMeta.ReadyTxCount,
		NotReadyTxCount: txpoolMeta.NotReadyTxCount,
	}

	return resp
}

func (api *TxPoolAPI) getSimpleContent() SimpleContentResponse {
	data := api.api.TxPool().GetMeta(false)

	txpoolMeta, ok := data.(*txpool.Meta[types.Transaction, *types.Transaction])
	if !ok {
		api.logger.Errorf("failed to get txpool meta")
		return SimpleContentResponse{}
	}

	SimpleContentM := make(map[string]SimpleAccountContent)
	for account, accMeta := range txpoolMeta.Accounts {
		pending := make(map[string]string)
		queued := make(map[string]string)
		lo.ForEach(accMeta.SimpleTxs, func(info *txpool.TxSimpleInfo, index int) {
			if info.Nonce > accMeta.PendingNonce {
				queued[fmt.Sprintf("%d", info.Nonce)] = info.Hash
			} else {
				pending[fmt.Sprintf("%d", info.Nonce)] = info.Hash
			}
		})
		SimpleContentM[account] = SimpleAccountContent{
			Pending:      pending,
			Queued:       queued,
			CommitNonce:  accMeta.CommitNonce,
			PendingNonce: accMeta.PendingNonce,
			TxCount:      accMeta.TxCount,
		}
	}
	resp := SimpleContentResponse{
		SimpleAccountContent: SimpleContentM,
		TxCountLimit:         txpoolMeta.TxCountLimit,
		TxCount:              txpoolMeta.TxCount,
		ReadyTxCount:         txpoolMeta.ReadyTxCount,
		NotReadyTxCount:      txpoolMeta.NotReadyTxCount,
	}

	return resp
}

func (api *TxPoolAPI) ContentFrom(addr common.Address) any {
	data := api.api.TxPool().GetAccountMeta(addr.String(), true)
	accountMeta, ok := data.(*txpool.AccountMeta[types.Transaction, *types.Transaction])
	if !ok {
		api.logger.Errorf("failed to get account meta: %s", addr.String())
		return AccountContentResponse{}
	}

	pending, queue := api.formatTxMeta(accountMeta)

	return AccountContentResponse{
		Pending:      pending,
		Queued:       queue,
		CommitNonce:  accountMeta.CommitNonce,
		PendingNonce: accountMeta.PendingNonce,
		TxCount:      accountMeta.TxCount,
	}
}

// Inspect retrieves the content of the transaction pool and flattens it into an easily inspectable list.
func (api *TxPoolAPI) Inspect() any {
	rawContent := api.getContent()

	// Define a formatter to flatten a transaction into a string
	var format = func(tx *rpctypes.RPCTransaction) string {
		if tx.To != nil {
			return fmt.Sprintf("%s: %v %s + %v gas × %v %s", tx.To, tx.Value, unit, tx.Gas, tx.GasPrice, unit)
		}
		return fmt.Sprintf("%s: %v %s + %v gas × %v %s", tx.To, tx.Value, unit, tx.Gas, tx.GasPrice, unit)
	}

	// Define to flatten pending and queued transactions
	var flatten = func(rawCt map[string]map[string]*rpctypes.RPCTransaction) map[string]map[string]string {
		flattenContent := make(map[string]map[string]string)
		for account, txs := range rawCt {
			dump := make(map[string]string)
			for nonce, tx := range txs {
				dump[nonce] = format(tx)
			}
			flattenContent[account] = dump
		}

		return flattenContent
	}

	pendingContent := flatten(rawContent.Pending)
	queuedContent := flatten(rawContent.Queued)
	return InspectResponse{
		Pending: pendingContent,
		Queued:  queuedContent,
	}
}

func (api *TxPoolAPI) Status() any {
	data := api.api.TxPool().GetMeta(false)
	meta, ok := data.(*txpool.Meta[types.Transaction, *types.Transaction])
	if !ok {
		api.logger.Errorf("failed to get txpool meta")
		return nil
	}
	return StatusResponse{
		Pending: meta.ReadyTxCount,
		Queued:  meta.NotReadyTxCount,
		Total:   meta.TxCount,
	}
}

func (api *TxPoolAPI) GetChainInfo() any {
	return api.api.TxPool().GetChainInfo()
}
