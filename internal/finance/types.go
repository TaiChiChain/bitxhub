package finance

import "github.com/pkg/errors"

var (
	ErrTxsOutOfRange        = errors.New("current txs is out of range")
	ErrGasOutOfRange        = errors.New("parent gas price is out of range")
	ErrNoIncentiveAddrFound = errors.New("incentive part in genesis config is required")
	ErrAXCContractType      = errors.New("axc contract not implement SystemContract Interface")
	// ErrMiningRewardExceeds = errors.New("the mining rewards exceeds remaining value")
)

type MiningRules struct {
	startBlock uint64
	endBlock   uint64
	percentage float64
}
