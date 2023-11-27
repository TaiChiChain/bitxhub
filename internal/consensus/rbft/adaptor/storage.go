package adaptor

import (
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	storeDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "axiom_ledger",
		Subsystem: "consensus",
		Name:      "storage_action",
		Buckets:   prometheus.ExponentialBuckets(0.0001, 2, 10),
	}, []string{"type"})
)

func init() {
	prometheus.MustRegister(storeDuration)
}

// StoreState stores a key,value pair to the database with the given namespace
func (a *RBFTAdaptor) StoreState(key string, value []byte) error {
	start := time.Now()
	err := a.store.StoreState(key, value)
	storeDuration.WithLabelValues("StoreState").Observe(time.Since(start).Seconds())
	return err
}

// DelState removes a key,value pair from the database with the given namespace
func (a *RBFTAdaptor) DelState(key string) error {
	start := time.Now()
	err := a.store.DelState(key)
	storeDuration.WithLabelValues("DelState").Observe(time.Since(start).Seconds())
	return err
}

// ReadState retrieves a value to a key from the database with the given namespace
func (a *RBFTAdaptor) ReadState(key string) ([]byte, error) {
	start := time.Now()
	b, err := a.store.ReadState(key)
	storeDuration.WithLabelValues("ReadState").Observe(time.Since(start).Seconds())
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, errors.New("not found")
	}
	return b, nil
}

// ReadStateSet retrieves all key-value pairs where the key starts with prefix from the database with the given namespace
func (a *RBFTAdaptor) ReadStateSet(prefix string) (map[string][]byte, error) {
	start := time.Now()
	ret, err := a.store.ReadStateSet(prefix)
	storeDuration.WithLabelValues("ReadStateSet").Observe(time.Since(start).Seconds())
	if err != nil {
		return nil, err
	}
	if len(ret) == 0 {
		return nil, errors.New("not found")
	}
	return ret, nil
}
