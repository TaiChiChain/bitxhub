package saccount

import (
	"math/big"
	"strings"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/interfaces"
)

// oracle runtime bytecode
const oracleBytecode = "608060405234801561000f575f80fd5b5060043610610029575f3560e01c80632a9ff49c1461002d575b5f80fd5b61004061003b3660046100c2565b610052565b60405190815260200160405180910390f35b5f6001600160a01b0383166100ad5760405162461bcd60e51b815260206004820152601d60248201527f7469636b657220616464726573732063616e277420626520656d707479000000604482015260640160405180910390fd5b6100bb6305f5e100836100f7565b9392505050565b5f80604083850312156100d3575f80fd5b82356001600160a01b03811681146100e9575f80fd5b946020939093013593505050565b5f8261011157634e487b7160e01b5f52601260045260245ffd5b50049056fea26469706673582212200b8ebb4149b0a36094fa961f9e747fbf002c109fa1840360b0101fcc5ea2262c64736f6c63430008140033"

// 8 decimals erc20 runtime bytecode
const anotherErc20Bytecode = "608060405234801561000f575f80fd5b50600436106100e5575f3560e01c806370a082311161008857806395d89b411161006357806395d89b41146101d1578063a9059cbb146101d9578063d505accf146101ec578063dd62ed3e146101ff575f80fd5b806370a082311461017b5780637ecebe00146101a357806384b0196e146101b6575f80fd5b806323b872dd116100c357806323b872dd1461013c578063313ce5671461014f5780633644e5151461015e57806340c10f1914610166575f80fd5b806306fdde03146100e9578063095ea7b31461010757806318160ddd1461012a575b5f80fd5b6100f1610237565b6040516100fe9190610c96565b60405180910390f35b61011a610115366004610cca565b6102c7565b60405190151581526020016100fe565b6002545b6040519081526020016100fe565b61011a61014a366004610cf2565b6102e0565b604051600881526020016100fe565b61012e610303565b610179610174366004610cca565b610311565b005b61012e610189366004610d2b565b6001600160a01b03165f9081526020819052604090205490565b61012e6101b1366004610d2b565b61031f565b6101be61033c565b6040516100fe9796959493929190610d44565b6100f161037e565b61011a6101e7366004610cca565b61038d565b6101796101fa366004610dd8565b61039a565b61012e61020d366004610e45565b6001600160a01b039182165f90815260016020908152604080832093909416825291909152205490565b60606003805461024690610e76565b80601f016020809104026020016040519081016040528092919081815260200182805461027290610e76565b80156102bd5780601f10610294576101008083540402835291602001916102bd565b820191905f5260205f20905b8154815290600101906020018083116102a057829003601f168201915b5050505050905090565b5f336102d48185856104d5565b60019150505b92915050565b5f336102ed8582856104e7565b6102f8858585610562565b506001949350505050565b5f61030c6105bf565b905090565b61031b82826106e8565b5050565b6001600160a01b0381165f908152600760205260408120546102da565b5f6060805f805f606061034d61071c565b610355610749565b604080515f80825260208201909252600f60f81b9b939a50919850469750309650945092509050565b60606004805461024690610e76565b5f336102d4818585610562565b834211156103c35760405163313c898160e11b8152600481018590526024015b60405180910390fd5b5f7f6e71edae12b1b97f4d1f60370fef10105fa2faae0126114a169c64845d6126c988888861040e8c6001600160a01b03165f90815260076020526040902080546001810190915590565b6040805160208101969096526001600160a01b0394851690860152929091166060840152608083015260a082015260c0810186905260e0016040516020818303038152906040528051906020012090505f61046882610776565b90505f610477828787876107a2565b9050896001600160a01b0316816001600160a01b0316146104be576040516325c0072360e11b81526001600160a01b0380831660048301528b1660248201526044016103ba565b6104c98a8a8a6104d5565b50505050505050505050565b6104e283838360016107ce565b505050565b6001600160a01b038381165f908152600160209081526040808320938616835292905220545f19811461055c578181101561054e57604051637dc7a0d960e11b81526001600160a01b038416600482015260248101829052604481018390526064016103ba565b61055c84848484035f6107ce565b50505050565b6001600160a01b03831661058b57604051634b637e8f60e11b81525f60048201526024016103ba565b6001600160a01b0382166105b45760405163ec442f0560e01b81525f60048201526024016103ba565b6104e28383836108a0565b5f306001600160a01b037f00000000000000000000000000000000000000000000000000000000000000001614801561061757507f000000000000000000000000000000000000000000000000000000000000000046145b1561064157507f000000000000000000000000000000000000000000000000000000000000000090565b61030c604080517f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f60208201527f0000000000000000000000000000000000000000000000000000000000000000918101919091527f000000000000000000000000000000000000000000000000000000000000000060608201524660808201523060a08201525f9060c00160405160208183030381529060405280519060200120905090565b6001600160a01b0382166107115760405163ec442f0560e01b81525f60048201526024016103ba565b61031b5f83836108a0565b606061030c7f000000000000000000000000000000000000000000000000000000000000000060056109c6565b606061030c7f000000000000000000000000000000000000000000000000000000000000000060066109c6565b5f6102da6107826105bf565b8360405161190160f01b8152600281019290925260228201526042902090565b5f805f806107b288888888610a6f565b9250925092506107c28282610b37565b50909695505050505050565b6001600160a01b0384166107f75760405163e602df0560e01b81525f60048201526024016103ba565b6001600160a01b03831661082057604051634a1406b160e11b81525f60048201526024016103ba565b6001600160a01b038085165f908152600160209081526040808320938716835292905220829055801561055c57826001600160a01b0316846001600160a01b03167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b9258460405161089291815260200190565b60405180910390a350505050565b6001600160a01b0383166108ca578060025f8282546108bf9190610eae565b9091555061093a9050565b6001600160a01b0383165f908152602081905260409020548181101561091c5760405163391434e360e21b81526001600160a01b038516600482015260248101829052604481018390526064016103ba565b6001600160a01b0384165f9081526020819052604090209082900390555b6001600160a01b03821661095657600280548290039055610974565b6001600160a01b0382165f9081526020819052604090208054820190555b816001600160a01b0316836001600160a01b03167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef836040516109b991815260200190565b60405180910390a3505050565b606060ff83146109e0576109d983610bef565b90506102da565b8180546109ec90610e76565b80601f0160208091040260200160405190810160405280929190818152602001828054610a1890610e76565b8015610a635780601f10610a3a57610100808354040283529160200191610a63565b820191905f5260205f20905b815481529060010190602001808311610a4657829003601f168201915b505050505090506102da565b5f80807f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a0841115610aa857505f91506003905082610b2d565b604080515f808252602082018084528a905260ff891692820192909252606081018790526080810186905260019060a0016020604051602081039080840390855afa158015610af9573d5f803e3d5ffd5b5050604051601f1901519150506001600160a01b038116610b2457505f925060019150829050610b2d565b92505f91508190505b9450945094915050565b5f826003811115610b4a57610b4a610ecd565b03610b53575050565b6001826003811115610b6757610b67610ecd565b03610b855760405163f645eedf60e01b815260040160405180910390fd5b6002826003811115610b9957610b99610ecd565b03610bba5760405163fce698f760e01b8152600481018290526024016103ba565b6003826003811115610bce57610bce610ecd565b0361031b576040516335e2f38360e21b8152600481018290526024016103ba565b60605f610bfb83610c2c565b6040805160208082528183019092529192505f91906020820181803683375050509182525060208101929092525090565b5f60ff8216601f8111156102da57604051632cd44ac360e21b815260040160405180910390fd5b5f81518084525f5b81811015610c7757602081850181015186830182015201610c5b565b505f602082860101526020601f19601f83011685010191505092915050565b602081525f610ca86020830184610c53565b9392505050565b80356001600160a01b0381168114610cc5575f80fd5b919050565b5f8060408385031215610cdb575f80fd5b610ce483610caf565b946020939093013593505050565b5f805f60608486031215610d04575f80fd5b610d0d84610caf565b9250610d1b60208501610caf565b9150604084013590509250925092565b5f60208284031215610d3b575f80fd5b610ca882610caf565b60ff60f81b881681525f602060e081840152610d6360e084018a610c53565b8381036040850152610d75818a610c53565b606085018990526001600160a01b038816608086015260a0850187905284810360c086015285518082528387019250908301905f5b81811015610dc657835183529284019291840191600101610daa565b50909c9b505050505050505050505050565b5f805f805f805f60e0888a031215610dee575f80fd5b610df788610caf565b9650610e0560208901610caf565b95506040880135945060608801359350608088013560ff81168114610e28575f80fd5b9699959850939692959460a0840135945060c09093013592915050565b5f8060408385031215610e56575f80fd5b610e5f83610caf565b9150610e6d60208401610caf565b90509250929050565b600181811c90821680610e8a57607f821691505b602082108103610ea857634e487b7160e01b5f52602260045260245ffd5b50919050565b808201808211156102da57634e487b7160e01b5f52601160045260245ffd5b634e487b7160e01b5f52602160045260245ffdfea2646970667358221220733a17d178e6af325fbf6ecef2ccec2addb464082ef4471b39b1b48468d3a63864736f6c63430008140033"

