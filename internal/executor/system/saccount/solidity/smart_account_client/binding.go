// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package smart_account_client

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

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

// BindingContractMetaData contains all meta data concerning the BindingContract contract.
var BindingContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIEntryPoint\",\"name\":\"anEntryPoint\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"contractIEntryPoint\",\"name\":\"entryPoint\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"SmartAccountInitialized\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"addDeposit\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"entryPoint\",\"outputs\":[{\"internalType\":\"contractIEntryPoint\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"dest\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"func\",\"type\":\"bytes\"}],\"name\":\"execute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"dest\",\"type\":\"address[]\"},{\"internalType\":\"bytes[]\",\"name\":\"func\",\"type\":\"bytes[]\"}],\"name\":\"executeBatch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getDeposit\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getNonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"anOwner\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"initCode\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"callData\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"callGasLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"verificationGasLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"preVerificationGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxPriorityFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"paymasterAndData\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"internalType\":\"structUserOperation\",\"name\":\"userOp\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"userOpHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"missingAccountFunds\",\"type\":\"uint256\"}],\"name\":\"validateUserOp\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"validationData\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"withdrawAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"withdrawDepositTo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
	Bin: "0x60a060405234801562000010575f80fd5b506040516200143a3803806200143a8339818101604052810190620000369190620000e9565b8073ffffffffffffffffffffffffffffffffffffffff1660808173ffffffffffffffffffffffffffffffffffffffff16815250505062000119565b5f80fd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f620000a08262000075565b9050919050565b5f620000b38262000094565b9050919050565b620000c581620000a7565b8114620000d0575f80fd5b50565b5f81519050620000e381620000ba565b92915050565b5f6020828403121562000101576200010062000071565b5b5f6200011084828501620000d3565b91505092915050565b608051611301620001395f395f818161047a01526109b201526113015ff3fe608060405260043610610094575f3560e01c8063b0d691fe11610058578063b0d691fe1461015f578063b61d27f614610189578063c399ec88146101b1578063c4d66de8146101db578063d087d288146102035761009b565b806318dfb3c71461009f5780633a871cdd146100c75780634a58db19146101035780634d44560d1461010d5780638da5cb5b146101355761009b565b3661009b57005b5f80fd5b3480156100aa575f80fd5b506100c560048036038101906100c09190610ad4565b61022d565b005b3480156100d2575f80fd5b506100ed60048036038101906100e89190610bdb565b610336565b6040516100fa9190610c56565b60405180910390f35b61010b610368565b005b348015610118575f80fd5b50610133600480360381019061012e9190610cc9565b6103d9565b005b348015610140575f80fd5b50610149610454565b6040516101569190610d27565b60405180910390f35b34801561016a575f80fd5b50610173610477565b6040516101809190610d9b565b60405180910390f35b348015610194575f80fd5b506101af60048036038101906101aa9190610e33565b61049e565b005b3480156101bc575f80fd5b506101c56104fa565b6040516101d29190610c56565b60405180910390f35b3480156101e6575f80fd5b5061020160048036038101906101fc9190610ea4565b61057f565b005b34801561020e575f80fd5b5061021761058b565b6040516102249190610c56565b60405180910390f35b610235610612565b81819050848490501461027d576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161027490610f29565b60405180910390fd5b5f5b8484905081101561032f576103228585838181106102a05761029f610f47565b5b90506020020160208101906102b59190610ea4565b5f8585858181106102c9576102c8610f47565b5b90506020028101906102db9190610f80565b8080601f0160208091040260200160405190810160405280939291908181526020018383808284375f81840152601f19601f820116905080830192505050505050506106de565b808060010191505061027f565b5050505050565b5f61033f61075e565b61034984846107d5565b905061035884602001356107dc565b610361826107df565b9392505050565b610370610477565b73ffffffffffffffffffffffffffffffffffffffff1663b760faf934306040518363ffffffff1660e01b81526004016103a99190610d27565b5f604051808303818588803b1580156103c0575f80fd5b505af11580156103d2573d5f803e3d5ffd5b5050505050565b6103e1610876565b6103e9610477565b73ffffffffffffffffffffffffffffffffffffffff1663205c287883836040518363ffffffff1660e01b8152600401610423929190610ff1565b5f604051808303815f87803b15801561043a575f80fd5b505af115801561044c573d5f803e3d5ffd5b505050505050565b5f8054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b5f7f0000000000000000000000000000000000000000000000000000000000000000905090565b6104a6610612565b6104f4848484848080601f0160208091040260200160405190810160405280939291908181526020018383808284375f81840152601f19601f820116905080830192505050505050506106de565b50505050565b5f610503610477565b73ffffffffffffffffffffffffffffffffffffffff166370a08231306040518263ffffffff1660e01b815260040161053b9190610d27565b602060405180830381865afa158015610556573d5f803e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061057a919061102c565b905090565b6105888161093b565b50565b5f610594610477565b73ffffffffffffffffffffffffffffffffffffffff166335567e1a305f6040518363ffffffff1660e01b81526004016105ce9291906110b3565b602060405180830381865afa1580156105e9573d5f803e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061060d919061102c565b905090565b61061a610477565b73ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16148061069d57505f8054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16145b6106dc576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016106d390611124565b60405180910390fd5b565b5f808473ffffffffffffffffffffffffffffffffffffffff16848460405161070691906111ae565b5f6040518083038185875af1925050503d805f8114610740576040519150601f19603f3d011682016040523d82523d5f602084013e610745565b606091505b50915091508161075757805160208201fd5b5050505050565b610766610477565b73ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146107d3576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016107ca9061120e565b60405180910390fd5b565b5f92915050565b50565b5f8114610873575f3373ffffffffffffffffffffffffffffffffffffffff16827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff9060405161082d9061124f565b5f60405180830381858888f193505050503d805f8114610868576040519150601f19603f3d011682016040523d82523d5f602084013e61086d565b606091505b50509050505b50565b5f8054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614806108fa57503073ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16145b610939576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610930906112ad565b60405180910390fd5b565b805f806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505f8054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff167f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff167fb7053def2fe3d2a5ecb12939fbfcc30f59b5f3efabd9addbe6537fbea7c2739860405160405180910390a350565b5f80fd5b5f80fd5b5f80fd5b5f80fd5b5f80fd5b5f8083601f840112610a3f57610a3e610a1e565b5b8235905067ffffffffffffffff811115610a5c57610a5b610a22565b5b602083019150836020820283011115610a7857610a77610a26565b5b9250929050565b5f8083601f840112610a9457610a93610a1e565b5b8235905067ffffffffffffffff811115610ab157610ab0610a22565b5b602083019150836020820283011115610acd57610acc610a26565b5b9250929050565b5f805f8060408587031215610aec57610aeb610a16565b5b5f85013567ffffffffffffffff811115610b0957610b08610a1a565b5b610b1587828801610a2a565b9450945050602085013567ffffffffffffffff811115610b3857610b37610a1a565b5b610b4487828801610a7f565b925092505092959194509250565b5f80fd5b5f6101608284031215610b6c57610b6b610b52565b5b81905092915050565b5f819050919050565b610b8781610b75565b8114610b91575f80fd5b50565b5f81359050610ba281610b7e565b92915050565b5f819050919050565b610bba81610ba8565b8114610bc4575f80fd5b50565b5f81359050610bd581610bb1565b92915050565b5f805f60608486031215610bf257610bf1610a16565b5b5f84013567ffffffffffffffff811115610c0f57610c0e610a1a565b5b610c1b86828701610b56565b9350506020610c2c86828701610b94565b9250506040610c3d86828701610bc7565b9150509250925092565b610c5081610ba8565b82525050565b5f602082019050610c695f830184610c47565b92915050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f610c9882610c6f565b9050919050565b610ca881610c8e565b8114610cb2575f80fd5b50565b5f81359050610cc381610c9f565b92915050565b5f8060408385031215610cdf57610cde610a16565b5b5f610cec85828601610cb5565b9250506020610cfd85828601610bc7565b9150509250929050565b5f610d1182610c6f565b9050919050565b610d2181610d07565b82525050565b5f602082019050610d3a5f830184610d18565b92915050565b5f819050919050565b5f610d63610d5e610d5984610c6f565b610d40565b610c6f565b9050919050565b5f610d7482610d49565b9050919050565b5f610d8582610d6a565b9050919050565b610d9581610d7b565b82525050565b5f602082019050610dae5f830184610d8c565b92915050565b610dbd81610d07565b8114610dc7575f80fd5b50565b5f81359050610dd881610db4565b92915050565b5f8083601f840112610df357610df2610a1e565b5b8235905067ffffffffffffffff811115610e1057610e0f610a22565b5b602083019150836001820283011115610e2c57610e2b610a26565b5b9250929050565b5f805f8060608587031215610e4b57610e4a610a16565b5b5f610e5887828801610dca565b9450506020610e6987828801610bc7565b935050604085013567ffffffffffffffff811115610e8a57610e89610a1a565b5b610e9687828801610dde565b925092505092959194509250565b5f60208284031215610eb957610eb8610a16565b5b5f610ec684828501610dca565b91505092915050565b5f82825260208201905092915050565b7f77726f6e67206172726179206c656e67746873000000000000000000000000005f82015250565b5f610f13601383610ecf565b9150610f1e82610edf565b602082019050919050565b5f6020820190508181035f830152610f4081610f07565b9050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52603260045260245ffd5b5f80fd5b5f80fd5b5f80fd5b5f8083356001602003843603038112610f9c57610f9b610f74565b5b80840192508235915067ffffffffffffffff821115610fbe57610fbd610f78565b5b602083019250600182023603831315610fda57610fd9610f7c565b5b509250929050565b610feb81610c8e565b82525050565b5f6040820190506110045f830185610fe2565b6110116020830184610c47565b9392505050565b5f8151905061102681610bb1565b92915050565b5f6020828403121561104157611040610a16565b5b5f61104e84828501611018565b91505092915050565b5f819050919050565b5f77ffffffffffffffffffffffffffffffffffffffffffffffff82169050919050565b5f61109d61109861109384611057565b610d40565b611060565b9050919050565b6110ad81611083565b82525050565b5f6040820190506110c65f830185610d18565b6110d360208301846110a4565b9392505050565b7f6163636f756e743a206e6f74204f776e6572206f7220456e747279506f696e745f82015250565b5f61110e602083610ecf565b9150611119826110da565b602082019050919050565b5f6020820190508181035f83015261113b81611102565b9050919050565b5f81519050919050565b5f81905092915050565b5f5b83811015611173578082015181840152602081019050611158565b5f8484015250505050565b5f61118882611142565b611192818561114c565b93506111a2818560208601611156565b80840191505092915050565b5f6111b9828461117e565b915081905092915050565b7f6163636f756e743a206e6f742066726f6d20456e747279506f696e74000000005f82015250565b5f6111f8601c83610ecf565b9150611203826111c4565b602082019050919050565b5f6020820190508181035f830152611225816111ec565b9050919050565b50565b5f61123a5f8361114c565b91506112458261122c565b5f82019050919050565b5f6112598261122f565b9150819050919050565b7f6f6e6c79206f776e6572000000000000000000000000000000000000000000005f82015250565b5f611297600a83610ecf565b91506112a282611263565b602082019050919050565b5f6020820190508181035f8301526112c48161128b565b905091905056fea26469706673582212201d916021f062fe33055d01898a12f0dac530b5caec24d4856c741fc9ef3526ab64736f6c63430008180033",
}

