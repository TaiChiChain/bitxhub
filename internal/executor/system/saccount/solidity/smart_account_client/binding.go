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
	AuthData             []byte
	ClientData           []byte
}

// BindingContractMetaData contains all meta data concerning the BindingContract contract.
var BindingContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIEntryPoint\",\"name\":\"anEntryPoint\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"contractIEntryPoint\",\"name\":\"entryPoint\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"SmartAccountInitialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"lockedTime\",\"type\":\"uint256\"}],\"name\":\"UserLocked\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"addDeposit\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"entryPoint\",\"outputs\":[{\"internalType\":\"contractIEntryPoint\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"dest\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"func\",\"type\":\"bytes\"}],\"name\":\"execute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"dest\",\"type\":\"address[]\"},{\"internalType\":\"bytes[]\",\"name\":\"func\",\"type\":\"bytes[]\"}],\"name\":\"executeBatch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getDeposit\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getNonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"anOwner\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"resetOwner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"setGuardian\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"},{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"name\":\"setPasskey\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"name\":\"setSession\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"initCode\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"callData\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"callGasLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"verificationGasLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"preVerificationGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxPriorityFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"paymasterAndData\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"authData\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"clientData\",\"type\":\"bytes\"}],\"internalType\":\"structUserOperation\",\"name\":\"userOp\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"userOpHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"missingAccountFunds\",\"type\":\"uint256\"}],\"name\":\"validateUserOp\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"validationData\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"withdrawAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"withdrawDepositTo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
	Bin: "0x60a060405234801562000010575f80fd5b506040516200168b3803806200168b8339818101604052810190620000369190620000e9565b8073ffffffffffffffffffffffffffffffffffffffff1660808173ffffffffffffffffffffffffffffffffffffffff16815250505062000119565b5f80fd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f620000a08262000075565b9050919050565b5f620000b38262000094565b9050919050565b620000c581620000a7565b8114620000d0575f80fd5b50565b5f81519050620000e381620000ba565b92915050565b5f6020828403121562000101576200010062000071565b5b5f6200011084828501620000d3565b91505092915050565b608051611552620001395f395f81816105650152610acf01526115525ff3fe6080604052600436106100e0575f3560e01c8063a3ca5fe31161007e578063b61d27f611610058578063b61d27f614610275578063c399ec881461029d578063c4d66de8146102c7578063d087d288146102ef576100e7565b8063a3ca5fe3146101e7578063b0d691fe1461020f578063b0fff5ca14610239576100e7565b80634d44560d116100ba5780634d44560d1461014557806373cc802a1461016d5780638a0dac4a146101955780638da5cb5b146101bd576100e7565b806318dfb3c7146100eb57806339d8bab5146101135780634a58db191461013b576100e7565b366100e757005b5f80fd5b3480156100f6575f80fd5b50610111600480360381019061010c9190610bf1565b610319565b005b34801561011e575f80fd5b5061013960048036038101906101349190610cfa565b610422565b005b61014361042f565b005b348015610150575f80fd5b5061016b60048036038101906101669190610de4565b6104a0565b005b348015610178575f80fd5b50610193600480360381019061018e9190610e5d565b61051b565b005b3480156101a0575f80fd5b506101bb60048036038101906101b69190610e5d565b610526565b005b3480156101c8575f80fd5b506101d1610531565b6040516101de9190610e97565b60405180910390f35b3480156101f2575f80fd5b5061020d60048036038101906102089190610eed565b610554565b005b34801561021a575f80fd5b50610223610562565b6040516102309190610fac565b60405180910390f35b348015610244575f80fd5b5061025f600480360381019061025a919061101b565b610589565b60405161026c9190611096565b60405180910390f35b348015610280575f80fd5b5061029b600480360381019061029691906110af565b6105bb565b005b3480156102a8575f80fd5b506102b1610617565b6040516102be9190611096565b60405180910390f35b3480156102d2575f80fd5b506102ed60048036038101906102e89190610e5d565b61069c565b005b3480156102fa575f80fd5b506103036106a8565b6040516103109190611096565b60405180910390f35b61032161072f565b818190508484905014610369576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016103609061117a565b60405180910390fd5b5f5b8484905081101561041b5761040e85858381811061038c5761038b611198565b5b90506020020160208101906103a19190610e5d565b5f8585858181106103b5576103b4611198565b5b90506020028101906103c791906111d1565b8080601f0160208091040260200160405190810160405280939291908181526020018383808284375f81840152601f19601f820116905080830192505050505050506107fb565b808060010191505061036b565b5050505050565b61042a61087b565b505050565b610437610562565b73ffffffffffffffffffffffffffffffffffffffff1663b760faf934306040518363ffffffff1660e01b81526004016104709190610e97565b5f604051808303818588803b158015610487575f80fd5b505af1158015610499573d5f803e3d5ffd5b5050505050565b6104a861087b565b6104b0610562565b73ffffffffffffffffffffffffffffffffffffffff1663205c287883836040518363ffffffff1660e01b81526004016104ea929190611242565b5f604051808303815f87803b158015610501575f80fd5b505af1158015610513573d5f803e3d5ffd5b505050505050565b61052361087b565b50565b61052e61087b565b50565b5f8054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b61055c61087b565b50505050565b5f7f0000000000000000000000000000000000000000000000000000000000000000905090565b5f610592610940565b61059c84846109b7565b90506105ab84602001356109be565b6105b4826109c1565b9392505050565b6105c361072f565b610611848484848080601f0160208091040260200160405190810160405280939291908181526020018383808284375f81840152601f19601f820116905080830192505050505050506107fb565b50505050565b5f610620610562565b73ffffffffffffffffffffffffffffffffffffffff166370a08231306040518263ffffffff1660e01b81526004016106589190610e97565b602060405180830381865afa158015610673573d5f803e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610697919061127d565b905090565b6106a581610a58565b50565b5f6106b1610562565b73ffffffffffffffffffffffffffffffffffffffff166335567e1a305f6040518363ffffffff1660e01b81526004016106eb929190611304565b602060405180830381865afa158015610706573d5f803e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061072a919061127d565b905090565b610737610562565b73ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614806107ba57505f8054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16145b6107f9576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016107f090611375565b60405180910390fd5b565b5f808473ffffffffffffffffffffffffffffffffffffffff16848460405161082391906113ff565b5f6040518083038185875af1925050503d805f811461085d576040519150601f19603f3d011682016040523d82523d5f602084013e610862565b606091505b50915091508161087457805160208201fd5b5050505050565b5f8054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614806108ff57503073ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16145b61093e576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016109359061145f565b60405180910390fd5b565b610948610562565b73ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146109b5576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016109ac906114c7565b60405180910390fd5b565b5f92915050565b50565b5f8114610a55575f3373ffffffffffffffffffffffffffffffffffffffff16827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff90604051610a0f90611508565b5f60405180830381858888f193505050503d805f8114610a4a576040519150601f19603f3d011682016040523d82523d5f602084013e610a4f565b606091505b50509050505b50565b805f806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505f8054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff167f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff167fb7053def2fe3d2a5ecb12939fbfcc30f59b5f3efabd9addbe6537fbea7c2739860405160405180910390a350565b5f80fd5b5f80fd5b5f80fd5b5f80fd5b5f80fd5b5f8083601f840112610b5c57610b5b610b3b565b5b8235905067ffffffffffffffff811115610b7957610b78610b3f565b5b602083019150836020820283011115610b9557610b94610b43565b5b9250929050565b5f8083601f840112610bb157610bb0610b3b565b5b8235905067ffffffffffffffff811115610bce57610bcd610b3f565b5b602083019150836020820283011115610bea57610be9610b43565b5b9250929050565b5f805f8060408587031215610c0957610c08610b33565b5b5f85013567ffffffffffffffff811115610c2657610c25610b37565b5b610c3287828801610b47565b9450945050602085013567ffffffffffffffff811115610c5557610c54610b37565b5b610c6187828801610b9c565b925092505092959194509250565b5f8083601f840112610c8457610c83610b3b565b5b8235905067ffffffffffffffff811115610ca157610ca0610b3f565b5b602083019150836001820283011115610cbd57610cbc610b43565b5b9250929050565b5f60ff82169050919050565b610cd981610cc4565b8114610ce3575f80fd5b50565b5f81359050610cf481610cd0565b92915050565b5f805f60408486031215610d1157610d10610b33565b5b5f84013567ffffffffffffffff811115610d2e57610d2d610b37565b5b610d3a86828701610c6f565b93509350506020610d4d86828701610ce6565b9150509250925092565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f610d8082610d57565b9050919050565b610d9081610d76565b8114610d9a575f80fd5b50565b5f81359050610dab81610d87565b92915050565b5f819050919050565b610dc381610db1565b8114610dcd575f80fd5b50565b5f81359050610dde81610dba565b92915050565b5f8060408385031215610dfa57610df9610b33565b5b5f610e0785828601610d9d565b9250506020610e1885828601610dd0565b9150509250929050565b5f610e2c82610d57565b9050919050565b610e3c81610e22565b8114610e46575f80fd5b50565b5f81359050610e5781610e33565b92915050565b5f60208284031215610e7257610e71610b33565b5b5f610e7f84828501610e49565b91505092915050565b610e9181610e22565b82525050565b5f602082019050610eaa5f830184610e88565b92915050565b5f67ffffffffffffffff82169050919050565b610ecc81610eb0565b8114610ed6575f80fd5b50565b5f81359050610ee781610ec3565b92915050565b5f805f8060808587031215610f0557610f04610b33565b5b5f610f1287828801610e49565b9450506020610f2387828801610dd0565b9350506040610f3487828801610ed9565b9250506060610f4587828801610ed9565b91505092959194509250565b5f819050919050565b5f610f74610f6f610f6a84610d57565b610f51565b610d57565b9050919050565b5f610f8582610f5a565b9050919050565b5f610f9682610f7b565b9050919050565b610fa681610f8c565b82525050565b5f602082019050610fbf5f830184610f9d565b92915050565b5f80fd5b5f6101a08284031215610fdf57610fde610fc5565b5b81905092915050565b5f819050919050565b610ffa81610fe8565b8114611004575f80fd5b50565b5f8135905061101581610ff1565b92915050565b5f805f6060848603121561103257611031610b33565b5b5f84013567ffffffffffffffff81111561104f5761104e610b37565b5b61105b86828701610fc9565b935050602061106c86828701611007565b925050604061107d86828701610dd0565b9150509250925092565b61109081610db1565b82525050565b5f6020820190506110a95f830184611087565b92915050565b5f805f80606085870312156110c7576110c6610b33565b5b5f6110d487828801610e49565b94505060206110e587828801610dd0565b935050604085013567ffffffffffffffff81111561110657611105610b37565b5b61111287828801610c6f565b925092505092959194509250565b5f82825260208201905092915050565b7f77726f6e67206172726179206c656e67746873000000000000000000000000005f82015250565b5f611164601383611120565b915061116f82611130565b602082019050919050565b5f6020820190508181035f83015261119181611158565b9050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52603260045260245ffd5b5f80fd5b5f80fd5b5f80fd5b5f80833560016020038436030381126111ed576111ec6111c5565b5b80840192508235915067ffffffffffffffff82111561120f5761120e6111c9565b5b60208301925060018202360383131561122b5761122a6111cd565b5b509250929050565b61123c81610d76565b82525050565b5f6040820190506112555f830185611233565b6112626020830184611087565b9392505050565b5f8151905061127781610dba565b92915050565b5f6020828403121561129257611291610b33565b5b5f61129f84828501611269565b91505092915050565b5f819050919050565b5f77ffffffffffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6112ee6112e96112e4846112a8565b610f51565b6112b1565b9050919050565b6112fe816112d4565b82525050565b5f6040820190506113175f830185610e88565b61132460208301846112f5565b9392505050565b7f6163636f756e743a206e6f74204f776e6572206f7220456e747279506f696e745f82015250565b5f61135f602083611120565b915061136a8261132b565b602082019050919050565b5f6020820190508181035f83015261138c81611353565b9050919050565b5f81519050919050565b5f81905092915050565b5f5b838110156113c45780820151818401526020810190506113a9565b5f8484015250505050565b5f6113d982611393565b6113e3818561139d565b93506113f38185602086016113a7565b80840191505092915050565b5f61140a82846113cf565b915081905092915050565b7f6f6e6c79206f776e6572000000000000000000000000000000000000000000005f82015250565b5f611449600a83611120565b915061145482611415565b602082019050919050565b5f6020820190508181035f8301526114768161143d565b9050919050565b7f6163636f756e743a206e6f742066726f6d20456e747279506f696e74000000005f82015250565b5f6114b1601c83611120565b91506114bc8261147d565b602082019050919050565b5f6020820190508181035f8301526114de816114a5565b9050919050565b50565b5f6114f35f8361139d565b91506114fe826114e5565b5f82019050919050565b5f611512826114e8565b915081905091905056fea2646970667358221220089b0959409683d6dcb58dfa9e2e7c1abfdadd13a2ade74a9bd6d7b0fb45499964736f6c63430008170033",
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

