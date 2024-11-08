package common

import (
	"fmt"

	kittypes "github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/chainstate"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common/metrics"
	"github.com/axiomesh/axiom-ledger/internal/consensus/epochmgr"
	consensustypes "github.com/axiomesh/axiom-ledger/internal/consensus/types"
	synccomm "github.com/axiomesh/axiom-ledger/internal/sync/common"
	api_events "github.com/axiomesh/axiom-ledger/pkg/events"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	"github.com/bcds/go-hpc-dagbft/common/types"
	"github.com/bcds/go-hpc-dagbft/common/types/protos"
	"github.com/ethereum/go-ethereum/event"
	"github.com/samber/lo"
)

func GetRemotePeers(epochChanges []*types.QuorumCheckpoint, chainstate *chainstate.ChainState) []*synccomm.Node {
	peersM := make(map[uint64]*synccomm.Node)

	// get the validator set of the current local epoch
	for _, validatorInfo := range chainstate.ValidatorSet {
		v, err := chainstate.GetNodeInfo(validatorInfo.ID)
		if err == nil {
			if v.NodeInfo.P2PID != chainstate.SelfNodeInfo.P2PID {
				peersM[validatorInfo.ID] = &synccomm.Node{Id: validatorInfo.ID, PeerID: v.P2PID}
			}
		}
	}

	// get the validator set of the remote latest epoch
	if len(epochChanges) != 0 {
		lo.ForEach(epochChanges[len(epochChanges)-1].Validators(), func(v *protos.Validator, _ int) {
			if _, ok := peersM[uint64(v.ValidatorId)]; !ok && uint64(v.GetValidatorId()) != chainstate.SelfNodeInfo.ID {
				info, err := chainstate.GetNodeInfo(uint64(v.ValidatorId))
				if err != nil {
					return
				}
				peersM[uint64(v.ValidatorId)] = &synccomm.Node{Id: uint64(v.ValidatorId), PeerID: info.P2PID}
			}
		})
	}
	// flatten peersM
	return lo.Values(peersM)
}

func PostAttestationEvent(checkpoint *types.QuorumCheckpoint, feed *event.Feed, getBlockFn func(height uint64) (*kittypes.Block, error), epochMgr *epochmgr.EpochManager) error {
	newBlock, err := getBlockFn(checkpoint.Height())
	if err != nil {
		return err
	}

	proof, err := encodeProof(&consensustypes.DagbftQuorumCheckpoint{QuorumCheckpoint: checkpoint})
	if err != nil {
		return err
	}
	newBlockBytes, err := newBlock.Marshal()
	if err != nil {
		return err
	}
	attestationEvent := api_events.AttestationEvent{
		AttestationData: &consensustypes.Attestation{
			Epoch:         newBlock.Header.Epoch,
			ConsensusType: repo.ConsensusTypeDagBft,
			Block:         newBlockBytes,
			Proof:         proof,
		},
	}
	feed.Send(attestationEvent)
	if err = epochMgr.StoreLatestProof(proof); err != nil {
		return err
	}
	metrics.AttestationCounter.WithLabelValues(consensustypes.Dagbft).Inc()
	return nil
}

func DecodeProof(proof []byte) (*consensustypes.DagbftQuorumCheckpoint, error) {
	ckpt := &consensustypes.DagbftQuorumCheckpoint{}
	if err := ckpt.Unmarshal(proof); err != nil {
		return nil, err
	}
	return ckpt, nil
}

func encodeProof(ckpt *consensustypes.DagbftQuorumCheckpoint) ([]byte, error) {
	data, err := ckpt.Marshal()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func GetValidators(state *chainstate.ChainState, useBls bool) ([]*protos.Validator, error) {
	var validators []*protos.Validator

	var innerErr error
	validators = lo.Map(state.ValidatorSet, func(item chainstate.ValidatorInfo, index int) *protos.Validator {
		var (
			pubBytes []byte
		)
		nodeInfo, err := state.GetNodeInfo(item.ID)
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

	if innerErr != nil {
		return nil, innerErr
	}

	return validators, nil
}
