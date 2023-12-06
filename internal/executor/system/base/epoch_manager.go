package base

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
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
	id                        = "ID"
	p2PNodeID                 = "P2PNodeID"
)

const epochManagerABIData = "[\n\t{\n\t\t\"inputs\": [],\n\t\t\"name\": \"currentEpoch\",\n\t\t\"outputs\": [\n\t\t\t{\n\t\t\t\t\"components\": [\n\t\t\t\t\t{\n\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\"name\": \"Version\",\n\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\"name\": \"Epoch\",\n\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\"name\": \"EpochPeriod\",\n\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\"name\": \"StartBlock\",\n\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"internalType\": \"string[]\",\n\t\t\t\t\t\t\"name\": \"P2PBootstrapNodeAddresses\",\n\t\t\t\t\t\t\"type\": \"string[]\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"components\": [\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"string\",\n\t\t\t\t\t\t\t\t\"name\": \"ValidatorElectionType\",\n\t\t\t\t\t\t\t\t\"type\": \"string\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"string\",\n\t\t\t\t\t\t\t\t\"name\": \"ProposerElectionType\",\n\t\t\t\t\t\t\t\t\"type\": \"string\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"CheckpointPeriod\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"HighWatermarkCheckpointPeriod\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"MaxValidatorNum\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"BlockMaxTxNum\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"bool\",\n\t\t\t\t\t\t\t\t\"name\": \"EnableTimedGenEmptyBlock\",\n\t\t\t\t\t\t\t\t\"type\": \"bool\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"int64\",\n\t\t\t\t\t\t\t\t\"name\": \"NotActiveWeight\",\n\t\t\t\t\t\t\t\t\"type\": \"int64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"AbnormalNodeExcludeView\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"AgainProposeIntervalBlockInValidatorsNumPercentage\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"ContinuousNullRequestToleranceNumber\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"ReBroadcastToleranceNumber\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t}\n\t\t\t\t\t\t],\n\t\t\t\t\t\t\"internalType\": \"struct ConsensusParams\",\n\t\t\t\t\t\t\"name\": \"ConsensusParams\",\n\t\t\t\t\t\t\"type\": \"tuple\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"components\": [\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"GasLimit\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"bool\",\n\t\t\t\t\t\t\t\t\"name\": \"StartGasPriceAvailable\",\n\t\t\t\t\t\t\t\t\"type\": \"bool\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"StartGasPrice\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"MaxGasPrice\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"MinGasPrice\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"GasChangeRateValue\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"GasChangeRateDecimals\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t}\n\t\t\t\t\t\t],\n\t\t\t\t\t\t\"internalType\": \"struct FinanceParams\",\n\t\t\t\t\t\t\"name\": \"FinanceParams\",\n\t\t\t\t\t\t\"type\": \"tuple\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"components\": [\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"TxMaxSize\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t}\n\t\t\t\t\t\t],\n\t\t\t\t\t\t\"internalType\": \"struct MiscParams\",\n\t\t\t\t\t\t\"name\": \"MiscParams\",\n\t\t\t\t\t\t\"type\": \"tuple\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"components\": [\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"ID\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"string\",\n\t\t\t\t\t\t\t\t\"name\": \"AccountAddress\",\n\t\t\t\t\t\t\t\t\"type\": \"string\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"string\",\n\t\t\t\t\t\t\t\t\"name\": \"P2PNodeID\",\n\t\t\t\t\t\t\t\t\"type\": \"string\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"int64\",\n\t\t\t\t\t\t\t\t\"name\": \"ConsensusVotingPower\",\n\t\t\t\t\t\t\t\t\"type\": \"int64\"\n\t\t\t\t\t\t\t}\n\t\t\t\t\t\t],\n\t\t\t\t\t\t\"internalType\": \"struct NodeInfo[]\",\n\t\t\t\t\t\t\"name\": \"ValidatorSet\",\n\t\t\t\t\t\t\"type\": \"tuple[]\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"components\": [\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"ID\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"string\",\n\t\t\t\t\t\t\t\t\"name\": \"AccountAddress\",\n\t\t\t\t\t\t\t\t\"type\": \"string\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"string\",\n\t\t\t\t\t\t\t\t\"name\": \"P2PNodeID\",\n\t\t\t\t\t\t\t\t\"type\": \"string\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"int64\",\n\t\t\t\t\t\t\t\t\"name\": \"ConsensusVotingPower\",\n\t\t\t\t\t\t\t\t\"type\": \"int64\"\n\t\t\t\t\t\t\t}\n\t\t\t\t\t\t],\n\t\t\t\t\t\t\"internalType\": \"struct NodeInfo[]\",\n\t\t\t\t\t\t\"name\": \"CandidateSet\",\n\t\t\t\t\t\t\"type\": \"tuple[]\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"components\": [\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"ID\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"string\",\n\t\t\t\t\t\t\t\t\"name\": \"AccountAddress\",\n\t\t\t\t\t\t\t\t\"type\": \"string\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"string\",\n\t\t\t\t\t\t\t\t\"name\": \"P2PNodeID\",\n\t\t\t\t\t\t\t\t\"type\": \"string\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"int64\",\n\t\t\t\t\t\t\t\t\"name\": \"ConsensusVotingPower\",\n\t\t\t\t\t\t\t\t\"type\": \"int64\"\n\t\t\t\t\t\t\t}\n\t\t\t\t\t\t],\n\t\t\t\t\t\t\"internalType\": \"struct NodeInfo[]\",\n\t\t\t\t\t\t\"name\": \"DataSyncerSet\",\n\t\t\t\t\t\t\"type\": \"tuple[]\"\n\t\t\t\t\t}\n\t\t\t\t],\n\t\t\t\t\"internalType\": \"struct EpochInfo\",\n\t\t\t\t\"name\": \"epochInfo\",\n\t\t\t\t\"type\": \"tuple\"\n\t\t\t}\n\t\t],\n\t\t\"stateMutability\": \"view\",\n\t\t\"type\": \"function\"\n\t},\n\t{\n\t\t\"inputs\": [\n\t\t\t{\n\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\"name\": \"epochID\",\n\t\t\t\t\"type\": \"uint64\"\n\t\t\t}\n\t\t],\n\t\t\"name\": \"historyEpoch\",\n\t\t\"outputs\": [\n\t\t\t{\n\t\t\t\t\"components\": [\n\t\t\t\t\t{\n\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\"name\": \"Version\",\n\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\"name\": \"Epoch\",\n\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\"name\": \"EpochPeriod\",\n\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\"name\": \"StartBlock\",\n\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"internalType\": \"string[]\",\n\t\t\t\t\t\t\"name\": \"P2PBootstrapNodeAddresses\",\n\t\t\t\t\t\t\"type\": \"string[]\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"components\": [\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"string\",\n\t\t\t\t\t\t\t\t\"name\": \"ValidatorElectionType\",\n\t\t\t\t\t\t\t\t\"type\": \"string\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"string\",\n\t\t\t\t\t\t\t\t\"name\": \"ProposerElectionType\",\n\t\t\t\t\t\t\t\t\"type\": \"string\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"CheckpointPeriod\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"HighWatermarkCheckpointPeriod\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"MaxValidatorNum\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"BlockMaxTxNum\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"bool\",\n\t\t\t\t\t\t\t\t\"name\": \"EnableTimedGenEmptyBlock\",\n\t\t\t\t\t\t\t\t\"type\": \"bool\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"int64\",\n\t\t\t\t\t\t\t\t\"name\": \"NotActiveWeight\",\n\t\t\t\t\t\t\t\t\"type\": \"int64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"AbnormalNodeExcludeView\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"AgainProposeIntervalBlockInValidatorsNumPercentage\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"ContinuousNullRequestToleranceNumber\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"ReBroadcastToleranceNumber\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t}\n\t\t\t\t\t\t],\n\t\t\t\t\t\t\"internalType\": \"struct ConsensusParams\",\n\t\t\t\t\t\t\"name\": \"ConsensusParams\",\n\t\t\t\t\t\t\"type\": \"tuple\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"components\": [\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"GasLimit\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"bool\",\n\t\t\t\t\t\t\t\t\"name\": \"StartGasPriceAvailable\",\n\t\t\t\t\t\t\t\t\"type\": \"bool\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"StartGasPrice\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"MaxGasPrice\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"MinGasPrice\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"GasChangeRateValue\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"GasChangeRateDecimals\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t}\n\t\t\t\t\t\t],\n\t\t\t\t\t\t\"internalType\": \"struct FinanceParams\",\n\t\t\t\t\t\t\"name\": \"FinanceParams\",\n\t\t\t\t\t\t\"type\": \"tuple\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"components\": [\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"TxMaxSize\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t}\n\t\t\t\t\t\t],\n\t\t\t\t\t\t\"internalType\": \"struct MiscParams\",\n\t\t\t\t\t\t\"name\": \"MiscParams\",\n\t\t\t\t\t\t\"type\": \"tuple\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"components\": [\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"ID\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"string\",\n\t\t\t\t\t\t\t\t\"name\": \"AccountAddress\",\n\t\t\t\t\t\t\t\t\"type\": \"string\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"string\",\n\t\t\t\t\t\t\t\t\"name\": \"P2PNodeID\",\n\t\t\t\t\t\t\t\t\"type\": \"string\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"int64\",\n\t\t\t\t\t\t\t\t\"name\": \"ConsensusVotingPower\",\n\t\t\t\t\t\t\t\t\"type\": \"int64\"\n\t\t\t\t\t\t\t}\n\t\t\t\t\t\t],\n\t\t\t\t\t\t\"internalType\": \"struct NodeInfo[]\",\n\t\t\t\t\t\t\"name\": \"ValidatorSet\",\n\t\t\t\t\t\t\"type\": \"tuple[]\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"components\": [\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"ID\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"string\",\n\t\t\t\t\t\t\t\t\"name\": \"AccountAddress\",\n\t\t\t\t\t\t\t\t\"type\": \"string\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"string\",\n\t\t\t\t\t\t\t\t\"name\": \"P2PNodeID\",\n\t\t\t\t\t\t\t\t\"type\": \"string\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"int64\",\n\t\t\t\t\t\t\t\t\"name\": \"ConsensusVotingPower\",\n\t\t\t\t\t\t\t\t\"type\": \"int64\"\n\t\t\t\t\t\t\t}\n\t\t\t\t\t\t],\n\t\t\t\t\t\t\"internalType\": \"struct NodeInfo[]\",\n\t\t\t\t\t\t\"name\": \"CandidateSet\",\n\t\t\t\t\t\t\"type\": \"tuple[]\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"components\": [\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"ID\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"string\",\n\t\t\t\t\t\t\t\t\"name\": \"AccountAddress\",\n\t\t\t\t\t\t\t\t\"type\": \"string\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"string\",\n\t\t\t\t\t\t\t\t\"name\": \"P2PNodeID\",\n\t\t\t\t\t\t\t\t\"type\": \"string\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"int64\",\n\t\t\t\t\t\t\t\t\"name\": \"ConsensusVotingPower\",\n\t\t\t\t\t\t\t\t\"type\": \"int64\"\n\t\t\t\t\t\t\t}\n\t\t\t\t\t\t],\n\t\t\t\t\t\t\"internalType\": \"struct NodeInfo[]\",\n\t\t\t\t\t\t\"name\": \"DataSyncerSet\",\n\t\t\t\t\t\t\"type\": \"tuple[]\"\n\t\t\t\t\t}\n\t\t\t\t],\n\t\t\t\t\"internalType\": \"struct EpochInfo\",\n\t\t\t\t\"name\": \"epochInfo\",\n\t\t\t\t\"type\": \"tuple\"\n\t\t\t}\n\t\t],\n\t\t\"stateMutability\": \"view\",\n\t\t\"type\": \"function\"\n\t},\n\t{\n\t\t\"inputs\": [],\n\t\t\"name\": \"nextEpoch\",\n\t\t\"outputs\": [\n\t\t\t{\n\t\t\t\t\"components\": [\n\t\t\t\t\t{\n\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\"name\": \"Version\",\n\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\"name\": \"Epoch\",\n\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\"name\": \"EpochPeriod\",\n\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\"name\": \"StartBlock\",\n\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"internalType\": \"string[]\",\n\t\t\t\t\t\t\"name\": \"P2PBootstrapNodeAddresses\",\n\t\t\t\t\t\t\"type\": \"string[]\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"components\": [\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"string\",\n\t\t\t\t\t\t\t\t\"name\": \"ValidatorElectionType\",\n\t\t\t\t\t\t\t\t\"type\": \"string\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"string\",\n\t\t\t\t\t\t\t\t\"name\": \"ProposerElectionType\",\n\t\t\t\t\t\t\t\t\"type\": \"string\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"CheckpointPeriod\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"HighWatermarkCheckpointPeriod\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"MaxValidatorNum\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"BlockMaxTxNum\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"bool\",\n\t\t\t\t\t\t\t\t\"name\": \"EnableTimedGenEmptyBlock\",\n\t\t\t\t\t\t\t\t\"type\": \"bool\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"int64\",\n\t\t\t\t\t\t\t\t\"name\": \"NotActiveWeight\",\n\t\t\t\t\t\t\t\t\"type\": \"int64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"AbnormalNodeExcludeView\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"AgainProposeIntervalBlockInValidatorsNumPercentage\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"ContinuousNullRequestToleranceNumber\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"ReBroadcastToleranceNumber\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t}\n\t\t\t\t\t\t],\n\t\t\t\t\t\t\"internalType\": \"struct ConsensusParams\",\n\t\t\t\t\t\t\"name\": \"ConsensusParams\",\n\t\t\t\t\t\t\"type\": \"tuple\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"components\": [\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"GasLimit\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"bool\",\n\t\t\t\t\t\t\t\t\"name\": \"StartGasPriceAvailable\",\n\t\t\t\t\t\t\t\t\"type\": \"bool\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"StartGasPrice\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"MaxGasPrice\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"MinGasPrice\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"GasChangeRateValue\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"GasChangeRateDecimals\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t}\n\t\t\t\t\t\t],\n\t\t\t\t\t\t\"internalType\": \"struct FinanceParams\",\n\t\t\t\t\t\t\"name\": \"FinanceParams\",\n\t\t\t\t\t\t\"type\": \"tuple\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"components\": [\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"TxMaxSize\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t}\n\t\t\t\t\t\t],\n\t\t\t\t\t\t\"internalType\": \"struct MiscParams\",\n\t\t\t\t\t\t\"name\": \"MiscParams\",\n\t\t\t\t\t\t\"type\": \"tuple\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"components\": [\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"ID\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"string\",\n\t\t\t\t\t\t\t\t\"name\": \"AccountAddress\",\n\t\t\t\t\t\t\t\t\"type\": \"string\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"string\",\n\t\t\t\t\t\t\t\t\"name\": \"P2PNodeID\",\n\t\t\t\t\t\t\t\t\"type\": \"string\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"int64\",\n\t\t\t\t\t\t\t\t\"name\": \"ConsensusVotingPower\",\n\t\t\t\t\t\t\t\t\"type\": \"int64\"\n\t\t\t\t\t\t\t}\n\t\t\t\t\t\t],\n\t\t\t\t\t\t\"internalType\": \"struct NodeInfo[]\",\n\t\t\t\t\t\t\"name\": \"ValidatorSet\",\n\t\t\t\t\t\t\"type\": \"tuple[]\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"components\": [\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"ID\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"string\",\n\t\t\t\t\t\t\t\t\"name\": \"AccountAddress\",\n\t\t\t\t\t\t\t\t\"type\": \"string\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"string\",\n\t\t\t\t\t\t\t\t\"name\": \"P2PNodeID\",\n\t\t\t\t\t\t\t\t\"type\": \"string\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"int64\",\n\t\t\t\t\t\t\t\t\"name\": \"ConsensusVotingPower\",\n\t\t\t\t\t\t\t\t\"type\": \"int64\"\n\t\t\t\t\t\t\t}\n\t\t\t\t\t\t],\n\t\t\t\t\t\t\"internalType\": \"struct NodeInfo[]\",\n\t\t\t\t\t\t\"name\": \"CandidateSet\",\n\t\t\t\t\t\t\"type\": \"tuple[]\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"components\": [\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"uint64\",\n\t\t\t\t\t\t\t\t\"name\": \"ID\",\n\t\t\t\t\t\t\t\t\"type\": \"uint64\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"string\",\n\t\t\t\t\t\t\t\t\"name\": \"AccountAddress\",\n\t\t\t\t\t\t\t\t\"type\": \"string\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"string\",\n\t\t\t\t\t\t\t\t\"name\": \"P2PNodeID\",\n\t\t\t\t\t\t\t\t\"type\": \"string\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\t\"internalType\": \"int64\",\n\t\t\t\t\t\t\t\t\"name\": \"ConsensusVotingPower\",\n\t\t\t\t\t\t\t\t\"type\": \"int64\"\n\t\t\t\t\t\t\t}\n\t\t\t\t\t\t],\n\t\t\t\t\t\t\"internalType\": \"struct NodeInfo[]\",\n\t\t\t\t\t\t\"name\": \"DataSyncerSet\",\n\t\t\t\t\t\t\"type\": \"tuple[]\"\n\t\t\t\t\t}\n\t\t\t\t],\n\t\t\t\t\"internalType\": \"struct EpochInfo\",\n\t\t\t\t\"name\": \"epochInfo\",\n\t\t\t\t\"type\": \"tuple\"\n\t\t\t}\n\t\t],\n\t\t\"stateMutability\": \"view\",\n\t\t\"type\": \"function\"\n\t}\n]"

