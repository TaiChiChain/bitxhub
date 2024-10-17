package data_syncer

import (
	"context"
	"testing"
	"time"

	"github.com/axiomesh/axiom-kit/types"
)

func TestNewClient(t *testing.T) {
	cli, err := newClient("127.0.0.1:9191", context.Background())
	if err != nil {
		t.Fatal(err)
	}

	newBlockCh := make(chan *types.Block)
	cli.listenNewBlock(newBlockCh)

	for {
		select {
		case newBlock := <-newBlockCh:
			t.Log(newBlock)
		case <-time.After(10 * time.Second):
			t.Fatal("timeout")
		}

	}

}
