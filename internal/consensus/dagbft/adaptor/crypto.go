package adaptor

import (
	"fmt"
	"runtime"
	"time"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/chainstate"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/consensus/precheck"
	"github.com/axiomesh/axiom-ledger/pkg/crypto"
	"github.com/bcds/go-hpc-dagbft/common/types/protos"
	"github.com/bcds/go-hpc-dagbft/common/utils/channel"
	"github.com/bcds/go-hpc-dagbft/common/utils/containers"
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
	precheck    precheck.PreCheck
	pool        *workerpool.WorkerPool
	logger      logrus.FieldLogger
	waitTimeout time.Duration
}

type EmptyBatchVerifier struct {
}

func (e EmptyBatchVerifier) VerifyTransactions(batchTxs []protocol.Transaction) error {
	return nil
}

func newBatchVerifier(check precheck.PreCheck, logger logrus.FieldLogger) *BatchVerifier {
	return &BatchVerifier{
		precheck: check,
		pool:     workerpool.New(concurrencyLimit),
		logger:   logger,
		// todo: make this configurable
		waitTimeout: 60 * time.Second,
	}
}

func (b *BatchVerifier) verifyBatchTransactions(batchTxs [][]byte, metrics *blockChainMetrics) []*types.Transaction {
	validTxs := make([]*types.Transaction, len(batchTxs))
	resultCh := make(chan containers.Tuple[*types.Transaction, int, error], len(batchTxs))
	closeCh := make(chan struct{})
	verified := 0
	lo.ForEach(batchTxs, func(raw []byte, index int) {
		b.pool.Submit(func() {
			var err error
			tx := &types.Transaction{}
			err = tx.Unmarshal(raw)
			if err != nil {
				b.logger.Errorf("failed to unmarshal transaction: %w, data: %v", err, raw)
				tx = nil
				channel.SafeSend(resultCh, containers.Pack3(tx, index, fmt.Errorf("failed to unmarshal transaction: %w", err)), closeCh)
				return
			}

			err = b.precheck.BasicCheckTx(tx)
			if err != nil {
				channel.SafeSend(resultCh, containers.Pack3(tx, index, fmt.Errorf("failed to basic check transaction: %w", err)), closeCh)
				return
			}

			err = b.precheck.VerifySignature(tx)
			if err != nil {
				channel.SafeSend(resultCh, containers.Pack3(tx, index, fmt.Errorf("failed to verify signature: %w", err)), closeCh)
				return
			}

			err = b.precheck.VerifyData(tx)
			if err != nil {
				channel.SafeSend(resultCh, containers.Pack3(tx, index, fmt.Errorf("failed to verify data: %w", err)), closeCh)
				return
			}

			// verify tx successfully
			channel.SafeSend(resultCh, containers.Pack3(tx, index, err), closeCh)
		})
	})

	for {
		select {
		case <-closeCh:
			return lo.Filter(validTxs, func(tx *types.Transaction, index int) bool {
				return tx != nil
			})
		case res := <-resultCh:
			verified++
			tx, index, err := res.Unpack()
			if err != nil {
				metrics.discardedTransactions.WithLabelValues(err.Error()).Inc()
				if tx != nil {
					b.logger.WithFields(logrus.Fields{
						"tx hash":  tx.RbftGetTxHash(),
						"tx nonce": tx.RbftGetNonce(),
						"tx from":  tx.RbftGetFrom(),
						"index":    index,
					}).Warningf("verify tx failed: %s", err)
				}
				validTxs[index] = nil
			} else {
				validTxs[index] = tx
			}

			if verified == len(batchTxs) {
				channel.SafeClose(closeCh)
			}
		case <-time.After(b.waitTimeout):
			b.logger.Errorf("verify batch txs timeout")
			channel.SafeClose(closeCh)
		}
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
	return &EmptyBatchVerifier{}
}
