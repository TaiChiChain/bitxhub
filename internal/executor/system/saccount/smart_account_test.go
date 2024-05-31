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
)

// erc20 runtime bin code
const erc20bytecode = "6080604052600436106100d4575f3560e01c806370a082311161007857806395d89b411161005557806395d89b411461023b578063a9059cbb1461024f578063d505accf1461026e578063dd62ed3e1461028d57005b806370a08231146101c15780637ecebe00146101f557806384b0196e1461021457005b806323b872dd116100b157806323b872dd14610154578063313ce567146101735780633644e5151461018e57806340c10f19146101a257005b806306fdde03146100dd578063095ea7b31461010757806318160ddd1461013657005b366100db57005b005b3480156100e8575f80fd5b506100f16102d1565b6040516100fe9190610d30565b60405180910390f35b348015610112575f80fd5b50610126610121366004610d64565b610361565b60405190151581526020016100fe565b348015610141575f80fd5b506002545b6040519081526020016100fe565b34801561015f575f80fd5b5061012661016e366004610d8c565b61037a565b34801561017e575f80fd5b50604051601281526020016100fe565b348015610199575f80fd5b5061014661039d565b3480156101ad575f80fd5b506100db6101bc366004610d64565b6103ab565b3480156101cc575f80fd5b506101466101db366004610dc5565b6001600160a01b03165f9081526020819052604090205490565b348015610200575f80fd5b5061014661020f366004610dc5565b6103b9565b34801561021f575f80fd5b506102286103d6565b6040516100fe9796959493929190610dde565b348015610246575f80fd5b506100f1610418565b34801561025a575f80fd5b50610126610269366004610d64565b610427565b348015610279575f80fd5b506100db610288366004610e75565b610434565b348015610298575f80fd5b506101466102a7366004610ee2565b6001600160a01b039182165f90815260016020908152604080832093909416825291909152205490565b6060600380546102e090610f13565b80601f016020809104026020016040519081016040528092919081815260200182805461030c90610f13565b80156103575780601f1061032e57610100808354040283529160200191610357565b820191905f5260205f20905b81548152906001019060200180831161033a57829003601f168201915b5050505050905090565b5f3361036e81858561056f565b60019150505b92915050565b5f33610387858285610581565b6103928585856105fc565b506001949350505050565b5f6103a6610659565b905090565b6103b58282610782565b5050565b6001600160a01b0381165f90815260076020526040812054610374565b5f6060805f805f60606103e76107b6565b6103ef6107e3565b604080515f80825260208201909252600f60f81b9b939a50919850469750309650945092509050565b6060600480546102e090610f13565b5f3361036e8185856105fc565b8342111561045d5760405163313c898160e11b8152600481018590526024015b60405180910390fd5b5f7f6e71edae12b1b97f4d1f60370fef10105fa2faae0126114a169c64845d6126c98888886104a88c6001600160a01b03165f90815260076020526040902080546001810190915590565b6040805160208101969096526001600160a01b0394851690860152929091166060840152608083015260a082015260c0810186905260e0016040516020818303038152906040528051906020012090505f61050282610810565b90505f6105118287878761083c565b9050896001600160a01b0316816001600160a01b031614610558576040516325c0072360e11b81526001600160a01b0380831660048301528b166024820152604401610454565b6105638a8a8a61056f565b50505050505050505050565b61057c8383836001610868565b505050565b6001600160a01b038381165f908152600160209081526040808320938616835292905220545f1981146105f657818110156105e857604051637dc7a0d960e11b81526001600160a01b03841660048201526024810182905260448101839052606401610454565b6105f684848484035f610868565b50505050565b6001600160a01b03831661062557604051634b637e8f60e11b81525f6004820152602401610454565b6001600160a01b03821661064e5760405163ec442f0560e01b81525f6004820152602401610454565b61057c83838361093a565b5f306001600160a01b037f00000000000000000000000033b1b5aa9aa4da83a332f0bc5cac6a903fde5d92161480156106b157507f000000000000000000000000000000000000000000000000000000000000054c46145b156106db57507f4b34c66d199e304a1d5eb341841b52049ce109fb644ffb01c4e43ed7c48a81f590565b6103a6604080517f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f60208201527f245c734e6d4ec044daf7beffa09d54d4bafba490113c199734d790b04a7390e5918101919091527fc89efdaa54c0f20c7adf612882df0950f5a951637e0307cdcb4c672f298b8bc660608201524660808201523060a08201525f9060c00160405160208183030381529060405280519060200120905090565b6001600160a01b0382166107ab5760405163ec442f0560e01b81525f6004820152602401610454565b6103b55f838361093a565b60606103a67f4d79546f6b656e000000000000000000000000000000000000000000000000076005610a60565b60606103a67f31000000000000000000000000000000000000000000000000000000000000016006610a60565b5f61037461081c610659565b8360405161190160f01b8152600281019290925260228201526042902090565b5f805f8061084c88888888610b09565b92509250925061085c8282610bd1565b50909695505050505050565b6001600160a01b0384166108915760405163e602df0560e01b81525f6004820152602401610454565b6001600160a01b0383166108ba57604051634a1406b160e11b81525f6004820152602401610454565b6001600160a01b038085165f90815260016020908152604080832093871683529290522082905580156105f657826001600160a01b0316846001600160a01b03167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b9258460405161092c91815260200190565b60405180910390a350505050565b6001600160a01b038316610964578060025f8282546109599190610f4b565b909155506109d49050565b6001600160a01b0383165f90815260208190526040902054818110156109b65760405163391434e360e21b81526001600160a01b03851660048201526024810182905260448101839052606401610454565b6001600160a01b0384165f9081526020819052604090209082900390555b6001600160a01b0382166109f057600280548290039055610a0e565b6001600160a01b0382165f9081526020819052604090208054820190555b816001600160a01b0316836001600160a01b03167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef83604051610a5391815260200190565b60405180910390a3505050565b606060ff8314610a7a57610a7383610c89565b9050610374565b818054610a8690610f13565b80601f0160208091040260200160405190810160405280929190818152602001828054610ab290610f13565b8015610afd5780601f10610ad457610100808354040283529160200191610afd565b820191905f5260205f20905b815481529060010190602001808311610ae057829003601f168201915b50505050509050610374565b5f80807f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a0841115610b4257505f91506003905082610bc7565b604080515f808252602082018084528a905260ff891692820192909252606081018790526080810186905260019060a0016020604051602081039080840390855afa158015610b93573d5f803e3d5ffd5b5050604051601f1901519150506001600160a01b038116610bbe57505f925060019150829050610bc7565b92505f91508190505b9450945094915050565b5f826003811115610be457610be4610f6a565b03610bed575050565b6001826003811115610c0157610c01610f6a565b03610c1f5760405163f645eedf60e01b815260040160405180910390fd5b6002826003811115610c3357610c33610f6a565b03610c545760405163fce698f760e01b815260048101829052602401610454565b6003826003811115610c6857610c68610f6a565b036103b5576040516335e2f38360e21b815260048101829052602401610454565b60605f610c9583610cc6565b6040805160208082528183019092529192505f91906020820181803683375050509182525060208101929092525090565b5f60ff8216601f81111561037457604051632cd44ac360e21b815260040160405180910390fd5b5f81518084525f5b81811015610d1157602081850181015186830182015201610cf5565b505f602082860101526020601f19601f83011685010191505092915050565b602081525f610d426020830184610ced565b9392505050565b80356001600160a01b0381168114610d5f575f80fd5b919050565b5f8060408385031215610d75575f80fd5b610d7e83610d49565b946020939093013593505050565b5f805f60608486031215610d9e575f80fd5b610da784610d49565b9250610db560208501610d49565b9150604084013590509250925092565b5f60208284031215610dd5575f80fd5b610d4282610d49565b60ff60f81b881681525f602060e06020840152610dfe60e084018a610ced565b8381036040850152610e10818a610ced565b606085018990526001600160a01b038816608086015260a0850187905284810360c0860152855180825260208088019350909101905f5b81811015610e6357835183529284019291840191600101610e47565b50909c9b505050505050505050505050565b5f805f805f805f60e0888a031215610e8b575f80fd5b610e9488610d49565b9650610ea260208901610d49565b95506040880135945060608801359350608088013560ff81168114610ec5575f80fd5b9699959850939692959460a0840135945060c09093013592915050565b5f8060408385031215610ef3575f80fd5b610efc83610d49565b9150610f0a60208401610d49565b90509250929050565b600181811c90821680610f2757607f821691505b602082108103610f4557634e487b7160e01b5f52602260045260245ffd5b50919050565b8082018082111561037457634e487b7160e01b5f52601160045260245ffd5b634e487b7160e01b5f52602160045260245ffdfea264697066735822122080a1fbd346348771e1cf7d26cd3ccdb31ba57421996d0d57102dfbb01d82d93c64736f6c63430008180033"

