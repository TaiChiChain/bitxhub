package common

import (
	_ "embed"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtype "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	vm "github.com/axiomesh/eth-kit/evm"
)

const (
	// ZeroAddress is a special address, no one has control
	ZeroAddress = "0x0000000000000000000000000000000000000000"

	// system contract address range 0x1000-0xffff, start from 1000, avoid conflicts with precompiled contracts

	// ProposalIDContractAddr is the contract to used to generate the proposal ID
	ProposalIDContractAddr = "0x0000000000000000000000000000000000001000"
	GovernanceContractAddr = "0x0000000000000000000000000000000000001001"

	// Addr2NameContractAddr for unique name mapping to address
	Addr2NameContractAddr           = "0x0000000000000000000000000000000000001003"
	WhiteListContractAddr           = "0x0000000000000000000000000000000000001004"
	NotFinishedProposalContractAddr = "0x0000000000000000000000000000000000001006"

	// EpochManagerContractAddr is the contract to used to manager chain epoch info
	EpochManagerContractAddr = "0x0000000000000000000000000000000000001007"

	ProposeMethod                = "propose"
	VoteMethod                   = "vote"
	GetLatestProposalIDMethod    = "getLatestProposalID"
	SubmitMethod                 = "submit"
	RemoveMethod                 = "remove"
	QueryAuthInfoMethod          = "queryAuthInfo"
	QueryWhiteListProviderMethod = "queryWhiteListProvider"
	CurrentEpochMethod           = "currentEpoch"
	NextEpochMethod              = "nextEpoch"
	HistoryEpochMethod           = "historyEpoch"
)

//go:embed sol/Governance.abi
var governanceABI string

//go:embed sol/WhiteList.abi
var whiteListABI string

//go:embed sol/EpochManager.abi
var epochManagerABI string

// governanceMethod2Sig is method name to method signature mapping
var governanceMethod2Sig = map[string]string{
	"propose":             "propose(uint8,string,string,uint64,bytes)",
	"vote":                "vote(uint64,uint8)",
	"proposal":            "proposal(uint64)",
	"getLatestProposalID": "getLatestProposalID()",
}

var whiteListMethod2Sig = map[string]string{
	SubmitMethod:                 "submit(bytes)",
	RemoveMethod:                 "remove(bytes)",
	QueryAuthInfoMethod:          "queryAuthInfo(bytes)",
	QueryWhiteListProviderMethod: "queryWhiteListProvider(bytes)",
}

var epochManagerMethod2Sig = map[string]string{
	CurrentEpochMethod: "currentEpoch()",
	NextEpochMethod:    "nextEpoch()",
	HistoryEpochMethod: "historyEpoch(uint64)",
}

// contractAddr2MethodSig is system contract address to method sig mapping
var contractAddr2MethodSig = map[string]map[string]string{
	GovernanceContractAddr:   governanceMethod2Sig,
	WhiteListContractAddr:    whiteListMethod2Sig,
	EpochManagerContractAddr: epochManagerMethod2Sig,
}

var contract2ABIFile = map[string]string{
	GovernanceContractAddr:   governanceABI,
	WhiteListContractAddr:    whiteListABI,
	EpochManagerContractAddr: epochManagerABI,
}

var Contract2MethodSig map[string]map[string][]byte
var Contract2ABI map[string]abi.ABI

func init() {
	Contract2MethodSig = make(map[string]map[string][]byte)
	for contractAddr, method2Sig := range contractAddr2MethodSig {
		m2sig := make(map[string][]byte)
		for methodName, methodSig := range method2Sig {
			m2sig[methodName] = crypto.Keccak256([]byte(methodSig))
		}
		Contract2MethodSig[contractAddr] = m2sig
	}

	Contract2ABI = make(map[string]abi.ABI)
	for contractAddr, abiFile := range contract2ABIFile {
		contractABI, err := abi.JSON(strings.NewReader(abiFile))
		if err != nil {
			panic(err)
		}
		Contract2ABI[contractAddr] = contractABI
	}
}

type SystemContractConfig struct {
	Logger logrus.FieldLogger
}

type SystemContractConstruct func(cfg *SystemContractConfig) SystemContract

type SystemContract interface {
	// IsSystemContract judge if is system contract
	IsSystemContract(addr *types.Address) bool

	// Reset the state of the system contract
	Reset(uint64, ledger.StateLedger)

	// Run the system contract
	Run(*vm.Message) (*vm.ExecutionResult, error)

	// EstimateGas estimate the gas cost of the system contract
	EstimateGas(*types.CallArgs) (uint64, error)

	// View return a view system contract
	View() SystemContract
}

type SystemContractContext struct {
	StateLedger   ledger.StateLedger
	CurrentHeight uint64
	CurrentLogs   *[]Log
	CurrentUser   *ethcommon.Address
}

// InnerSystemContract must be implemented by all system contract
type InnerSystemContract interface {
	SetContext(*SystemContractContext)
}

func IsInSlice[T ~uint8 | ~string](value T, slice []T) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}

	return false
}

func RemoveFirstMatchStrInSlice(slice []string, val string) []string {
	for i, v := range slice {
		if v == val {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

type Log struct {
	Address *types.Address
	Topics  []*types.Hash
	Data    []byte
	Removed bool
}

func CalculateDynamicGas(bytes []byte) uint64 {
	gas, _ := vm.IntrinsicGas(bytes, []ethtype.AccessTuple{}, false, true, true, true)
	return gas
}

func ParseContractCallArgs(contractAbi *abi.ABI, data []byte, methodSig2ArgsReceiverConstructor map[string]func() any) (any, *abi.Method, error) {
	if len(data) < 4 {
		return nil, nil, errors.New("gabi: data is invalid")
	}

	method, err := contractAbi.MethodById(data[:4])
	if err != nil {
		return nil, nil, errors.Errorf("gabi: not found method: %v", err)
	}
	argsReceiverConstructor, ok := methodSig2ArgsReceiverConstructor[method.Sig]
	if !ok {
		return nil, nil, errors.Errorf("gabi: not support method: %v", method.Name)
	}
	args := argsReceiverConstructor()
	unpacked, err := method.Inputs.Unpack(data[4:])
	if err != nil {
		return nil, nil, errors.Errorf("gabi: decode method[%s] args failed: %v", method.Name, err)
	}
	if err = method.Inputs.Copy(args, unpacked); err != nil {
		return nil, nil, errors.Errorf("gabi: decode method[%s] args failed: %v", method.Name, err)
	}
	return args, method, nil
}
