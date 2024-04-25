package packer

import (
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/types"
)

type EventTest struct {
	User   common.Address
	Amount *big.Int
}

func (_event *EventTest) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return PackEvent(_event, abi.Events["Test"])
}

func TestPackEvent(t *testing.T) {
	abiStr := "[\n\t{\n\t\t\"anonymous\": false,\n\t\t\"inputs\": [\n\t\t\t{\n\t\t\t\t\"indexed\": true,\n\t\t\t\t\"internalType\": \"address\",\n\t\t\t\t\"name\": \"user\",\n\t\t\t\t\"type\": \"address\"\n\t\t\t},\n\t\t\t{\n\t\t\t\t\"indexed\": false,\n\t\t\t\t\"internalType\": \"uint256\",\n\t\t\t\t\"name\": \"amount\",\n\t\t\t\t\"type\": \"uint256\"\n\t\t\t}\n\t\t],\n\t\t\"name\": \"Test\",\n\t\t\"type\": \"event\"\n\t}\n]"
	innerABI, err := abi.JSON(strings.NewReader(abiStr))
	assert.Nil(t, err)
	input := EventTest{
		User:   common.HexToAddress("0x0000000000000000000000000000000000001000"),
		Amount: big.NewInt(100),
	}
	log, err := input.Pack(innerABI)
	assert.Nil(t, err)
	assert.Equal(t, innerABI.Events["Test"].ID, log.Topics[0].ETHHash())
	data, err := innerABI.Events["Test"].Inputs.Unpack(log.Data)
	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(100), data[0])
	var topics []common.Hash
	topics = append(topics, log.Topics[1].ETHHash())
	parseEv := &EventTest{}
	err = abi.ParseTopics(parseEv,
		abi.Arguments{
			innerABI.Events["Test"].Inputs[0],
		}, topics)
	assert.Nil(t, err)
	assert.Equal(t, parseEv.User, input.User)
}

type ErrorTest struct {
	User   common.Address
	Amount *big.Int
}

func (_error *ErrorTest) Pack(abi abi.ABI) error {
	return PackError(_error, abi.Errors["testError"])
}

func TestPackError(t *testing.T) {
	abiStr := "[\n\t{\n\t\t\"inputs\": [\n\t\t\t{\n\t\t\t\t\"internalType\": \"address\",\n\t\t\t\t\"name\": \"user\",\n\t\t\t\t\"type\": \"address\"\n\t\t\t},\n\t\t\t{\n\t\t\t\t\"internalType\": \"uint256\",\n\t\t\t\t\"name\": \"amount\",\n\t\t\t\t\"type\": \"uint256\"\n\t\t\t}\n\t\t],\n\t\t\"name\": \"testError\",\n\t\t\"type\": \"error\"\n\t}\n]"
	innerAbi, err := abi.JSON(strings.NewReader(abiStr))
	assert.Nil(t, err)
	errTest := &ErrorTest{
		User:   common.HexToAddress("0x0000000000000000000000000000000000001000"),
		Amount: big.NewInt(100),
	}
	var revertErr *RevertError
	err = errTest.Pack(innerAbi)
	ok := errors.As(err, &revertErr)
	assert.True(t, ok)
}
