package common

import (
	"github.com/samber/lo"

	"github.com/axiomesh/axiom-bft/common/consensus"
	"github.com/axiomesh/axiom-bft/txpool"
	"github.com/axiomesh/axiom-kit/types"
)

const (
	LocalTxEvent = iota
	RemoteTxEvent
)

// UncheckedTxEvent represents misc event sent by local modules
type UncheckedTxEvent struct {
	EventType int
	Event     any
}

type TxWithResp struct {
	Tx     *types.Transaction
	RespCh chan *TxResp
}

type TxResp struct {
	Status   bool
	ErrorMsg string
}

type CommitEvent struct {
	Block                  *types.Block
	StateUpdatedCheckpoint *consensus.Checkpoint
}

type TxSimpleInfo struct {
	Hash        string
	Nonce       uint64
	Size        int
	Local       bool
	LifeTime    int64
	ArrivedTime int64
}

type TxInfo struct {
	Tx          *types.Transaction
	Local       bool
	LifeTime    int64
	ArrivedTime int64
}

type AccountMeta struct {
	CommitNonce  uint64
	PendingNonce uint64
	TxCount      uint64
	Txs          []*TxInfo
	SimpleTxs    []*TxSimpleInfo
}

func AccountMetaFromTxpool(res *txpool.AccountMeta[types.Transaction, *types.Transaction]) *AccountMeta {
	if res == nil {
		return nil
	}

	return &AccountMeta{
		CommitNonce:  res.CommitNonce,
		PendingNonce: res.PendingNonce,
		TxCount:      res.TxCount,
		Txs: lo.Map(res.Txs, func(tx *txpool.TxInfo[types.Transaction, *types.Transaction], index int) *TxInfo {
			return &TxInfo{
				Tx:          tx.Tx,
				Local:       tx.Local,
				LifeTime:    tx.LifeTime,
				ArrivedTime: tx.ArrivedTime,
			}
		}),
		SimpleTxs: lo.Map(res.SimpleTxs, func(tx *txpool.TxSimpleInfo, index int) *TxSimpleInfo {
			return &TxSimpleInfo{
				Hash:        tx.Hash,
				Nonce:       tx.Nonce,
				Size:        tx.Size,
				Local:       tx.Local,
				LifeTime:    tx.LifeTime,
				ArrivedTime: tx.ArrivedTime,
			}
		}),
	}
}

type BatchSimpleInfo struct {
	TxCount   uint64
	Txs       []*TxSimpleInfo
	Timestamp int64
}

type Meta struct {
	TxCountLimit    uint64
	TxCount         uint64
	ReadyTxCount    uint64
	Batches         map[string]*BatchSimpleInfo
	MissingBatchTxs map[string]map[uint64]string
	Accounts        map[string]*AccountMeta
}

func MetaFromTxpool(res *txpool.Meta[types.Transaction, *types.Transaction]) *Meta {
	if res == nil {
		return nil
	}

	return &Meta{
		TxCountLimit: res.TxCountLimit,
		TxCount:      res.TxCount,
		ReadyTxCount: res.ReadyTxCount,
		Batches: lo.MapEntries(res.Batches, func(key string, value *txpool.BatchSimpleInfo) (string, *BatchSimpleInfo) {
			return key, &BatchSimpleInfo{
				TxCount: value.TxCount,
				Txs: lo.Map(value.Txs, func(tx *txpool.TxSimpleInfo, index int) *TxSimpleInfo {
					return &TxSimpleInfo{
						Hash:        tx.Hash,
						Nonce:       tx.Nonce,
						Size:        tx.Size,
						Local:       tx.Local,
						LifeTime:    tx.LifeTime,
						ArrivedTime: tx.ArrivedTime,
					}
				}),
				Timestamp: value.Timestamp,
			}
		}),
		MissingBatchTxs: res.MissingBatchTxs,
		Accounts: lo.MapEntries(res.Accounts, func(key string, value *txpool.AccountMeta[types.Transaction, *types.Transaction]) (string, *AccountMeta) {
			return key, AccountMetaFromTxpool(value)
		}),
	}
}
