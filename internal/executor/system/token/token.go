package token

import (
	"math/big"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
)

type Token interface {
	Init(lg ledger.StateLedger, config *Config) error
	// Name Returns the name of the token
	Name() string

	// Symbol Returns the symbol of the token
	Symbol() string

	// Decimals Number of decimal this token has
	Decimals() uint8

	// TotalSupply Returns the amount of tokens in existence
	TotalSupply() *big.Int

	// BalanceOf Returns the balance of the account
	BalanceOf(account *types.Address) *big.Int

	// Mint tokens for account
	Mint(account *types.Address, amount *big.Int) error

	// Burn tokens for account, return error if account have not enough balance
	Burn(account *types.Address, amount *big.Int) error

	// Allowance Returns the amount which `spender` is still allowed to withdraw from `owner`
	// This is zero by default, This value changes when {approve} or {transferFrom} are called.
	Allowance(owner, spender *types.Address) *big.Int

	// IncreaseAllowance Atomically increases the allowance granted to `spender` by the caller.
	IncreaseAllowance(owner, spender *types.Address, amount *big.Int) error

	// DecreaseAllowance Atomically decreases the allowance granted to `spender` by the caller.
	DecreaseAllowance(owner, spender *types.Address, amount *big.Int) error

	// Approve Sets `amount` as the allowance of `spender` over the caller's tokens
	Approve(owner, spender *types.Address, amount *big.Int) error

	// Transfer transfers `amount` tokens from token contract's account to `recipient`
	Transfer(recipient *types.Address, amount *big.Int) error

	// TransferFrom moves `amount` tokens from `sender` to `recipient` using the allowance mechanism,
	// 'amount' is then deducted from the caller's allowance.
	TransferFrom(sender, recipient *types.Address, amount *big.Int) error
}
