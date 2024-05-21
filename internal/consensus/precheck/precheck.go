package precheck

import (
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
)

//go:generate mockgen -destination mock_precheck/mock_precheck.go -package mock_precheck -source precheck.go -typed
type PreCheck interface {
	// Start the precheck service
	Start()

	// PostUncheckedTxEvent posts unchecked tx event to precheckMgr
	PostUncheckedTxEvent(ev *common.UncheckedTxEvent)

	UpdateEpochInfo(epoch *types.EpochInfo)
}
