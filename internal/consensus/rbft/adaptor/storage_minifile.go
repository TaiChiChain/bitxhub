package adaptor

import (
	"github.com/axiomesh/axiom-kit/storage/minifile"
)

type MinifileAdaptor struct {
	store *minifile.MiniFile
}

func OpenMinifile(path string) (*MinifileAdaptor, error) {
	store, err := minifile.New(path)
	if err != nil {
		return nil, err
	}
	return &MinifileAdaptor{
		store: store,
	}, nil
}

func (a *MinifileAdaptor) StoreState(key string, value []byte) error {
	return a.store.Put(key, value)
}

func (a *MinifileAdaptor) DelState(key string) error {
	return a.store.Delete(key)
}

func (a *MinifileAdaptor) ReadState(key string) ([]byte, error) {
	b, err := a.store.Get(key)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// ReadStateSet retrieves all key-value pairs where the key starts with prefix from the database with the given namespace
func (a *MinifileAdaptor) ReadStateSet(prefix string) (map[string][]byte, error) {
	ret, err := a.store.GetAll(prefix)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
