package precheck

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/txpool"
	"github.com/ethereum/go-ethereum/params"
	"github.com/gammazero/workerpool"
	"github.com/sirupsen/logrus"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	vm "github.com/axiomesh/eth-kit/evm"
)

var _ PreCheck = (*TxPreCheckMgr)(nil)

const (
	defaultTxPreCheckSize = 10000
	PrecheckError         = "verify tx err"
	ErrTxSign             = "tx signature verify failed"
	ErrTo                 = "tx from and to address is same"
	ErrTxEventType        = "invalid tx event type"
	ErrParseTxEventType   = "parse tx event type error"
	ErrGasPriceTooLow     = "gas price too low"
)

var concurrencyLimit = runtime.NumCPU()

type ValidTxs struct {
	Local       bool
	Txs         []*types.Transaction
	LocalRespCh chan *common.TxResp
}

type TxPreCheckMgr struct {
	basicCheckCh chan *common.UncheckedTxEvent
	verifyDataCh chan *common.UncheckedTxEvent
	validTxsCh   chan *ValidTxs
	logger       logrus.FieldLogger

	BaseFee        *big.Int // current is 0
	getBalanceFn   func(address *types.Address) *big.Int
	getChainMetaFn func() *types.ChainMeta

	ctx       context.Context
	evmConfig repo.EVM
	txMaxSize atomic.Uint64
}

func (tp *TxPreCheckMgr) UpdateEpochInfo(epoch *rbft.EpochInfo) {
	tp.txMaxSize.Store(epoch.ConfigParams.TxMaxSize)
}

func (tp *TxPreCheckMgr) PostUncheckedTxEvent(ev *common.UncheckedTxEvent) {
	tp.basicCheckCh <- ev
}

func (tp *TxPreCheckMgr) pushValidTxs(ev *ValidTxs) {
	tp.validTxsCh <- ev
}

func (tp *TxPreCheckMgr) CommitValidTxs() chan *ValidTxs {
	return tp.validTxsCh
}

func NewTxPreCheckMgr(ctx context.Context, conf *common.Config) *TxPreCheckMgr {
	tp := &TxPreCheckMgr{
		basicCheckCh:   make(chan *common.UncheckedTxEvent, defaultTxPreCheckSize),
		verifyDataCh:   make(chan *common.UncheckedTxEvent, defaultTxPreCheckSize),
		validTxsCh:     make(chan *ValidTxs, defaultTxPreCheckSize),
		logger:         conf.Logger,
		ctx:            ctx,
		BaseFee:        big.NewInt(0),
		getBalanceFn:   conf.GetAccountBalance,
		getChainMetaFn: conf.GetChainMetaFunc,
		evmConfig:      conf.EVMConfig,
	}

	if conf.GenesisEpochInfo.ConfigParams.TxMaxSize == 0 {
		tp.txMaxSize.Store(repo.DefaultTxMaxSize)
	} else {
		tp.txMaxSize.Store(conf.GenesisEpochInfo.ConfigParams.TxMaxSize)
	}

	return tp
}

func (tp *TxPreCheckMgr) Start() {
	go tp.dispatchTxEvent()
	go tp.dispatchVerifyDataEvent()
	tp.logger.Info("tx precheck manager started")
}

func (tp *TxPreCheckMgr) dispatchTxEvent() {
	wp := workerpool.New(concurrencyLimit)

	for {
		select {
		case <-tp.ctx.Done():
			wp.StopWait()
			close(tp.verifyDataCh)
			return
		case ev := <-tp.basicCheckCh:
			wp.Submit(func() {
				switch ev.EventType {
				case common.LocalTxEvent:
					txWithResp, ok := ev.Event.(*common.TxWithResp)
					if !ok {
						tp.logger.Errorf("%s:%s", ErrParseTxEventType, "receive invalid local TxEvent")
						return
					}
					if err := tp.basicCheckTx(txWithResp.Tx); err != nil {
						txWithResp.RespCh <- &common.TxResp{
							Status:   false,
							ErrorMsg: fmt.Errorf("%s:%w", PrecheckError, err).Error(),
						}
						return
					}
					tp.verifyDataCh <- ev

				case common.RemoteTxEvent:
					txSet, ok := ev.Event.([]*types.Transaction)
					if !ok {
						tp.logger.Errorf("%s:%s", ErrParseTxEventType, "receive invalid remote TxEvent")
						return
					}
					validSignTxs := make([]*types.Transaction, 0)
					for _, tx := range txSet {
						if err := tp.basicCheckTx(tx); err != nil {
							tp.logger.Warningf("basic check remote tx err:%s", err)
							continue
						}
						validSignTxs = append(validSignTxs, tx)
					}
					ev.Event = validSignTxs
					tp.verifyDataCh <- ev
				default:
					tp.logger.Errorf(ErrTxEventType)
					return
				}
			})
		}
	}
}

