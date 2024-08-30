package common

import (
	"errors"

	"github.com/axiomesh/axiom-kit/storage/kv"
)

func StoreEpochState(epochStore kv.Storage, key string, value []byte) error {
	epochStore.Put([]byte("epoch."+key), value)
	return nil
}

func ReadEpochState(epochStore kv.Storage, key string) ([]byte, error) {
	b := epochStore.Get([]byte("epoch." + key))
	if b == nil {
		return nil, errors.New("not found")
	}
	return b, nil
}