func TestSmartAccount_Execute(t *testing.T) {
	entrypoint, factory := initEntryPoint(t)
	// deploy erc20
	erc20ContractAddr := ethcommon.HexToAddress("0x82C6D3ed4cD33d8EC1E51d0B5Cc1d822Eaa0c3dC")
	contractAccount := entrypoint.stateLedger.GetOrCreateAccount(types.NewAddress(erc20ContractAddr.Bytes()))
	contractAccount.SetCodeAndHash(ethcommon.Hex2Bytes(erc20bytecode))
	// deploy oracle
	oracleContractAddr := ethcommon.HexToAddress("0x3EBD66861C1d8F298c20ED56506b063206103227")
	contractAccount = entrypoint.stateLedger.GetOrCreateAccount(types.NewAddress(oracleContractAddr.Bytes()))
	contractAccount.SetCodeAndHash(ethcommon.Hex2Bytes(oracleBytecode))

	sk, _ := crypto.GenerateKey()
	owner := crypto.PubkeyToAddress(sk.PublicKey)
	accountAddr := factory.GetAddress(owner, big.NewInt(1))

	InitializeTokenPaymaster(entrypoint.stateLedger, owner)

	tp := NewTokenPaymaster(entrypoint)
	// owner set oracle
	tp.SetContext(&common.VMContext{
		StateLedger:   entrypoint.stateLedger,
		CurrentHeight: 1,
		CurrentLogs:   entrypoint.currentLogs,
		CurrentUser:   &owner,
		CurrentEVM:    entrypoint.currentEVM,
	})
	err := tp.AddToken(erc20ContractAddr, oracleContractAddr)
	assert.Nil(t, err)

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

	// no callFunc
	callData := ethcommon.Hex2Bytes("b61d27f600000000000000000000000082c6d3ed4cd33d8ec1e51d0b5cc1d822eaa0c3dc000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000000")
	isExist, _, totalUsedValue, err := JudgeOrCallInnerMethod(callData, sa)
	assert.Nil(t, err)
	assert.True(t, isExist)
	assert.Equal(t, uint64(10), totalUsedValue.Uint64())

	// totalSupply
	callData = ethcommon.Hex2Bytes("18160ddd")
	_, totalUsedValue, err = sa.Execute(erc20ContractAddr, big.NewInt(0), callData)
	assert.Nil(t, err)
	assert.EqualValues(t, 0, totalUsedValue.Uint64())

	// balanceOf
	callData = ethcommon.Hex2Bytes("70a082310000000000000000000000001cd68912e00c5a02ea0b8b2d972a9c13906b6650")
	_, _, err = sa.Execute(erc20ContractAddr, big.NewInt(0), callData)
	assert.Nil(t, err)

	// mint erc20
	// 0x40c10f190000000000000000000000001cd68912e00c5a02ea0b8b2d972a9c13906b66500000000000000000000000000000000000000000000000000000000000002710
	callData = ethcommon.Hex2Bytes("40c10f190000000000000000000000001cd68912e00c5a02ea0b8b2d972a9c13906b66500fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	_, _, err = sa.Execute(erc20ContractAddr, big.NewInt(0), callData)
	assert.Nil(t, err)

	// balanceOf after mint
	callData = ethcommon.Hex2Bytes("70a082310000000000000000000000001cd68912e00c5a02ea0b8b2d972a9c13906b6650")
	_, _, err = sa.Execute(erc20ContractAddr, big.NewInt(0), callData)
	assert.Nil(t, err)

	// transfer erc20
	callData = ethcommon.Hex2Bytes("b61d27f600000000000000000000000082c6d3ed4cd33d8ec1e51d0b5cc1d822eaa0c3dc000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000044a9059cbb000000000000000000000000c7f999b83af6df9e67d0a37ee7e900bf38b3d013000000000000000000000000000000000000000000000000000000000000000700000000000000000000000000000000000000000000000000000000")
	// replace caller to smart account address
	sa.selfAddr.SetBytes(ethcommon.Hex2Bytes("1cd68912e00c5a02ea0b8b2d972a9c13906b6650"))
	sa.remainingGas = big.NewInt(MaxCallGasLimit)
	isExist, _, totalUsedValue, err = JudgeOrCallInnerMethod(callData, sa)
	assert.Nil(t, err)
	assert.True(t, isExist)
	assert.Equal(t, uint64(7), totalUsedValue.Uint64())

	// transfer max erc20
	callData = ethcommon.Hex2Bytes("b61d27f600000000000000000000000082c6d3ed4cd33d8ec1e51d0b5cc1d822eaa0c3dc000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000044a9059cbb00000000000000000000000079b698e9499a9d6d80c5d8ae180cd23386a7c26d0000000000000000000000000000001d6329f1c35ca4d4d85b977bb4fac0031cv")
	// replace caller to smart account address
	sa.selfAddr.SetBytes(ethcommon.Hex2Bytes("1cd68912e00c5a02ea0b8b2d972a9c13906b6650"))
	sa.remainingGas = big.NewInt(MaxCallGasLimit)
	isExist, _, totalUsedValue, err = JudgeOrCallInnerMethod(callData, sa)
	assert.ErrorContains(t, err, "transfer value exceeds 128 bits, would cause overflow")

	// test smart account
	sa2 := NewSmartAccount(entrypoint.logger, entrypoint)
	sk2, _ := crypto.GenerateKey()
	owner2 := crypto.PubkeyToAddress(sk2.PublicKey)
	accountAddr2 := factory.GetAddress(owner2, big.NewInt(1))
	sa2.SetContext(&common.VMContext{
		StateLedger:   entrypoint.stateLedger,
		CurrentHeight: 1,
		CurrentLogs:   entrypoint.currentLogs,
		CurrentUser:   &entrypointAddr,
		CurrentEVM:    entrypoint.currentEVM,
	})
	sa2.InitializeOrLoad(accountAddr2, owner2, ethcommon.Address{}, big.NewInt(MaxCallGasLimit))

	// sa2 transfer to sa
	// no callFunc
	str := "b61d27f600000000000000000000000082c6d3ed4cd33d8ec1e51d0b5cc1d822eaa0c3dc000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000000"
	dest := strings.ReplaceAll(str, "82c6d3ed4cd33d8ec1e51d0b5cc1d822eaa0c3dc", strings.TrimLeft(accountAddr.Hex(), "0x"))
	callData = ethcommon.Hex2Bytes(dest)
	isExist, _, totalUsedValue, err = JudgeOrCallInnerMethod(callData, sa2)
	assert.Nil(t, err)
	assert.True(t, isExist)
	assert.Equal(t, uint64(10), totalUsedValue.Uint64())
}
