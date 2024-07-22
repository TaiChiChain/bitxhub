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

// PassKey is an auto generated low-level Go binding around an user-defined struct.
type PassKey struct {
	PubKeyX *big.Int
	PubKeyY *big.Int
	Algo    uint8
}

// SessionKey is an auto generated low-level Go binding around an user-defined struct.
type SessionKey struct {
	Addr          common.Address
	SpendingLimit *big.Int
	SpentAmount   *big.Int
	ValidUntil    uint64
	ValidAfter    uint64
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
	AuthData             []byte
	ClientData           []byte
}

// BindingContractMetaData contains all meta data concerning the BindingContract contract.
var BindingContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIEntryPoint\",\"name\":\"anEntryPoint\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"contractIEntryPoint\",\"name\":\"entryPoint\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"SmartAccountInitialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"lockedTime\",\"type\":\"uint256\"}],\"name\":\"UserLocked\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"entryPoint\",\"outputs\":[{\"internalType\":\"contractIEntryPoint\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"dest\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"func\",\"type\":\"bytes\"}],\"name\":\"execute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"dest\",\"type\":\"address[]\"},{\"internalType\":\"bytes[]\",\"name\":\"func\",\"type\":\"bytes[]\"}],\"name\":\"executeBatch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getGuardian\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getNonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getPasskeys\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"pubKeyX\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pubKeyY\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"algo\",\"type\":\"uint8\"}],\"internalType\":\"structPassKey[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getSessions\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"spendingLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"spentAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint64\",\"name\":\"validUntil\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"validAfter\",\"type\":\"uint64\"}],\"internalType\":\"structSessionKey[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getStatus\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"anOwner\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"resetOwner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"setGuardian\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"},{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"name\":\"setPasskey\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"name\":\"setSession\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"initCode\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"callData\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"callGasLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"verificationGasLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"preVerificationGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxPriorityFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"paymasterAndData\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"authData\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"clientData\",\"type\":\"bytes\"}],\"internalType\":\"structUserOperation\",\"name\":\"userOp\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"userOpHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"missingAccountFunds\",\"type\":\"uint256\"}],\"name\":\"validateUserOp\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"validationData\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
	Bin: "0x60a060405234801562000010575f80fd5b50604051620017a7380380620017a78339818101604052810190620000369190620000e9565b8073ffffffffffffffffffffffffffffffffffffffff1660808173ffffffffffffffffffffffffffffffffffffffff16815250505062000119565b5f80fd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f620000a08262000075565b9050919050565b5f620000b38262000094565b9050919050565b620000c581620000a7565b8114620000d0575f80fd5b50565b5f81519050620000e381620000ba565b92915050565b5f6020828403121562000101576200010062000071565b5b5f6200011084828501620000d3565b91505092915050565b60805161166e620001395f395f81816105160152610a00015261166e5ff3fe6080604052600436106100f6575f3560e01c8063a3ca5fe311610089578063b61d27f611610058578063b61d27f614610301578063c4d66de814610329578063d087d28814610351578063e4093c221461037b576100fd565b8063a3ca5fe314610249578063a75b87d214610271578063b0d691fe1461029b578063b0fff5ca146102c5576100fd565b806373cc802a116100c557806373cc802a146101a5578063893d20e8146101cd5780638a0dac4a146101f75780638da5cb5b1461021f576100fd565b806318dfb3c71461010157806339d8bab5146101295780634e69d5601461015157806361503e451461017b576100fd565b366100fd57005b5f80fd5b34801561010c575f80fd5b5061012760048036038101906101229190610b22565b6103a5565b005b348015610134575f80fd5b5061014f600480360381019061014a9190610c2b565b6104ae565b005b34801561015c575f80fd5b506101656104bb565b6040516101729190610caa565b60405180910390f35b348015610186575f80fd5b5061018f6104bf565b60405161019c9190610e37565b60405180910390f35b3480156101b0575f80fd5b506101cb60048036038101906101c69190610e81565b6104c4565b005b3480156101d8575f80fd5b506101e16104cf565b6040516101ee9190610ebb565b60405180910390f35b348015610202575f80fd5b5061021d60048036038101906102189190610e81565b6104d3565b005b34801561022a575f80fd5b506102336104de565b6040516102409190610ebb565b60405180910390f35b348015610254575f80fd5b5061026f600480360381019061026a9190610f28565b610501565b005b34801561027c575f80fd5b5061028561050f565b6040516102929190610ebb565b60405180910390f35b3480156102a6575f80fd5b506102af610513565b6040516102bc9190610fe7565b60405180910390f35b3480156102d0575f80fd5b506102eb60048036038101906102e69190611056565b61053a565b6040516102f891906110d1565b60405180910390f35b34801561030c575f80fd5b50610327600480360381019061032291906110ea565b61056c565b005b348015610334575f80fd5b5061034f600480360381019061034a9190610e81565b6105c8565b005b34801561035c575f80fd5b506103656105d4565b60405161037291906110d1565b60405180910390f35b348015610386575f80fd5b5061038f61065b565b60405161039c9190611252565b60405180910390f35b6103ad610660565b8181905084849050146103f5576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016103ec906112cc565b60405180910390fd5b5f5b848490508110156104a75761049a858583818110610418576104176112ea565b5b905060200201602081019061042d9190610e81565b5f858585818110610441576104406112ea565b5b90506020028101906104539190611323565b8080601f0160208091040260200160405190810160405280939291908181526020018383808284375f81840152601f19601f8201169050808301925050505050505061072c565b80806001019150506103f7565b5050505050565b6104b66107ac565b505050565b5f90565b606090565b6104cc6107ac565b50565b5f90565b6104db6107ac565b50565b5f8054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b6105096107ac565b50505050565b5f90565b5f7f0000000000000000000000000000000000000000000000000000000000000000905090565b5f610543610871565b61054d84846108e8565b905061055c84602001356108ef565b610565826108f2565b9392505050565b610574610660565b6105c2848484848080601f0160208091040260200160405190810160405280939291908181526020018383808284375f81840152601f19601f8201169050808301925050505050505061072c565b50505050565b6105d181610989565b50565b5f6105dd610513565b73ffffffffffffffffffffffffffffffffffffffff166335567e1a305f6040518363ffffffff1660e01b81526004016106179291906113e1565b602060405180830381865afa158015610632573d5f803e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610656919061141c565b905090565b606090565b610668610513565b73ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614806106eb57505f8054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16145b61072a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161072190611491565b60405180910390fd5b565b5f808473ffffffffffffffffffffffffffffffffffffffff168484604051610754919061151b565b5f6040518083038185875af1925050503d805f811461078e576040519150601f19603f3d011682016040523d82523d5f602084013e610793565b606091505b5091509150816107a557805160208201fd5b5050505050565b5f8054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16148061083057503073ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16145b61086f576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016108669061157b565b60405180910390fd5b565b610879610513565b73ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146108e6576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016108dd906115e3565b60405180910390fd5b565b5f92915050565b50565b5f8114610986575f3373ffffffffffffffffffffffffffffffffffffffff16827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff9060405161094090611624565b5f60405180830381858888f193505050503d805f811461097b576040519150601f19603f3d011682016040523d82523d5f602084013e610980565b606091505b50509050505b50565b805f806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505f8054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff167f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff167fb7053def2fe3d2a5ecb12939fbfcc30f59b5f3efabd9addbe6537fbea7c2739860405160405180910390a350565b5f80fd5b5f80fd5b5f80fd5b5f80fd5b5f80fd5b5f8083601f840112610a8d57610a8c610a6c565b5b8235905067ffffffffffffffff811115610aaa57610aa9610a70565b5b602083019150836020820283011115610ac657610ac5610a74565b5b9250929050565b5f8083601f840112610ae257610ae1610a6c565b5b8235905067ffffffffffffffff811115610aff57610afe610a70565b5b602083019150836020820283011115610b1b57610b1a610a74565b5b9250929050565b5f805f8060408587031215610b3a57610b39610a64565b5b5f85013567ffffffffffffffff811115610b5757610b56610a68565b5b610b6387828801610a78565b9450945050602085013567ffffffffffffffff811115610b8657610b85610a68565b5b610b9287828801610acd565b925092505092959194509250565b5f8083601f840112610bb557610bb4610a6c565b5b8235905067ffffffffffffffff811115610bd257610bd1610a70565b5b602083019150836001820283011115610bee57610bed610a74565b5b9250929050565b5f60ff82169050919050565b610c0a81610bf5565b8114610c14575f80fd5b50565b5f81359050610c2581610c01565b92915050565b5f805f60408486031215610c4257610c41610a64565b5b5f84013567ffffffffffffffff811115610c5f57610c5e610a68565b5b610c6b86828701610ba0565b93509350506020610c7e86828701610c17565b9150509250925092565b5f67ffffffffffffffff82169050919050565b610ca481610c88565b82525050565b5f602082019050610cbd5f830184610c9b565b92915050565b5f81519050919050565b5f82825260208201905092915050565b5f819050602082019050919050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f610d1582610cec565b9050919050565b610d2581610d0b565b82525050565b5f819050919050565b610d3d81610d2b565b82525050565b610d4c81610c88565b82525050565b60a082015f820151610d665f850182610d1c565b506020820151610d796020850182610d34565b506040820151610d8c6040850182610d34565b506060820151610d9f6060850182610d43565b506080820151610db26080850182610d43565b50505050565b5f610dc38383610d52565b60a08301905092915050565b5f602082019050919050565b5f610de582610cc3565b610def8185610ccd565b9350610dfa83610cdd565b805f5b83811015610e2a578151610e118882610db8565b9750610e1c83610dcf565b925050600181019050610dfd565b5085935050505092915050565b5f6020820190508181035f830152610e4f8184610ddb565b905092915050565b610e6081610d0b565b8114610e6a575f80fd5b50565b5f81359050610e7b81610e57565b92915050565b5f60208284031215610e9657610e95610a64565b5b5f610ea384828501610e6d565b91505092915050565b610eb581610d0b565b82525050565b5f602082019050610ece5f830184610eac565b92915050565b610edd81610d2b565b8114610ee7575f80fd5b50565b5f81359050610ef881610ed4565b92915050565b610f0781610c88565b8114610f11575f80fd5b50565b5f81359050610f2281610efe565b92915050565b5f805f8060808587031215610f4057610f3f610a64565b5b5f610f4d87828801610e6d565b9450506020610f5e87828801610eea565b9350506040610f6f87828801610f14565b9250506060610f8087828801610f14565b91505092959194509250565b5f819050919050565b5f610faf610faa610fa584610cec565b610f8c565b610cec565b9050919050565b5f610fc082610f95565b9050919050565b5f610fd182610fb6565b9050919050565b610fe181610fc7565b82525050565b5f602082019050610ffa5f830184610fd8565b92915050565b5f80fd5b5f6101a0828403121561101a57611019611000565b5b81905092915050565b5f819050919050565b61103581611023565b811461103f575f80fd5b50565b5f813590506110508161102c565b92915050565b5f805f6060848603121561106d5761106c610a64565b5b5f84013567ffffffffffffffff81111561108a57611089610a68565b5b61109686828701611004565b93505060206110a786828701611042565b92505060406110b886828701610eea565b9150509250925092565b6110cb81610d2b565b82525050565b5f6020820190506110e45f8301846110c2565b92915050565b5f805f806060858703121561110257611101610a64565b5b5f61110f87828801610e6d565b945050602061112087828801610eea565b935050604085013567ffffffffffffffff81111561114157611140610a68565b5b61114d87828801610ba0565b925092505092959194509250565b5f81519050919050565b5f82825260208201905092915050565b5f819050602082019050919050565b61118d81610bf5565b82525050565b606082015f8201516111a75f850182610d34565b5060208201516111ba6020850182610d34565b5060408201516111cd6040850182611184565b50505050565b5f6111de8383611193565b60608301905092915050565b5f602082019050919050565b5f6112008261115b565b61120a8185611165565b935061121583611175565b805f5b8381101561124557815161122c88826111d3565b9750611237836111ea565b925050600181019050611218565b5085935050505092915050565b5f6020820190508181035f83015261126a81846111f6565b905092915050565b5f82825260208201905092915050565b7f77726f6e67206172726179206c656e67746873000000000000000000000000005f82015250565b5f6112b6601383611272565b91506112c182611282565b602082019050919050565b5f6020820190508181035f8301526112e3816112aa565b9050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52603260045260245ffd5b5f80fd5b5f80fd5b5f80fd5b5f808335600160200384360303811261133f5761133e611317565b5b80840192508235915067ffffffffffffffff8211156113615761136061131b565b5b60208301925060018202360383131561137d5761137c61131f565b5b509250929050565b5f819050919050565b5f77ffffffffffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6113cb6113c66113c184611385565b610f8c565b61138e565b9050919050565b6113db816113b1565b82525050565b5f6040820190506113f45f830185610eac565b61140160208301846113d2565b9392505050565b5f8151905061141681610ed4565b92915050565b5f6020828403121561143157611430610a64565b5b5f61143e84828501611408565b91505092915050565b7f6163636f756e743a206e6f74204f776e6572206f7220456e747279506f696e745f82015250565b5f61147b602083611272565b915061148682611447565b602082019050919050565b5f6020820190508181035f8301526114a88161146f565b9050919050565b5f81519050919050565b5f81905092915050565b5f5b838110156114e05780820151818401526020810190506114c5565b5f8484015250505050565b5f6114f5826114af565b6114ff81856114b9565b935061150f8185602086016114c3565b80840191505092915050565b5f61152682846114eb565b915081905092915050565b7f6f6e6c79206f776e6572000000000000000000000000000000000000000000005f82015250565b5f611565600a83611272565b915061157082611531565b602082019050919050565b5f6020820190508181035f83015261159281611559565b9050919050565b7f6163636f756e743a206e6f742066726f6d20456e747279506f696e74000000005f82015250565b5f6115cd601c83611272565b91506115d882611599565b602082019050919050565b5f6020820190508181035f8301526115fa816115c1565b9050919050565b50565b5f61160f5f836114b9565b915061161a82611601565b5f82019050919050565b5f61162e82611604565b915081905091905056fea2646970667358221220b01667695bfbd93b9a17ef5a8970cd9483e75301f0d2ea73c718be8c7fbf864364736f6c63430008170033",
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

