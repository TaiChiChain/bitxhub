package base

import (
	"encoding/binary"
	"fmt"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	vm "github.com/axiomesh/eth-kit/evm"
)

const (
	nodeIDGeneratorKey        = "nodeIDGeneratorKey"
	nextEpochInfoKey          = "nextEpochInfoKey"
	historyEpochInfoKeyPrefix = "historyEpochInfoKeyPrefix"
)

var _ common.SystemContract = (*EpochManager)(nil)

type EpochManager struct {
	logger  logrus.FieldLogger
	account ledger.IAccount
}

func NewEpochManager(cfg *common.SystemContractConfig) *EpochManager {
	return &EpochManager{
		logger: cfg.Logger,
	}
}

func (m *EpochManager) Reset(lastHeight uint64, stateLedger ledger.StateLedger) {
	m.account = stateLedger.GetOrCreateAccount(types.NewAddressByStr(common.EpochManagerContractAddr))
}

func (m *EpochManager) Run(msg *vm.Message) (*vm.ExecutionResult, error) {
	// TODO: add query method
	return nil, errors.New("unsupported method")
}

func (m *EpochManager) EstimateGas(callArgs *types.CallArgs) (uint64, error) {
	return 0, errors.New("unsupported method")
}

func historyEpochInfoKey(epoch uint64) []byte {
	return []byte(fmt.Sprintf("%s_%d", historyEpochInfoKeyPrefix, epoch))
}

func InitEpochInfo(lg ledger.StateLedger, epochInfo *rbft.EpochInfo) error {
	account := lg.GetOrCreateAccount(types.NewAddressByStr(common.EpochManagerContractAddr))
	epochInfo = epochInfo.Clone()

	maxNodeID := uint64(0)
	nodeIDMap := make(map[uint64]struct{})
	nodeAccountAddrMap := make(map[string]struct{})
	nodeP2PIDMap := make(map[string]struct{})
	checkNodes := func(nodes []*rbft.NodeInfo) error {
		for _, n := range nodes {
			if err := checkNodeInfo(n); err != nil {
				return err
			}

			if _, ok := nodeIDMap[n.ID]; ok {
				return errors.Errorf("duplicate node id: %d", n.ID)
			}
			if _, ok := nodeAccountAddrMap[n.AccountAddress]; ok {
				return errors.Errorf("duplicate node account addr: %s", n.AccountAddress)
			}
			if _, ok := nodeP2PIDMap[n.P2PNodeID]; ok {
				return errors.Errorf("duplicate p2p node id: %s", n.P2PNodeID)
			}
			nodeIDMap[n.ID] = struct{}{}
			nodeAccountAddrMap[n.AccountAddress] = struct{}{}
			nodeP2PIDMap[n.P2PNodeID] = struct{}{}
			if n.ID > maxNodeID {
				maxNodeID = n.ID
			}
		}
		return nil
	}
	if err := checkNodes(epochInfo.ValidatorSet); err != nil {
		return err
	}
	if err := checkNodes(epochInfo.CandidateSet); err != nil {
		return err
	}
	if err := checkNodes(epochInfo.DataSyncerSet); err != nil {
		return err
	}

	setNodeIDGenerator(lg, maxNodeID+1)

	c, err := epochInfo.Marshal()
	if err != nil {
		return err
	}
	account.SetState(historyEpochInfoKey(epochInfo.Epoch), c)

	epochInfo.Epoch++
	epochInfo.StartBlock += epochInfo.EpochPeriod
	c, err = epochInfo.Marshal()
	if err != nil {
		return err
	}
	// set history state
	account.SetState([]byte(nextEpochInfoKey), c)
	return nil
}

func setNodeIDGenerator(lg ledger.StateLedger, newNodeIDGenerator uint64) {
	var nodeIDGenerator []byte
	nodeIDGenerator = binary.BigEndian.AppendUint64(nodeIDGenerator, newNodeIDGenerator)
	account := lg.GetOrCreateAccount(types.NewAddressByStr(common.EpochManagerContractAddr))
	account.SetState([]byte(nodeIDGeneratorKey), nodeIDGenerator)
}