// BindingContractABI is the input ABI used to generate the binding from.
// Deprecated: Use BindingContractMetaData.ABI instead.
var BindingContractABI = BindingContractMetaData.ABI

// BindingContractBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use BindingContractMetaData.Bin instead.
var BindingContractBin = BindingContractMetaData.Bin

// DeployBindingContract deploys a new Ethereum contract, binding an instance of BindingContract to it.
func DeployBindingContract(auth *bind.TransactOpts, backend bind.ContractBackend, anEntryPoint common.Address) (common.Address, *types.Transaction, *BindingContract, error) {
	parsed, err := BindingContractMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(BindingContractBin), backend, anEntryPoint)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &BindingContract{BindingContractCaller: BindingContractCaller{contract: contract}, BindingContractTransactor: BindingContractTransactor{contract: contract}, BindingContractFilterer: BindingContractFilterer{contract: contract}}, nil
}

// BindingContract is an auto generated Go binding around an Ethereum contract.
type BindingContract struct {
	BindingContractCaller     // Read-only binding to the contract
	BindingContractTransactor // Write-only binding to the contract
	BindingContractFilterer   // Log filterer for contract events
}

// BindingContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type BindingContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BindingContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BindingContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BindingContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BindingContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BindingContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BindingContractSession struct {
	Contract     *BindingContract  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BindingContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BindingContractCallerSession struct {
	Contract *BindingContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// BindingContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BindingContractTransactorSession struct {
	Contract     *BindingContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// BindingContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type BindingContractRaw struct {
	Contract *BindingContract // Generic contract binding to access the raw methods on
}

// BindingContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BindingContractCallerRaw struct {
	Contract *BindingContractCaller // Generic read-only contract binding to access the raw methods on
}

// BindingContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BindingContractTransactorRaw struct {
	Contract *BindingContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBindingContract creates a new instance of BindingContract, bound to a specific deployed contract.
func NewBindingContract(address common.Address, backend bind.ContractBackend) (*BindingContract, error) {
	contract, err := bindBindingContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BindingContract{BindingContractCaller: BindingContractCaller{contract: contract}, BindingContractTransactor: BindingContractTransactor{contract: contract}, BindingContractFilterer: BindingContractFilterer{contract: contract}}, nil
}

// NewBindingContractCaller creates a new read-only instance of BindingContract, bound to a specific deployed contract.
func NewBindingContractCaller(address common.Address, caller bind.ContractCaller) (*BindingContractCaller, error) {
	contract, err := bindBindingContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BindingContractCaller{contract: contract}, nil
}

// NewBindingContractTransactor creates a new write-only instance of BindingContract, bound to a specific deployed contract.
func NewBindingContractTransactor(address common.Address, transactor bind.ContractTransactor) (*BindingContractTransactor, error) {
	contract, err := bindBindingContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BindingContractTransactor{contract: contract}, nil
}