var epochManagerABI *abi.ABI

var methodSig2ArgsReceiverConstructor = map[string]func() any{
	"currentEpoch()": func() any {
		return &currentEpochMethodArgs{}
	},
	"nextEpoch()": func() any {
		return &nextEpochMethodArgs{}
	},
	"historyEpoch(uint64)": func() any {
		return &historyEpochMethodArgs{}
	},
}

type currentEpochMethodArgs struct {
}

type nextEpochMethodArgs struct {
}

type historyEpochMethodArgs struct {
	EpochID uint64
}

func init() {
	emAbi, err := abi.JSON(strings.NewReader(epochManagerABIData))
	if err != nil {
		panic(err)
	}
	epochManagerABI = &emAbi
}

var _ common.SystemContract = (*EpochManager)(nil)

type EpochManager struct {
	logger      logrus.FieldLogger
	account     ledger.IAccount
	stateLedger ledger.StateLedger
}

func NewEpochManager(cfg *common.SystemContractConfig) *EpochManager {
	return &EpochManager{
		logger: cfg.Logger,
	}
}

func (m *EpochManager) Reset(lastHeight uint64, stateLedger ledger.StateLedger) {
	m.account = stateLedger.GetOrCreateAccount(types.NewAddressByStr(common.EpochManagerContractAddr))
	m.stateLedger = stateLedger
}