func getNodeIDGenerator(lg ledger.StateLedger) (uint64, error) {
	account := lg.GetOrCreateAccount(types.NewAddressByStr(common.EpochManagerContractAddr))
	ok, nodeIDGeneratorBytes := account.GetState([]byte(nodeIDGeneratorKey))
	if !ok {
		return 0, errors.New("not found node id generator")
	}
	nodeIDGenerator := binary.BigEndian.Uint64(nodeIDGeneratorBytes)
	return nodeIDGenerator, nil
}

func getEpoch(lg ledger.StateLedger, key []byte) (*rbft.EpochInfo, error) {
	account := lg.GetOrCreateAccount(types.NewAddressByStr(common.EpochManagerContractAddr))
	success, data := account.GetState(key)
	if success {
		e := &rbft.EpochInfo{}
		if err := e.Unmarshal(data); err != nil {
			return nil, err
		}
		return e, nil
	}
	return nil, errors.New("not found epoch info")
}

func GetNextEpochInfo(lg ledger.StateLedger) (*rbft.EpochInfo, error) {
	return getEpoch(lg, []byte(nextEpochInfoKey))
}

func setNextEpochInfo(lg ledger.StateLedger, n *rbft.EpochInfo) error {
	c, err := n.Marshal()
	if err != nil {
		return err
	}
	account := lg.GetOrCreateAccount(types.NewAddressByStr(common.EpochManagerContractAddr))
	// set  epoch info
	account.SetState([]byte(nextEpochInfoKey), c)
	return nil
}

func setEpochInfo(lg ledger.StateLedger, n *rbft.EpochInfo) error {
	c, err := n.Marshal()
	if err != nil {
		return err
	}
	account := lg.GetOrCreateAccount(types.NewAddressByStr(common.EpochManagerContractAddr))
	// set epoch info
	account.SetState(historyEpochInfoKey(n.Epoch), c)
	return nil
}

func GetEpochInfo(lg ledger.StateLedger, epoch uint64) (*rbft.EpochInfo, error) {
	return getEpoch(lg, historyEpochInfoKey(epoch))
}

func GetCurrentEpochInfo(lg ledger.StateLedger) (*rbft.EpochInfo, error) {
	next, err := GetNextEpochInfo(lg)
	if err != nil {
		return nil, err
	}
	return getEpoch(lg, historyEpochInfoKey(next.Epoch-1))
}

// TurnIntoNewEpoch when execute epoch last, return new current epoch info
func TurnIntoNewEpoch(electValidatorsByWrfSeed []byte, lg ledger.StateLedger) (*rbft.EpochInfo, error) {
	account := lg.GetOrCreateAccount(types.NewAddressByStr(common.EpochManagerContractAddr))
	success, data := account.GetState([]byte(nextEpochInfoKey))
	if success {
		e := &rbft.EpochInfo{}
		if err := e.Unmarshal(data); err != nil {
			return nil, err
		}
		if err := e.ElectValidators(electValidatorsByWrfSeed); err != nil {
			return nil, err
		}
		validatorIDs := lo.Map(e.ValidatorSet, func(item *rbft.NodeInfo, index int) uint64 {
			return item.ID
		})
		loggers.Logger(loggers.Epoch).Infof("Elect new Validators: %v", validatorIDs)

		if err := setEpochInfo(lg, e); err != nil {
			return nil, err
		}

		n := e.Clone()
		n.Epoch++
		n.StartBlock += n.EpochPeriod
		if err := setNextEpochInfo(lg, n); err != nil {
			return nil, err
		}

		// return current
		return e, nil
	}
	return nil, errors.New("not found current epoch info")
}

func checkNodeInfo(node *rbft.NodeInfo) error {
	if !ethcommon.IsHexAddress(node.AccountAddress) {
		return errors.Errorf("invalid account address: %s", node.AccountAddress)
	}
	if _, err := peer.Decode(node.P2PNodeID); err != nil {
		return errors.Errorf("invalid p2p node id: %s", node.P2PNodeID)
	}

	return nil
}

