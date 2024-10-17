package data_syncer

import (
	"fmt"
	"sync"

	"github.com/axiomesh/axiom-ledger/internal/components"
	"github.com/axiomesh/axiom-ledger/internal/consensus/dagbft/adaptor"
	"github.com/axiomesh/axiom-ledger/internal/consensus/types"
	"github.com/bcds/go-hpc-dagbft/common/utils/channel"
	"github.com/sirupsen/logrus"
)

type verifier struct {
	*adaptor.ValidatorVerify
	wg      sync.WaitGroup
	logger  logrus.FieldLogger
	taskNum int
	errCh   chan error
	closeC  chan struct{}
}

func newVerifier(logger logrus.FieldLogger, verify *adaptor.ValidatorVerify, closeC chan struct{}) *verifier {
	taskNum := 3
	return &verifier{
		ValidatorVerify: verify,
		logger:          logger,
		taskNum:         taskNum,
		errCh:           make(chan error, taskNum),
		closeC:          closeC,
	}
}

func (v *verifier) verifyAttestation(info *types.AttestationAndBlock) error {
	cp, ok := info.AttestationData.(*types.DagbftQuorumCheckpoint)
	if !ok {
		return fmt.Errorf("unexpected checkpoint type: %T", info.AttestationData)
	}

	v.wg.Add(v.taskNum)
	// 1. verify block header
	go func(errCh chan error) {
		defer v.wg.Done()
		quorumHash := cp.GetCheckpoint().GetExecuteState().GetStateRoot()
		recvBlockHash := info.Block.Hash().String()
		if quorumHash != recvBlockHash {
			channel.SafeSend(errCh, fmt.Errorf("quorum hash mismatch: quorum:%s != recv blcok:%s", quorumHash, recvBlockHash), v.closeC)
		}
	}(v.errCh)

	// 2. verify block body
	go func(errCh chan error) {
		defer v.wg.Done()
		txs := info.Block.Transactions
		txRoot := info.Block.Header.TxRoot.String()

		// validate txRoot
		calcTxRoot, err := components.CalcTxsMerkleRoot(txs)
		if err != nil {
			channel.SafeSend(errCh, fmt.Errorf("failed to calculate txs merkle root: %w", err), v.closeC)
			return
		}
		if calcTxRoot.String() != txRoot {
			channel.SafeSend(errCh, fmt.Errorf("invalid txs root, "+
				"caculate txRoot is %s, but remote block txRoot is %s", calcTxRoot, txRoot), v.closeC)
		}
	}(v.errCh)

	// 3. verify attestation
	go func(errCh chan error) {
		defer v.wg.Done()
		err := v.ValidatorVerify.QuorumVerify(cp.Checkpoint().Digest().Bytes(), cp.Signatures())
		if err != nil {
			channel.SafeSend(errCh, err, v.closeC)
		}
	}(v.errCh)

	v.wg.Wait()

	select {
	case err := <-v.errCh:
		return err
	default:
		return nil
	}
}
