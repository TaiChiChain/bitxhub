package token

import "math/big"

type Config struct {
	Name        string
	Symbol      string
	Decimals    uint8
	TotalSupply *big.Int
}