// AddNode adds a new node to the ledger.
// It takes a StateLedger instance and a pointer to a NodeInfo struct as parameters.
// It returns an error if any validation fails.
func AddNode(lg ledger.StateLedger, newNode *rbft.NodeInfo) error {
	// Clone the newNode to avoid modifying the original instance
	newNode = newNode.Clone()

	// Check if the new node info is valid
	if err := checkNodeInfo(newNode); err != nil {
		return err
	}

	// Get the next epoch information from the ledger
	nextEpochInfo, err := GetNextEpochInfo(lg)
	if err != nil {
		return err
	}

	// Function to check for duplicate node information
	checkNodeInfoDuplicate := func(nodes []*rbft.NodeInfo) error {
		for _, n := range nodes {
			if n.ID == newNode.ID {
				return errors.Errorf("duplicate node id: %d", n.ID)
			}
			if n.AccountAddress == newNode.AccountAddress {
				return errors.Errorf("duplicate node account addr: %s", n.AccountAddress)
			}
			if n.P2PNodeID == newNode.P2PNodeID {
				return errors.Errorf("duplicate p2p node id: %s", n.P2PNodeID)
			}
		}
		return nil
	}

	// Check for duplicate node info in the validator set, candidate set, and data syncer set
	if err := checkNodeInfoDuplicate(nextEpochInfo.ValidatorSet); err != nil {
		return err
	}
	if err := checkNodeInfoDuplicate(nextEpochInfo.CandidateSet); err != nil {
		return err
	}
	if err := checkNodeInfoDuplicate(nextEpochInfo.DataSyncerSet); err != nil {
		return err
	}

	// Get the node ID generator from the ledger
	nodeIDGenerator, err := getNodeIDGenerator(lg)
	if err != nil {
		return err
	}

	// Automatically assign a self-increasing ID to the new node
	newNode.ID = nodeIDGenerator

	// Update the ID generator
	setNodeIDGenerator(lg, nodeIDGenerator+1)

	// Update the next epoch info by adding the new node to the candidate set
	nextEpochInfo.CandidateSet = append(nextEpochInfo.CandidateSet, newNode)

	// Set the updated next epoch info in the ledger
	if err := setNextEpochInfo(lg, nextEpochInfo); err != nil {
		return err
	}

	return nil
}

// RemoveNode removes a node from the validator set, candidate set, or data syncer set.
// It takes a StateLedger and the ID of the node to remove as input.
// It returns an error if the node ID is not found in any of the sets.
func RemoveNode(lg ledger.StateLedger, removeNodeID uint64) error {
	// Get the next epoch info from the ledger
	nextEpochInfo, err := GetNextEpochInfo(lg)
	if err != nil {
		return err
	}

	// Function to remove a node from a set
	removeNode := func(nodes []*rbft.NodeInfo) (bool, []*rbft.NodeInfo) {
		var matchedIdx int
		var matched bool

		// Find the index of the node with the given ID in the set
		for idx, n := range nodes {
			if n.ID == removeNodeID {
				matched = true
				matchedIdx = idx
				break
			}
		}

		// If the node is found, remove it from the set
		if matched {
			return true, append(nodes[:matchedIdx], nodes[matchedIdx+1:]...)
		}

		return false, nodes
	}

	var removed bool

	// Remove the node from the validator set
	removed, nextEpochInfo.ValidatorSet = removeNode(nextEpochInfo.ValidatorSet)
	if removed {
		return setNextEpochInfo(lg, nextEpochInfo)
	}

	// Remove the node from the candidate set
	removed, nextEpochInfo.CandidateSet = removeNode(nextEpochInfo.CandidateSet)
	if removed {
		return setNextEpochInfo(lg, nextEpochInfo)
	}

	// Remove the node from the data syncer set
	removed, nextEpochInfo.DataSyncerSet = removeNode(nextEpochInfo.DataSyncerSet)
	if removed {
		return setNextEpochInfo(lg, nextEpochInfo)
	}

	// If the node is not found in any of the sets, return an error
	return errors.Errorf("failed to remove node, node id not found: %d", removeNodeID)
}
