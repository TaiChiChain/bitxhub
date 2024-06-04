package executor

import (
	"errors"
	"math/big"
	"strings"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/types"
)

func TestNewEVMBlockContextAdaptor(t *testing.T) {
	Block := &types.Block{
		Header: &types.BlockHeader{
			Number:    1,
			Timestamp: 0,
		},
	}
	getHashFunc := func(n uint64) ethcommon.Hash {
		hash := Block.Hash()
		if hash == nil {
			return ethcommon.Hash{}
		}
		return ethcommon.BytesToHash(hash.Bytes())
	}

	adaptor := NewEVMBlockContextAdaptor(Block.Height(), uint64(Block.Header.Timestamp), "0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013", getHashFunc)
	assert.Equal(t, strings.ToLower("0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013"), strings.ToLower(adaptor.Coinbase.Hex()))
	assert.Equal(t, Block.Height(), adaptor.BlockNumber.Uint64())
	assert.Equal(t, uint64(Block.Header.Timestamp), adaptor.Time)
}

func TestTransactionToMessage(t *testing.T) {
	s, err := types.GenerateSigner()
	assert.Nil(t, err)
	data := "This is a long string"

	keyTo, err := ethcrypto.GenerateKey()
	assert.Nil(t, err)
	addrTo := ethcrypto.PubkeyToAddress(keyTo.PublicKey)

	types.InitEIP155Signer(big.NewInt(1356))
	testSuit := []struct {
		tx                *types.Transaction
		SkipAccountChecks bool
		isFake            bool
	}{
		{
			tx: &types.Transaction{
				Inner: &types.DynamicFeeTx{
					ChainID:   big.NewInt(1356),
					Nonce:     0,
					GasTipCap: big.NewInt(1000),
					GasFeeCap: big.NewInt(2000),
					Gas:       200000,
					To:        nil,
					Value:     big.NewInt(0),
					Data:      []byte(data),
				},
			},
			SkipAccountChecks: false,
			isFake:            false,
		},
		{
			tx: &types.Transaction{
				Inner: &types.DynamicFeeTx{
					ChainID:   big.NewInt(1356),
					Nonce:     1,
					GasTipCap: big.NewInt(1000),
					GasFeeCap: big.NewInt(2000),
					Gas:       200000,
					To:        &addrTo,
					Value:     big.NewInt(10000),
					Data:      []byte(data),
				},
			},
			SkipAccountChecks: false,
			isFake:            false,
		},
		{
			tx: &types.Transaction{
				Inner: &types.DynamicFeeTx{
					ChainID:   big.NewInt(1356),
					Nonce:     2,
					GasTipCap: big.NewInt(1000),
					GasFeeCap: big.NewInt(2000),
					Gas:       200000,
					To:        &addrTo,
					Data:      []byte(data),
					Value:     big.NewInt(10000),
					AccessList: ethtypes.AccessList{
						ethtypes.AccessTuple{
							Address: addrTo,
							StorageKeys: []ethcommon.Hash{
								ethcommon.BytesToHash([]byte("key1")),
								ethcommon.BytesToHash([]byte("key2")),
							},
						},
					},
				},
			},
			SkipAccountChecks: true,
			isFake:            true,
		},
	}

	for _, test := range testSuit {
		err = test.tx.SignByTxType(s.Sk)
		assert.Nil(t, err)
		assert.NotNil(t, test.tx.GetFrom())
		if test.isFake {
			test.tx.Inner.(*types.DynamicFeeTx).V = nil
		}
		message := TransactionToMessage(test.tx)
		assert.Equal(t, test.SkipAccountChecks, message.SkipAccountChecks)
		assert.Equal(t, test.tx.Inner.GetNonce(), message.Nonce)
		assert.Equal(t, test.tx.Inner.GetTo(), message.To)
		assert.Equal(t, test.tx.Inner.GetValue(), message.Value)
		assert.Equal(t, test.tx.Inner.GetData(), message.Data)
		assert.Equal(t, test.tx.GetFrom().ETHAddress(), message.From)
		assert.Equal(t, test.tx.Inner.GetAccessList(), message.AccessList)

		// check strcut field address is not equals
		value := test.tx.Inner.GetValue()
		data := test.tx.Inner.GetData()
		accessList := test.tx.Inner.GetAccessList()
		if value == message.Value {
			assert.Fail(t, "address of value is equals")
		}

		if &data[0] == &message.Data[0] {
			assert.Fail(t, "address of data is equals")
		}

		if accessList != nil {
			if &accessList == &message.AccessList {
				assert.Fail(t, "address of accessList is equals")
			}

			if &accessList[0].Address == &message.AccessList[0].Address {
				assert.Fail(t, "address of accessList.Address is equals")

			}

			if &accessList[0].StorageKeys == &message.AccessList[0].StorageKeys {
				assert.Fail(t, "address of accessList.StorageKeys is equals")
			}
		}

	}
}

