package adaptor

import (
	"bytes"

	"github.com/pkg/errors"
)

// StoreState stores a key,value pair to the database with the given namespace
func (a *RBFTAdaptor) StoreState(key string, value []byte) error {
	a.store.Put([]byte("consensus."+key), value)
	return nil
}

// DelState removes a key,value pair from the database with the given namespace
func (a *RBFTAdaptor) DelState(key string) error {
	a.store.Delete([]byte("consensus." + key))
	return nil
}

// ReadState retrieves a value to a key from the database with the given namespace
func (a *RBFTAdaptor) ReadState(key string) ([]byte, error) {
	b := a.store.Get([]byte("consensus." + key))
	if b == nil {
		return nil, errors.New("not found")
	}
	return b, nil
}

// ReadStateSet retrieves all key-value pairs where the key starts with prefix from the database with the given namespace
func (a *RBFTAdaptor) ReadStateSet(prefix string) (map[string][]byte, error) {
	prefixRaw := []byte("consensus." + prefix)

	ret := make(map[string][]byte)
	it := a.store.Prefix(prefixRaw)
	if it == nil {
		return nil, errors.New("can't get Iterator")
	}

	if !it.Seek(prefixRaw) {
		return nil, errors.New("not found")
	}

	for bytes.HasPrefix(it.Key(), prefixRaw) {
		key := string(it.Key())
		key = key[len("consensus."):]
		ret[key] = append([]byte(nil), it.Value()...)
		if !it.Next() {
			break
		}
	}
	return ret, nil
}
