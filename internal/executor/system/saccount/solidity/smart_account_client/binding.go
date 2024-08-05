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
	PublicKey []byte
	Algo      uint8
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
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIEntryPoint\",\"name\":\"anEntryPoint\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"contractIEntryPoint\",\"name\":\"entryPoint\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"SmartAccountInitialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"lockedTime\",\"type\":\"uint256\"}],\"name\":\"UserLocked\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"entryPoint\",\"outputs\":[{\"internalType\":\"contractIEntryPoint\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"dest\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"func\",\"type\":\"bytes\"}],\"name\":\"execute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"dest\",\"type\":\"address[]\"},{\"internalType\":\"bytes[]\",\"name\":\"func\",\"type\":\"bytes[]\"}],\"name\":\"executeBatch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"dest\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"value\",\"type\":\"uint256[]\"},{\"internalType\":\"bytes[]\",\"name\":\"func\",\"type\":\"bytes[]\"}],\"name\":\"executeBatch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getGuardian\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getNonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getPasskeys\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes\",\"name\":\"publicKey\",\"type\":\"bytes\"},{\"internalType\":\"uint8\",\"name\":\"algo\",\"type\":\"uint8\"}],\"internalType\":\"structPassKey[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getSessions\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"spendingLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"spentAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint64\",\"name\":\"validUntil\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"validAfter\",\"type\":\"uint64\"}],\"internalType\":\"structSessionKey[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getStatus\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"anOwner\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"resetOwner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"setGuardian\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"},{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"name\":\"setPasskey\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"name\":\"setSession\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"initCode\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"callData\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"callGasLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"verificationGasLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"preVerificationGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxPriorityFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"paymasterAndData\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"authData\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"clientData\",\"type\":\"bytes\"}],\"internalType\":\"structUserOperation\",\"name\":\"userOp\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"userOpHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"missingAccountFunds\",\"type\":\"uint256\"}],\"name\":\"validateUserOp\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"validationData\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
	Bin: "0x60a060405234801562000010575f80fd5b506040516200194c3803806200194c8339818101604052810190620000369190620000e9565b8073ffffffffffffffffffffffffffffffffffffffff1660808173ffffffffffffffffffffffffffffffffffffffff16815250505062000119565b5f80fd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f620000a08262000075565b9050919050565b5f620000b38262000094565b9050919050565b620000c581620000a7565b8114620000d0575f80fd5b50565b5f81519050620000e381620000ba565b92915050565b5f6020828403121562000101576200010062000071565b5b5f6200011084828501620000d3565b91505092915050565b608051611813620001395f395f81816105510152610a3b01526118135ff3fe608060405260043610610101575f3560e01c80638da5cb5b11610094578063b0fff5ca11610063578063b0fff5ca146102f8578063b61d27f614610334578063c4d66de81461035c578063d087d28814610384578063e4093c22146103ae57610108565b80638da5cb5b14610252578063a3ca5fe31461027c578063a75b87d2146102a4578063b0d691fe146102ce57610108565b806361503e45116100d057806361503e45146101ae57806373cc802a146101d8578063893d20e8146102005780638a0dac4a1461022a57610108565b806318dfb3c71461010c57806339d8bab51461013457806347e1da2a1461015c5780634e69d5601461018457610108565b3661010857005b5f80fd5b348015610117575f80fd5b50610132600480360381019061012d9190610b5d565b6103d8565b005b34801561013f575f80fd5b5061015a60048036038101906101559190610c66565b6104e1565b005b348015610167575f80fd5b50610182600480360381019061017d9190610d18565b6104ee565b005b34801561018f575f80fd5b506101986104f6565b6040516101a59190610dea565b60405180910390f35b3480156101b9575f80fd5b506101c26104fa565b6040516101cf9190610f77565b60405180910390f35b3480156101e3575f80fd5b506101fe60048036038101906101f99190610fc1565b6104ff565b005b34801561020b575f80fd5b5061021461050a565b6040516102219190610ffb565b60405180910390f35b348015610235575f80fd5b50610250600480360381019061024b9190610fc1565b61050e565b005b34801561025d575f80fd5b50610266610519565b6040516102739190610ffb565b60405180910390f35b348015610287575f80fd5b506102a2600480360381019061029d9190611068565b61053c565b005b3480156102af575f80fd5b506102b861054a565b6040516102c59190610ffb565b60405180910390f35b3480156102d9575f80fd5b506102e261054e565b6040516102ef9190611127565b60405180910390f35b348015610303575f80fd5b5061031e60048036038101906103199190611196565b610575565b60405161032b9190611211565b60405180910390f35b34801561033f575f80fd5b5061035a6004803603810190610355919061122a565b6105a7565b005b348015610367575f80fd5b50610382600480360381019061037d9190610fc1565b610603565b005b34801561038f575f80fd5b5061039861060f565b6040516103a59190611211565b60405180910390f35b3480156103b9575f80fd5b506103c2610696565b6040516103cf9190611429565b60405180910390f35b6103e061069b565b818190508484905014610428576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161041f906114a3565b60405180910390fd5b5f5b848490508110156104da576104cd85858381811061044b5761044a6114c1565b5b90506020020160208101906104609190610fc1565b5f858585818110610474576104736114c1565b5b905060200281019061048691906114fa565b8080601f0160208091040260200160405190810160405280939291908181526020018383808284375f81840152601f19601f82011690508083019250505050505050610767565b808060010191505061042a565b5050505050565b6104e96107e7565b505050565b505050505050565b5f90565b606090565b6105076107e7565b50565b5f90565b6105166107e7565b50565b5f8054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b6105446107e7565b50505050565b5f90565b5f7f0000000000000000000000000000000000000000000000000000000000000000905090565b5f61057e6108ac565b6105888484610923565b9050610597846020013561092a565b6105a08261092d565b9392505050565b6105af61069b565b6105fd848484848080601f0160208091040260200160405190810160405280939291908181526020018383808284375f81840152601f19601f82011690508083019250505050505050610767565b50505050565b61060c816109c4565b50565b5f61061861054e565b73ffffffffffffffffffffffffffffffffffffffff166335567e1a305f6040518363ffffffff1660e01b81526004016106529291906115b8565b602060405180830381865afa15801561066d573d5f803e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061069191906115f3565b905090565b606090565b6106a361054e565b73ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16148061072657505f8054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16145b610765576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161075c90611668565b60405180910390fd5b565b5f808473ffffffffffffffffffffffffffffffffffffffff16848460405161078f91906116c0565b5f6040518083038185875af1925050503d805f81146107c9576040519150601f19603f3d011682016040523d82523d5f602084013e6107ce565b606091505b5091509150816107e057805160208201fd5b5050505050565b5f8054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16148061086b57503073ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16145b6108aa576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016108a190611720565b60405180910390fd5b565b6108b461054e565b73ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610921576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161091890611788565b60405180910390fd5b565b5f92915050565b50565b5f81146109c1575f3373ffffffffffffffffffffffffffffffffffffffff16827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff9060405161097b906117c9565b5f60405180830381858888f193505050503d805f81146109b6576040519150601f19603f3d011682016040523d82523d5f602084013e6109bb565b606091505b50509050505b50565b805f806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505f8054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff167f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff167fb7053def2fe3d2a5ecb12939fbfcc30f59b5f3efabd9addbe6537fbea7c2739860405160405180910390a350565b5f80fd5b5f80fd5b5f80fd5b5f80fd5b5f80fd5b5f8083601f840112610ac857610ac7610aa7565b5b8235905067ffffffffffffffff811115610ae557610ae4610aab565b5b602083019150836020820283011115610b0157610b00610aaf565b5b9250929050565b5f8083601f840112610b1d57610b1c610aa7565b5b8235905067ffffffffffffffff811115610b3a57610b39610aab565b5b602083019150836020820283011115610b5657610b55610aaf565b5b9250929050565b5f805f8060408587031215610b7557610b74610a9f565b5b5f85013567ffffffffffffffff811115610b9257610b91610aa3565b5b610b9e87828801610ab3565b9450945050602085013567ffffffffffffffff811115610bc157610bc0610aa3565b5b610bcd87828801610b08565b925092505092959194509250565b5f8083601f840112610bf057610bef610aa7565b5b8235905067ffffffffffffffff811115610c0d57610c0c610aab565b5b602083019150836001820283011115610c2957610c28610aaf565b5b9250929050565b5f60ff82169050919050565b610c4581610c30565b8114610c4f575f80fd5b50565b5f81359050610c6081610c3c565b92915050565b5f805f60408486031215610c7d57610c7c610a9f565b5b5f84013567ffffffffffffffff811115610c9a57610c99610aa3565b5b610ca686828701610bdb565b93509350506020610cb986828701610c52565b9150509250925092565b5f8083601f840112610cd857610cd7610aa7565b5b8235905067ffffffffffffffff811115610cf557610cf4610aab565b5b602083019150836020820283011115610d1157610d10610aaf565b5b9250929050565b5f805f805f8060608789031215610d3257610d31610a9f565b5b5f87013567ffffffffffffffff811115610d4f57610d4e610aa3565b5b610d5b89828a01610ab3565b9650965050602087013567ffffffffffffffff811115610d7e57610d7d610aa3565b5b610d8a89828a01610cc3565b9450945050604087013567ffffffffffffffff811115610dad57610dac610aa3565b5b610db989828a01610b08565b92509250509295509295509295565b5f67ffffffffffffffff82169050919050565b610de481610dc8565b82525050565b5f602082019050610dfd5f830184610ddb565b92915050565b5f81519050919050565b5f82825260208201905092915050565b5f819050602082019050919050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f610e5582610e2c565b9050919050565b610e6581610e4b565b82525050565b5f819050919050565b610e7d81610e6b565b82525050565b610e8c81610dc8565b82525050565b60a082015f820151610ea65f850182610e5c565b506020820151610eb96020850182610e74565b506040820151610ecc6040850182610e74565b506060820151610edf6060850182610e83565b506080820151610ef26080850182610e83565b50505050565b5f610f038383610e92565b60a08301905092915050565b5f602082019050919050565b5f610f2582610e03565b610f2f8185610e0d565b9350610f3a83610e1d565b805f5b83811015610f6a578151610f518882610ef8565b9750610f5c83610f0f565b925050600181019050610f3d565b5085935050505092915050565b5f6020820190508181035f830152610f8f8184610f1b565b905092915050565b610fa081610e4b565b8114610faa575f80fd5b50565b5f81359050610fbb81610f97565b92915050565b5f60208284031215610fd657610fd5610a9f565b5b5f610fe384828501610fad565b91505092915050565b610ff581610e4b565b82525050565b5f60208201905061100e5f830184610fec565b92915050565b61101d81610e6b565b8114611027575f80fd5b50565b5f8135905061103881611014565b92915050565b61104781610dc8565b8114611051575f80fd5b50565b5f813590506110628161103e565b92915050565b5f805f80608085870312156110805761107f610a9f565b5b5f61108d87828801610fad565b945050602061109e8782880161102a565b93505060406110af87828801611054565b92505060606110c087828801611054565b91505092959194509250565b5f819050919050565b5f6110ef6110ea6110e584610e2c565b6110cc565b610e2c565b9050919050565b5f611100826110d5565b9050919050565b5f611111826110f6565b9050919050565b61112181611107565b82525050565b5f60208201905061113a5f830184611118565b92915050565b5f80fd5b5f6101a0828403121561115a57611159611140565b5b81905092915050565b5f819050919050565b61117581611163565b811461117f575f80fd5b50565b5f813590506111908161116c565b92915050565b5f805f606084860312156111ad576111ac610a9f565b5b5f84013567ffffffffffffffff8111156111ca576111c9610aa3565b5b6111d686828701611144565b93505060206111e786828701611182565b92505060406111f88682870161102a565b9150509250925092565b61120b81610e6b565b82525050565b5f6020820190506112245f830184611202565b92915050565b5f805f806060858703121561124257611241610a9f565b5b5f61124f87828801610fad565b94505060206112608782880161102a565b935050604085013567ffffffffffffffff81111561128157611280610aa3565b5b61128d87828801610bdb565b925092505092959194509250565b5f81519050919050565b5f82825260208201905092915050565b5f819050602082019050919050565b5f81519050919050565b5f82825260208201905092915050565b5f5b838110156112fb5780820151818401526020810190506112e0565b5f8484015250505050565b5f601f19601f8301169050919050565b5f611320826112c4565b61132a81856112ce565b935061133a8185602086016112de565b61134381611306565b840191505092915050565b61135781610c30565b82525050565b5f604083015f8301518482035f8601526113778282611316565b915050602083015161138c602086018261134e565b508091505092915050565b5f6113a2838361135d565b905092915050565b5f602082019050919050565b5f6113c08261129b565b6113ca81856112a5565b9350836020820285016113dc856112b5565b805f5b8581101561141757848403895281516113f88582611397565b9450611403836113aa565b925060208a019950506001810190506113df565b50829750879550505050505092915050565b5f6020820190508181035f83015261144181846113b6565b905092915050565b5f82825260208201905092915050565b7f77726f6e67206172726179206c656e67746873000000000000000000000000005f82015250565b5f61148d601383611449565b915061149882611459565b602082019050919050565b5f6020820190508181035f8301526114ba81611481565b9050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52603260045260245ffd5b5f80fd5b5f80fd5b5f80fd5b5f8083356001602003843603038112611516576115156114ee565b5b80840192508235915067ffffffffffffffff821115611538576115376114f2565b5b602083019250600182023603831315611554576115536114f6565b5b509250929050565b5f819050919050565b5f77ffffffffffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6115a261159d6115988461155c565b6110cc565b611565565b9050919050565b6115b281611588565b82525050565b5f6040820190506115cb5f830185610fec565b6115d860208301846115a9565b9392505050565b5f815190506115ed81611014565b92915050565b5f6020828403121561160857611607610a9f565b5b5f611615848285016115df565b91505092915050565b7f6163636f756e743a206e6f74204f776e6572206f7220456e747279506f696e745f82015250565b5f611652602083611449565b915061165d8261161e565b602082019050919050565b5f6020820190508181035f83015261167f81611646565b9050919050565b5f81905092915050565b5f61169a826112c4565b6116a48185611686565b93506116b48185602086016112de565b80840191505092915050565b5f6116cb8284611690565b915081905092915050565b7f6f6e6c79206f776e6572000000000000000000000000000000000000000000005f82015250565b5f61170a600a83611449565b9150611715826116d6565b602082019050919050565b5f6020820190508181035f830152611737816116fe565b9050919050565b7f6163636f756e743a206e6f742066726f6d20456e747279506f696e74000000005f82015250565b5f611772601c83611449565b915061177d8261173e565b602082019050919050565b5f6020820190508181035f83015261179f81611766565b9050919050565b50565b5f6117b45f83611686565b91506117bf826117a6565b5f82019050919050565b5f6117d3826117a9565b915081905091905056fea26469706673582212204a6a967773043b45b7f8c8a04524afcd196082148de9efb6b5b494090798d39c64736f6c63430008170033",
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
// Solidity: function getPasskeys() view returns((bytes,uint8)[])
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
// Solidity: function getPasskeys() view returns((bytes,uint8)[])
func (_BindingContract *BindingContractSession) GetPasskeys() ([]PassKey, error) {
	return _BindingContract.Contract.GetPasskeys(&_BindingContract.CallOpts)
}

