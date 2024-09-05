package precheck

import (
	"github.com/axiomesh/axiom-kit/types"
	consensustypes "github.com/axiomesh/axiom-ledger/internal/consensus/types"
)

//go:generate mockgen -destination mock_precheck/mock_precheck.go -package mock_precheck -source precheck.go -typed
type PreCheck interface {
	// Start the precheck service
	Start()

	// PostUncheckedTxEvent posts unchecked tx event to precheckMgr
	PostUncheckedTxEvent(ev *consensustypes.UncheckedTxEvent)

	UpdateEpochInfo(epoch *types.EpochInfo)

	BasicCheckTx(tx *types.Transaction) error

	VerifySignature(tx *types.Transaction) error
}
