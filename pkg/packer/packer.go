package packer

import (
	"fmt"
	"reflect"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type Event interface {
	Pack(abi abi.ABI) (*types.EvmLog, error)
}

type Error interface {
	Pack(abi abi.ABI) error
}

func PackEvent(eventStruct any, event abi.Event) (*types.EvmLog, error) {
	if eventStruct == nil {
		return nil, errors.New("event struct is nil")
	}
	// references: https://medium.com/mycrypto/understanding-event-logs-on-the-ethereum-blockchain-f4ae7ba50378
	var noIndexedArgs []any
	topicArgs := [][]any{
		{event.ID},
	}
	v := reflect.ValueOf(eventStruct).Elem()
	for _, input := range event.Inputs {
		if !input.Indexed {
			noIndexedArgs = append(noIndexedArgs, v.FieldByName(abi.ToCamelCase(input.Name)).Interface())
		} else {
			topicArgs = append(topicArgs, []any{v.FieldByName(abi.ToCamelCase(input.Name)).Interface()})
		}
	}

	topics, err := abi.MakeTopics(topicArgs...)
	if err != nil {
		return nil, errors.Wrapf(err, "event %s make topics error", event.Name)
	}

	packedData, err := event.Inputs.NonIndexed().Pack(noIndexedArgs...)
	if err != nil {
		return nil, errors.Wrapf(err, "event %s pack args error", event.Name)
	}

	return &types.EvmLog{
		Topics: lo.Map(topics, func(t []common.Hash, i int) *types.Hash {
			return types.NewHash(t[0].Bytes())
		}),
		Data:    packedData,
		Removed: false,
	}, nil
}

type RevertError struct {
	Err error

	// Data is encoded reverted reason, or result
	Data []byte

	// reverted result
	Str string
}

func (e *RevertError) Error() string {
	return fmt.Sprintf("%s errdata %s", e.Err.Error(), e.Str)
}

func PackError(errStruct any, abiErr abi.Error) error {
	if errStruct == nil {
		return errors.New("error struct is nil")
	}
	selector := common.CopyBytes(abiErr.ID.Bytes()[:4])
	var args []any
	v := reflect.ValueOf(errStruct).Elem()
	for _, input := range abiErr.Inputs {
		args = append(args, v.FieldByName(abi.ToCamelCase(input.Name)).Interface())
	}
	packed, err := abiErr.Inputs.Pack(args...)
	if err != nil {
		return err
	}

	return &RevertError{
		Err:  vm.ErrExecutionReverted,
		Data: append(selector, packed...),
		Str:  fmt.Sprintf("%s, args: %v", abiErr.String(), args),
	}
}
