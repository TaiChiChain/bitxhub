package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/axiomesh/axiom-kit/fileutil"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/txpool"
)

var txpoolCMD = &cli.Command{
	Name:  "txpool",
	Usage: "The txpool manage commands",
	Subcommands: []*cli.Command{
		{
			Name:   "txrecords",
			Usage:  "Get all txs in txrecords",
			Action: getAllTxRecords,
		},
	},
}

func getAllTxRecords(ctx *cli.Context) error {
	p, err := getRootPath(ctx)
	if err != nil {
		return err
	}
	p = filepath.Join(p, "storage/txpool")
	r, err := prepareRepo(ctx)
	if err != nil {
		return err
	}
	p = filepath.Join(p, r.ConsensusConfig.TxPool.TxRecordsFile)
	if !fileutil.Exist(p) {
		fmt.Println("axiom-ledger is not starting, please run axiom-ledger first, " + p)
		return nil
	}
	records, err := txpool.GetAllTxRecords(p)
	if err != nil {
		return err
	}

	var res []*types.Transaction
	for _, record := range records {
		tmp := &types.Transaction{}
		err := tmp.RbftUnmarshal(record)
		if err != nil {
			continue
		}
		res = append(res, tmp)
	}
	fmt.Println(json.Marshal(res))
	return nil
}