// NewBindingContractFilterer creates a new log filterer instance of BindingContract, bound to a specific deployed contract.
func NewBindingContractFilterer(address common.Address, filterer bind.ContractFilterer) (*BindingContractFilterer, error) {
	contract, err := bindBindingContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BindingContractFilterer{contract: contract}, nil
}

// bindBindingContract binds a generic wrapper to an already deployed contract.
func bindBindingContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BindingContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BindingContract *BindingContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BindingContract.Contract.BindingContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BindingContract *BindingContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BindingContract.Contract.BindingContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BindingContract *BindingContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BindingContract.Contract.BindingContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BindingContract *BindingContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BindingContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BindingContract *BindingContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BindingContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BindingContract *BindingContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BindingContract.Contract.contract.Transact(opts, method, params...)
}

// EntryPoint is a free data retrieval call binding the contract method 0xb0d691fe.
//
// Solidity: function entryPoint() view returns(address)
func (_BindingContract *BindingContractCaller) EntryPoint(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "entryPoint")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// EntryPoint is a free data retrieval call binding the contract method 0xb0d691fe.
//
// Solidity: function entryPoint() view returns(address)
func (_BindingContract *BindingContractSession) EntryPoint() (common.Address, error) {
	return _BindingContract.Contract.EntryPoint(&_BindingContract.CallOpts)
}

// EntryPoint is a free data retrieval call binding the contract method 0xb0d691fe.
//
// Solidity: function entryPoint() view returns(address)
func (_BindingContract *BindingContractCallerSession) EntryPoint() (common.Address, error) {
	return _BindingContract.Contract.EntryPoint(&_BindingContract.CallOpts)
}

