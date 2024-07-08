package main

import (
	"bufio"
	"fmt"
	"io"
	"math/big"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/urfave/cli/v2"

	"github.com/axiomesh/axiom-bft/common/consensus"
	"github.com/axiomesh/axiom-kit/fileutil"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/cmd/axiom-ledger/common"
	"github.com/axiomesh/axiom-ledger/internal/app"
	syscommon "github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/utils"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

// maxBatchSize defines the maximum size of the data in single batch write operation, which is 64 MB.
const maxBatchSize = 64 * 1024 * 1024

var ledgerGetBlockArgs = struct {
	Number uint64
	Hash   string
	Full   bool
}{}

var ledgerGetTxArgs = struct {
	Hash string
}{}

var ledgerSimpleRollbackArgs = struct {
	TargetBlockNumber uint64
	Force             bool
	DisableBackup     bool
}{}

var ledgerGenerateTrieArgs = struct {
	TargetBlockNumber uint64
	TargetStoragePath string
	remotePeers       cli.StringSlice
}{}

var ledgerImportAccountsArgs = struct {
	TargetFilePath string
	Balance        string
	BatchSize      uint
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
			Name:   "tx",
			Usage:  "Get tx info by hash",
			Action: getTx,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "hash",
					Usage:       "tx hash",
					Destination: &ledgerGetTxArgs.Hash,
					Required:    true,
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
				&cli.BoolFlag{
					Name:        "disable-backup",
					Aliases:     []string{"d"},
					Usage:       "disable backup original ledger folder",
					Destination: &ledgerSimpleRollbackArgs.DisableBackup,
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
				&cli.StringSliceFlag{
					Name:        "peers",
					Usage:       `list peers which have the same state, format: "1:p2pID"`,
					Aliases:     []string{`p`},
					Destination: &ledgerGenerateTrieArgs.remotePeers,
					Required:    true,
				},
			},
		},
		{
			Name:   "import-accounts",
			Usage:  "used after generating the genesis block, where large number of accounts with preset balances are inserted into ledger. This process aims to assist testers in initializing accounts prior to conducting tests. The file should contain one Ethereum Hexadecimal Address per line.",
			Action: importAccounts,
			Flags: []cli.Flag{
				&cli.PathFlag{
					Name:        "account-file",
					Usage:       "get accounts from this file",
					Required:    true,
					Destination: &ledgerImportAccountsArgs.TargetFilePath,
					Aliases:     []string{"path"},
				},
				&cli.StringFlag{
					Name:        "balance",
					Usage:       "add this balance for all accounts",
					Required:    true,
					Destination: &ledgerImportAccountsArgs.Balance,
				},
				&cli.UintFlag{
					Name:        "batch-size",
					Usage:       "number of accounts to import in one block",
					Required:    false,
					Destination: &ledgerImportAccountsArgs.BatchSize,
				},
			},
		},
	},
}

func importAccounts(ctx *cli.Context) error {
	p, err := common.GetRootPath(ctx)
	if err != nil {
		return err
	}

	if !fileutil.Exist(ledgerImportAccountsArgs.TargetFilePath) {
		return errors.New("target account file not exist")
	}

	rep, err := repo.Load(p)
	if err != nil {
		return err
	}

	if err = app.PrepareAxiomLedger(rep); err != nil {
		return err
	}

	rwLdg, err := ledger.NewLedger(rep)
	if err != nil {
		return err
	}
	// check genesis block
	if _, err = rwLdg.ChainLedger.GetBlock(0); err != nil {
		return errors.New("genesis block is needed before importing accounts from file")
	}

	return importAccountsFromFile(rep, rwLdg, ledgerImportAccountsArgs.TargetFilePath, ledgerImportAccountsArgs.Balance, ledgerImportAccountsArgs.BatchSize)
}