// ResetOwner is a paid mutator transaction binding the contract method 0x73cc802a.
//
// Solidity: function resetOwner(address ) returns()
func (_BindingContract *BindingContractTransactor) ResetOwner(opts *bind.TransactOpts, arg0 common.Address) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "resetOwner", arg0)
}

// ResetOwner is a paid mutator transaction binding the contract method 0x73cc802a.
//
// Solidity: function resetOwner(address ) returns()
func (_BindingContract *BindingContractSession) ResetOwner(arg0 common.Address) (*types.Transaction, error) {
	return _BindingContract.Contract.ResetOwner(&_BindingContract.TransactOpts, arg0)
}

// ResetOwner is a paid mutator transaction binding the contract method 0x73cc802a.
//
// Solidity: function resetOwner(address ) returns()
func (_BindingContract *BindingContractTransactorSession) ResetOwner(arg0 common.Address) (*types.Transaction, error) {
	return _BindingContract.Contract.ResetOwner(&_BindingContract.TransactOpts, arg0)
}

// SetGuardian is a paid mutator transaction binding the contract method 0x8a0dac4a.
//
// Solidity: function setGuardian(address ) returns()
func (_BindingContract *BindingContractTransactor) SetGuardian(opts *bind.TransactOpts, arg0 common.Address) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "setGuardian", arg0)
}