func TestCallArgsToMessage(t *testing.T) {
	from, err := ethcrypto.GenerateKey()
	assert.Nil(t, err)
	fromAddr := ethcrypto.PubkeyToAddress(from.PublicKey)

	to, err := ethcrypto.GenerateKey()
	assert.Nil(t, err)
	toAddr := ethcrypto.PubkeyToAddress(to.PublicKey)

	data := "This is a long string"
	b := []byte(data)
	hexData := hexutil.Bytes(b)

	newGas := big.NewInt(1000)
	hexGas := hexutil.Big(*newGas)

	newUint := uint64(1000)
	hexUint := hexutil.Uint64(newUint)

	newChainId := big.NewInt(1356)
	hexChainId := hexutil.Big(*newChainId)

	accessTuple := ethtypes.AccessTuple{
		Address:     fromAddr,
		StorageKeys: []ethcommon.Hash{ethcommon.BytesToHash([]byte{1})},
	}
	accessList := ethtypes.AccessList{accessTuple}

	testCases := []struct {
		args         *types.CallArgs
		globalGasCap uint64
		baseFee      *big.Int
		expect       error
	}{
		{
			args: &types.CallArgs{
				GasPrice:     &hexGas,
				MaxFeePerGas: &hexGas,
			},
			expect: errors.New("both gasPrice and (maxFeePerGas or maxPriorityFeePerGas) specified"),
		},
		{
			args: &types.CallArgs{
				From:       &fromAddr,
				To:         &toAddr,
				Gas:        &hexUint,
				GasPrice:   &hexGas,
				Value:      &hexGas,
				Nonce:      &hexUint,
				Data:       &hexData,
				ChainID:    &hexChainId,
				AccessList: &accessList,
			},
			baseFee: big.NewInt(10),
			expect:  nil,
		},
		{ // if globalGasCap != 0 && globalGasCap < gas
			args: &types.CallArgs{
				From:       &fromAddr,
				To:         &toAddr,
				Gas:        &hexUint,
				GasPrice:   &hexGas,
				Value:      &hexGas,
				Nonce:      &hexUint,
				Data:       &hexData,
				ChainID:    &hexChainId,
				AccessList: &accessList,
			},
			globalGasCap: 1,
			expect:       nil,
		},
		{ // MaxFeePerGas: &hexGas,
			args: &types.CallArgs{
				From:         &fromAddr,
				To:           &toAddr,
				Gas:          &hexUint,
				MaxFeePerGas: &hexGas,
				Value:        &hexGas,
				Nonce:        &hexUint,
				Data:         &hexData,
				ChainID:      &hexChainId,
				AccessList:   &accessList,
			},
			globalGasCap: 1,
			baseFee:      big.NewInt(10),
			expect:       nil,
		},
		{ // MaxPriorityFeePerGas: &hexGas
			args: &types.CallArgs{
				From:                 &fromAddr,
				To:                   &toAddr,
				Gas:                  &hexUint,
				MaxPriorityFeePerGas: &hexGas,
				Value:                &hexGas,
				Nonce:                &hexUint,
				Data:                 &hexData,
				ChainID:              &hexChainId,
				AccessList:           &accessList,
			},
			globalGasCap: 1,
			baseFee:      big.NewInt(10),
			expect:       nil,
		},
	}
	for _, test := range testCases {
		_, err = CallArgsToMessage(test.args, test.globalGasCap, test.baseFee)
		assert.Equal(t, test.expect, err)
	}
}