func importAccountsFromFile(r *repo.Repo, lg *ledger.Ledger, filePath string, balanceStr string, size uint) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("account list file does not exist: %s, error: %v", filePath, err)
	}
	totalLines, err := countLines(filePath)
	if err != nil {
		return fmt.Errorf("failed to count lines in account list file: %v", err)
	}
	batchSize := int(size)
	if batchSize == 0 {
		batchSize = 10000
	}
	totalBatches := totalLines / batchSize
	if totalLines%batchSize > 0 {
		totalBatches++
	}
	epochSize := r.GenesisConfig.EpochInfo.EpochPeriod
	currentHeight := lg.ChainLedger.GetChainMeta().Height
	currentEpochNumber := (currentHeight - 1) / epochSize
	endOfCurrentEpoch := (currentEpochNumber + 1) * epochSize
	remainingBlocksInEpoch := endOfCurrentEpoch - currentHeight
	if totalBatches > int(remainingBlocksInEpoch) {
		return fmt.Errorf("the number of batches %d is larger than the epoch period %d, please increase the epochsize", totalBatches, r.GenesisConfig.EpochInfo.EpochPeriod)
	}
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open account list file: %v", err)
	}
	defer func() {
		_ = file.Close()
	}()
	scanner := bufio.NewScanner(file)
	currentLine := 0
	for batch := 0; batch < totalBatches; batch++ {
		currentHeight := lg.ChainLedger.GetChainMeta().Height
		fmt.Printf("current height: %d\n", currentHeight+1)
		parentBlockHeader, err := lg.ChainLedger.GetBlockHeader(currentHeight)
		if err != nil {
			return err
		}
		lg.StateLedger.PrepareBlock(parentBlockHeader.StateRoot, currentHeight+1)
		startLine := batch * batchSize
		endLine := startLine + batchSize
		if endLine > totalLines {
			endLine = totalLines
		}
		for currentLine < endLine && scanner.Scan() {
			address := scanner.Text()
			account := lg.StateLedger.GetOrCreateAccount(types.NewAddressByStr(address))
			balance, ok := new(big.Int).SetString(balanceStr, 10)
			if !ok {
				return fmt.Errorf("failed to parse balance value: %s", balanceStr)
			}
			account.AddBalance(balance)
			currentLine++
		}
		if err := persistBlock4Test(r, lg, currentHeight, parentBlockHeader); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error scanning file: %v", err)
	}
	return nil
}

func persistBlock4Test(r *repo.Repo, lg *ledger.Ledger, currentHeight uint64, parentBlockHeader *types.BlockHeader) error {
	lg.StateLedger.Finalise()
	stateRoot, err := lg.StateLedger.Commit()
	if err != nil {
		return err
	}
	g := r.GenesisConfig
	block := &types.Block{
		Header: &types.BlockHeader{
			Number:         currentHeight + 1,
			StateRoot:      stateRoot,
			TxRoot:         &types.Hash{},
			ReceiptRoot:    &types.Hash{},
			ParentHash:     parentBlockHeader.Hash(),
			Timestamp:      g.Timestamp,
			Epoch:          g.EpochInfo.Epoch,
			Bloom:          new(types.Bloom),
			GasUsed:        0,
			ProposerNodeID: 0,
			TotalGasFee:    big.NewInt(0),
			GasFeeReward:   big.NewInt(0),
		},
		Transactions: []*types.Transaction{},
	}
	blockData := &ledger.BlockData{
		Block: block,
	}

	lg.PersistBlockData(blockData)
	return nil
}

func countLines(filePath string) (int, error) {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		return countLinesWC(filePath)
	}
	return countLinesBuffered(filePath)
}

func countLinesBuffered(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = file.Close()
	}()

	reader := bufio.NewReader(file)
	lineCount := 0

	for {
		_, isPrefix, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}
		if !isPrefix {
			lineCount++
		}
	}

	return lineCount, nil
}

func countLinesWC(filePath string) (int, error) {
	cmd := exec.Command("wc", "-l", filePath)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	var lineCount int
	_, err = fmt.Sscanf(string(output), "%d", &lineCount)
	if err != nil {
		return 0, err
	}
	return lineCount, nil
}