// GetDeposit is a free data retrieval call binding the contract method 0xc399ec88.
//
// Solidity: function getDeposit() view returns(uint256)
func (_BindingContract *BindingContractCaller) GetDeposit(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "getDeposit")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetDeposit is a free data retrieval call binding the contract method 0xc399ec88.
//
// Solidity: function getDeposit() view returns(uint256)
func (_BindingContract *BindingContractSession) GetDeposit() (*big.Int, error) {
	return _BindingContract.Contract.GetDeposit(&_BindingContract.CallOpts)
}

// GetDeposit is a free data retrieval call binding the contract method 0xc399ec88.
//
// Solidity: function getDeposit() view returns(uint256)
func (_BindingContract *BindingContractCallerSession) GetDeposit() (*big.Int, error) {
	return _BindingContract.Contract.GetDeposit(&_BindingContract.CallOpts)
}

// GetNonce is a free data retrieval call binding the contract method 0xd087d288.
//
// Solidity: function getNonce() view returns(uint256)
func (_BindingContract *BindingContractCaller) GetNonce(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "getNonce")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetNonce is a free data retrieval call binding the contract method 0xd087d288.
//
// Solidity: function getNonce() view returns(uint256)
func (_BindingContract *BindingContractSession) GetNonce() (*big.Int, error) {
	return _BindingContract.Contract.GetNonce(&_BindingContract.CallOpts)
}

// GetNonce is a free data retrieval call binding the contract method 0xd087d288.
//
// Solidity: function getNonce() view returns(uint256)
func (_BindingContract *BindingContractCallerSession) GetNonce() (*big.Int, error) {
	return _BindingContract.Contract.GetNonce(&_BindingContract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BindingContract *BindingContractCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BindingContract *BindingContractSession) Owner() (common.Address, error) {
	return _BindingContract.Contract.Owner(&_BindingContract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BindingContract *BindingContractCallerSession) Owner() (common.Address, error) {
	return _BindingContract.Contract.Owner(&_BindingContract.CallOpts)
}

// AddDeposit is a paid mutator transaction binding the contract method 0x4a58db19.
//
// Solidity: function addDeposit() payable returns()
func (_BindingContract *BindingContractTransactor) AddDeposit(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "addDeposit")
}

// AddDeposit is a paid mutator transaction binding the contract method 0x4a58db19.
//
// Solidity: function addDeposit() payable returns()
func (_BindingContract *BindingContractSession) AddDeposit() (*types.Transaction, error) {
	return _BindingContract.Contract.AddDeposit(&_BindingContract.TransactOpts)
}

// AddDeposit is a paid mutator transaction binding the contract method 0x4a58db19.
//
// Solidity: function addDeposit() payable returns()
func (_BindingContract *BindingContractTransactorSession) AddDeposit() (*types.Transaction, error) {
	return _BindingContract.Contract.AddDeposit(&_BindingContract.TransactOpts)
}

// Execute is a paid mutator transaction binding the contract method 0xb61d27f6.
//
// Solidity: function execute(address dest, uint256 value, bytes func) returns()
func (_BindingContract *BindingContractTransactor) Execute(opts *bind.TransactOpts, dest common.Address, value *big.Int, arg2 []byte) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "execute", dest, value, arg2)
}

// Execute is a paid mutator transaction binding the contract method 0xb61d27f6.
//
// Solidity: function execute(address dest, uint256 value, bytes func) returns()
func (_BindingContract *BindingContractSession) Execute(dest common.Address, value *big.Int, arg2 []byte) (*types.Transaction, error) {
	return _BindingContract.Contract.Execute(&_BindingContract.TransactOpts, dest, value, arg2)
}

// Execute is a paid mutator transaction binding the contract method 0xb61d27f6.
//
// Solidity: function execute(address dest, uint256 value, bytes func) returns()
func (_BindingContract *BindingContractTransactorSession) Execute(dest common.Address, value *big.Int, arg2 []byte) (*types.Transaction, error) {
	return _BindingContract.Contract.Execute(&_BindingContract.TransactOpts, dest, value, arg2)
}

// ExecuteBatch is a paid mutator transaction binding the contract method 0x18dfb3c7.
//
// Solidity: function executeBatch(address[] dest, bytes[] func) returns()
func (_BindingContract *BindingContractTransactor) ExecuteBatch(opts *bind.TransactOpts, dest []common.Address, arg1 [][]byte) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "executeBatch", dest, arg1)
}

