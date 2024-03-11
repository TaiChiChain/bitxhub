package txpool

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
)

// devNull mimic the behavior of the Unix /dev/null.
// It's a WriteCloser that effectively ignores anything written to it, just like a data black hole.
type devNull struct{}

const (
	TxRecordPrefixLength = 8
	TxRecordsBatchSize   = 1000
	TxRecordsFile        = "tx_records.pb"
	DecodeTxRecordsFile  = "decode_tx_records.json"
)

func (*devNull) Write(p []byte) (n int, err error) { return len(p), nil }

func (*devNull) Close() error { return nil }

type txRecords[T any, Constraint types.TXConstraint[T]] struct {
	logger   logrus.FieldLogger
	filePath string
	writer   io.WriteCloser
}

func newTxRecords[T any, Constraint types.TXConstraint[T]](filePath string, logger logrus.FieldLogger) *txRecords[T, Constraint] {
	return &txRecords[T, Constraint]{
		filePath: filePath,
		logger:   logger,
	}
}

func (r *txRecords[T, Constraint]) load(input *os.File, taskDoneCh chan struct{}) chan []*T {
	batchCh := make(chan []*T, 1024)

	r.writer = new(devNull)
	defer func() { r.writer = nil }()

	buf := bufio.NewReader(input)
	var txNums uint64
	batch := make([]*T, 0, TxRecordsBatchSize)

	go func(txNums uint64) {
		for {
			lengthBytes, err := buf.Peek(TxRecordPrefixLength)
			if err != nil {
				if errors.Is(err, io.EOF) {
					if len(batch) > 0 {
						batchCh <- batch
					}

				} else {
					r.logger.Errorf("TxRecords load failed to peek transaction size: %v", err)
				}
				r.logger.Infof("TxRecords loaded %d transactions from %s", txNums, r.filePath)
				taskDoneCh <- struct{}{}
				return
			}

			length := binary.LittleEndian.Uint64(lengthBytes)
			_, _ = buf.Discard(TxRecordPrefixLength)

			data := make([]byte, length)
			if _, err := io.ReadFull(buf, data); err != nil {
				r.logger.Errorf("TxRecords load failed to error reading transaction data: %v", err)
				continue
			}

			tx := new(T)
			if err = Constraint(tx).RbftUnmarshal(data); err != nil {
				r.logger.Errorf("TxRecords load failed to unmarshal transaction: %v", err)
				continue
			}

			batch = append(batch, tx)
			if len(batch) >= TxRecordsBatchSize {
				getBatch := make([]*T, len(batch))
				copy(getBatch, batch)
				batchCh <- getBatch
				// Get a batch from the pool
				batch = make([]*T, 0, TxRecordsBatchSize)
			}
			txNums++
		}
	}(txNums)

	return batchCh
}

func (r *txRecords[T, Constraint]) insert(tx *T) error {
	if r.writer == nil {
		return errors.New("no active txRecords")
	}
	b, err := Constraint(tx).RbftMarshal()
	if err != nil {
		return err
	}
	length := uint64(len(b))
	var lengthBytes [TxRecordPrefixLength]byte
	binary.LittleEndian.PutUint64(lengthBytes[:], length)
	if _, err := r.writer.Write(lengthBytes[:]); err != nil {
		return err
	}
	if _, err := r.writer.Write(b); err != nil {
		return err
	}
	return nil
}

func (r *txRecords[T, Constraint]) rotate(all map[string]*txSortedMap[T, Constraint]) error {
	// Close the current records (if any is open)
	if r.writer != nil {
		if err := r.writer.Close(); err != nil {
			return err
		}
		r.writer = nil
	}
	dir := filepath.Dir(r.filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	replacement, err := os.OpenFile(r.filePath+".new", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	var batch []byte
	batchCount := 0
	record := 0
	for _, txMap := range all {
		for _, internalTx := range txMap.items {
			if !internalTx.local {
				continue
			}
			tx := internalTx.rawTx
			b, err := Constraint(tx).RbftMarshal()
			if err != nil {
				r.logger.Errorf("TxRecords rotate failed to marshal transaction: %v", internalTx.getHash())
				continue
			}
			length := uint64(len(b))
			var lengthBytes [TxRecordPrefixLength]byte
			binary.LittleEndian.PutUint64(lengthBytes[:], length)
			batch = append(batch, lengthBytes[:]...)
			batch = append(batch, b...)
			batchCount++
			record++
			if batchCount >= TxRecordsBatchSize || record == len(all) {
				if _, err := replacement.Write(batch); err != nil {
					r.logger.Errorf("TxRecords rotate failed to write batch to file: %v", err)
				}
				batch = nil
				batchCount = 0
			}
		}
	}
	if len(batch) > 0 {
		if _, err := replacement.Write(batch); err != nil {
			r.logger.Errorf("TxRecords rotate failed to write remaining batch to file: %v", err)
		}
	}
	replacement.Close()

	if err = os.Rename(r.filePath+".new", r.filePath); err != nil {
		return err
	}
	sink, err := os.OpenFile(r.filePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	r.writer = sink
	r.logger.Infof("TxRecords rotated and regenerated txRecords, wrote transactions: %d, accounts: %d", record, len(all))

	return nil
}

func GetAllTxRecords(filePath string) ([][]byte, error) {
	input, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer input.Close()
	buf := bufio.NewReader(input)
	var res [][]byte
	for {
		lengthBytes, err := buf.Peek(TxRecordPrefixLength)
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		length := binary.LittleEndian.Uint64(lengthBytes)
		_, _ = buf.Discard(TxRecordPrefixLength)
		data := make([]byte, length)
		if _, err := io.ReadFull(buf, data); err != nil {
			continue
		}
		res = append(res, data)
	}
	return res, nil
}

func (r *txRecords[T, Constraint]) close() error {
	var err error

	if r.writer != nil {
		err = r.writer.Close()
		r.writer = nil
	}
	return err
}