func getBlock(ctx *cli.Context) error {
	r, err := common.PrepareRepo(ctx)
	if err != nil {
		return err
	}

	chainLedger, err := ledger.NewChainLedger(r, "")
	if err != nil {
		return fmt.Errorf("init chain ledger failed: %w", err)
	}

	var blockHeader *types.BlockHeader
	if ctx.IsSet("number") {
		blockHeader, err = chainLedger.GetBlockHeader(ledgerGetBlockArgs.Number)
		if err != nil {
			return err
		}
	} else {
		var blockNumber uint64
		if ctx.IsSet("hash") {
			blockNumber, err = chainLedger.GetBlockNumberByHash(types.NewHashByStr(ledgerGetBlockArgs.Hash))
			if err != nil {
				return err
			}
		} else {
			blockNumber = chainLedger.GetChainMeta().Height
		}
		blockHeader, err = chainLedger.GetBlockHeader(blockNumber)
		if err != nil {
			return err
		}
	}

	bloom, _ := blockHeader.Bloom.ETHBloom().MarshalText()
	blockInfo := map[string]any{
		"number":         blockHeader.Number,
		"hash":           blockHeader.Hash().String(),
		"state_root":     blockHeader.StateRoot.String(),
		"tx_root":        blockHeader.TxRoot.String(),
		"receipt_root":   blockHeader.ReceiptRoot.String(),
		"parent_hash":    blockHeader.ParentHash.String(),
		"timestamp":      blockHeader.Timestamp,
		"epoch":          blockHeader.Epoch,
		"bloom":          string(bloom),
		"gas_used":       blockHeader.GasUsed,
		"gas_price":      blockHeader.GasPrice,
		"proposer_node":  blockHeader.ProposerNodeID,
		"total_gas_fee":  blockHeader.TotalGasFee.String(),
		"gas_fee_reward": blockHeader.GasFeeReward.String(),
	}

	if ledgerGetBlockArgs.Full {
		txs, err := chainLedger.GetBlockTxList(blockHeader.Number)
		if err != nil {
			return err
		}

		blockInfo["transactions"] = lo.Map(txs, func(item *types.Transaction, index int) string {
			return item.GetHash().String()
		})
		blockInfo["tx_count"] = len(txs)
	} else {
		txs, err := chainLedger.GetBlockTxHashList(blockHeader.Number)
		if err != nil {
			return err
		}
		blockInfo["transactions"] = txs
		blockInfo["tx_count"] = len(txs)
	}

	return common.Pretty(blockInfo)
}

func getTx(ctx *cli.Context) error {
	rep, err := common.PrepareRepo(ctx)
	if err != nil {
		return err
	}

	chainLedger, err := ledger.NewChainLedger(rep, "")
	if err != nil {
		return fmt.Errorf("init chain ledger failed: %w", err)
	}

	tx, err := chainLedger.GetTransaction(types.NewHashByStr(ledgerGetTxArgs.Hash))
	if err != nil {
		return err
	}

	to := "0x0000000000000000000000000000000000000000"
	if tx.GetTo() != nil {
		to = tx.GetTo().String()
	}
	from := "0x0000000000000000000000000000000000000000"
	if tx.GetFrom() != nil {
		to = tx.GetFrom().String()
	}
	v, r, s := tx.GetRawSignature()
	txInfo := map[string]any{
		"type":      tx.GetType(),
		"from":      from,
		"gas":       tx.GetGas(),
		"gas_price": tx.GetGasPrice(),
		"hash":      tx.GetHash().String(),
		"input":     hexutil.Bytes(tx.GetPayload()),
		"nonce":     tx.GetNonce(),
		"to":        to,
		"value":     (*hexutil.Big)(tx.GetValue()),
		"v":         (*hexutil.Big)(v),
		"r":         (*hexutil.Big)(r),
		"s":         (*hexutil.Big)(s),
	}
	return common.Pretty(txInfo)
}

func getLatestChainMeta(ctx *cli.Context) error {
	r, err := common.PrepareRepo(ctx)
	if err != nil {
		return err
	}

	chainLedger, err := ledger.NewChainLedger(r, "")
	if err != nil {
		return fmt.Errorf("init chain ledger failed: %w", err)
	}

	meta := chainLedger.GetChainMeta()
	return common.Pretty(meta)
}

