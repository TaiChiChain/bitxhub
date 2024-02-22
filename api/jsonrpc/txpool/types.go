package txpool

import (
	rpctypes "github.com/axiomesh/axiom-ledger/api/jsonrpc/types"
)

type ContentResponse struct {
	Pending         map[string]map[string]*rpctypes.RPCTransaction `json:"pending"`
	Queued          map[string]map[string]*rpctypes.RPCTransaction `json:"queued"`
	TxCountLimit    uint64                                         `json:"txCountLimit"`
	TxCount         uint64                                         `json:"txCount"`
	ReadyTxCount    uint64                                         `json:"readyTxCount"`
	NotReadyTxCount uint64                                         `json:"notReadyTxCount"`
}

type SimpleContentResponse struct {
	SimpleAccountContent map[string]SimpleAccountContent `json:"simpleAccountContent"`
	TxCountLimit         uint64                          `json:"txCountLimit"`
	TxCount              uint64                          `json:"txCount"`
	ReadyTxCount         uint64                          `json:"readyTxCount"`
	NotReadyTxCount      uint64                          `json:"notReadyTxCount"`
}

type SimpleAccountContent struct {
	Pending      map[string]string `json:"pending"`
	Queued       map[string]string `json:"queued"`
	CommitNonce  uint64            `json:"commitNonce"`
	PendingNonce uint64            `json:"pendingNonce"`
	TxCount      uint64            `json:"txCount"`
}

type AccountContentResponse struct {
	Pending      map[string]*rpctypes.RPCTransaction `json:"pending"`
	Queued       map[string]*rpctypes.RPCTransaction `json:"queued"`
	CommitNonce  uint64                              `json:"commitNonce"`
	PendingNonce uint64                              `json:"pendingNonce"`
	TxCount      uint64                              `json:"txCount"`
}

type InspectResponse struct {
	Pending map[string]map[string]string `json:"pending"`
	Queued  map[string]map[string]string `json:"queued"`
}

type StatusResponse struct {
	Pending uint64 `json:"pending"`
	Queued  uint64 `json:"queued"`
	Total   uint64 `json:"total"`
}
