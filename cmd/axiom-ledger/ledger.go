package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/axiomesh/axiom-kit/fileutil"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/app"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/executor"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/base"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/genesis"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

var ledgerGetBlockArgs = struct {
	Number uint64
	Hash   string
	Full   bool
}{}

var ledgerSimpleRollbackArgs = struct {
	TargetBlockNumber uint64
	Force             bool
}{}

var ledgerSimpleSyncArgs = struct {
	TargetBlockNumber uint64
	SourceStorage     string
	TargetStorage     string
	Force             bool
}{}

var ledgerGenerateTrieArgs = struct {
	TargetBlockNumber uint64
	TargetStoragePath string
}{}

var ledgerCMD = &cli.Command{
	Name:  "ledger",
	Usage: "The ledger manage commands",
	Subcommands: []*cli.Command{
		{
			Name:   "block",
			Usage:  "Get block info by number or hash, if not specified, get the latest",
			Action: getBlock,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:        "full",
					Aliases:     []string{"f"},
					Usage:       "additionally display transactions",
					Destination: &ledgerGetBlockArgs.Full,
					Required:    false,
				},
				&cli.Uint64Flag{
					Name:        "number",
					Aliases:     []string{"n"},
					Usage:       "block number",
					Destination: &ledgerGetBlockArgs.Number,
					Required:    false,
				},
				&cli.StringFlag{
					Name:        "hash",
					Usage:       "block hash",
					Destination: &ledgerGetBlockArgs.Hash,
					Required:    false,
				},
			},
		},
		{
			Name:   "chain-meta",
			Usage:  "Get latest chain meta info",
			Action: getLatestChainMeta,
		},
		{
			Name:   "rollback",
			Usage:  "Rollback ledger to the specific block history height",
			Action: rollback,
			Flags: []cli.Flag{
				&cli.Uint64Flag{
					Name:        "target-block-number",
					Aliases:     []string{"b"},
					Usage:       "rollback target block number, must be less than the current latest block height and greater than 1",
					Destination: &ledgerSimpleRollbackArgs.TargetBlockNumber,
					Required:    true,
				},
				&cli.BoolFlag{
					Name:        "force",
					Aliases:     []string{"f"},
					Usage:       "disable interactive confirmation and remove existing rollback storage directory of the same height",
					Destination: &ledgerSimpleRollbackArgs.Force,
					Required:    false,
				},
			},
		},
		{
			Name:   "simple-sync",
			Usage:  "Sync to the specified block height from the target storage(by replaying transactions)",
			Action: simpleSync,
			Flags: []cli.Flag{
				&cli.Uint64Flag{
					Name:        "target-block-number",
					Aliases:     []string{"b"},
					Usage:       "sync target block number, must be less than or equal to the source-storage latest block height and greater than the target-storage latest block height",
					Destination: &ledgerSimpleSyncArgs.TargetBlockNumber,
					Required:    true,
				},
				&cli.StringFlag{
					Name:        "source-storage",
					Aliases:     []string{"s"},
					Usage:       "sync from storage dir",
					Destination: &ledgerSimpleSyncArgs.SourceStorage,
					Required:    true,
				},
				&cli.StringFlag{
					Name:        "target-storage",
					Aliases:     []string{"t"},
					Usage:       "sync to storage dir",
					Destination: &ledgerSimpleSyncArgs.TargetStorage,
					Required:    true,
				},
				&cli.BoolFlag{
					Name:        "force",
					Aliases:     []string{"f"},
					Usage:       "disable interactive confirmation",
					Destination: &ledgerSimpleSyncArgs.Force,
					Required:    false,
				},
			},
		},
		{
			Name:   "generate-trie",
			Usage:  "Generate world state trie at specific block",
			Action: generateTrie,
			Flags: []cli.Flag{
				&cli.Uint64Flag{
					Name:        "target-block-number",
					Aliases:     []string{"b"},
					Usage:       "block number of target trie, must be less than or equal to the latest block height",
					Destination: &ledgerGenerateTrieArgs.TargetBlockNumber,
					Required:    true,
				},
				&cli.StringFlag{
					Name:        "target-storage",
					Aliases:     []string{"t"},
					Usage:       "directory to store trie instance",
					Destination: &ledgerGenerateTrieArgs.TargetStoragePath,
					Required:    true,
				},
			},
		},
	},
}

