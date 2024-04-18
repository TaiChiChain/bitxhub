package saccount

import (
	"fmt"
	"math/big"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/solidity/ipaymaster_client"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/interfaces"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

const (
	tokenOwnerKey  = "owner"
	tokenOracleKey = "oracle"
)

var (
	balanceOfSig          = crypto.Keccak256([]byte("balanceOf(address)"))[:4]
	transferFromSig       = crypto.Keccak256([]byte("transferFrom(address,address,uint256)"))[:4]
	getTokenValueOfAxcSig = crypto.Keccak256([]byte("getTokenValueOfAXC(address,uint256)"))[:4]
	decimalsSig           = crypto.Keccak256([]byte("decimals()"))[:4]

	contextArg = abi.Arguments{
		{Name: "account", Type: common.AddressType},
		{Name: "token", Type: common.AddressType},
		{Name: "gasPriceUserOp", Type: common.BigIntType},
		{Name: "maxTokenCost", Type: common.BigIntType},
		{Name: "maxCost", Type: common.BigIntType},
	}
)

var TokenPaymasterBuildConfig = &common.SystemContractBuildConfig[*TokenPaymaster]{
	Name:    "saccount_token_paymaster",
	Address: common.TokenPaymasterContractAddr,
	AbiStr:  ipaymaster_client.BindingContractMetaData.ABI,
	Constructor: func(systemContractBase common.SystemContractBase) *TokenPaymaster {
		return &TokenPaymaster{
			SystemContractBase: systemContractBase,
		}
	},
}

var _ interfaces.IPaymaster = (*TokenPaymaster)(nil)

type TokenPaymaster struct {
	common.SystemContractBase

	owner *common.VMSlot[ethcommon.Address]

	// token -> tokenPriceOracle
	oracles *common.VMMap[ethcommon.Address, ethcommon.Address]
}

func (tp *TokenPaymaster) GenesisInit(genesis *repo.GenesisConfig) error {
	if !ethcommon.IsHexAddress(genesis.SmartAccountAdmin) {
		return errors.New("invalid admin address")
	}

	if err := tp.owner.Put(ethcommon.HexToAddress(genesis.SmartAccountAdmin)); err != nil {
		return err
	}
	return nil
}

func (tp *TokenPaymaster) SetContext(ctx *common.VMContext) {
	tp.SystemContractBase.SetContext(ctx)

	tp.owner = common.NewVMSlot[ethcommon.Address](tp.StateAccount, tokenOwnerKey)
	tp.oracles = common.NewVMMap[ethcommon.Address, ethcommon.Address](tp.StateAccount, tokenOracleKey, func(key ethcommon.Address) string {
		return key.String()
	})
}

func (tp *TokenPaymaster) AddToken(token, tokenPriceOracle ethcommon.Address) error {
	owner, err := tp.owner.MustGet()
	if err != nil {
		return err
	}

	if tp.Ctx.From != owner {
		return errors.New("only owner can add token")
	}

	if tp.oracles.Has(token) {
		return errors.New("token already set")
	}

	if err := tp.oracles.Put(token, tokenPriceOracle); err != nil {
		return err
	}
	return nil
}

func (tp *TokenPaymaster) GetToken(token ethcommon.Address) ethcommon.Address {
	res, _ := tp.oracles.MustGet(token)
	return res
}

// PostOp implements interfaces.IPaymaster.
func (tp *TokenPaymaster) PostOp(mode interfaces.PostOpMode, context []byte, actualGasCost *big.Int) error {
	if tp.Ctx.From != ethcommon.HexToAddress(common.EntryPointContractAddr) {
		return errors.New("only entrypoint can call post op")
	}

	account, tokenAddr, _, maxTokenCost, maxCost, err := decodeContext(context)
	if err != nil {
		return fmt.Errorf("token paymaster: decode context failed: %s", err.Error())
	}

	if maxCost.Sign() == 0 {
		// max cost is 0, no need to transfer token
		return nil
	}

	// use same conversion rate as used for validation.
	actualTokenCost := new(big.Int).Div(new(big.Int).Mul(actualGasCost, maxTokenCost), maxCost)
	// transfer token as gas fee from account to paymaster
	return tp.transferFrom(account, tokenAddr, actualTokenCost)
}

// ValidatePaymasterUserOp implements interfaces.IPaymaster.
func (tp *TokenPaymaster) ValidatePaymasterUserOp(userOp interfaces.UserOperation, userOpHash [32]byte, maxCost *big.Int) (context []byte, validationData *big.Int, err error) {
	if tp.Ctx.From != ethcommon.HexToAddress(common.EntryPointContractAddr) {
		return nil, nil, errors.New("only entrypoint can call validate paymaster user op")
	}

	context, validation, err := tp.validatePaymasterUserOp(userOp, userOpHash, maxCost)
	// nolint
	return context, big.NewInt(int64(validation.SigValidation)), err
}

// nolint
func (tp *TokenPaymaster) validatePaymasterUserOp(userOp interfaces.UserOperation, userOpHash [32]byte, maxCost *big.Int) (context []byte, validation *interfaces.Validation, err error) {
	validation = &interfaces.Validation{SigValidation: interfaces.SigValidationFailed}
	paymasterAndData := userOp.PaymasterAndData
	if len(paymasterAndData) < 20+20 {
		return nil, validation, errors.New("token paymaster: paymasterAndData must specify token")
	}

	tokenAddr := ethcommon.BytesToAddress(paymasterAndData[20:40])
	maxTokenCost, err := tp.getTokenValueOfAxc(tokenAddr, maxCost)
	if err != nil {
		return nil, validation, fmt.Errorf(fmt.Sprintf("token paymaster: get token %s value failed: %s", tokenAddr.String(), err.Error()))
	}
	balance, err := tp.getTokenBalance(userOp.Sender, tokenAddr)
	if err != nil {
		return nil, validation, fmt.Errorf("token paymaster: get token balance failed: %s", err.Error())
	}
	if balance.Cmp(maxTokenCost) < 0 {
		return nil, validation, fmt.Errorf("token paymaster: not enough token balance, balance: %s, maxTokenCost: %s", balance.String(), maxTokenCost.String())
	}
	context, err = encodeContext(userOp.Sender, tokenAddr, interfaces.GetGasPrice(&userOp), maxTokenCost, maxCost)
	if err != nil {
		return nil, validation, fmt.Errorf("token paymaster: encode context failed: %s", err.Error())
	}
	tp.Logger.Infof("token paymaster: validate paymaster user op success, sender: %s, token addr: %s, maxTokenCost: %s, maxCost: %s, balance: %s", userOp.Sender.String(), tokenAddr.String(), maxTokenCost.String(), maxCost.String(), balance.String())
	validation.SigValidation = interfaces.SigValidationSucceeded
	return context, validation, err
}

// getTokenValueOfAxc translate the give axc value to token value
func (tp *TokenPaymaster) getTokenValueOfAxc(token ethcommon.Address, value *big.Int) (*big.Int, error) {
	exist, oracleAddr, err := tp.oracles.Get(token)
	if err != nil {
		return nil, err
	}
	if exist {
		callData := append(getTokenValueOfAxcSig, ethcommon.LeftPadBytes(token.Bytes(), 32)...)
		callData = append(callData, ethcommon.LeftPadBytes(value.Bytes(), 32)...)
		// return value of token is 10 decimals
		tokenValueRes, _, err := tp.CrossCallEVMContract(big.NewInt(MaxCallGasLimit), oracleAddr, callData)
		if err != nil {
			return nil, err
		}
		tokenValue := new(big.Int).SetBytes(tokenValueRes)

		// get token decimals
		callData = decimalsSig
		tokenDecimalsRes, _, err := tp.CrossCallEVMContract(big.NewInt(MaxCallGasLimit), token, callData)
		if err != nil {
			return nil, err
		}
		tokenDecimals := new(big.Int).SetBytes(tokenDecimalsRes)

		// tokenValue / 10^10 * 10^tokenDecimals = tokenValue * 10^(tokenDecimals - 10)
		if tokenDecimals.Cmp(big.NewInt(10)) >= 0 {
			return new(big.Int).Mul(tokenValue, new(big.Int).Exp(big.NewInt(10), new(big.Int).Sub(tokenDecimals, big.NewInt(10)), nil)), nil
		}
		// tokenValue * 10^(tokenDecimals - 10) = tokenValue / 10^(10 - tokenDecimals)
		return new(big.Int).Div(tokenValue, new(big.Int).Exp(big.NewInt(10), new(big.Int).Sub(big.NewInt(10), tokenDecimals), nil)), nil
	}

	return big.NewInt(0), fmt.Errorf("token paymaster: token %s not set oracle", token.String())
}

func (tp *TokenPaymaster) getTokenBalance(account, token ethcommon.Address) (*big.Int, error) {
	callData := append(balanceOfSig, ethcommon.LeftPadBytes(account.Bytes(), 32)...)
	result, _, err := tp.CrossCallEVMContract(big.NewInt(MaxCallGasLimit), token, callData)
	if err != nil {
		return nil, err
	}

	return new(big.Int).SetBytes(result), nil
}

func (tp *TokenPaymaster) transferFrom(account, token ethcommon.Address, tokenValue *big.Int) error {
	callData := append(transferFromSig, ethcommon.LeftPadBytes(account.Bytes(), 32)...)
	callData = append(callData, ethcommon.LeftPadBytes(tp.EthAddress.Bytes(), 32)...)
	callData = append(callData, ethcommon.LeftPadBytes(tokenValue.Bytes(), 32)...)
	_, _, err := tp.CrossCallEVMContract(big.NewInt(MaxCallGasLimit), token, callData)
	if err != nil {
		return fmt.Errorf("token paymaster: call transferFrom failed: %s", err.Error())
	}

	return nil
}

func encodeContext(account, token ethcommon.Address, gasPriceUserOp, maxTokenCost, maxCost *big.Int) ([]byte, error) {
	return contextArg.Pack(account, token, gasPriceUserOp, maxTokenCost, maxCost)
}

func decodeContext(context []byte) (account, token ethcommon.Address, gasPriceUserOp, maxTokenCost, maxCost *big.Int, err error) {
	args, err := contextArg.Unpack(context)
	if err != nil {
		return ethcommon.Address{}, ethcommon.Address{}, nil, nil, nil, err
	}

	return args[0].(ethcommon.Address), args[1].(ethcommon.Address), args[2].(*big.Int), args[3].(*big.Int), args[4].(*big.Int), nil
}
