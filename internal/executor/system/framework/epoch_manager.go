package framework

import (
	"strconv"

	"github.com/pkg/errors"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework/solidity/epoch_manager"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework/solidity/epoch_manager_client"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

const (
	EpochManagerCurrentEpochIDStorageKey   = "currentEpochID"
	EpochManagerNextEpochInfoStorageKey    = "nextEpochInfo"
	EpochManagerHistoryEpochInfoStorageKey = "historyEpochInfo"
)

var EpochManagerBuildConfig = &common.SystemContractBuildConfig[*EpochManager]{
	Name:    "framework_epoch_manager",
	Address: common.EpochManagerContractAddr,
	AbiStr:  epoch_manager_client.BindingContractMetaData.ABI,
	Constructor: func(systemContractBase common.SystemContractBase) *EpochManager {
		return &EpochManager{
			SystemContractBase: systemContractBase,
		}
	},
}

// var _ epoch_manager.EpochManager = (*EpochManager)(nil)
type EpochManager struct {
	common.SystemContractBase

	currentEpochID *common.VMSlot[uint64]
	nextEpochInfo  *common.VMSlot[epoch_manager.EpochInfo]

	// history epoch id -> history epoch info
	historyEpochInfoMap *common.VMMap[uint64, epoch_manager.EpochInfo]
}

func (m *EpochManager) GenesisInit(genesis *repo.GenesisConfig) error {
	if genesis.EpochInfo.StartBlock != 0 {
		return errors.Errorf("genesis epoch start block must be 0")
	}
	if genesis.EpochInfo.Epoch != 1 {
		return errors.Errorf("genesis epoch number must be 1")
	}

	if err := genesis.EpochInfo.Validate(); err != nil {
		return errors.Wrapf(err, "invalid genesis epoch info")
	}

	epochInfo := genesis.EpochInfo.Clone()
	if err := m.historyEpochInfoMap.Put(epochInfo.Epoch, epoch_manager.FromTypesEpoch(*epochInfo)); err != nil {
		return err
	}

	if err := m.currentEpochID.Put(epochInfo.Epoch); err != nil {
		return err
	}

	epochInfo.Epoch++
	epochInfo.StartBlock += epochInfo.EpochPeriod
	if err := m.nextEpochInfo.Put(epoch_manager.FromTypesEpoch(*epochInfo)); err != nil {
		return err
	}
	return nil
}

func (m *EpochManager) SetContext(context *common.VMContext) {
	m.SystemContractBase.SetContext(context)

	m.currentEpochID = common.NewVMSlot[uint64](m.StateAccount, EpochManagerCurrentEpochIDStorageKey)
	m.nextEpochInfo = common.NewVMSlot[epoch_manager.EpochInfo](m.StateAccount, EpochManagerNextEpochInfoStorageKey)
	m.historyEpochInfoMap = common.NewVMMap[uint64, epoch_manager.EpochInfo](m.StateAccount, EpochManagerHistoryEpochInfoStorageKey, func(id uint64) string {
		return strconv.FormatUint(id, 10)
	})
}

func (m *EpochManager) CurrentEpoch() (epoch_manager.EpochInfo, error) {
	currentEpochID, err := m.currentEpochID.MustGet()
	if err != nil {
		return epoch_manager.EpochInfo{}, err
	}

	return m.HistoryEpoch(currentEpochID)
}

func (m *EpochManager) NextEpoch() (epoch_manager.EpochInfo, error) {
	epochInfo, err := m.nextEpochInfo.MustGet()
	if err != nil {
		return epoch_manager.EpochInfo{}, err
	}
	return epochInfo, nil
}

func (m *EpochManager) HistoryEpoch(epochID uint64) (epoch_manager.EpochInfo, error) {
	epochInfo, err := m.historyEpochInfoMap.MustGet(epochID)
	if err != nil {
		return epoch_manager.EpochInfo{}, err
	}
	return epochInfo, nil
}

func (m *EpochManager) UpdateNextEpoch(nextEpochInfo epoch_manager.EpochInfo) error {
	return m.nextEpochInfo.Put(nextEpochInfo)
}

func (m *EpochManager) TurnIntoNewEpoch() (epoch_manager.EpochInfo, error) {
	oldNextEpochInfo, err := m.nextEpochInfo.MustGet()
	if err != nil {
		return epoch_manager.EpochInfo{}, err
	}

	if err := m.historyEpochInfoMap.Put(oldNextEpochInfo.Epoch, oldNextEpochInfo); err != nil {
		return epoch_manager.EpochInfo{}, err
	}

	if err := m.currentEpochID.Put(oldNextEpochInfo.Epoch); err != nil {
		return epoch_manager.EpochInfo{}, err
	}

	newNextEpoch := oldNextEpochInfo
	newNextEpoch.Epoch++
	newNextEpoch.StartBlock += newNextEpoch.EpochPeriod
	if err := m.nextEpochInfo.Put(newNextEpoch); err != nil {
		return epoch_manager.EpochInfo{}, err
	}
	return oldNextEpochInfo, nil
}