func getBlock(ctx *cli.Context) error {
	r, err := prepareRepo(ctx)
	if err != nil {
		return err
	}

	chainLedger, err := ledger.NewChainLedger(r, "")
	if err != nil {
		return fmt.Errorf("init chain ledger failed: %w", err)
	}

	var block *types.Block

	if ctx.IsSet("number") {
		block, err = chainLedger.GetBlock(ledgerGetBlockArgs.Number)
		if err != nil {
			return err
		}
	} else {
		if ctx.IsSet("hash") {
			block, err = chainLedger.GetBlockByHash(types.NewHashByStr(ledgerGetBlockArgs.Hash))
			if err != nil {
				return err
			}
		} else {
			block, err = chainLedger.GetBlock(chainLedger.GetChainMeta().Height)
			if err != nil {
				return err
			}
		}
	}

	bloom, _ := block.BlockHeader.Bloom.ETHBloom().MarshalText()
	blockInfo := map[string]any{
		"number":           block.BlockHeader.Number,
		"hash":             block.Hash().String(),
		"state_root":       block.BlockHeader.StateRoot.String(),
		"tx_root":          block.BlockHeader.TxRoot.String(),
		"receipt_root":     block.BlockHeader.ReceiptRoot.String(),
		"parent_hash":      block.BlockHeader.ParentHash.String(),
		"timestamp":        block.BlockHeader.Timestamp,
		"epoch":            block.BlockHeader.Epoch,
		"bloom":            string(bloom),
		"gas_price":        block.BlockHeader.GasPrice,
		"proposer_account": block.BlockHeader.ProposerAccount,
		"tx_count":         len(block.Transactions),
	}

	if ledgerGetBlockArgs.Full {
		blockInfo["transactions"] = lo.Map(block.Transactions, func(item *types.Transaction, index int) string {
			return item.GetHash().String()
		})
	}

	return pretty(blockInfo)
}

func getLatestChainMeta(ctx *cli.Context) error {
	r, err := prepareRepo(ctx)
	if err != nil {
		return err
	}

	chainLedger, err := ledger.NewChainLedger(r, "")
	if err != nil {
		return fmt.Errorf("init chain ledger failed: %w", err)
	}

	meta := chainLedger.GetChainMeta()
	return pretty(meta)
}

func fetchAndExecuteBlocks(ctx context.Context, r *repo.Repo, targetLedger *ledger.Ledger, fetchBlock func(n uint64) (*types.Block, error), targetBlockNumber uint64) error {
	if targetLedger.ChainLedger.GetChainMeta().Height == 0 {
		if err := genesis.Initialize(r.GenesisConfig, targetLedger); err != nil {
			return err
		}
		logger := loggers.Logger(loggers.App)
		logger.WithFields(logrus.Fields{
			"genesis block hash": targetLedger.ChainLedger.GetChainMeta().BlockHash,
		}).Info("Initialize genesis")
	}

	e, err := executor.New(r, targetLedger)
	if err != nil {
		return fmt.Errorf("init executor failed: %w", err)
	}

	blockCh := make(chan *common.CommitEvent, 100)
	go func() {
		for i := targetLedger.ChainLedger.GetChainMeta().Height + 1; i <= targetBlockNumber; i++ {
			b, err := fetchBlock(i)
			if err != nil {
				panic(errors.Wrapf(err, "failed to get block %d", i))
			}
			select {
			case <-ctx.Done():
				return
			case blockCh <- &common.CommitEvent{
				Block: b,
			}:
			}
		}
		blockCh <- nil
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		case b := <-blockCh:
			if b == nil {
				return nil
			}
			var originBlockHash string
			if b.Block.Height()%10 == 0 || b.Block.Height() == targetBlockNumber {
				originBlockHash = b.Block.BlockHeader.Hash().String()
			}

			e.ExecuteBlock(b)

			// check hash
			if b.Block.Height()%10 == 0 || b.Block.Height() == targetBlockNumber {
				if originBlockHash != targetLedger.ChainLedger.GetChainMeta().BlockHash.String() {
					panic(fmt.Sprintf("inconsistent block %d hash, get %s, but want %s", b.Block.Height(), targetLedger.ChainLedger.GetChainMeta().BlockHash.String(), originBlockHash))
				}
			}
		}
	}
}