// SetGuardian is a paid mutator transaction binding the contract method 0x8a0dac4a.
//
// Solidity: function setGuardian(address ) returns()
func (_BindingContract *BindingContractSession) SetGuardian(arg0 common.Address) (*types.Transaction, error) {
	return _BindingContract.Contract.SetGuardian(&_BindingContract.TransactOpts, arg0)
}

// SetGuardian is a paid mutator transaction binding the contract method 0x8a0dac4a.
//
// Solidity: function setGuardian(address ) returns()
func (_BindingContract *BindingContractTransactorSession) SetGuardian(arg0 common.Address) (*types.Transaction, error) {
	return _BindingContract.Contract.SetGuardian(&_BindingContract.TransactOpts, arg0)
}

// SetPasskey is a paid mutator transaction binding the contract method 0x39d8bab5.
//
// Solidity: function setPasskey(bytes , uint8 ) returns()
func (_BindingContract *BindingContractTransactor) SetPasskey(opts *bind.TransactOpts, arg0 []byte, arg1 uint8) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "setPasskey", arg0, arg1)
}

// SetPasskey is a paid mutator transaction binding the contract method 0x39d8bab5.
//
// Solidity: function setPasskey(bytes , uint8 ) returns()
func (_BindingContract *BindingContractSession) SetPasskey(arg0 []byte, arg1 uint8) (*types.Transaction, error) {
	return _BindingContract.Contract.SetPasskey(&_BindingContract.TransactOpts, arg0, arg1)
}

