package precheck

import (
	"context"
	"fmt"
	"math/big"
	"runtime"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/params"
	"github.com/gammazero/workerpool"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	vm "github.com/axiomesh/eth-kit/evm"
)

const (
	defaultTxPreCheckSize = 10000
	ErrTxSign             = "tx signature verify failed"
	ErrTo                 = "tx from and to address is same"
	ErrTxEventType        = "invalid tx event type"
	ErrParseTxEventType   = "parse tx event type error"
)

var concurrencyLimit = runtime.NumCPU()

type ValidTxs struct {
	Local       bool
	Txs         []*types.Transaction
	LocalRespCh chan *common.TxResp
}

type TxPreCheckMgr struct {
	uncheckedCh  chan *common.UncheckedTxEvent
	verifyDataCh chan *common.UncheckedTxEvent
	validTxsCh   chan *ValidTxs
	logger       logrus.FieldLogger

	BaseFee      *big.Int // current is 0
	getBalanceFn func(address *types.Address) *big.Int

	ctx       context.Context
	evmConfig repo.EVM
}

func (tp *TxPreCheckMgr) PostUncheckedTxEvent(ev *common.UncheckedTxEvent) {
	tp.uncheckedCh <- ev
}

func (tp *TxPreCheckMgr) CommitValidTxs() chan *ValidTxs {
	return tp.validTxsCh
}

func NewTxPreCheckMgr(ctx context.Context, evmConfig repo.EVM, logger logrus.FieldLogger, fn func(address *types.Address) *big.Int) *TxPreCheckMgr {
	return &TxPreCheckMgr{
		uncheckedCh:  make(chan *common.UncheckedTxEvent, defaultTxPreCheckSize),
		verifyDataCh: make(chan *common.UncheckedTxEvent, defaultTxPreCheckSize),
		validTxsCh:   make(chan *ValidTxs, defaultTxPreCheckSize),
		logger:       logger,
		ctx:          ctx,
		BaseFee:      big.NewInt(0),
		getBalanceFn: fn,
		evmConfig:    evmConfig,
	}
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
		case ev := <-tp.uncheckedCh:
			wp.Submit(func() {
				switch ev.EventType {
				case common.LocalTxEvent:
					txWithResp, ok := ev.Event.(*common.TxWithResp)
					if !ok {
						tp.logger.Errorf("%s:%s", ErrParseTxEventType, "receive invalid local TxEvent")
						return
					}
					if err := tp.verifySignature(txWithResp.Tx); err != nil {
						txWithResp.RespCh <- &common.TxResp{
							Status:   false,
							ErrorMsg: err.Error(),
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
						if err := tp.verifySignature(tx); err != nil {
							tp.logger.Warningf("verify remote tx err:%s", err)
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
				switch ev.EventType {
				case common.LocalTxEvent:
					txWithResp := ev.Event.(*common.TxWithResp)
					if err := tp.verifyData(txWithResp.Tx); err != nil {
						txWithResp.RespCh <- &common.TxResp{
							Status:   false,
							ErrorMsg: err.Error(),
						}
					} else {
						tp.validTxsCh <- &ValidTxs{
							Local:       true,
							Txs:         []*types.Transaction{txWithResp.Tx},
							LocalRespCh: txWithResp.RespCh,
						}
					}

				case common.RemoteTxEvent:
					txSet, ok := ev.Event.([]*types.Transaction)
					if !ok {
						tp.logger.Errorf("receive invalid remote TxEvent")
						return
					}
					validDataTxs := make([]*types.Transaction, 0)
					for _, tx := range txSet {
						if err := tp.verifyData(tx); err != nil {
							tp.logger.Errorf("verify remote tx data failed: %v", err)
						} else {
							validDataTxs = append(validDataTxs, tx)
						}
					}
					tp.validTxsCh <- &ValidTxs{
						Txs:   validDataTxs,
						Local: false,
					}
				}
			})
		}
	}
}

func (tp *TxPreCheckMgr) verifySignature(tx *types.Transaction) error {
	if err := tx.VerifySignature(); err != nil {
		return fmt.Errorf("verify tx err: %s", ErrTxSign)
	}

	if tx.GetTo() != nil {
		if tx.GetFrom().String() == tx.GetTo().String() {
			err := fmt.Errorf("verify tx err: %s", ErrTo)
			tp.logger.Errorf(err.Error())
			return err
		}
	}
	return nil
}

func (tp *TxPreCheckMgr) verifyData(tx *types.Transaction) error {
	// 1. the gas parameters's format are valid
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

	// 2. account has enough balance to cover transaction fee(gaslimit * gasprice)
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

	// 3.1 the purchased gas is enough to cover intrinsic usage
	// 3.2 there is no overflow when calculating intrinsic gas
	gas, err := vm.IntrinsicGas(tx.GetPayload(), tx.GetInner().GetAccessList(), isContractCreation, true, true, true)
	if err != nil {
		return err
	}
	if gasRemaining < gas {
		return fmt.Errorf("%w: have %d, want %d", core.ErrIntrinsicGas, gasRemaining, gas)
	}

	// 4. account has enough balance to cover asset transfer for **topmost** call
	if tx.GetValue().Sign() > 0 && balanceRemaining.Cmp(tx.GetValue()) < 0 {
		return fmt.Errorf("%w: address %v", core.ErrInsufficientFundsForTransfer, tx.GetFrom())
	}

	// 5. if deployed a contract, Check whether the init code size has been exceeded.
	if isContractCreation && len(tx.GetPayload()) > params.MaxInitCodeSize && !tp.evmConfig.DisableMaxCodeSizeLimit {
		return fmt.Errorf("%w: code size %v limit %v", core.ErrMaxInitCodeSizeExceeded, len(tx.GetPayload()), params.MaxInitCodeSize)
	}

	return nil
}
