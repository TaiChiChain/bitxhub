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
		return err
	}
	defer input.Close()

	r.writer = new(devNull)
	defer func() { r.writer = nil }()

	buf := bufio.NewReader(input)
	var txNums uint64
	for {
		lengthBytes, err := buf.Peek(8)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		length := binary.LittleEndian.Uint64(lengthBytes)
		_, _ = buf.Discard(8)

		data := make([]byte, length)
		if _, err := io.ReadFull(buf, data); err != nil {
			return err
		}

		tx := new(T)
		if err = Constraint(tx).RbftUnmarshal(data); err != nil {
			return err
		}

		if errs := add([]*T{tx}, true); len(errs) > 0 {
			return errs[0]
		}
		txNums++
	}
	r.logger.Infof("local tx persist: Loaded %d transactions from %s", txNums, r.filePath)
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
	var lengthBytes [8]byte
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
	// Close the current journal (if any is open)
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
	journaled := 0
	for _, txMap := range all {
		for _, internalTx := range txMap.items {
			tx := internalTx.rawTx

			b, err := Constraint(tx).RbftMarshal()
			if err != nil {
				return err
			}

			length := uint64(len(b))
			var lengthBytes [8]byte
			binary.LittleEndian.PutUint64(lengthBytes[:], length)
			if _, err := replacement.Write(lengthBytes[:]); err != nil {
				return err
			}

			if _, err := replacement.Write(b); err != nil {
				return err
			}
		}
		journaled += len(txMap.items)
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
	r.logger.Infof("local tx persist: Regenerated txRecords, wrote transactions: %d, accounts: %d", journaled, len(all))

	return nil
}

func (r *txRecords[T, Constraint]) close() error {
	var err error

	if r.writer != nil {
		err = r.writer.Close()
		r.writer = nil
	}
	return err
}