func TestTokenPaymaster_ValidatePaymasterUserOp(t *testing.T) {
	entrypoint, factory := initEntryPoint(t)
	// deploy erc20
	erc20ContractAddr := ethcommon.HexToAddress("0x82C6D3ed4cD33d8EC1E51d0B5Cc1d822Eaa0c3dC")
	contractAccount := entrypoint.stateLedger.GetOrCreateAccount(types.NewAddress(erc20ContractAddr.Bytes()))
	contractAccount.SetCodeAndHash(ethcommon.Hex2Bytes(erc20bytecode))
	// deploy oracle
	oracleContractAddr := ethcommon.HexToAddress("0x3EBD66861C1d8F298c20ED56506b063206103227")
	contractAccount = entrypoint.stateLedger.GetOrCreateAccount(types.NewAddress(oracleContractAddr.Bytes()))
	contractAccount.SetCodeAndHash(ethcommon.Hex2Bytes(oracleBytecode))
	// deploy erc20
	anotherErc20ContractAddr := ethcommon.HexToAddress("0x067c804bb006836469379D4A2A69a81803bd1F45")
	contractAccount = entrypoint.stateLedger.GetOrCreateAccount(types.NewAddress(anotherErc20ContractAddr.Bytes()))
	contractAccount.SetCodeAndHash(ethcommon.Hex2Bytes(anotherErc20Bytecode))

	sk, _ := crypto.GenerateKey()
	owner := crypto.PubkeyToAddress(sk.PublicKey)
	accountAddr := factory.GetAddress(owner, big.NewInt(1))

	InitializeTokenPaymaster(entrypoint.stateLedger, owner)

	t.Logf("account address: %s", accountAddr)

	sa := NewSmartAccount(entrypoint.logger, entrypoint)
	entrypointAddr := entrypoint.selfAddress().ETHAddress()
	sa.SetContext(&common.VMContext{
		StateLedger:   entrypoint.stateLedger,
		CurrentHeight: 1,
		CurrentLogs:   entrypoint.currentLogs,
		CurrentUser:   &entrypointAddr,
		CurrentEVM:    entrypoint.currentEVM,
	})
	sa.InitializeOrLoad(accountAddr, owner, ethcommon.Address{}, big.NewInt(MaxCallGasLimit))

	// mint erc20
	mintSig := crypto.Keccak256([]byte("mint(address,uint256)"))[:4]
	t.Logf("mint sig: %x", mintSig) // 40c10f19
	callData := append(mintSig, ethcommon.LeftPadBytes(accountAddr.Bytes(), 32)...)
	callData = append(callData, ethcommon.LeftPadBytes(big.NewInt(1500000000000000000).Bytes(), 32)...)

	_, totalUsedValue, err := sa.Execute(erc20ContractAddr, big.NewInt(0), callData)
	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(0), totalUsedValue)

	// mint another erc20
	sa.remainingGas = big.NewInt(MaxCallGasLimit)
	_, _, err = sa.Execute(anotherErc20ContractAddr, big.NewInt(0), callData)
	assert.Nil(t, err)

	userOp := interfaces.UserOperation{
		Sender:               accountAddr,
		Nonce:                big.NewInt(0),
		InitCode:             nil,
		CallData:             nil,
		CallGasLimit:         big.NewInt(10000000),
		VerificationGasLimit: big.NewInt(10000000),
		PreVerificationGas:   big.NewInt(10000000),
		MaxFeePerGas:         big.NewInt(1000000000000),
		MaxPriorityFeePerGas: big.NewInt(1000000000000),
	}
	userOp.PaymasterAndData = append(ethcommon.Hex2Bytes(strings.TrimPrefix(common.TokenPaymasterContractAddr, "0x")), erc20ContractAddr.Bytes()...)

	tp := NewTokenPaymaster(entrypoint)
	tp.SetContext(&common.VMContext{
		StateLedger:   entrypoint.stateLedger,
		CurrentHeight: 1,
		CurrentLogs:   entrypoint.currentLogs,
		CurrentUser:   &entrypointAddr,
		CurrentEVM:    entrypoint.currentEVM,
	})
	_, _, err = tp.ValidatePaymasterUserOp(userOp, nil, big.NewInt(100000000))
	assert.Contains(t, err.Error(), "not set oracle")

	// owner set oracle
	tp.SetContext(&common.VMContext{
		StateLedger:   entrypoint.stateLedger,
		CurrentHeight: 1,
		CurrentLogs:   entrypoint.currentLogs,
		CurrentUser:   &owner,
		CurrentEVM:    entrypoint.currentEVM,
	})
	err = tp.AddToken(erc20ContractAddr, oracleContractAddr)
	assert.Nil(t, err)
	err = tp.AddToken(anotherErc20ContractAddr, oracleContractAddr)
	assert.Nil(t, err)

	tp.SetContext(&common.VMContext{
		StateLedger:   entrypoint.stateLedger,
		CurrentHeight: 1,
		CurrentLogs:   entrypoint.currentLogs,
		CurrentUser:   &entrypointAddr,
		CurrentEVM:    entrypoint.currentEVM,
	})
	context, validateData, err := tp.ValidatePaymasterUserOp(userOp, nil, big.NewInt(100000000))
	assert.Nil(t, err)
	assert.EqualValues(t, interfaces.SigValidationSucceeded, validateData.Uint64())
	account, token, _, maxTokenCost, maxCost, err := decodeContext(context)
	assert.Nil(t, err)
	assert.Equal(t, account.String(), accountAddr.String())
	assert.Equal(t, token.String(), erc20ContractAddr.String())
	assert.EqualValues(t, 100000000, maxTokenCost.Uint64())
	assert.EqualValues(t, 100000000, maxCost.Uint64())

	copy(userOp.PaymasterAndData[20:], anotherErc20ContractAddr.Bytes())

	context, validateData, err = tp.ValidatePaymasterUserOp(userOp, nil, big.NewInt(10000000000))
	assert.Nil(t, err)
	assert.EqualValues(t, interfaces.SigValidationSucceeded, validateData.Uint64())
	account, token, _, maxTokenCost, maxCost, err = decodeContext(context)
	assert.Nil(t, err)
	assert.Equal(t, account.String(), accountAddr.String())
	assert.Equal(t, token.String(), anotherErc20ContractAddr.String())
	assert.EqualValues(t, 1, maxTokenCost.Uint64())
	assert.EqualValues(t, 10000000000, maxCost.Uint64())
}