// ExecuteBatch is a paid mutator transaction binding the contract method 0x18dfb3c7.
//
// Solidity: function executeBatch(address[] dest, bytes[] func) returns()
func (_BindingContract *BindingContractSession) ExecuteBatch(dest []common.Address, arg1 [][]byte) (*types.Transaction, error) {
	return _BindingContract.Contract.ExecuteBatch(&_BindingContract.TransactOpts, dest, arg1)
}

// ExecuteBatch is a paid mutator transaction binding the contract method 0x18dfb3c7.
//
// Solidity: function executeBatch(address[] dest, bytes[] func) returns()
func (_BindingContract *BindingContractTransactorSession) ExecuteBatch(dest []common.Address, arg1 [][]byte) (*types.Transaction, error) {
	return _BindingContract.Contract.ExecuteBatch(&_BindingContract.TransactOpts, dest, arg1)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address anOwner) returns()
func (_BindingContract *BindingContractTransactor) Initialize(opts *bind.TransactOpts, anOwner common.Address) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "initialize", anOwner)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address anOwner) returns()
func (_BindingContract *BindingContractSession) Initialize(anOwner common.Address) (*types.Transaction, error) {
	return _BindingContract.Contract.Initialize(&_BindingContract.TransactOpts, anOwner)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address anOwner) returns()
func (_BindingContract *BindingContractTransactorSession) Initialize(anOwner common.Address) (*types.Transaction, error) {
	return _BindingContract.Contract.Initialize(&_BindingContract.TransactOpts, anOwner)
}

// ValidateUserOp is a paid mutator transaction binding the contract method 0x3a871cdd.
//
// Solidity: function validateUserOp((address,uint256,bytes,bytes,uint256,uint256,uint256,uint256,uint256,bytes,bytes) userOp, bytes32 userOpHash, uint256 missingAccountFunds) returns(uint256 validationData)
func (_BindingContract *BindingContractTransactor) ValidateUserOp(opts *bind.TransactOpts, userOp UserOperation, userOpHash [32]byte, missingAccountFunds *big.Int) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "validateUserOp", userOp, userOpHash, missingAccountFunds)
}

// ValidateUserOp is a paid mutator transaction binding the contract method 0x3a871cdd.
//
// Solidity: function validateUserOp((address,uint256,bytes,bytes,uint256,uint256,uint256,uint256,uint256,bytes,bytes) userOp, bytes32 userOpHash, uint256 missingAccountFunds) returns(uint256 validationData)
func (_BindingContract *BindingContractSession) ValidateUserOp(userOp UserOperation, userOpHash [32]byte, missingAccountFunds *big.Int) (*types.Transaction, error) {
	return _BindingContract.Contract.ValidateUserOp(&_BindingContract.TransactOpts, userOp, userOpHash, missingAccountFunds)
}

// ValidateUserOp is a paid mutator transaction binding the contract method 0x3a871cdd.
//
// Solidity: function validateUserOp((address,uint256,bytes,bytes,uint256,uint256,uint256,uint256,uint256,bytes,bytes) userOp, bytes32 userOpHash, uint256 missingAccountFunds) returns(uint256 validationData)
func (_BindingContract *BindingContractTransactorSession) ValidateUserOp(userOp UserOperation, userOpHash [32]byte, missingAccountFunds *big.Int) (*types.Transaction, error) {
	return _BindingContract.Contract.ValidateUserOp(&_BindingContract.TransactOpts, userOp, userOpHash, missingAccountFunds)
}

// WithdrawDepositTo is a paid mutator transaction binding the contract method 0x4d44560d.
//
// Solidity: function withdrawDepositTo(address withdrawAddress, uint256 amount) returns()
func (_BindingContract *BindingContractTransactor) WithdrawDepositTo(opts *bind.TransactOpts, withdrawAddress common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "withdrawDepositTo", withdrawAddress, amount)
}