func rollback(ctx *cli.Context) error {
	r, err := common.PrepareRepo(ctx)
	if err != nil {
		return err
	}
	logger := loggers.Logger(loggers.App)

	// back up storage dir
	backupStorageDir := path.Join(r.RepoRoot, "storage-backup")
	if !ledgerSimpleRollbackArgs.DisableBackup {
		if fileutil.Exist(backupStorageDir) {
			if !ledgerSimpleRollbackArgs.Force {
				return errors.Errorf("backup dir %s already exists\n", backupStorageDir)
			}
			if err := os.RemoveAll(backupStorageDir); err != nil {
				return err
			}
		}
		if err := os.MkdirAll(backupStorageDir, os.ModePerm); err != nil {
			return errors.Errorf("mkdir storage-backup dir error: %v", err.Error())
		}
		if err := copyDir(repo.GetStoragePath(r.RepoRoot), backupStorageDir); err != nil {
			return errors.Errorf("backup original storage dir error: %v", err.Error())
		}
		logger.Infof("backup original storage success")
	}

	logger.Infof("This operation will REMOVE original snapshot/consensus/epoch data, and MODIFY original ledger/blockchain/blockfile, you'd better back up those data if you need. Continue? y/n\n")
	if err := common.WaitUserConfirm(); err != nil {
		return err
	}

	// remove original snapshot
	if err := os.RemoveAll(repo.GetStoragePath(r.RepoRoot, storagemgr.Snapshot)); err != nil {
		return err
	}
	if err := os.RemoveAll(repo.GetStoragePath(r.RepoRoot, storagemgr.Consensus)); err != nil {
		return err
	}
	logger.Infof("remove original snapshot successfully")

	originChainLedger, err := ledger.NewChainLedger(r, "")
	if err != nil {
		return fmt.Errorf("init chain ledger failed: %w", err)
	}
	originStateLedger, err := ledger.NewStateLedger(r, "")
	if err != nil {
		return fmt.Errorf("init state ledger failed: %w", err)
	}

	// check if target height is legal
	targetBlockNumber := ledgerSimpleRollbackArgs.TargetBlockNumber
	if targetBlockNumber < 0 {
		return errors.New("target-block-number must be greater than or equal to 0")
	}
	chainMeta := originChainLedger.GetChainMeta()
	if targetBlockNumber >= chainMeta.Height {
		return errors.Errorf("target-block-number %d must be less than the current latest block height %d\n", targetBlockNumber, chainMeta.Height)
	}
	if r.Config.Ledger.EnablePrune {
		minHeight, maxHeight := originStateLedger.GetHistoryRange()
		if targetBlockNumber < minHeight || targetBlockNumber > maxHeight {
			return errors.Errorf("this is a prune node, target-block-number %d must be within valid range, which is from %d to %d\n", targetBlockNumber, minHeight, maxHeight)
		}
	}

	targetBlockHeader, err := originChainLedger.GetBlockHeader(targetBlockNumber)
	if err != nil {
		return fmt.Errorf("get target block failed: %w", err)
	}

	logger.Infof("current chain meta info height: %d, hash: %s, will rollback to the target height %d, hash: %s, confirm? y/n\n", chainMeta.Height, chainMeta.BlockHash, targetBlockNumber, targetBlockHeader.Hash())
	if err := common.WaitUserConfirm(); err != nil {
		return err
	}

	// If we need to back up rollback info, then write rollback logs into rollback dir in KV format.
	if !ledgerSimpleRollbackArgs.DisableBackup {
		rollbackDir := path.Join(backupStorageDir, "rollback-log")
		rollbackStorage, err := storagemgr.Open(rollbackDir)
		if err != nil {
			return fmt.Errorf("create rollback log dir failed: %w", err)
		}
		batch := rollbackStorage.NewBatch()

		// write stale blocks
		for i := chainMeta.Height; i > targetBlockNumber; i-- {
			block, err := originChainLedger.GetBlock(i)
			if err != nil {
				return fmt.Errorf("get rollback block failed: %w", err)
			}

			logger.Infof("[rollback] stale block%v=%v\n", i, block)

			blockBlob, err := block.Marshal()
			if err != nil {
				return fmt.Errorf("marshal rollback block failed: %w", err)
			}
			batch.Put(utils.CompositeKey(utils.RollbackBlockKey, i), blockBlob)
			if batch.Size() > maxBatchSize {
				batch.Commit()
				batch.Reset()
				logger.Infof("[rollback] write batch periodically")
			}
		}

		// write stale ledger state deltas
		if r.Config.Ledger.EnablePrune {
			_, maxHeight := originStateLedger.GetHistoryRange()
			for i := maxHeight; i > targetBlockNumber; i-- {
				stateDelta := originStateLedger.GetStateDelta(i)
				batch.Put(utils.CompositeKey(utils.RollbackStateKey, i), stateDelta.Encode())
				if batch.Size() > maxBatchSize {
					batch.Commit()
					batch.Reset()
					logger.Infof("[rollback] write batch periodically")
				}
			}

		}
		batch.Commit()
	}

	// wait for generating snapshot of target block
	errC := make(chan error)
	go originStateLedger.GenerateSnapshot(targetBlockHeader, errC)
	err = <-errC
	if err != nil {
		return fmt.Errorf("generate snapshot failed: %w", err)
	}

	// rollback directly on the original ledger
	if err := originChainLedger.RollbackBlockChain(targetBlockNumber); err != nil {
		return errors.Errorf("rollback chain ledger error: %v", err.Error())
	}

	if err := originStateLedger.RollbackState(targetBlockNumber, targetBlockHeader.StateRoot); err != nil {
		return fmt.Errorf("rollback state ledger failed: %w", err)
	}

	logger.Info("rollback chain ledger and state ledger success")

	return nil
}