// SetPasskey is a paid mutator transaction binding the contract method 0x39d8bab5.
//
// Solidity: function setPasskey(bytes , uint8 ) returns()
func (_BindingContract *BindingContractTransactorSession) SetPasskey(arg0 []byte, arg1 uint8) (*types.Transaction, error) {
	return _BindingContract.Contract.SetPasskey(&_BindingContract.TransactOpts, arg0, arg1)
}

// SetSession is a paid mutator transaction binding the contract method 0xa3ca5fe3.
//
// Solidity: function setSession(address , uint256 , uint64 , uint64 ) returns()
func (_BindingContract *BindingContractTransactor) SetSession(opts *bind.TransactOpts, arg0 common.Address, arg1 *big.Int, arg2 uint64, arg3 uint64) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "setSession", arg0, arg1, arg2, arg3)
}

// SetSession is a paid mutator transaction binding the contract method 0xa3ca5fe3.
//
// Solidity: function setSession(address , uint256 , uint64 , uint64 ) returns()
func (_BindingContract *BindingContractSession) SetSession(arg0 common.Address, arg1 *big.Int, arg2 uint64, arg3 uint64) (*types.Transaction, error) {
	return _BindingContract.Contract.SetSession(&_BindingContract.TransactOpts, arg0, arg1, arg2, arg3)
}

// SetSession is a paid mutator transaction binding the contract method 0xa3ca5fe3.
//
// Solidity: function setSession(address , uint256 , uint64 , uint64 ) returns()
func (_BindingContract *BindingContractTransactorSession) SetSession(arg0 common.Address, arg1 *big.Int, arg2 uint64, arg3 uint64) (*types.Transaction, error) {
	return _BindingContract.Contract.SetSession(&_BindingContract.TransactOpts, arg0, arg1, arg2, arg3)
}

