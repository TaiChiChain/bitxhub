package main

import (
	"fmt"
	"io"
	"os"

	"github.com/axiomesh/axiom-ledger/pkg/bind"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"
)

var (
	// Flags needed by abigen
	abiFlag = &cli.StringFlag{
		Name:  "abi",
		Usage: "Path to the system contract ABI json to bind, - for STDIN",
	}
	pkgFlag = &cli.StringFlag{
		Name:  "pkg",
		Usage: "Package name to generate the binding into",
	}
	outFlag = &cli.StringFlag{
		Name:  "out",
		Usage: "Output file for the generated binding (default = stdout)",
	}
)

var app = cli.NewApp()

func init() {
	app.Name = "axmgen"
	app.Flags = []cli.Flag{
		abiFlag,
		pkgFlag,
		outFlag,
	}
	app.Action = abigen
}

func abigen(c *cli.Context) error {
	if c.String(pkgFlag.Name) == "" {
		utils.Fatalf("No destination package specified (--pkg)")
	}
	// If the entire solidity code was specified, build and bind based on that
	var (
		abis  []string
		types []string
		sigs  []map[string]string
		libs  = make(map[string]string)
	)
	if c.String(abiFlag.Name) != "" {
		// Load up the ABI, optional bytecode and type name from the parameters
		var (
			abi []byte
			err error
		)
		input := c.String(abiFlag.Name)
		if input == "-" {
			abi, err = io.ReadAll(os.Stdin)
		} else {
			abi, err = os.ReadFile(input)
		}
		if err != nil {
			utils.Fatalf("Failed to read input ABI: %v", err)
		}
		abis = append(abis, string(abi))

		kind := c.String(pkgFlag.Name)
		types = append(types, kind)
	}
	// Generate the contract binding
	code, err := bind.Bind(types, abis, sigs, c.String(pkgFlag.Name), libs)
	if err != nil {
		utils.Fatalf("Failed to generate ABI binding: %v", err)
	}
	// Either flush it out to a file or display on the standard output
	if !c.IsSet(outFlag.Name) {
		fmt.Printf("%s\n", code)
		return nil
	}
	if err := os.WriteFile(c.String(outFlag.Name), []byte(code), 0600); err != nil {
		utils.Fatalf("Failed to write ABI binding: %v", err)
	}
	return nil
}

func main() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlInfo, log.StreamHandler(os.Stderr, log.TerminalFormat(true))))

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
