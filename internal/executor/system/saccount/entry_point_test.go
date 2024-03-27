package saccount

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/interfaces"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"go.uber.org/mock/gomock"
)

func initEntryPoint(t *testing.T) *EntryPoint {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(2, types.NewAddressByStr(common.GovernanceContractAddr))
	account.SetBalance(big.NewInt(3000000000000000000))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().GetNonce(gomock.Any()).Return(0).AnyTimes()
	stateLedger.EXPECT().Snapshot().AnyTimes()
	stateLedger.EXPECT().Commit().Return(types.NewHash([]byte("")), nil).AnyTimes()
	stateLedger.EXPECT().Clear().AnyTimes()
	stateLedger.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)).AnyTimes()
	stateLedger.EXPECT().SetNonce(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().Finalise().AnyTimes()
	stateLedger.EXPECT().Snapshot().Return(0).AnyTimes()
	stateLedger.EXPECT().RevertToSnapshot(0).AnyTimes()
	stateLedger.EXPECT().SetTxContext(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().GetLogs(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	stateLedger.EXPECT().PrepareBlock(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().GetBalance(gomock.Any()).Return(big.NewInt(3000000000000000000)).AnyTimes()
	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SubBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().AddBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().GetCodeHash(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().Exist(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().GetRefund().AnyTimes()
	stateLedger.EXPECT().GetCode(gomock.Any()).AnyTimes()

	coinbase := "0xed17543171C1459714cdC6519b58fFcC29A3C3c9"
	blkCtx := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     nil,
		Coinbase:    ethcommon.HexToAddress(coinbase),
		BlockNumber: new(big.Int).SetUint64(0),
		Time:        uint64(time.Now().Unix()),
		Difficulty:  big.NewInt(0x2000),
		BaseFee:     big.NewInt(0),
		GasLimit:    0x2fefd8,
		Random:      &ethcommon.Hash{},
	}

	evm := vm.NewEVM(blkCtx, vm.TxContext{}, &ledger.EvmStateDBAdaptor{
		StateLedger: stateLedger,
	}, &params.ChainConfig{
		ChainID: big.NewInt(1356),
	}, vm.Config{})

	entryPoint := NewEntryPoint(&common.SystemContractConfig{})
	entryPoint.SetContext(&common.VMContext{
		CurrentEVM:    evm,
		CurrentHeight: 1,
		CurrentLogs:   new([]common.Log),
		CurrentUser:   nil,
		StateLedger:   stateLedger,
	})
	return entryPoint
}

func TestEntryPoint_SimulateHandleOp(t *testing.T) {
	entryPoint := initEntryPoint(t)
	sk, _ := crypto.GenerateKey()
	owner := crypto.PubkeyToAddress(sk.PublicKey)
	initCode := ethcommon.Hex2Bytes("00000000000000000000000000000000000010095fbfb9cf0000000000000000000000009965507d1a55bcc2695c58ba16fb37d819b0a4dc0000000000000000000000000000000000000000000000000000000000015b43")
	callData := ethcommon.Hex2Bytes("b61d27f6000000000000000000000000ed17543171c1459714cdc6519b58ffcc29a3c3c9000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000044a9059cbb000000000000000000000000c7f999b83af6df9e67d0a37ee7e900bf38b3d013000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000")
	// replace owner
	copy(initCode[20+4+(32-20):], owner.Bytes())

	senderAddr, _, err := entryPoint.createSender(big.NewInt(10000), initCode)
	assert.Nil(t, err)
	copy(initCode[20+4+(32-20):], owner.Bytes())
	userOp := &interfaces.UserOperation{
		Sender:               senderAddr,
		Nonce:                big.NewInt(0),
		InitCode:             initCode,
		CallData:             callData,
		CallGasLimit:         big.NewInt(0),
		VerificationGasLimit: big.NewInt(9000000),
		PreVerificationGas:   big.NewInt(21000),
		MaxFeePerGas:         big.NewInt(2000000000000),
		MaxPriorityFeePerGas: big.NewInt(2000000000000),
		PaymasterAndData:     nil,
	}
	userOpHash := entryPoint.GetUserOpHash(*userOp)
	sig, err := crypto.Sign(userOpHash, sk)
	assert.Nil(t, err)
	userOp.Signature = sig

	err = entryPoint.SimulateHandleOp(*userOp, ethcommon.Address{}, nil)
	executeResultErr, ok := err.(*common.RevertError)
	assert.True(t, ok, "can't convert err to RevertError")
	assert.Equal(t, executeResultErr.GetError(), vm.ErrExecutionReverted)
}

func TestEntryPoint_call(t *testing.T) {
	entryPoint := initEntryPoint(t)
	// test call method
	to := ethcommon.HexToAddress(common.AccountFactoryContractAddr)
	res, err := call(entryPoint.stateLedger, entryPoint.currentEVM, big.NewInt(300000), types.NewAddressByStr(common.EntryPointContractAddr), &to, []byte(""))
	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func TestEntryPoint_GetUserOpHash(t *testing.T) {
	entryPoint := initEntryPoint(t)
	dstUserOpHash := "c02fe0170a5402c0d08502e2140d43803ce1672500b9100ac95aea6e7e2bd508"
	sk, err := crypto.HexToECDSA("8b3a350cf5c34c9194ca85829a2df0ec3153be0318b5e2d3348e872092edffba")
	assert.Nil(t, err)
	owner := crypto.PubkeyToAddress(sk.PublicKey)

	t.Logf("owner addr: %s", owner.String())

	initCode := ethcommon.Hex2Bytes("00000000000000000000000000000000000010095fbfb9cf0000000000000000000000009965507d1a55bcc2695c58ba16fb37d819b0a4dc0000000000000000000000000000000000000000000000000000000000015b43")
	// replace owner
	copy(initCode[20+4+(32-20):], owner.Bytes())
	t.Logf("initCode addr: %v", initCode)
	callData := ethcommon.Hex2Bytes("b61d27f60000000000000000000000008464135c8f25da09e49bc8782676a84730c318bc000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000044a9059cbb000000000000000000000000c7f999b83af6df9e67d0a37ee7e900bf38b3d013000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000")

	senderAddr, _, err := entryPoint.createSender(big.NewInt(10000), initCode)
	assert.Nil(t, err)
	t.Logf("sender addr: %s", senderAddr.String())

	userOp := &interfaces.UserOperation{
		Sender:               senderAddr,
		Nonce:                big.NewInt(0),
		InitCode:             initCode,
		CallData:             callData,
		CallGasLimit:         big.NewInt(12100),
		VerificationGasLimit: big.NewInt(21971),
		PreVerificationGas:   big.NewInt(50579),
		MaxFeePerGas:         big.NewInt(5875226332498),
		MaxPriorityFeePerGas: big.NewInt(0),
		PaymasterAndData:     nil,
		Signature:            ethcommon.Hex2Bytes("70aae8a31ee824b05e449e082f9fcf848218fe4563988d426e72d4675d1179b9735c026f847eb22e6b0f1a8dc81825678f383b07189c5df6a4a513a972ad71191c"),
	}

	userOpHash := entryPoint.GetUserOpHash(*userOp)
	t.Logf("op hash: %x", userOpHash)
	t.Logf("chain id: %s", entryPoint.currentEVM.ChainConfig().ChainID.Text(10))
	assert.Equal(t, ethcommon.Hex2Bytes(dstUserOpHash), userOpHash)

	hash := accounts.TextHash(userOpHash)
	// 0xc911ab31b05daa531d9ef77fc2670e246a9489610e48e88d8296604c26dcd5c4
	t.Logf("eth hash: %x", hash)
	sig, err := crypto.Sign(hash, sk)
	assert.Nil(t, err)
	// js return r|s|v, v only 1 byte
	// golang return rid, v = rid +27
	msgSig := make([]byte, 65)
	copy(msgSig, sig)
	msgSig[64] += 27
	assert.Equal(t, msgSig, userOp.Signature)

	recoveredPub, err := crypto.Ecrecover(hash, sig)
	assert.Nil(t, err)
	pubKey, _ := crypto.UnmarshalPubkey(recoveredPub)
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	assert.Equal(t, owner, recoveredAddr)
}

func TestEntryPoint_createSender(t *testing.T) {
	entryPoint := initEntryPoint(t)
	initCode := ethcommon.Hex2Bytes("00000000000000000000000000000000000010095fbfb9cf0000000000000000000000009965507d1a55bcc2695c58ba16fb37d819b0a4dc0000000000000000000000000000000000000000000000000000000000015b43")
	addr, _, err := entryPoint.createSender(big.NewInt(10000), initCode)
	assert.Nil(t, err)

	assert.NotNil(t, addr)
	assert.NotEqual(t, addr, ethcommon.Address{})

	addr2, _, err := entryPoint.createSender(big.NewInt(10000), initCode)
	assert.Nil(t, err)
	assert.Equal(t, addr, addr2)
}

func TestEntryPoint_ExecutionResult(t *testing.T) {
	err := interfaces.ExecutionResult(big.NewInt(1000), big.NewInt(999), big.NewInt(0), big.NewInt(0), false, nil)
	_, ok := err.(*common.RevertError)
	assert.True(t, ok)
	t.Logf("err is : %s", err)
}

func TestEntryPoint_userOpEvent(t *testing.T) {
	userOpEvent := abi.NewEvent("UserOperationEvent", "UserOperationEvent", false, abi.Arguments{
		{Name: "userOpHash", Type: common.Bytes32Type, Indexed: true},
		{Name: "sender", Type: common.AddressType, Indexed: true},
		{Name: "paymaster", Type: common.AddressType, Indexed: true},
		{Name: "nonce", Type: common.BigIntType},
		{Name: "success", Type: common.BoolType},
		{Name: "actualGasCost", Type: common.BigIntType},
		{Name: "actualGasUsed", Type: common.BigIntType},
	})

	assert.Equal(t, "0x49628fd1471006c1482da88028e9ce4dbb080b815c9b0344d39e5a8e6ec1419f", types.NewHash(userOpEvent.ID.Bytes()).String())
}

// getDepositInfo test