func rollback(ctx *cli.Context) error {
	r, err := prepareRepo(ctx)
	if err != nil {
		return err
	}
	logger := loggers.Logger(loggers.App)
	originBlockchainDir := repo.GetStoragePath(r.RepoRoot, storagemgr.BlockChain)
	originBlockfileDir := repo.GetStoragePath(r.RepoRoot, storagemgr.Blockfile)
	originStateLedgerDir := repo.GetStoragePath(r.RepoRoot, storagemgr.Ledger)

	originChainLedger, err := ledger.NewChainLedger(r, "")
	if err != nil {
		return fmt.Errorf("init chain ledger failed: %w", err)
	}

	// check if target height is legal
	targetBlockNumber := ledgerSimpleRollbackArgs.TargetBlockNumber
	if targetBlockNumber <= 1 {
		return errors.New("target-block-number must be greater than 1")
	}
	chainMeta := originChainLedger.GetChainMeta()
	if targetBlockNumber >= chainMeta.Height {
		return errors.Errorf("target-block-number %d must be less than the current latest block height %d\n", targetBlockNumber, chainMeta.Height)
	}

	rollbackDir := path.Join(r.RepoRoot, fmt.Sprintf("storage-rollback-%d", targetBlockNumber))
	force := ledgerSimpleRollbackArgs.Force
	if fileutil.Exist(rollbackDir) {
		if !force {
			return errors.Errorf("rollback dir %s already exists\n", rollbackDir)
		}
		if err := os.RemoveAll(rollbackDir); err != nil {
			return err
		}
	}
	targetBlock, err := originChainLedger.GetBlock(targetBlockNumber)
	if err != nil {
		return fmt.Errorf("get target block failed: %w", err)
	}
	if force {
		logger.Infof("current chain meta info height: %d, hash: %s, will rollback to the target height %d, hash: %s, rollback storage dir: %s\n", chainMeta.Height, chainMeta.BlockHash, targetBlockNumber, targetBlock.Hash(), rollbackDir)
	} else {
		logger.Infof("current chain meta info height: %d, hash: %s, will rollback to the target height %d, hash: %s, rollback storage dir: %s, confirm? y/n\n", chainMeta.Height, chainMeta.BlockHash, targetBlockNumber, targetBlock.Hash(), rollbackDir)
		if err := waitUserConfirm(); err != nil {
			return err
		}
	}

	// copy blockchain dir
	targetBlockchainDir := path.Join(rollbackDir, storagemgr.BlockChain)
	if err := os.MkdirAll(targetBlockchainDir, os.ModePerm); err != nil {
		return errors.Errorf("mkdir blockchain dir error: %v", err.Error())
	}
	if err := copyDir(originBlockchainDir, targetBlockchainDir); err != nil {
		return errors.Errorf("copy blockchain dir to rollback dir error: %v", err.Error())
	}
	logger.Infof("copy blockchain dir to rollback dir success")

	// copy blockfile dir
	targetBlockfileDir := path.Join(rollbackDir, storagemgr.Blockfile)
	if err := os.MkdirAll(targetBlockfileDir, os.ModePerm); err != nil {
		return errors.Errorf("mkdir blockfile dir error: %v", err.Error())
	}
	if err := copyDir(originBlockfileDir, targetBlockfileDir); err != nil {
		return errors.Errorf("copy blockfile dir to rollback dir error: %v", err.Error())
	}
	logger.Infof("copy blockfile dir to rollback dir success")

	// copy state ledger dir
	targetStateLedgerDir := path.Join(rollbackDir, storagemgr.Ledger)
	if err := os.MkdirAll(targetStateLedgerDir, os.ModePerm); err != nil {
		return errors.Errorf("mkdir state ledger dir error: %v", err.Error())
	}
	if err := copyDir(originStateLedgerDir, targetStateLedgerDir); err != nil {
		return errors.Errorf("copy state ledger dir to rollback dir error: %v", err.Error())
	}
	logger.Infof("copy state ledger dir to rollback dir success")

	// rollback chain ledger
	rollbackChainLedger, err := ledger.NewChainLedger(r, rollbackDir)
	if err != nil {
		return fmt.Errorf("init rollback chain ledger failed: %w", err)
	}
	if err := rollbackChainLedger.RollbackBlockChain(targetBlockNumber); err != nil {
		return errors.Errorf("rollback chain ledger error: %v", err.Error())
	}

	// rollback state ledger
	rollbackStateLedger, err := ledger.NewStateLedger(r, rollbackDir)
	if err != nil {
		return fmt.Errorf("init rollback state ledger failed: %w", err)
	}
	if err := rollbackStateLedger.RollbackState(targetBlockNumber, targetBlock.BlockHeader.StateRoot); err != nil {
		return fmt.Errorf("rollback state ledger failed: %w", err)
	}

	// wait for generating snapshot of target block
	errC := make(chan error)
	go rollbackStateLedger.GenerateSnapshot(targetBlock, errC)
	err = <-errC
	if err != nil {
		return fmt.Errorf("generate snapshot failed: %w", err)
	}

	return nil
}

