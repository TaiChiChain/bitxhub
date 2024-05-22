package saccount

import (
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/interfaces"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

func TestVerifyingPaymaster_ValidatePaymasterUserOp(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	entrypoint := EntryPointBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	factory := SmartAccountFactoryBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	verifyingPaymaster := VerifyingPaymasterBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(verifyingPaymaster)

	owner := ethcommon.HexToAddress(testNVM.Rep.GenesisConfig.SmartAccountAdmin)
	accountAddr, _ := factory.GetAddress(owner, big.NewInt(1))
	sa := SmartAccountBuildConfig.BuildWithAddress(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}), accountAddr).SetRemainingGas(big.NewInt(MaxCallGasLimit))

	testNVM.RunSingleTX(sa, entrypoint.EthAddress, func() error {
		err := sa.Initialize(owner, ethcommon.Address{})
		assert.Nil(t, err)
		return err
	})

	verifyingPaymasterKey := repo.MockOperatorKeys[0]

	userOp := interfaces.UserOperation{
		Sender:               accountAddr,
		Nonce:                big.NewInt(0),
		InitCode:             nil,
		CallData:             nil,
		CallGasLimit:         big.NewInt(10000000),
		VerificationGasLimit: big.NewInt(10000000),
		PreVerificationGas:   big.NewInt(10000000),
		MaxFeePerGas:         big.NewInt(1000000000000),
		MaxPriorityFeePerGas: big.NewInt(1000000000000),
	}
	arg := abi.Arguments{
		{Name: "validUntil", Type: common.UInt48Type},
		{Name: "validAfter", Type: common.UInt48Type},
	}
	validUntil := new(big.Int).SetInt64(int64(time.Now().Add(time.Hour).Second()))
	validAfter := new(big.Int).SetInt64(int64(time.Now().Second()))
	argBytes, err := arg.Pack(validUntil, validAfter)
	assert.Nil(t, err)

	userOp.PaymasterAndData = append(ethcommon.Hex2Bytes(strings.TrimPrefix(common.VerifyingPaymasterContractAddr, "0x")), argBytes...)

	testNVM.RunSingleTX(verifyingPaymaster, entrypoint.EthAddress, func() error {
		sk := crypto.ToECDSAUnsafe(ethcommon.Hex2Bytes(verifyingPaymasterKey))
		hash := verifyingPaymaster.getHash(userOp, validUntil, validAfter)
		ethHash := accounts.TextHash(hash[:])
		sig, err := crypto.Sign(ethHash, sk)
		assert.Nil(t, err)
		userOp.PaymasterAndData = append(userOp.PaymasterAndData, sig...)

		_, validateData, err := verifyingPaymaster.ValidatePaymasterUserOp(userOp, [32]byte{}, big.NewInt(100000000))
		assert.Nil(t, err)
		validation := interfaces.ParseValidationData(validateData)
		assert.EqualValues(t, interfaces.SigValidationSucceeded, validation.SigValidation)
		return err
	})
}