// ValidateUserOp is a paid mutator transaction binding the contract method 0xb0fff5ca.
//
// Solidity: function validateUserOp((address,uint256,bytes,bytes,uint256,uint256,uint256,uint256,uint256,bytes,bytes,bytes,bytes) userOp, bytes32 userOpHash, uint256 missingAccountFunds) returns(uint256 validationData)
func (_BindingContract *BindingContractTransactor) ValidateUserOp(opts *bind.TransactOpts, userOp UserOperation, userOpHash [32]byte, missingAccountFunds *big.Int) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "validateUserOp", userOp, userOpHash, missingAccountFunds)
}

// ValidateUserOp is a paid mutator transaction binding the contract method 0xb0fff5ca.
//
// Solidity: function validateUserOp((address,uint256,bytes,bytes,uint256,uint256,uint256,uint256,uint256,bytes,bytes,bytes,bytes) userOp, bytes32 userOpHash, uint256 missingAccountFunds) returns(uint256 validationData)
func (_BindingContract *BindingContractSession) ValidateUserOp(userOp UserOperation, userOpHash [32]byte, missingAccountFunds *big.Int) (*types.Transaction, error) {
	return _BindingContract.Contract.ValidateUserOp(&_BindingContract.TransactOpts, userOp, userOpHash, missingAccountFunds)
}

// ValidateUserOp is a paid mutator transaction binding the contract method 0xb0fff5ca.
//
// Solidity: function validateUserOp((address,uint256,bytes,bytes,uint256,uint256,uint256,uint256,uint256,bytes,bytes,bytes,bytes) userOp, bytes32 userOpHash, uint256 missingAccountFunds) returns(uint256 validationData)
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

// BindingContractUserLockedIterator is returned from FilterUserLocked and is used to iterate over the raw logs and unpacked data for UserLocked events raised by the BindingContract contract.
type BindingContractUserLockedIterator struct {
	Event *BindingContractUserLocked // Event containing the contract specifics and raw log

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
func (it *BindingContractUserLockedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingContractUserLocked)
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
		it.Event = new(BindingContractUserLocked)
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
func (it *BindingContractUserLockedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingContractUserLockedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingContractUserLocked represents a UserLocked event raised by the BindingContract contract.
type BindingContractUserLocked struct {
	Sender     common.Address
	NewOwner   common.Address
	LockedTime *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterUserLocked is a free log retrieval operation binding the contract event 0xae6bdfca4d8c6bf2a59efb0bbb7f8bdecd8795878222b0d7660413a29c4531ec.
//
// Solidity: event UserLocked(address indexed sender, address indexed newOwner, uint256 lockedTime)
func (_BindingContract *BindingContractFilterer) FilterUserLocked(opts *bind.FilterOpts, sender []common.Address, newOwner []common.Address) (*BindingContractUserLockedIterator, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BindingContract.contract.FilterLogs(opts, "UserLocked", senderRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BindingContractUserLockedIterator{contract: _BindingContract.contract, event: "UserLocked", logs: logs, sub: sub}, nil
}

// WatchUserLocked is a free log subscription operation binding the contract event 0xae6bdfca4d8c6bf2a59efb0bbb7f8bdecd8795878222b0d7660413a29c4531ec.
//
// Solidity: event UserLocked(address indexed sender, address indexed newOwner, uint256 lockedTime)
func (_BindingContract *BindingContractFilterer) WatchUserLocked(opts *bind.WatchOpts, sink chan<- *BindingContractUserLocked, sender []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BindingContract.contract.WatchLogs(opts, "UserLocked", senderRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingContractUserLocked)
				if err := _BindingContract.contract.UnpackLog(event, "UserLocked", log); err != nil {
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

// ParseUserLocked is a log parse operation binding the contract event 0xae6bdfca4d8c6bf2a59efb0bbb7f8bdecd8795878222b0d7660413a29c4531ec.
//
// Solidity: event UserLocked(address indexed sender, address indexed newOwner, uint256 lockedTime)
func (_BindingContract *BindingContractFilterer) ParseUserLocked(log types.Log) (*BindingContractUserLocked, error) {
	event := new(BindingContractUserLocked)
	if err := _BindingContract.contract.UnpackLog(event, "UserLocked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