func simpleSync(ctx *cli.Context) error {
	r, err := prepareRepo(ctx)
	if err != nil {
		return err
	}

	sourceChainLedger, err := ledger.NewChainLedger(r, ledgerSimpleSyncArgs.SourceStorage)
	if err != nil {
		return fmt.Errorf("init source chain ledger failed: %w", err)
	}

	targetChainLedger, err := ledger.NewChainLedger(r, ledgerSimpleSyncArgs.TargetStorage)
	if err != nil {
		return fmt.Errorf("init target chain ledger failed: %w", err)
	}

	targetBlockNumber := ledgerSimpleSyncArgs.TargetBlockNumber
	sourceChainMeta := sourceChainLedger.GetChainMeta()
	targetChainMeta := targetChainLedger.GetChainMeta()
	if targetBlockNumber > sourceChainMeta.Height || targetBlockNumber <= targetChainMeta.Height {
		return errors.Errorf("target-block-number %d must be less than or equal to the source-storage latest block height %d and greater than the target-storage latest block height %d\n", targetBlockNumber, sourceChainMeta.Height, targetChainMeta.Height)
	}
	targetBlockHash := targetChainLedger.GetBlockHash(targetBlockNumber).String()
	logger := loggers.Logger(loggers.App)
	if ledgerSimpleSyncArgs.Force {
		logger.Infof("target storage current chain meta info height: %d, hash: %s, will sync to the target height %d, hash: %s, target storage dir: %s, source storage dir: %s\n", targetChainMeta.Height, targetChainMeta.BlockHash, targetBlockNumber, targetBlockHash, ledgerSimpleSyncArgs.TargetStorage, ledgerSimpleSyncArgs.SourceStorage)
	} else {
		logger.Infof("target storage current chain meta info height: %d, hash: %s, will sync to the target height %d, hash: %s, target storage dir: %s, source storage dir: %s, confirm? y/n\n", targetChainMeta.Height, targetChainMeta.BlockHash, targetBlockNumber, targetBlockHash, ledgerSimpleSyncArgs.TargetStorage, ledgerSimpleSyncArgs.SourceStorage)
		if err := waitUserConfirm(); err != nil {
			return err
		}
	}

	targetStateLedger, err := ledger.NewStateLedger(r, ledgerSimpleSyncArgs.TargetStorage)
	if err != nil {
		return fmt.Errorf("init target state ledger failed: %w", err)
	}

	return fetchAndExecuteBlocks(ctx.Context, r, &ledger.Ledger{
		ChainLedger: targetChainLedger,
		StateLedger: targetStateLedger,
	}, func(n uint64) (*types.Block, error) {
		return sourceChainLedger.GetBlock(n)
	}, targetBlockNumber)
}