func (tp *TxPreCheckMgr) dispatchVerifyDataEvent() {
	wp := workerpool.New(concurrencyLimit)
	for {
		select {
		case <-tp.ctx.Done():
			wp.StopWait()
			close(tp.validTxsCh)
			return
		case ev := <-tp.verifyDataCh:
			wp.Submit(func() {
				var (
					validDataTxs []*types.Transaction
					local        bool
					localRespCh  chan *common.TxResp
				)

				switch ev.EventType {
				case common.LocalTxEvent:
					local = true
					txWithResp := ev.Event.(*common.TxWithResp)
					localRespCh = txWithResp.RespCh
					// 1. check signature
					if err := tp.verifySignature(txWithResp.Tx); err != nil {
						txWithResp.RespCh <- &common.TxResp{
							Status:   false,
							ErrorMsg: fmt.Errorf("%s:%w", PrecheckError, err).Error(),
						}
						return
					}

					// 2. check balance
					if err := tp.verifyInsufficientBalance(txWithResp.Tx); err != nil {
						txWithResp.RespCh <- &common.TxResp{
							Status:   false,
							ErrorMsg: fmt.Errorf("%s:%w", PrecheckError, err).Error(),
						}
						return
					}

					validDataTxs = append(validDataTxs, txWithResp.Tx)

				case common.RemoteTxEvent:
					txSet, ok := ev.Event.([]*types.Transaction)
					if !ok {
						tp.logger.Errorf("receive invalid remote TxEvent")
						return
					}
					for _, tx := range txSet {
						if err := tp.verifySignature(tx); err != nil {
							tp.logger.Warningf("verify remote tx signature failed: %v", err)
							continue
						}

						if err := tp.verifyInsufficientBalance(tx); err != nil {
							tp.logger.Warningf("verify remote tx balance failed: %v", err)
							continue
						}

						validDataTxs = append(validDataTxs, tx)
					}
				}

				validTxs := &ValidTxs{
					Local: local,
					Txs:   validDataTxs,
				}
				if local {
					validTxs.LocalRespCh = localRespCh
				}

				tp.pushValidTxs(validTxs)
			})
		}
	}
}

func (tp *TxPreCheckMgr) verifyInsufficientBalance(tx *types.Transaction) error {
	// 1. account has enough balance to cover transaction fee(gaslimit * gasprice)
	mgval := new(big.Int).SetUint64(tx.GetGas())
	mgval = mgval.Mul(mgval, tx.GetGasPrice())
	balanceCheck := mgval
	if tx.GetGasFeeCap() != nil {
		balanceCheck = new(big.Int).SetUint64(tx.GetGas())
		balanceCheck = balanceCheck.Mul(balanceCheck, tx.GetGasFeeCap())
		balanceCheck.Add(balanceCheck, tx.GetValue())
	}
	balanceRemaining := new(big.Int).Set(tp.getBalanceFn(tx.GetFrom()))
	if have, want := balanceRemaining, balanceCheck; have.Cmp(want) < 0 {
		return fmt.Errorf("%w: address %v have %v want %v", core.ErrInsufficientFunds, tx.GetFrom(), have, want)
	}

	// sub gas fee temporarily
	balanceRemaining.Sub(balanceRemaining, mgval)

	gasRemaining := tx.GetGas()

	var isContractCreation bool
	if tx.GetTo() == nil {
		isContractCreation = true
	}

	// 2.1 the purchased gas is enough to cover intrinsic usage
	// 2.2 there is no overflow when calculating intrinsic gas
	gas, err := vm.IntrinsicGas(tx.GetPayload(), tx.GetInner().GetAccessList(), isContractCreation, true, true, true)
	if err != nil {
		return err
	}
	if gasRemaining < gas {
		return fmt.Errorf("%w: have %d, want %d", core.ErrIntrinsicGas, gasRemaining, gas)
	}

	// 3. account has enough balance to cover asset transfer for **topmost** call
	if tx.GetValue().Sign() > 0 && balanceRemaining.Cmp(tx.GetValue()) < 0 {
		return fmt.Errorf("%w: address %v", core.ErrInsufficientFundsForTransfer, tx.GetFrom())
	}
	return nil
}

