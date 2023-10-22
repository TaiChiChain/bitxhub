package system

import (
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/access"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/base"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/governance"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

// addr2ContractConstruct is address to system contract
var addr2ContractConstruct map[types.Address]common.SystemContractConstruct

var globalCfg = &common.SystemContractConfig{
	Logger: logrus.New(),
}

func init() {
	addr2ContractConstruct = map[types.Address]common.SystemContractConstruct{
		*types.NewAddressByStr(common.EpochManagerContractAddr): func(cfg *common.SystemContractConfig) common.SystemContract {
			return base.NewEpochManager(cfg)
		},
		*types.NewAddressByStr(common.NodeManagerContractAddr): func(cfg *common.SystemContractConfig) common.SystemContract {
			return governance.NewNodeManager(cfg)
		},
		*types.NewAddressByStr(common.CouncilManagerContractAddr): func(cfg *common.SystemContractConfig) common.SystemContract {
			return governance.NewCouncilManager(cfg)
		},
		*types.NewAddressByStr(common.WhiteListProviderManagerContractAddr): func(cfg *common.SystemContractConfig) common.SystemContract {
			return governance.NewWhiteListProviderManager(cfg)
		},
		*types.NewAddressByStr(common.WhiteListContractAddr): func(cfg *common.SystemContractConfig) common.SystemContract {
			return access.NewWhiteList(cfg)
		},
	}
}

func Initialize(logger logrus.FieldLogger) {
	globalCfg.Logger = logger
}

// GetSystemContract get system contract
// return true if system contract, false if not
func GetSystemContract(addr *types.Address) (common.SystemContract, bool) {
	if addr == nil {
		return nil, false
	}

	if contractConstruct, ok := addr2ContractConstruct[*addr]; ok {
		return contractConstruct(globalCfg), true
	}
	return nil, false
}

func InitGenesisData(genesis *repo.Genesis, lg ledger.StateLedger) error {
	if err := base.InitEpochInfo(lg, genesis.EpochInfo.Clone()); err != nil {
		return err
	}
	if err := governance.InitCouncilMembers(lg, genesis.Admins, genesis.Balance); err != nil {
		return err
	}

	admins := lo.Map[*repo.Admin, string](genesis.Admins, func(x *repo.Admin, _ int) string {
		return x.Address
	})
	totalLength := len(admins) + len(genesis.InitWhiteListProviders) + len(genesis.Accounts)
	combined := make([]string, 0, totalLength)
	combined = append(combined, admins...)
	combined = append(combined, genesis.InitWhiteListProviders...)
	combined = append(combined, genesis.Accounts...)
	if err := access.InitProvidersAndWhiteList(lg, combined, genesis.InitWhiteListProviders); err != nil {
		return err
	}
	return nil
}

// CheckAndUpdateAllState check and update all system contract state if need
func CheckAndUpdateAllState(lastHeight uint64, stateLedger ledger.StateLedger) {
	for _, contractConstruct := range addr2ContractConstruct {
		contractConstruct(globalCfg).CheckAndUpdateState(lastHeight, stateLedger)
	}
}
