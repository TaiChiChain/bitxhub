package txpool

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

// Config defines the txpool config items.
type Config struct {
	RepoRoot               string
	ConsensusMode          string
	Logger                 logrus.FieldLogger
	PoolSize               uint64
	ToleranceNonceGap      uint64
	ToleranceTime          time.Duration
	ToleranceRemoveTime    time.Duration
	CleanEmptyAccountTime  time.Duration
	RotateTxLocalsInterval time.Duration
	GetAccountNonce        GetAccountNonceFunc
	GetAccountBalance      GetAccountBalanceFunc
	EnableLocalsPersist    bool
	PriceLimit             uint64
	PriceBump              uint64
	GenerateBatchType      string
}

// sanitize checks the provided user configurations and changes anything that's
// unreasonable or unworkable.
func (c *Config) sanitize() {
	if c.ConsensusMode == "" {
		c.ConsensusMode = repo.ConsensusTypeDagBft
	}
	if c.PoolSize == 0 {
		c.PoolSize = DefaultPoolSize
	}
	if c.ToleranceTime == 0 {
		c.ToleranceTime = DefaultToleranceTime
	}
	if c.ToleranceRemoveTime == 0 {
		c.ToleranceRemoveTime = DefaultToleranceRemoveTime
	}
	if c.CleanEmptyAccountTime == 0 {
		c.CleanEmptyAccountTime = DefaultCleanEmptyAccountTime
	}
	if c.ToleranceNonceGap == 0 {
		c.ToleranceNonceGap = DefaultToleranceNonceGap
	}
	if c.RotateTxLocalsInterval == 0 {
		c.RotateTxLocalsInterval = DefaultRotateTxLocalsInterval
	}

	if c.GenerateBatchType != repo.GenerateBatchByTime && c.GenerateBatchType != repo.GenerateBatchByGasPrice {
		c.GenerateBatchType = repo.GenerateBatchByTime
	}
	if c.PriceBump < DefaultPriceBump {
		c.PriceBump = DefaultPriceBump
	}
}