func (m *EpochManager) Run(msg *vm.Message) (*vm.ExecutionResult, error) {
	result := &vm.ExecutionResult{}
	ret, err := func() ([]byte, error) {
		args, method, err := common.ParseContractCallArgs(epochManagerABI, msg.Data, methodSig2ArgsReceiverConstructor)
		if err != nil {
			return nil, err
		}
		switch t := args.(type) {
		case *currentEpochMethodArgs:
			epochInfo, err := GetCurrentEpochInfo(m.stateLedger)
			if err != nil {
				return nil, err
			}
			return method.Outputs.Pack(epochInfo)
		case *nextEpochMethodArgs:
			epochInfo, err := GetNextEpochInfo(m.stateLedger)
			if err != nil {
				return nil, err
			}
			return method.Outputs.Pack(epochInfo)
		case *historyEpochMethodArgs:
			epochInfo, err := getEpoch(m.stateLedger, historyEpochInfoKey(t.EpochID))
			if err != nil {
				return nil, err
			}
			return method.Outputs.Pack(epochInfo)
		default:
			return nil, errors.Errorf("%v: not support method", vm.ErrExecutionReverted)
		}
	}()
	if err != nil {
		result.Err = vm.ErrExecutionReverted
		result.ReturnData = []byte(err.Error())
	} else {
		result.ReturnData = ret
	}
	result.UsedGas = common.CalculateDynamicGas(msg.Data)
	return result, nil
}

