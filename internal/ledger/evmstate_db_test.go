package ledger

import (
	"math/big"
	"testing"

	"github.com/holiman/uint256"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/ethereum/go-ethereum/common"
	etherTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

func TestEvmStateDBAdaptor_TestSnapshot(t *testing.T) {
	lg, _ := initLedger(t, "", "pebble")
	sl := lg.StateLedger.(*StateLedgerImpl)
	evmStateDB := EvmStateDBAdaptor{StateLedger: sl}
	// create an account
	account := types.NewAddress(LeftPadBytes([]byte{110}, 20))
	// change the balance
	input := uint256.NewInt(1000)
	evmStateDB.AddBalance(account.ETHAddress(), input)
	// keep the snapshot
	ss := evmStateDB.Snapshot()
	// change the balance twice
	evmStateDB.AddBalance(account.ETHAddress(), input)
	// revert the snapshot
	evmStateDB.RevertToSnapshot(ss)
	// check the balance
	balance := evmStateDB.GetBalance(account.ETHAddress())
	assert.Equal(t, input, balance)
}

func TestEvmStateDBAdaptor_PrepareEVMAccessList(t *testing.T) {
	lg, _ := initLedger(t, "", "pebble")
	sl := lg.StateLedger.(*StateLedgerImpl)
	evmStateDB := EvmStateDBAdaptor{StateLedger: sl}
	account := common.BytesToAddress(LeftPadBytes([]byte{111}, 20))
	at1 := etherTypes.AccessTuple{
		Address:     account,
		StorageKeys: []common.Hash{common.BytesToHash([]byte{1})},
	}
	evmStateDB.PrepareEVMAccessList(account, &account, []common.Address{account}, etherTypes.AccessList{at1})
	isIn := evmStateDB.AddressInAccessList(account)
	assert.Equal(t, true, isIn)
}

func TestEvmStateDBAdaptor_Selfdestruct6780(t *testing.T) {
	lg, _ := initLedger(t, "", "pebble")
	sl := lg.StateLedger.(*StateLedgerImpl)
	evmStateDB := EvmStateDBAdaptor{StateLedger: sl}
	address := common.BytesToAddress(LeftPadBytes([]byte{111}, 20))
	// case for non-existent account
	evmStateDB.Selfdestruct6780(address)

	// case for existing account
	account := sl.GetOrCreateAccount(types.NewAddress(address.Bytes()))
	account.AddBalance(big.NewInt(1000))
	evmStateDB.Selfdestruct6780(address)
	iAccount := sl.GetAccount(types.NewAddress(address.Bytes()))
	assert.Equal(t, iAccount.GetBalance(), big.NewInt(0))
	sa := iAccount.(*SimpleAccount)
	assert.Equal(t, sa.selfDestructed, true)
}
