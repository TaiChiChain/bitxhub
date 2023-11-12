package adaptor

import (
	"bytes"

	"github.com/pkg/errors"
	"github.com/rosedblabs/rosedb/v2"
)

// StoreState stores a key,value pair to the database with the given namespace
func (a *RBFTAdaptor) StoreState(key string, value []byte) error {
	return a.store.Put([]byte(key), value)
}

// DelState removes a key,value pair from the database with the given namespace
func (a *RBFTAdaptor) DelState(key string) error {
	return a.store.Delete([]byte(key))
}

// ReadState retrieves a value to a key from the database with the given namespace
func (a *RBFTAdaptor) ReadState(key string) ([]byte, error) {
	b, err := a.store.Get([]byte(key))
	if err != nil {
		if errors.Is(err, rosedb.ErrKeyNotFound) {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	if b == nil {
		return nil, errors.New("not found")
	}
	return b, nil
}

// ReadStateSet retrieves all key-value pairs where the key starts with prefix from the database with the given namespace
func (a *RBFTAdaptor) ReadStateSet(prefix string) (map[string][]byte, error) {
	prefixRaw := []byte(prefix)

	ret := make(map[string][]byte)
	a.store.Ascend(func(k []byte, v []byte) (bool, error) {
		if bytes.HasPrefix(k, prefixRaw) {
			ret[string(k)] = append([]byte(nil), v...)
		}
		return true, nil
	})
	if len(ret) == 0 {
		return nil, errors.New("not found")
	}
	return ret, nil
}
