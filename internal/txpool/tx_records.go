package txpool

import (
	"bufio"
	"encoding/binary"
	"errors"
	"path/filepath"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/sirupsen/logrus"

	"io"
	"io/fs"
	"os"
)

// devNull mimic the behavior of the Unix /dev/null.
// It's a WriteCloser that effectively ignores anything written to it, just like a data black hole.
type devNull struct{}

const TxRecordPrefixLength = 8

const TxRecordsBatchSize = 1000

func (*devNull) Write(p []byte) (n int, err error) { return len(p), nil }
func (*devNull) Close() error                      { return nil }

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

func (r *txRecords[T, Constraint]) load(add func(txs []*T, local bool) []error) error {
	input, err := os.Open(r.filePath)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		r.logger.Errorf("Failed to open tx records file: %v", err)
		return err
	}
	defer input.Close()

	r.writer = new(devNull)
	defer func() { r.writer = nil }()

	buf := bufio.NewReader(input)
	var txNums uint64
	var batch []*T

	for {
		lengthBytes, err := buf.Peek(TxRecordPrefixLength)
		if err != nil {
			if len(batch) > 0 {
				_ = add(batch, true)
			}
			break
		}
		length := binary.LittleEndian.Uint64(lengthBytes)
		_, _ = buf.Discard(TxRecordPrefixLength)

		data := make([]byte, length)
		if _, err := io.ReadFull(buf, data); err != nil {
			r.logger.Errorf("TxRecords load failed to error reading transaction data: %v", err)
			return err
		}

		tx := new(T)
		if err = Constraint(tx).RbftUnmarshal(data); err != nil {
			r.logger.Errorf("TxRecords load failed to unmarshal transaction: %v", err)
			continue
		}

		batch = append(batch, tx)
		if len(batch) >= TxRecordsBatchSize {
			_ = add(batch, true)
			batch = nil
		}
		txNums++
	}

	r.logger.Infof("TxRecords loaded %d transactions from %s", txNums, r.filePath)
	return nil
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
		if err := os.MkdirAll(dir, 0755); err != nil {
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
