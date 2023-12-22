package adaptor

import (
	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
)

func (a *RBFTAdaptor) GetCurrentEpochInfo() (*rbft.EpochInfo, error) {
	return a.config.GetCurrentEpochInfoFromEpochMgrContractFunc()
}

func (a *RBFTAdaptor) GetEpochInfo(epoch uint64) (*rbft.EpochInfo, error) {
	return a.config.GetEpochInfoFromEpochMgrContractFunc(epoch)
}

func (a *RBFTAdaptor) StoreEpochState(key string, value []byte) error {
	return common.StoreEpochState(a.epochStore, key, value)
}

func (a *RBFTAdaptor) ReadEpochState(key string) ([]byte, error) {
	return common.ReadEpochState(a.epochStore, key)
}
