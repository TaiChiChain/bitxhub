// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package ientry_point

import (
	"math/big"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/pkg/packer"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = common.Big1
	_ = types.AxcUnit
	_ = abi.ConvertType
	_ = packer.RevertError{}
)

// IEntryPointAggregatorStakeInfo is an auto generated low-level Go binding around an user-defined struct.
type IEntryPointAggregatorStakeInfo struct {
	Aggregator common.Address
	StakeInfo  IStakeManagerStakeInfo
}

// IEntryPointReturnInfo is an auto generated low-level Go binding around an user-defined struct.
type IEntryPointReturnInfo struct {
	PreOpGas         *big.Int
	Prefund          *big.Int
	SigFailed        bool
	ValidAfter       *big.Int
	ValidUntil       *big.Int
	PaymasterContext []byte
}

// IEntryPointUserOpsPerAggregator is an auto generated low-level Go binding around an user-defined struct.
type IEntryPointUserOpsPerAggregator struct {
	UserOps    []UserOperation
	Aggregator common.Address
	Signature  []byte
}

// IStakeManagerDepositInfo is an auto generated low-level Go binding around an user-defined struct.
type IStakeManagerDepositInfo struct {
	Deposit         *big.Int
	Staked          bool
	Stake           *big.Int
	UnstakeDelaySec uint32
	WithdrawTime    *big.Int
}

// IStakeManagerStakeInfo is an auto generated low-level Go binding around an user-defined struct.
type IStakeManagerStakeInfo struct {
	Stake           *big.Int
	UnstakeDelaySec *big.Int
}

// UserOperation is an auto generated low-level Go binding around an user-defined struct.
type UserOperation struct {
	Sender               common.Address
	Nonce                *big.Int
	InitCode             []byte
	CallData             []byte
	CallGasLimit         *big.Int
	VerificationGasLimit *big.Int
	PreVerificationGas   *big.Int
	MaxFeePerGas         *big.Int
	MaxPriorityFeePerGas *big.Int
	PaymasterAndData     []byte
	Signature            []byte
}

