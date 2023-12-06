package token

import (
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

type IToken interface { // Name Returns the name of the token
	Name() string

	// Symbol Returns the symbol of the token
	Symbol() string

	// Decimals Number of decimal this token has
	Decimals() uint8

	// TotalSupply Returns the Value of tokens in existence
	TotalSupply() *big.Int

	// BalanceOf Returns the balance of the account
	BalanceOf(account ethcommon.Address) *big.Int

	// Mint tokens for account
	Mint(value *big.Int) error

	// Burn tokens for account, return error if account have not enough balance
	Burn(value *big.Int) error

	// Allowance Returns the Value which `spender` is still allowed to withdraw from `owner`
	// This is zero by default, This Value changes when {approve} or {transferFrom} are called.
	Allowance(owner, spender ethcommon.Address) *big.Int

	// Approve Sets `Value` as the allowance of `spender` over the caller's tokens
	Approve(spender ethcommon.Address, value *big.Int) error

	// Transfer transfers `Value` tokens from token contract's account to `recipient`
	Transfer(recipient ethcommon.Address, value *big.Int) error

	// TransferFrom moves `Value` tokens from `sender` to `recipient` using the allowance mechanism,
	// 'Value' is then deducted from the caller's allowance.
	TransferFrom(sender, recipient ethcommon.Address, value *big.Int) error
}