func (m *EpochManager) EstimateGas(callArgs *types.CallArgs) (uint64, error) {
	var data []byte
	if callArgs.Data != nil {
		data = *callArgs.Data
	}

	_, _, err := common.ParseContractCallArgs(epochManagerABI, data, methodSig2ArgsReceiverConstructor)
	if err != nil {
		return 0, errors.Errorf("%v: %v", vm.ErrExecutionReverted, err)
	}

	return common.CalculateDynamicGas(*callArgs.Data), nil
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
	checkNodes := func(nodes []rbft.NodeInfo) error {
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

func SetNextEpochInfo(lg ledger.StateLedger, n *rbft.EpochInfo) error {
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
		lastEpochInfo, err := GetEpochInfo(lg, e.Epoch-1)
		if err != nil {
			return nil, err
		}
		if err := e.ElectValidators(lastEpochInfo, electValidatorsByWrfSeed); err != nil {
			return nil, err
		}
		validatorIDs := lo.Map(e.ValidatorSet, func(item rbft.NodeInfo, index int) uint64 {
			return item.ID
		})
		loggers.Logger(loggers.Epoch).Infof("Elect new Validators: %v", validatorIDs)

		if err := setEpochInfo(lg, e); err != nil {
			return nil, err
		}

		n := e.Clone()
		n.Epoch++
		n.StartBlock += n.EpochPeriod
		n.FinanceParams.StartGasPrice = 0
		n.FinanceParams.StartGasPriceAvailable = false
		if err := SetNextEpochInfo(lg, n); err != nil {
			return nil, err
		}

		// return current
		return e, nil
	}
	return nil, errors.New("not found current epoch info")
}

func checkNodeInfo(node rbft.NodeInfo) error {
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
func AddNode(lg ledger.StateLedger, newNode rbft.NodeInfo) (uint64, error) {
	// Clone the newNode to avoid modifying the original instance
	newNode = newNode.Clone()

	// Check if the new node info is valid
	if err := checkNodeInfo(newNode); err != nil {
		return 0, err
	}

	// Get the next epoch information from the ledger
	nextEpochInfo, err := GetNextEpochInfo(lg)
	if err != nil {
		return 0, err
	}

	// Function to check for duplicate node information
	checkNodeInfoDuplicate := func(nodes []rbft.NodeInfo) error {
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
		return 0, err
	}
	if err := checkNodeInfoDuplicate(nextEpochInfo.CandidateSet); err != nil {
		return 0, err
	}
	if err := checkNodeInfoDuplicate(nextEpochInfo.DataSyncerSet); err != nil {
		return 0, err
	}

	// Get the node ID generator from the ledger
	nodeIDGenerator, err := getNodeIDGenerator(lg)
	if err != nil {
		return 0, err
	}

	// Automatically assign a self-increasing ID to the new node
	newNode.ID = nodeIDGenerator

	// Update the ID generator
	setNodeIDGenerator(lg, nodeIDGenerator+1)

	// Update the next epoch info by adding the new node to the candidate set
	nextEpochInfo.CandidateSet = append(nextEpochInfo.CandidateSet, newNode)

	// Set the updated next epoch info in the ledger
	if err := SetNextEpochInfo(lg, nextEpochInfo); err != nil {
		return 0, err
	}

	return newNode.ID, nil
}

// RemoveNode removes a node from the validator set, candidate set, or data syncer set.
// It takes a StateLedger and the ID of the node to remove as input.
// It returns an error if the node ID is not found in any of the sets.
func RemoveNode(lg ledger.StateLedger, removeNodeID uint64) error {
	return removeNodeByField(lg, removeNodeID, id)
}

func RemoveNodeByP2PNodeID(lg ledger.StateLedger, removeByP2PNodeID string) error {
	return removeNodeByField(lg, removeByP2PNodeID, p2PNodeID)
}

func removeNodeByField(lg ledger.StateLedger, fieldValue any, fieldName string) error {
	// Get the next epoch info from the ledger
	nextEpochInfo, err := GetNextEpochInfo(lg)
	if err != nil {
		return err
	}

	// Function to remove a node from a set
	removeNode := func(nodes []rbft.NodeInfo) (bool, []rbft.NodeInfo) {
		var matchedIdx int
		var matched bool

		// Find the index of the node with the given field value in the set
		for idx, n := range nodes {
			switch fieldName {
			case id:
				if n.ID == fieldValue {
					matched = true
				}
			case p2PNodeID:
				if n.P2PNodeID == fieldValue {
					matched = true
				}
			}

			if matched {
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
		return SetNextEpochInfo(lg, nextEpochInfo)
	}

	// Remove the node from the candidate set
	removed, nextEpochInfo.CandidateSet = removeNode(nextEpochInfo.CandidateSet)
	if removed {
		return SetNextEpochInfo(lg, nextEpochInfo)
	}

	// Remove the node from the data syncer set
	removed, nextEpochInfo.DataSyncerSet = removeNode(nextEpochInfo.DataSyncerSet)
	if removed {
		return SetNextEpochInfo(lg, nextEpochInfo)
	}

	// If the node is not found in any of the sets, return an error
	return errors.Errorf("failed to remove node, node %s not found: %v", fieldName, fieldValue)
}
