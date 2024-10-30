package data_syncer

import (
	"sync"
	"testing"
	"time"

	"github.com/axiomesh/axiom-kit/log"
	consensus_types "github.com/axiomesh/axiom-ledger/internal/consensus/types"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	closeC := make(chan struct{})
	recvCh := make(chan *consensus_types.Attestation, 1)
	cli, err := newClient("127.0.0.1:9991", closeC, recvCh)
	if err != nil {
		t.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	logger := log.NewWithModule("consensus")
	ap, err := newAsyncPool(10, logger)
	assert.Nil(t, err)

	err = cli.start(wg, ap)
	assert.Nil(t, err)

	for {
		select {
		case newBlock := <-recvCh:
			t.Log(newBlock)
		case <-time.After(10 * time.Second):
			t.Fatal("timeout")
		}

	}

}
