package adaptor

import (
	"github.com/pkg/errors"

	rbft "github.com/axiomesh/axiom-bft"
)

func (a *RBFTAdaptor) GetCurrentEpochInfo() (*rbft.EpochInfo, error) {
	return a.config.GetCurrentEpochInfoFromEpochMgrContractFunc()
}

func (a *RBFTAdaptor) GetEpochInfo(epoch uint64) (*rbft.EpochInfo, error) {
	return a.config.GetEpochInfoFromEpochMgrContractFunc(epoch)
}

func (a *RBFTAdaptor) StoreEpochState(key string, value []byte) error {
	a.epochStore.Put([]byte("epoch."+key), value)
	return nil
}

func (a *RBFTAdaptor) ReadEpochState(key string) ([]byte, error) {
	b := a.epochStore.Get([]byte("epoch." + key))
	if b == nil {
		return nil, errors.New("not found")
	}
	return b, nil
}
