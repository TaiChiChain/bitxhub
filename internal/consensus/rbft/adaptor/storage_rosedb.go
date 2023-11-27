package adaptor

import (
	"bytes"
	"time"

	"github.com/pkg/errors"
	"github.com/rosedblabs/rosedb/v2"
)

type RosedbAdaptor struct {
	store *rosedb.DB
}

func OpenRosedb(path string) (*RosedbAdaptor, error) {
	options := rosedb.DefaultOptions
	options.DirPath = path
	options.Sync = false
	options.BlockCache = 50 * 1024 * 1024
	options.SegmentSize = 50 * 1024 * 1024
	store, err := rosedb.Open(options)
	if err != nil {
		return nil, err
	}

	// merge log files
	if err := store.Merge(true); err != nil {
		return nil, err
	}
	go func() {
		tk := time.NewTicker(5 * time.Minute)
		defer tk.Stop()

		for {
			<-tk.C
			if err := store.Merge(true); err != nil {
				continue
			}
		}
	}()

	return &RosedbAdaptor{
		store: store,
	}, nil
}

func (a *RosedbAdaptor) StoreState(key string, value []byte) error {
	return a.store.Put([]byte(key), value)
}

func (a *RosedbAdaptor) DelState(key string) error {
	return a.store.Delete([]byte(key))
}

func (a *RosedbAdaptor) ReadState(key string) ([]byte, error) {
	b, err := a.store.Get([]byte(key))
	if err != nil {
		if errors.Is(err, rosedb.ErrKeyNotFound) {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	return b, nil
}

// ReadStateSet retrieves all key-value pairs where the key starts with prefix from the database with the given namespace
func (a *RosedbAdaptor) ReadStateSet(prefix string) (map[string][]byte, error) {
	prefixRaw := []byte(prefix)
	ret := make(map[string][]byte)
	a.store.Ascend(func(k []byte, v []byte) (bool, error) {
		if bytes.HasPrefix(k, prefixRaw) {
			ret[string(k)] = append([]byte(nil), v...)
		}
		return true, nil
	})
	return ret, nil
}