// GetPasskeys is a free data retrieval call binding the contract method 0xe4093c22.
//
// Solidity: function getPasskeys() view returns((bytes,uint8)[])
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

// ExecuteBatch0 is a paid mutator transaction binding the contract method 0x47e1da2a.
//
// Solidity: function executeBatch(address[] dest, uint256[] value, bytes[] func) returns()
func (_BindingContract *BindingContractTransactor) ExecuteBatch0(opts *bind.TransactOpts, dest []common.Address, value []*big.Int, arg2 [][]byte) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "executeBatch0", dest, value, arg2)
}

// ExecuteBatch0 is a paid mutator transaction binding the contract method 0x47e1da2a.
//
// Solidity: function executeBatch(address[] dest, uint256[] value, bytes[] func) returns()
func (_BindingContract *BindingContractSession) ExecuteBatch0(dest []common.Address, value []*big.Int, arg2 [][]byte) (*types.Transaction, error) {
	return _BindingContract.Contract.ExecuteBatch0(&_BindingContract.TransactOpts, dest, value, arg2)
}

// ExecuteBatch0 is a paid mutator transaction binding the contract method 0x47e1da2a.
//
// Solidity: function executeBatch(address[] dest, uint256[] value, bytes[] func) returns()
func (_BindingContract *BindingContractTransactorSession) ExecuteBatch0(dest []common.Address, value []*big.Int, arg2 [][]byte) (*types.Transaction, error) {
	return _BindingContract.Contract.ExecuteBatch0(&_BindingContract.TransactOpts, dest, value, arg2)
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
