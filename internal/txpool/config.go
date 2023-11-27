package txpool

import (
	"time"

	"github.com/sirupsen/logrus"
)

// Config defines the txpool config items.
type Config struct {
	Logger                logrus.FieldLogger
	BatchSize             uint64
	PoolSize              uint64
	BatchMemLimit         bool
	BatchMaxMem           uint64
	IsTimed               bool
	ToleranceNonceGap     uint64
	ToleranceTime         time.Duration
	ToleranceRemoveTime   time.Duration
	CleanEmptyAccountTime time.Duration
	GetAccountNonce       GetAccountNonceFunc
}
