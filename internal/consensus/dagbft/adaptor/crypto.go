package adaptor

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/chainstate"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common/metrics"
	"github.com/axiomesh/axiom-ledger/internal/consensus/precheck"
	consensustypes "github.com/axiomesh/axiom-ledger/internal/consensus/types"
	"github.com/axiomesh/axiom-ledger/pkg/crypto"
	"github.com/bcds/go-hpc-dagbft/common/types/protos"
	"github.com/bcds/go-hpc-dagbft/common/utils/channel"
	"github.com/bcds/go-hpc-dagbft/protocol"
	"github.com/bcds/go-hpc-dagbft/protocol/layer"
	"github.com/gammazero/workerpool"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

var _ layer.Crypto = (*CryptoImpl)(nil)

var _ protocol.ValidatorSigner = (*ValidatorSigner)(nil)

type CryptoImpl struct {
	useBls            bool
	signer            *ValidatorSigner
	validatorVerifier *ValidatorVerify
	batchVerifier     *BatchVerifier
	logger            logrus.FieldLogger
}

func NewCryptoImpl(conf *common.Config, precheck precheck.PreCheck) (*CryptoImpl, error) {
	logger := conf.Logger
	var (
		useBls bool
		priv   crypto.PrivateKey
	)
	if conf.Repo.Config.Consensus.UseBlsKey {
		useBls = true
		priv = conf.Repo.ConsensusKeystore.PrivateKey
	} else {
		useBls = false
		priv = conf.Repo.P2PKeystore.PrivateKey
	}
	var validators []*protos.Validator

	var innerErr error
	validators = lo.Map(conf.ChainState.ValidatorSet, func(item chainstate.ValidatorInfo, index int) *protos.Validator {
		var (
			pubBytes []byte
		)
		nodeInfo, err := conf.ChainState.GetNodeInfo(item.ID)
		if err != nil {
			innerErr = fmt.Errorf("failed to get node info: %w", err)
		}

		if useBls {
			pubBytes, err = nodeInfo.ConsensusPubKey.Marshal()
		} else {
			pubBytes, err = nodeInfo.P2PPubKey.Marshal()
		}
		if err != nil {
			innerErr = fmt.Errorf("failed to marshal public key: %w", err)
		}
		return &protos.Validator{
			Hostname:    nodeInfo.Primary,
			PubKey:      pubBytes,
			ValidatorId: uint32(item.ID),
			VotePower:   uint64(item.ConsensusVotingPower),
			Workers:     nodeInfo.Workers,
		}
	})

	return &CryptoImpl{
		useBls:            useBls,
		signer:            newValidatorSigner(priv, logger),
		validatorVerifier: newValidatorVerify(validators, useBls, logger),
		batchVerifier:     newBatchVerifier(precheck),
		logger:            logger,
	}, innerErr
}

type ValidatorSigner struct {
	priKey crypto.PrivateKey
	logger logrus.FieldLogger
}

func newValidatorSigner(priKey crypto.PrivateKey, logger logrus.FieldLogger) *ValidatorSigner {
	return &ValidatorSigner{
		priKey: priKey,
		logger: logger,
	}
}

func (v *ValidatorSigner) Sign(msg []byte) []byte {
	result, err := v.priKey.Sign(msg)
	if err != nil {
		v.logger.Error("failed to sign message", "err", err)
	}
	return result
}

var concurrencyLimit = runtime.NumCPU()

type BatchVerifier struct {
	precheck precheck.PreCheck
	pool     *workerpool.WorkerPool
}

func newBatchVerifier(check precheck.PreCheck) *BatchVerifier {
	return &BatchVerifier{
		precheck: check,
		pool:     workerpool.New(concurrencyLimit),
	}
}

func (b *BatchVerifier) VerifyTransactions(batchTxs []protocol.Transaction) error {
	start := time.Now()
	errCh := make(chan error, len(batchTxs))
	closeCh := make(chan struct{})
	verified := atomic.Uint64{}
	lo.ForEach(batchTxs, func(raw protocol.Transaction, index int) {
		b.pool.Submit(func() {
			tx := &types.Transaction{}
			err := tx.Unmarshal(raw)
			if err != nil {
				channel.SafeSend(errCh, fmt.Errorf("failed to unmarshal transaction: %w", err), closeCh)
			}

			err = b.precheck.BasicCheckTx(tx)
			if err != nil {
				channel.SafeSend(errCh, fmt.Errorf("failed to basic check transaction: %w", err), closeCh)
			}

			err = b.precheck.VerifySignature(tx)
			if err != nil {
				channel.SafeSend(errCh, fmt.Errorf("failed to verify signature: %w", err), closeCh)
			}

			verified.Add(1)
			if verified.Load() == uint64(len(batchTxs)) {
				channel.SafeClose(closeCh)
			}
		})
	})

	select {
	case err := <-errCh:
		channel.SafeClose(closeCh)
		b.pool.StopWait()
		return err
	case <-closeCh:
		metrics.BatchVerifyTime.WithLabelValues(consensustypes.Dagbft).Observe(time.Since(start).Seconds())
		return nil
	}
}

func (c *CryptoImpl) GetValidatorSigner() protocol.ValidatorSigner {
	return c.signer
}

func (c *CryptoImpl) GetValidatorVerifier() protocol.ValidatorVerifier {
	return c.validatorVerifier
}

func (c *CryptoImpl) GetVerifierByValidators(validators protocol.Validators) (protocol.ValidatorVerifier, error) {
	return newValidatorVerify(validators, c.useBls, c.logger), nil
}

func (c *CryptoImpl) GetBatchVerifier() protocol.BatchVerifier {
	return c.batchVerifier
}