// GetGuardian is a free data retrieval call binding the contract method 0xa75b87d2.
//
// Solidity: function getGuardian() view returns(address)
func (_BindingContract *BindingContractCaller) GetGuardian(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "getGuardian")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetGuardian is a free data retrieval call binding the contract method 0xa75b87d2.
//
// Solidity: function getGuardian() view returns(address)
func (_BindingContract *BindingContractSession) GetGuardian() (common.Address, error) {
	return _BindingContract.Contract.GetGuardian(&_BindingContract.CallOpts)
}

// GetGuardian is a free data retrieval call binding the contract method 0xa75b87d2.
//
// Solidity: function getGuardian() view returns(address)
func (_BindingContract *BindingContractCallerSession) GetGuardian() (common.Address, error) {
	return _BindingContract.Contract.GetGuardian(&_BindingContract.CallOpts)
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

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
func (_BindingContract *BindingContractCaller) GetOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "getOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
func (_BindingContract *BindingContractSession) GetOwner() (common.Address, error) {
	return _BindingContract.Contract.GetOwner(&_BindingContract.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
func (_BindingContract *BindingContractCallerSession) GetOwner() (common.Address, error) {
	return _BindingContract.Contract.GetOwner(&_BindingContract.CallOpts)
}

// GetPasskeys is a free data retrieval call binding the contract method 0xe4093c22.
//
// Solidity: function getPasskeys() view returns((uint256,uint256,uint8)[])
func (_BindingContract *BindingContractCaller) GetPasskeys(opts *bind.CallOpts) ([]PassKey, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "getPasskeys")

	if err != nil {
		return *new([]PassKey), err
	}

	out0 := *abi.ConvertType(out[0], new([]PassKey)).(*[]PassKey)

	return out0, err

}

// GetPasskeys is a free data retrieval call binding the contract method 0xe4093c22.
//
// Solidity: function getPasskeys() view returns((uint256,uint256,uint8)[])
func (_BindingContract *BindingContractSession) GetPasskeys() ([]PassKey, error) {
	return _BindingContract.Contract.GetPasskeys(&_BindingContract.CallOpts)
}

// GetPasskeys is a free data retrieval call binding the contract method 0xe4093c22.
//
// Solidity: function getPasskeys() view returns((uint256,uint256,uint8)[])
func (_BindingContract *BindingContractCallerSession) GetPasskeys() ([]PassKey, error) {
	return _BindingContract.Contract.GetPasskeys(&_BindingContract.CallOpts)
}

// GetSessions is a free data retrieval call binding the contract method 0x61503e45.
//
// Solidity: function getSessions() view returns((address,uint256,uint256,uint64,uint64)[])
func (_BindingContract *BindingContractCaller) GetSessions(opts *bind.CallOpts) ([]SessionKey, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "getSessions")

	if err != nil {
		return *new([]SessionKey), err
	}

	out0 := *abi.ConvertType(out[0], new([]SessionKey)).(*[]SessionKey)

	return out0, err

}

// GetSessions is a free data retrieval call binding the contract method 0x61503e45.
//
// Solidity: function getSessions() view returns((address,uint256,uint256,uint64,uint64)[])
func (_BindingContract *BindingContractSession) GetSessions() ([]SessionKey, error) {
	return _BindingContract.Contract.GetSessions(&_BindingContract.CallOpts)
}

// GetSessions is a free data retrieval call binding the contract method 0x61503e45.
//
// Solidity: function getSessions() view returns((address,uint256,uint256,uint64,uint64)[])
func (_BindingContract *BindingContractCallerSession) GetSessions() ([]SessionKey, error) {
	return _BindingContract.Contract.GetSessions(&_BindingContract.CallOpts)
}

// GetStatus is a free data retrieval call binding the contract method 0x4e69d560.
//
// Solidity: function getStatus() view returns(uint64)
func (_BindingContract *BindingContractCaller) GetStatus(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "getStatus")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// GetStatus is a free data retrieval call binding the contract method 0x4e69d560.
//
// Solidity: function getStatus() view returns(uint64)
func (_BindingContract *BindingContractSession) GetStatus() (uint64, error) {
	return _BindingContract.Contract.GetStatus(&_BindingContract.CallOpts)
}

// GetStatus is a free data retrieval call binding the contract method 0x4e69d560.
//
// Solidity: function getStatus() view returns(uint64)
func (_BindingContract *BindingContractCallerSession) GetStatus() (uint64, error) {
	return _BindingContract.Contract.GetStatus(&_BindingContract.CallOpts)
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
