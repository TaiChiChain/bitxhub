package common

import (
	"testing"

	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

func TestRevertError(t *testing.T) {
	//0x6ca7b80600000000000000000000000027989c08e2cbb2979f8fbb398c6259c3c160d3c7
	sender := types.NewAddressByStr("0x8464135c8F25Da09e49BC8782676a84730C318bC").ETHAddress()
	revertErr := NewRevertError("SenderAddressResult", abi.Arguments{
		abi.Argument{
			Name: "sender",
			Type: AddressType,
		},
	}, []any{sender})

	t.Logf("%s", revertErr.(*RevertError).Data())
	t.Logf("%s", hexutil.Bytes(revertErr.(*RevertError).Data()))

	reason, errUnpack := abi.UnpackRevert(revertErr.(*RevertError).Data())
	t.Logf("reason: %s, data: %s, err: %s", reason, hexutil.Encode(revertErr.(*RevertError).Data()), errUnpack)
}
