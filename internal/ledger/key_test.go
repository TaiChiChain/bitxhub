package ledger

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/types"
)

func TestCompositeKey(t *testing.T) {
	assert.Equal(t, compositeKey(blockKey, 1), []byte("block-1"))
	assert.Equal(t, compositeKey(blockHashKey, "0x112233"), []byte("block-hash-0x112233"))
	assert.Equal(t, compositeKey(transactionKey, "0x112233"), []byte("tx-0x112233"))
}

func TestCompositeStorageKey(t *testing.T) {
	CompositeStorageKey(types.NewAddressByStr("0x5f9f18f7c3a6e5e4c0b877fe3e688ab08840b997"), []byte("12345fsdfdssd"))
	CompositeStorageKey(types.NewAddressByStr("0x5f9f18f7c3a6e5e4c0b877fe3e688ab08840b997"), []byte("1das2345fsdfdssd"))
	CompositeStorageKey(types.NewAddressByStr("0x5f9f18f7c3a6e5e4c0b877fe3e688ab08840b997"), []byte("12ad345fsdfdssd"))
	CompositeStorageKey(types.NewAddressByStr("0x5f9f18f7c3a6e5e4c0b877fe3e688ab08840b998"), []byte("12345fsdfdssd"))
}
