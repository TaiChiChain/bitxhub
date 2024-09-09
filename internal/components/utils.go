package components

import (
	"fmt"
	"math/big"

	"github.com/cbergoon/merkletree"
	"github.com/ethereum/go-ethereum/core"
	"github.com/samber/lo"

	"github.com/axiomesh/axiom-kit/types"
)

func VerifyInsufficientBalance[T any, Constraint types.TXConstraint[T]](tx *T, getBalanceFn func(address string) *big.Int) error {
	// 1. account has enough balance to cover transaction fee(gaslimit * gasprice), gasprice is the chain's latest gas price
	txGasLimit := Constraint(tx).RbftGetGasLimit()
	txValue := Constraint(tx).RbftGetValue()
	txFrom := Constraint(tx).RbftGetFrom()
	txTo := Constraint(tx).RbftGetTo()
	txData := Constraint(tx).RbftGetData()
	txAccessList := Constraint(tx).RbftGetAccessList()

	mgval := new(big.Int).SetUint64(txGasLimit)
	mgval = mgval.Mul(mgval, Constraint(tx).RbftGetGasPrice())
	balanceCheck := mgval
	balanceRemaining := new(big.Int).Set(getBalanceFn(txFrom))
	if have, want := balanceRemaining, balanceCheck; have.Cmp(want) < 0 {
		return fmt.Errorf("%w: address %v have %v want %v", core.ErrInsufficientFunds, txFrom, have, want)
	}

	// sub gas fee temporarily
	balanceRemaining.Sub(balanceRemaining, mgval)

	gasRemaining := txGasLimit

	var isContractCreation bool
	if txTo == "" {
		isContractCreation = true
	}

	// 2.1 the purchased gas is enough to cover intrinsic usage
	// 2.2 there is no overflow when calculating intrinsic gas
	gas, err := core.IntrinsicGas(txData, txAccessList.ToEthAccessList(), isContractCreation, true, true, true)
	if err != nil {
		return err
	}
	if gasRemaining < gas {
		return fmt.Errorf("%w: have %d, want %d", core.ErrIntrinsicGas, gasRemaining, gas)
	}

	// 3. account has enough balance to cover asset transfer for **topmost** call
	if txValue.Sign() > 0 && balanceRemaining.Cmp(txValue) < 0 {
		return fmt.Errorf("%w: address %v", core.ErrInsufficientFundsForTransfer, txFrom)
	}
	return nil
}

func CalcTxsMerkleRoot(txs []*types.Transaction) (*types.Hash, error) {
	hash, err := calcMerkleRoot(lo.Map(txs, func(item *types.Transaction, index int) merkletree.Content {
		return item.GetHash()
	}))
	if err != nil {
		return nil, err
	}

	return hash, nil
}

func CalcReceiptMerkleRoot(receipts []*types.Receipt) (*types.Hash, error) {
	receiptHashes := make([]merkletree.Content, 0, len(receipts))
	for _, receipt := range receipts {
		receiptHashes = append(receiptHashes, receipt.Hash())
	}
	receiptRoot, err := calcMerkleRoot(receiptHashes)
	if err != nil {
		return nil, err
	}

	return receiptRoot, nil
}

func calcMerkleRoot(contents []merkletree.Content) (*types.Hash, error) {
	if len(contents) == 0 {
		// compatible with Ethereum
		return types.NewHashByStr("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"), nil
	}

	tree, err := merkletree.NewTree(contents)
	if err != nil {
		return nil, err
	}

	return types.NewHash(tree.MerkleRoot()), nil
}

func ReadableSize(size int) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d Bytes", size)
	}
}