type IentryPoint interface {

	// AddStake is a paid mutator transaction binding the contract method 0x0396cb60.
	//
	// Solidity: function addStake(uint32 _unstakeDelaySec) payable returns()
	AddStake(_unstakeDelaySec uint32) error

	// DepositTo is a paid mutator transaction binding the contract method 0xb760faf9.
	//
	// Solidity: function depositTo(address account) payable returns()
	DepositTo(account common.Address) error

	// GetSenderAddress is a paid mutator transaction binding the contract method 0x9b249f69.
	//
	// Solidity: function getSenderAddress(bytes initCode) returns()
	GetSenderAddress(initCode []byte) error

	// HandleAggregatedOps is a paid mutator transaction binding the contract method 0x4b1d7cf5.
	//
	// Solidity: function handleAggregatedOps(((address,uint256,bytes,bytes,uint256,uint256,uint256,uint256,uint256,bytes,bytes)[],address,bytes)[] opsPerAggregator, address beneficiary) returns()
	HandleAggregatedOps(opsPerAggregator []IEntryPointUserOpsPerAggregator, beneficiary common.Address) error

	// HandleOps is a paid mutator transaction binding the contract method 0x1fad948c.
	//
	// Solidity: function handleOps((address,uint256,bytes,bytes,uint256,uint256,uint256,uint256,uint256,bytes,bytes)[] ops, address beneficiary) returns()
	HandleOps(ops []UserOperation, beneficiary common.Address) error

	// IncrementNonce is a paid mutator transaction binding the contract method 0x0bd28e3b.
	//
	// Solidity: function incrementNonce(uint192 key) returns()
	IncrementNonce(key *big.Int) error

	// SimulateHandleOp is a paid mutator transaction binding the contract method 0xd6383f94.
	//
	// Solidity: function simulateHandleOp((address,uint256,bytes,bytes,uint256,uint256,uint256,uint256,uint256,bytes,bytes) op, address target, bytes targetCallData) returns()
	SimulateHandleOp(op UserOperation, target common.Address, targetCallData []byte) error

	// SimulateValidation is a paid mutator transaction binding the contract method 0xee219423.
	//
	// Solidity: function simulateValidation((address,uint256,bytes,bytes,uint256,uint256,uint256,uint256,uint256,bytes,bytes) userOp) returns()
	SimulateValidation(userOp UserOperation) error

	// UnlockStake is a paid mutator transaction binding the contract method 0xbb9fe6bf.
	//
	// Solidity: function unlockStake() returns()
	UnlockStake() error

	// WithdrawStake is a paid mutator transaction binding the contract method 0xc23a5cea.
	//
	// Solidity: function withdrawStake(address withdrawAddress) returns()
	WithdrawStake(withdrawAddress common.Address) error

	// WithdrawTo is a paid mutator transaction binding the contract method 0x205c2878.
	//
	// Solidity: function withdrawTo(address withdrawAddress, uint256 withdrawAmount) returns()
	WithdrawTo(withdrawAddress common.Address, withdrawAmount *big.Int) error

	// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
	//
	// Solidity: function balanceOf(address account) view returns(uint256)
	BalanceOf(account common.Address) (*big.Int, error)

	// GetDepositInfo is a free data retrieval call binding the contract method 0x5287ce12.
	//
	// Solidity: function getDepositInfo(address account) view returns((uint112,bool,uint112,uint32,uint48) info)
	GetDepositInfo(account common.Address) (IStakeManagerDepositInfo, error)

	// GetNonce is a free data retrieval call binding the contract method 0x35567e1a.
	//
	// Solidity: function getNonce(address sender, uint192 key) view returns(uint256 nonce)
	GetNonce(sender common.Address, key *big.Int) (*big.Int, error)

	// GetUserOpHash is a free data retrieval call binding the contract method 0xa6193531.
	//
	// Solidity: function getUserOpHash((address,uint256,bytes,bytes,uint256,uint256,uint256,uint256,uint256,bytes,bytes) userOp) view returns(bytes32)
	GetUserOpHash(userOp UserOperation) ([32]byte, error)
}

// EventAccountDeployed represents a AccountDeployed event raised by the IentryPoint contract.
type EventAccountDeployed struct {
	UserOpHash [32]byte
	Sender     common.Address
	Factory    common.Address
	Paymaster  common.Address
}

func (_event *EventAccountDeployed) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["AccountDeployed"])
}

// EventBeforeExecution represents a BeforeExecution event raised by the IentryPoint contract.
type EventBeforeExecution struct {
}

func (_event *EventBeforeExecution) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["BeforeExecution"])
}

// EventDeposited represents a Deposited event raised by the IentryPoint contract.
type EventDeposited struct {
	Account      common.Address
	TotalDeposit *big.Int
}

func (_event *EventDeposited) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["Deposited"])
}

// EventSignatureAggregatorChanged represents a SignatureAggregatorChanged event raised by the IentryPoint contract.
type EventSignatureAggregatorChanged struct {
	Aggregator common.Address
}

func (_event *EventSignatureAggregatorChanged) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["SignatureAggregatorChanged"])
}

// EventStakeLocked represents a StakeLocked event raised by the IentryPoint contract.
type EventStakeLocked struct {
	Account         common.Address
	TotalStaked     *big.Int
	UnstakeDelaySec *big.Int
}

func (_event *EventStakeLocked) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["StakeLocked"])
}

// EventStakeUnlocked represents a StakeUnlocked event raised by the IentryPoint contract.
type EventStakeUnlocked struct {
	Account      common.Address
	WithdrawTime *big.Int
}

func (_event *EventStakeUnlocked) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["StakeUnlocked"])
}

