package adaptor

import (
	"fmt"
	"math"

	"github.com/axiomesh/axiom-ledger/pkg/crypto"
	"github.com/bcds/go-hpc-dagbft/common/errors"
	"github.com/bcds/go-hpc-dagbft/common/types"
	"github.com/bcds/go-hpc-dagbft/common/utils/concurrency"
	"github.com/bcds/go-hpc-dagbft/protocol"
	"github.com/sirupsen/logrus"
)

type ValidatorVerify struct {
	useBls        bool
	nodes         protocol.Validators
	totalPower    protocol.VotePower
	quorumPower   protocol.VotePower
	validityPower protocol.VotePower
	logger        logrus.FieldLogger
}

func newValidatorVerify(validators []*protocol.Validator, useBls bool, logger logrus.FieldLogger) *ValidatorVerify {
	totalPower := protocol.VotePower(0)
	for _, validator := range validators {
		totalPower += validator.VotePower
	}
	faultPower := (totalPower - 1) / 3
	validityPower := faultPower + 1
	quorumPower := protocol.VotePower(math.Ceil(float64(totalPower+faultPower+1) / float64(2)))

	return &ValidatorVerify{
		useBls:        useBls,
		nodes:         validators,
		totalPower:    totalPower,
		quorumPower:   quorumPower,
		validityPower: validityPower,
		logger:        logger,
	}
}

func id2index(id types.ValidatorID) uint32 {
	return id - 1
}

func (m *ValidatorVerify) GetValidators() []*protocol.Validator { return m.nodes }

func (m *ValidatorVerify) TotalPower() protocol.VotePower { return m.totalPower }

func (m *ValidatorVerify) QuorumPower() protocol.VotePower { return m.quorumPower }

func (m *ValidatorVerify) ValidityPower() protocol.VotePower { return m.validityPower }

func (m *ValidatorVerify) Verify(message []byte, signature protocol.Signature) error {
	var err error
	defer func() {
		if err != nil {
			m.logger.Errorf("failed to verify signature: %w", err)
		}
	}()
	index := id2index(signature.Signer)
	if index >= uint32(len(m.nodes)) {
		return fmt.Errorf("signer %d is not in validators", signature.Signer)
	}
	node := m.nodes[index]
	if signature.Signer != node.ValidatorId {
		return fmt.Errorf("invalid signer，expected %d, got %d, nodes: %v", node.ValidatorId, signature.Signer, m.nodes)
	}

	if err := m.verifySignature(node.PubKey, message, signature); err != nil {
		return fmt.Errorf("failed to verify signature by signer %d, err: %w", signature.Signer, err)
	}
	return nil
}

func (m *ValidatorVerify) verifySignature(PubKeyBytes, message []byte, signature protocol.Signature) error {
	var valid bool
	if m.useBls {
		pubKey := &crypto.Bls12381PublicKey{}
		err := pubKey.Unmarshal(PubKeyBytes)
		if err != nil {
			return err
		}
		valid = pubKey.Verify(message, signature.Signature)
	} else {
		pubKey := &crypto.Ed25519PublicKey{}
		err := pubKey.Unmarshal(PubKeyBytes)
		if err != nil {
			return err
		}
		valid = pubKey.Verify(message, signature.Signature)
	}
	if !valid {
		return fmt.Errorf("invalid signature by signer %d", signature.Signer)
	}
	return nil
}

func (m *ValidatorVerify) QuorumVerify(message []byte, signatures protocol.Signatures) error {
	switch signature := signatures.(type) {
	case types.MultiSignature:
		var weights protocol.VotePower
		verifies := make([]func() error, 0, signature.Len())
		votes := make(map[uint32]bool, signature.Len())
		signature.Range(func(sig protocol.Signature, verified bool) {
			weights += m.nodes[id2index(sig.Signer)].VotePower
			votes[sig.Signer] = true
			if !verified {
				verifies = append(verifies, func() error { return m.Verify(message, sig) })
			}
		})
		if len(votes) < signature.Len() {
			return errors.New("duplicate signer")
		}
		if weights < m.quorumPower {
			return fmt.Errorf("insufficient voting power %d for quorum %d", weights, m.quorumPower)
		}
		return concurrency.Parallel(verifies...)
	default:
		return fmt.Errorf("[DEMO] unsupported signature type %T", signature)
	}
}