// WithdrawDepositTo is a paid mutator transaction binding the contract method 0x4d44560d.
//
// Solidity: function withdrawDepositTo(address withdrawAddress, uint256 amount) returns()
func (_BindingContract *BindingContractSession) WithdrawDepositTo(withdrawAddress common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BindingContract.Contract.WithdrawDepositTo(&_BindingContract.TransactOpts, withdrawAddress, amount)
}

// WithdrawDepositTo is a paid mutator transaction binding the contract method 0x4d44560d.
//
// Solidity: function withdrawDepositTo(address withdrawAddress, uint256 amount) returns()
func (_BindingContract *BindingContractTransactorSession) WithdrawDepositTo(withdrawAddress common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BindingContract.Contract.WithdrawDepositTo(&_BindingContract.TransactOpts, withdrawAddress, amount)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_BindingContract *BindingContractTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BindingContract.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_BindingContract *BindingContractSession) Receive() (*types.Transaction, error) {
	return _BindingContract.Contract.Receive(&_BindingContract.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_BindingContract *BindingContractTransactorSession) Receive() (*types.Transaction, error) {
	return _BindingContract.Contract.Receive(&_BindingContract.TransactOpts)
}

// BindingContractSmartAccountInitializedIterator is returned from FilterSmartAccountInitialized and is used to iterate over the raw logs and unpacked data for SmartAccountInitialized events raised by the BindingContract contract.
type BindingContractSmartAccountInitializedIterator struct {
	Event *BindingContractSmartAccountInitialized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BindingContractSmartAccountInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingContractSmartAccountInitialized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BindingContractSmartAccountInitialized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BindingContractSmartAccountInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingContractSmartAccountInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingContractSmartAccountInitialized represents a SmartAccountInitialized event raised by the BindingContract contract.
type BindingContractSmartAccountInitialized struct {
	EntryPoint common.Address
	Owner      common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterSmartAccountInitialized is a free log retrieval operation binding the contract event 0xb7053def2fe3d2a5ecb12939fbfcc30f59b5f3efabd9addbe6537fbea7c27398.
//
// Solidity: event SmartAccountInitialized(address indexed entryPoint, address indexed owner)
func (_BindingContract *BindingContractFilterer) FilterSmartAccountInitialized(opts *bind.FilterOpts, entryPoint []common.Address, owner []common.Address) (*BindingContractSmartAccountInitializedIterator, error) {

	var entryPointRule []interface{}
	for _, entryPointItem := range entryPoint {
		entryPointRule = append(entryPointRule, entryPointItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _BindingContract.contract.FilterLogs(opts, "SmartAccountInitialized", entryPointRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &BindingContractSmartAccountInitializedIterator{contract: _BindingContract.contract, event: "SmartAccountInitialized", logs: logs, sub: sub}, nil
}

// WatchSmartAccountInitialized is a free log subscription operation binding the contract event 0xb7053def2fe3d2a5ecb12939fbfcc30f59b5f3efabd9addbe6537fbea7c27398.
//
// Solidity: event SmartAccountInitialized(address indexed entryPoint, address indexed owner)
func (_BindingContract *BindingContractFilterer) WatchSmartAccountInitialized(opts *bind.WatchOpts, sink chan<- *BindingContractSmartAccountInitialized, entryPoint []common.Address, owner []common.Address) (event.Subscription, error) {

	var entryPointRule []interface{}
	for _, entryPointItem := range entryPoint {
		entryPointRule = append(entryPointRule, entryPointItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _BindingContract.contract.WatchLogs(opts, "SmartAccountInitialized", entryPointRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingContractSmartAccountInitialized)
				if err := _BindingContract.contract.UnpackLog(event, "SmartAccountInitialized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSmartAccountInitialized is a log parse operation binding the contract event 0xb7053def2fe3d2a5ecb12939fbfcc30f59b5f3efabd9addbe6537fbea7c27398.
//
// Solidity: event SmartAccountInitialized(address indexed entryPoint, address indexed owner)
func (_BindingContract *BindingContractFilterer) ParseSmartAccountInitialized(log types.Log) (*BindingContractSmartAccountInitialized, error) {
	event := new(BindingContractSmartAccountInitialized)
	if err := _BindingContract.contract.UnpackLog(event, "SmartAccountInitialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