// EventStakeWithdrawn represents a StakeWithdrawn event raised by the IentryPoint contract.
type EventStakeWithdrawn struct {
	Account         common.Address
	WithdrawAddress common.Address
	Amount          *big.Int
}

func (_event *EventStakeWithdrawn) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["StakeWithdrawn"])
}

// EventUserOperationEvent represents a UserOperationEvent event raised by the IentryPoint contract.
type EventUserOperationEvent struct {
	UserOpHash    [32]byte
	Sender        common.Address
	Paymaster     common.Address
	Nonce         *big.Int
	Success       bool
	ActualGasCost *big.Int
	ActualGasUsed *big.Int
}

func (_event *EventUserOperationEvent) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["UserOperationEvent"])
}

// EventUserOperationRevertReason represents a UserOperationRevertReason event raised by the IentryPoint contract.
type EventUserOperationRevertReason struct {
	UserOpHash   [32]byte
	Sender       common.Address
	Nonce        *big.Int
	RevertReason []byte
}

func (_event *EventUserOperationRevertReason) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["UserOperationRevertReason"])
}

// EventWithdrawn represents a Withdrawn event raised by the IentryPoint contract.
type EventWithdrawn struct {
	Account         common.Address
	WithdrawAddress common.Address
	Amount          *big.Int
}

func (_event *EventWithdrawn) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["Withdrawn"])
}

// ErrorExecutionResult represents a ExecutionResult error raised by the IentryPoint contract.
type ErrorExecutionResult struct {
	PreOpGas      *big.Int
	Paid          *big.Int
	ValidAfter    *big.Int
	ValidUntil    *big.Int
	TargetSuccess bool
	TargetResult  []byte
}

func (_error *ErrorExecutionResult) Pack(abi abi.ABI) error {
	return packer.PackError(_error, abi.Errors["ExecutionResult"])
}

// ErrorFailedOp represents a FailedOp error raised by the IentryPoint contract.
type ErrorFailedOp struct {
	OpIndex *big.Int
	Reason  string
}

func (_error *ErrorFailedOp) Pack(abi abi.ABI) error {
	return packer.PackError(_error, abi.Errors["FailedOp"])
}

// ErrorSenderAddressResult represents a SenderAddressResult error raised by the IentryPoint contract.
type ErrorSenderAddressResult struct {
	Sender common.Address
}

func (_error *ErrorSenderAddressResult) Pack(abi abi.ABI) error {
	return packer.PackError(_error, abi.Errors["SenderAddressResult"])
}

// ErrorSignatureValidationFailed represents a SignatureValidationFailed error raised by the IentryPoint contract.
type ErrorSignatureValidationFailed struct {
	Aggregator common.Address
}

func (_error *ErrorSignatureValidationFailed) Pack(abi abi.ABI) error {
	return packer.PackError(_error, abi.Errors["SignatureValidationFailed"])
}

// ErrorValidationResult represents a ValidationResult error raised by the IentryPoint contract.
type ErrorValidationResult struct {
	ReturnInfo    IEntryPointReturnInfo
	SenderInfo    IStakeManagerStakeInfo
	FactoryInfo   IStakeManagerStakeInfo
	PaymasterInfo IStakeManagerStakeInfo
}

func (_error *ErrorValidationResult) Pack(abi abi.ABI) error {
	return packer.PackError(_error, abi.Errors["ValidationResult"])
}

// ErrorValidationResultWithAggregation represents a ValidationResultWithAggregation error raised by the IentryPoint contract.
type ErrorValidationResultWithAggregation struct {
	ReturnInfo     IEntryPointReturnInfo
	SenderInfo     IStakeManagerStakeInfo
	FactoryInfo    IStakeManagerStakeInfo
	PaymasterInfo  IStakeManagerStakeInfo
	AggregatorInfo IEntryPointAggregatorStakeInfo
}

func (_error *ErrorValidationResultWithAggregation) Pack(abi abi.ABI) error {
	return packer.PackError(_error, abi.Errors["ValidationResultWithAggregation"])
}
