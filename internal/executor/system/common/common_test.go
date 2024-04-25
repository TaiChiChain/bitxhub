package common

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/pkg/packer"
)

func TestRevertError(t *testing.T) {
	// 0x6ca7b80600000000000000000000000027989c08e2cbb2979f8fbb398c6259c3c160d3c7
	sender := types.NewAddressByStr("0x8464135c8F25Da09e49BC8782676a84730C318bC").ETHAddress()
	revertErr := NewRevertError("SenderAddressResult", abi.Arguments{
		abi.Argument{
			Name: "sender",
			Type: AddressType,
		},
	}, []any{sender})

	t.Logf("%v", revertErr.(*packer.RevertError).Data)
	t.Logf("%v", hexutil.Bytes(revertErr.(*packer.RevertError).Data))

	reason, errUnpack := abi.UnpackRevert(revertErr.(*packer.RevertError).Data)
	t.Logf("reason: %s, Data: %s, Err: %s", reason, hexutil.Encode(revertErr.(*packer.RevertError).Data), errUnpack)
}

type MockStakeEvent struct {
	PoolID uint64
	Owner  ethcommon.Address
	Amount *big.Int
}

func TestEmitEvent(t *testing.T) {
	mockAbi := `[{
			"anonymous": false,
			"inputs": [
				{
					"indexed": true,
					"internalType": "uint64",
					"name": "poolID",
					"type": "uint64"
				},
				{
					"indexed": true,
					"internalType": "address",
					"name": "owner",
					"type": "address"
				},
				{
					"indexed": false,
					"internalType": "uint256",
					"name": "amount",
					"type": "uint256"
				}
			],
			"name": "Stake",
			"type": "event"
		}]`
	parseAbi, err := abi.JSON(strings.NewReader(mockAbi))
	assert.Nil(t, err)

	log, err := packer.PackEvent(&MockStakeEvent{
		PoolID: 10,
		Owner:  ethcommon.Address{},
		Amount: big.NewInt(10),
	}, parseAbi.Events["Stake"])
	assert.Nil(t, err)

	assert.NotNil(t, log)
	assert.Equal(t, log.Address.String(), "")
	assert.Equal(t, len(log.Topics), 3)
	assert.Equal(t, log.Topics[0].String(), parseAbi.Events["Stake"].ID.String())
	assert.Equal(t, []byte{log.Data[len(log.Data)-1]}, big.NewInt(10).Bytes())
}
