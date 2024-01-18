package txpool

import (
	"time"

	commonpool "github.com/axiomesh/axiom-kit/txpool"
	"github.com/sirupsen/logrus"
)

// Config defines the txpool config items.
type Config struct {
	Logger                 logrus.FieldLogger
	ChainInfo              *commonpool.ChainInfo
	BatchSize              uint64
	PoolSize               uint64
	BatchMemLimit          bool
	BatchMaxMem            uint64
	IsTimed                bool
	ToleranceNonceGap      uint64
	ToleranceTime          time.Duration
	ToleranceRemoveTime    time.Duration
	CleanEmptyAccountTime  time.Duration
	RotateTxLocalsInterval time.Duration
	GetAccountNonce        GetAccountNonceFunc
	GetAccountBalance      GetAccountBalanceFunc
	EnableLocalsPersist    bool
	TxRecordsFile          string
}