func (tp *TxPreCheckMgr) verifySignature(tx *types.Transaction) error {
	if err := tx.VerifySignature(); err != nil {
		return errors.New(ErrTxSign)
	}

	// check to address
	if tx.GetTo() != nil {
		if tx.GetFrom().String() == tx.GetTo().String() {
			err := errors.New(ErrTo)
			tp.logger.Errorf(err.Error())
			return err
		}
	}
	return nil
}

func (tp *TxPreCheckMgr) basicCheckTx(tx *types.Transaction) error {
	// 1. reject transactions over defined size to prevent DOS attacks
	if uint64(tx.Size()) > tp.txMaxSize.Load() {
		return txpool.ErrOversizedData
	}

	gasPrice := tp.getChainMetaFn().GasPrice
	// ignore gas price if it's 0 or nil
	if tx.GetGasPrice() != nil {
		if tx.GetGasPrice().Uint64() != 0 && tx.GetGasPrice().Cmp(gasPrice) < 0 {
			return fmt.Errorf("%s: expect min gasPrice: %v, get price %v", ErrGasPriceTooLow, gasPrice, tx.GetGasPrice())
		}
	}

	// 2. check the gas parameters's format are valid
	if tx.GetType() == types.DynamicFeeTxType {
		if tx.GetGasFeeCap().BitLen() > 0 || tx.GetGasTipCap().BitLen() > 0 {
			if l := tx.GetGasFeeCap().BitLen(); l > 256 {
				return fmt.Errorf("%w: address %v, maxFeePerGas bit length: %d", core.ErrFeeCapVeryHigh,
					tx.GetFrom(), l)
			}
			if l := tx.GetGasTipCap().BitLen(); l > 256 {
				return fmt.Errorf("%w: address %v, maxPriorityFeePerGas bit length: %d", core.ErrTipVeryHigh,
					tx.GetFrom(), l)
			}

			if tx.GetGasFeeCap().Cmp(tx.GetGasTipCap()) < 0 {
				return fmt.Errorf("%w: address %v, maxPriorityFeePerGas: %s, maxFeePerGas: %s", core.ErrTipAboveFeeCap,
					tx.GetFrom(), tx.GetGasTipCap(), tx.GetGasFeeCap())
			}

			// This will panic if baseFee is nil, but basefee presence is verified
			// as part of header validation.
			// TODO: modify tp.BaseFee synchronously if baseFee changed
			if tx.GetGasFeeCap().Cmp(tp.BaseFee) < 0 {
				return fmt.Errorf("%w: address %v, maxFeePerGas: %s baseFee: %s", core.ErrFeeCapTooLow,
					tx.GetFrom(), tx.GetGasFeeCap(), tp.BaseFee)
			}
		}
	}

	var isContractCreation bool
	if tx.GetTo() == nil {
		isContractCreation = true
	}
	// 5. if deployed a contract, Check whether the init code size has been exceeded.
	if isContractCreation && len(tx.GetPayload()) > params.MaxInitCodeSize && !tp.evmConfig.DisableMaxCodeSizeLimit {
		return fmt.Errorf("%w: code size %v limit %v", core.ErrMaxInitCodeSizeExceeded, len(tx.GetPayload()), params.MaxInitCodeSize)
	}

	return nil
}