func generateTrie(ctx *cli.Context) error {
	logger := loggers.Logger(loggers.App)
	logger.Infof("start generating trie at height: %v\n", ledgerGenerateTrieArgs.TargetBlockNumber)

	r, err := prepareRepo(ctx)
	if err != nil {
		return err
	}

	chainLedger, err := ledger.NewChainLedger(r, "")
	if err != nil {
		return fmt.Errorf("init chain ledger failed: %w", err)
	}
	block, err := chainLedger.GetBlock(ledgerGenerateTrieArgs.TargetBlockNumber)
	if err != nil {
		return fmt.Errorf("get block failed: %w", err)
	}

	targetStateStoragePath := path.Join(repo.GetStoragePath(r.RepoRoot, storagemgr.Sync), ledgerGenerateTrieArgs.TargetStoragePath)
	targetStateStorage, err := storagemgr.Open(targetStateStoragePath)
	if err != nil {
		return fmt.Errorf("create targetStateStorage: %w", err)
	}

	originStateLedger, err := ledger.NewStateLedger(r, "")
	if err != nil {
		return fmt.Errorf("init state ledger failed: %w", err)
	}
	originStateLedger.(*ledger.StateLedgerImpl).WithGetEpochInfoFunc(base.GetEpochInfo)

	errC := make(chan error)
	go originStateLedger.IterateTrie(block, targetStateStorage, errC)
	err = <-errC

	logger.Infof("success generating trie at height: %v\n", ledgerGenerateTrieArgs.TargetBlockNumber)

	return err
}

func prepareRepo(ctx *cli.Context) (*repo.Repo, error) {
	p, err := getRootPath(ctx)
	if err != nil {
		return nil, err
	}
	if !fileutil.Exist(filepath.Join(p, repo.CfgFileName)) {
		return nil, errors.New("axiom-ledger repo not exist")
	}

	r, err := repo.Load(configGenerateArgs.Auth, p, false)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%s-repo: %s\n", repo.AppName, r.RepoRoot)

	if err := loggers.Initialize(ctx.Context, r, false); err != nil {
		return nil, err
	}

	if err := app.PrepareAxiomLedger(r); err != nil {
		return nil, fmt.Errorf("prepare axiom-ledger failed: %w", err)
	}
	return r, nil
}

func waitUserConfirm() error {
	var choice string
	if _, err := fmt.Scanln(&choice); err != nil {
		return err
	}
	if choice != "y" {
		return errors.New("interrupt by user")
	}
	return nil
}

func pretty(d any) error {
	res, err := json.MarshalIndent(d, "", "\t")
	if err != nil {
		return err
	}
	fmt.Println(string(res))
	return nil
}

func copyDir(src, dest string) error {
	files, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, file := range files {
		srcPath := filepath.Join(src, file.Name())
		destPath := filepath.Join(dest, file.Name())

		if file.IsDir() {
			if err := copyDir(srcPath, destPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, destPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}
