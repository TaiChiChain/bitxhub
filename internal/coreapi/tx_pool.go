package coreapi

import (
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/coreapi/api"
)

type TxPoolAPI CoreAPI

var _ api.TxPoolAPI = (*TxPoolAPI)(nil)

func (api *TxPoolAPI) GetTotalPendingTxCount() uint64 {
	if api.axiomLedger.Repo.ReadonlyMode {
		return 0
	}
	return api.axiomLedger.Consensus.GetTotalPendingTxCount()
}

func (api *TxPoolAPI) GetPendingTxCountByAccount(account string) uint64 {
	if api.axiomLedger.Repo.ReadonlyMode {
		return 0
	}
	return api.axiomLedger.Consensus.GetPendingTxCountByAccount(account)
}

func (api *TxPoolAPI) GetTransaction(hash *types.Hash) *types.Transaction {
	if api.axiomLedger.Repo.ReadonlyMode {
		return nil
	}
	return api.axiomLedger.Consensus.GetPendingTxByHash(hash)
}

func (api *TxPoolAPI) GetAccountMeta(account string, full bool) any {
	if api.axiomLedger.Repo.ReadonlyMode {
		return nil
	}

	return api.axiomLedger.Consensus.GetAccountPoolMeta(account, full)
}

func (api *TxPoolAPI) GetMeta(full bool) any {
	if api.axiomLedger.Repo.ReadonlyMode {
		return nil
	}
	return api.axiomLedger.Consensus.GetPoolMeta(full)
}
