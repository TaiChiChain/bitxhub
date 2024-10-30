package data_syncer

import (
	"fmt"
	"sync"

	"github.com/axiomesh/axiom-ledger/internal/components"
	"github.com/axiomesh/axiom-ledger/internal/consensus/dagbft/common"
	"github.com/axiomesh/axiom-ledger/internal/consensus/types"
	"github.com/bcds/go-hpc-dagbft/common/utils/channel"
	"github.com/bcds/go-hpc-dagbft/common/utils/concurrency"
	"github.com/bcds/go-hpc-dagbft/protocol"
	"github.com/sirupsen/logrus"
)

type verifier struct {
	cryptoVerifier protocol.ValidatorVerifier
	wg             sync.WaitGroup
	logger         logrus.FieldLogger
	taskNum        int
	errChPool      *channel.Pool[error]
	closeC         chan struct{}
}

func newVerifier(logger logrus.FieldLogger, verify protocol.ValidatorVerifier, closeC chan struct{}) *verifier {
	taskNum := 3
	return &verifier{
		cryptoVerifier: verify,
		logger:         logger,
		taskNum:        taskNum,
		errChPool:      channel.NewPool[error](taskNum),
		closeC:         closeC,
	}
}

func (v *verifier) verifyAttestation(info *types.Attestation) error {
	proof := info.Proof
	cp, err := common.DecodeProof(proof)
	if err != nil {
		return fmt.Errorf("failed to decode proof: %w", err)
	}

	verifies := make([]func() error, 0, v.taskNum)

	// 1. verify block header
	verifyHeader := func() error {
		quorumHash := cp.GetCheckpoint().GetExecuteState().GetStateRoot()
		recvBlockHash := info.Block.Hash().String()
		if quorumHash != recvBlockHash {
			return fmt.Errorf("quorum hash mismatch: quorum:%s != recv blcok:%s", quorumHash, recvBlockHash)
		}
		return nil
	}
	verifies = append(verifies, verifyHeader)

	// 2. verify block body
	verifyBody := func() error {
		txs := info.Block.Transactions
		txRoot := info.Block.Header.TxRoot.String()

		// validate txRoot
		calcTxRoot, err := components.CalcTxsMerkleRoot(txs)
		if err != nil {
			return fmt.Errorf("failed to calculate txs merkle root: %w", err)
		}
		if calcTxRoot.String() != txRoot {
			return fmt.Errorf("invalid txs root,caculate txRoot is %s, but remote block txRoot is %s", calcTxRoot, txRoot)
		}
		return nil
	}
	verifies = append(verifies, verifyBody)

	// 3. verify quorumCheckpoint signature
	verifies = append(verifies, func() error {
		return v.verifyProof(cp)
	})

	return concurrency.Parallel(verifies...)
}

func (v *verifier) verifyProof(proof *types.DagbftQuorumCheckpoint) error {
	return v.cryptoVerifier.QuorumVerify(proof.Checkpoint().Digest().Bytes(), proof.Signatures())
}
