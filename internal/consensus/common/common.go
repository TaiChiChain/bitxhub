package common

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	dagtypes "github.com/bcds/go-hpc-dagbft/common/types"
)

const (
	EpochStatePrefix = "epoch_q_chkpt."
	EpochIndexKey    = "epoch_latest_idx"
)

const MaxChainSize = 1000

func NeedChangeEpoch(height uint64, epochInfo types.EpochInfo) bool {
	return height == (epochInfo.StartBlock + epochInfo.EpochPeriod - 1)
}

func CalFaulty(N uint64) uint64 {
	f := (N - 1) / 3
	return f
}

func CalQuorum(N uint64) uint64 {
	f := (N - 1) / 3
	return (N + f + 2) / 2
}

func GenNodeDbPath(config *Config, name string, epoch dagtypes.Epoch) string {
	storePath := repo.GetStoragePath(config.Repo.RepoRoot, storagemgr.Consensus)
	var fileDir string
	if epoch > 0 {
		fileDir = filepath.Join(storePath, fmt.Sprintf("%s-%d", storagemgr.Epoch, epoch), name)
	} else {
		fileDir = filepath.Join(storePath, storagemgr.Ledger, name)
	}
	err := os.MkdirAll(fileDir, os.ModePerm)
	if err != nil {
		panic(err)
	}
	return fileDir
}

func DrainChannel[T any](ch chan T) {
DrainLoop:
	for {
		select {
		case _, ok := <-ch:
			if !ok { //ch is closed //immediately return err
				break DrainLoop
			}
		default: //all other case not-ready: means nothing in ch for now
			break DrainLoop
		}
	}
}