func decodePeers(peersStr []string) (*consensus.QuorumValidators, error) {
	var peers []*consensus.QuorumValidator
	if len(peersStr) == 0 {
		return nil, errors.New("peers cannot be empty")
	}
	for _, p := range peersStr {
		// spilt id and peerId by :
		data := strings.Split(p, ":")
		if len(data) != 2 {
			return nil, fmt.Errorf("invalid peer: %s, should be id:peerId", p)
		}
		id, err := strconv.ParseUint(data[0], 10, 64)
		if err != nil {
			return nil, err
		}
		peers = append(peers, &consensus.QuorumValidator{
			Id:     id,
			PeerId: data[1],
		})
	}

	return &consensus.QuorumValidators{Validators: peers}, nil
}

func generateTrie(ctx *cli.Context) error {
	logger := loggers.Logger(loggers.App)
	logger.Infof("start generating trie at height: %v\n", ledgerGenerateTrieArgs.TargetBlockNumber)

	peers, err := decodePeers(ledgerGenerateTrieArgs.remotePeers.Value())
	if err != nil {
		return fmt.Errorf("decode peers failed: %w", err)
	}

	r, err := common.PrepareRepo(ctx)
	if err != nil {
		return err
	}

	chainLedger, err := ledger.NewChainLedger(r, "")
	if err != nil {
		return fmt.Errorf("init chain ledger failed: %w", err)
	}
	blockHeader, err := chainLedger.GetBlockHeader(ledgerGenerateTrieArgs.TargetBlockNumber)
	if err != nil {
		return fmt.Errorf("get block failed: %w", err)
	}

	targetStateStoragePath := path.Join(r.RepoRoot, ledgerGenerateTrieArgs.TargetStoragePath)
	targetStateStorage, err := storagemgr.Open(targetStateStoragePath)
	if err != nil {
		return fmt.Errorf("create targetStateStorage: %w", err)
	}

	originStateLedger, err := ledger.NewStateLedger(r, "")
	if err != nil {
		return fmt.Errorf("init state ledger failed: %w", err)
	}
	if r.Config.Ledger.EnablePrune {
		minHeight, maxHeight := originStateLedger.GetHistoryRange()
		if ledgerGenerateTrieArgs.TargetBlockNumber < minHeight || ledgerGenerateTrieArgs.TargetBlockNumber > maxHeight {
			return errors.Errorf("This is a prune node, target-block-number %d must be within valid range, which is from %d to %d\n", ledgerGenerateTrieArgs.TargetBlockNumber, minHeight, maxHeight)
		}
	}

	epochManagerContract := framework.EpochManagerBuildConfig.Build(syscommon.NewViewVMContext(originStateLedger))
	epochInfo, err := epochManagerContract.HistoryEpoch(blockHeader.Epoch)
	if err != nil {
		return fmt.Errorf("get epoch info failed: %w", err)
	}

	errC := make(chan error)
	go originStateLedger.IterateTrie(&ledger.SnapshotMeta{
		BlockHeader: blockHeader,
		EpochInfo:   epochInfo.ToTypesEpoch(),
		Nodes:       peers,
	}, targetStateStorage, errC)
	err = <-errC

	logger.Infof("finish generating trie at height: %v\n", ledgerGenerateTrieArgs.TargetBlockNumber)

	return err
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
			if err := os.MkdirAll(destPath, os.ModePerm); err != nil {
				return errors.Errorf("mkdir %s dir error: %v", destPath, err.Error())
			}
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
	defer func() {
		_ = srcFile.Close()
	}()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() {
		_ = destFile.Close()
	}()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}
