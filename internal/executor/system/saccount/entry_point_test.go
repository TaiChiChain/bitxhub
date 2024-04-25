package saccount

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/interfaces"
	"github.com/axiomesh/axiom-ledger/pkg/packer"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

const (
	salt           = 88899
	accountBalance = 3000000000000000000
)

func buildUserOp(t *testing.T, testNVM *common.TestNVM, entrypoint *EntryPoint, accountFactory *SmartAccountFactory) (*interfaces.UserOperation, *ecdsa.PrivateKey) {
	sk, _ := crypto.GenerateKey()
	owner := crypto.PubkeyToAddress(sk.PublicKey)
	initCode := ethcommon.Hex2Bytes("00000000000000000000000000000000000010095fbfb9cf0000000000000000000000009965507d1a55bcc2695c58ba16fb37d819b0a4dc0000000000000000000000000000000000000000000000000000000000015b43")
	callData := ethcommon.Hex2Bytes("b61d27f6000000000000000000000000ed17543171c1459714cdc6519b58ffcc29a3c3c9000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000044a9059cbb000000000000000000000000c7f999b83af6df9e67d0a37ee7e900bf38b3d013000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000")
	// replace owner
	copy(initCode[20+4+(32-20):], owner.Bytes())

	senderAddr := accountFactory.GetAddress(owner, big.NewInt(salt))
	userOp := &interfaces.UserOperation{
		Sender:               senderAddr,
		Nonce:                big.NewInt(0),
		InitCode:             initCode,
		CallData:             callData,
		CallGasLimit:         big.NewInt(0),
		VerificationGasLimit: big.NewInt(9000000),
		PreVerificationGas:   big.NewInt(21000),
		MaxFeePerGas:         big.NewInt(2000000000000),
		MaxPriorityFeePerGas: big.NewInt(2000000000000),
		PaymasterAndData:     nil,
	}

	var userOpHash ethcommon.Hash
	testNVM.Call(entrypoint, ethcommon.Address{}, func() {
		userOpHash = entrypoint.GetUserOpHash(*userOp)
	})

	hash := accounts.TextHash(userOpHash.Bytes())
	sig, err := crypto.Sign(hash, sk)
	assert.Nil(t, err)
	userOp.Signature = sig
	return userOp, sk
}

func TestEntryPoint_SimulateHandleOp(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	entrypoint := EntryPointBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	accountFactory := SmartAccountFactoryBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	tokenPaymaster := TokenPaymasterBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(tokenPaymaster)

	userOp, sk := buildUserOp(t, testNVM, entrypoint, accountFactory)
	testNVM.Call(entrypoint, ethcommon.Address{}, func() {
		err := entrypoint.SimulateHandleOp(*userOp, ethcommon.Address{}, nil)
		executeResultErr, ok := err.(*packer.RevertError)
		assert.True(t, ok, "can't convert err to RevertError")
		assert.Equal(t, executeResultErr.Err, vm.ErrExecutionReverted)
		assert.Contains(t, executeResultErr.Error(), "AA21 didn't pay prefund")
	})

	userOp.InitCode = nil
	userOp.VerificationGasLimit = big.NewInt(90000)
	userOpHash := entrypoint.GetUserOpHash(*userOp)
	hash := accounts.TextHash(userOpHash.Bytes())
	sig, err := crypto.Sign(hash, sk)
	assert.Nil(t, err)
	userOp.Signature = sig

	testNVM.Call(entrypoint, ethcommon.Address{}, func() {
		err := entrypoint.SimulateHandleOp(*userOp, ethcommon.Address{}, nil)
		executeResultErr, ok := err.(*packer.RevertError)
		assert.True(t, ok, "can't convert err to RevertError")
		assert.Equal(t, executeResultErr.Err, vm.ErrExecutionReverted)
		assert.ErrorContains(t, executeResultErr, "account not deployed")
	})
}

func TestEntryPoint_HandleOp_Error(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	entrypoint := EntryPointBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	accountFactory := SmartAccountFactoryBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	tokenPaymaster := TokenPaymasterBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	verifyingPaymaster := VerifyingPaymasterBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(tokenPaymaster, verifyingPaymaster, accountFactory, entrypoint)

	userOp, sk := buildUserOp(t, testNVM, entrypoint, accountFactory)

	var err error
	testcases := []struct {
		BeforeRun  func(userOp *interfaces.UserOperation)
		IsSkipSign bool
		ErrMsg     string
	}{
		{
			BeforeRun: func(userOp *interfaces.UserOperation) {},

			ErrMsg: "AA21 didn't pay prefund",
		},
		{
			BeforeRun: func(userOp *interfaces.UserOperation) {
				userOp.VerificationGasLimit = big.NewInt(90000)

				testNVM.RunSingleTX(entrypoint, ethcommon.Address{}, func() error {
					_, _, err := entrypoint.createSender(userOp.VerificationGasLimit, userOp.InitCode)
					return err
				})
			},
			ErrMsg: "AA10 sender already constructed",
		},
		{
			BeforeRun: func(userOp *interfaces.UserOperation) {
				userOp.VerificationGasLimit = big.NewInt(90000)
				userOp.InitCode = nil
			},
			IsSkipSign: true,
			ErrMsg:     "signature error",
		},
		{
			// not set paymaster signature
			BeforeRun: func(userOp *interfaces.UserOperation) {
				userOp.VerificationGasLimit = big.NewInt(90000)
				userOp.InitCode = nil
				userOp.Nonce, err = entrypoint.GetNonce(userOp.Sender, big.NewInt(salt))
				assert.Nil(t, err)

				paymasterAndData := ethcommon.HexToAddress(common.VerifyingPaymasterContractAddr).Bytes()
				arg := abi.Arguments{
					{Name: "validUntil", Type: common.UInt48Type},
					{Name: "validAfter", Type: common.UInt48Type},
				}
				validUntil := big.NewInt(time.Now().Add(10 * time.Second).Unix())
				validAfter := big.NewInt(time.Now().Add(-time.Second).Unix())
				timeData, _ := arg.Pack(validUntil, validAfter)
				paymasterAndData = append(paymasterAndData, timeData...)
				userOp.PaymasterAndData = paymasterAndData
			},
			ErrMsg: "invalid signature length in paymasterAndData",
		},
		{
			// wrong paymaster signature
			BeforeRun: func(userOp *interfaces.UserOperation) {
				userOp.VerificationGasLimit = big.NewInt(90000)
				userOp.InitCode = nil
				userOp.Nonce, err = entrypoint.GetNonce(userOp.Sender, big.NewInt(salt))
				assert.Nil(t, err)

				paymasterAndData := ethcommon.HexToAddress(common.VerifyingPaymasterContractAddr).Bytes()
				arg := abi.Arguments{
					{Name: "validUntil", Type: common.UInt48Type},
					{Name: "validAfter", Type: common.UInt48Type},
				}
				validUntil := big.NewInt(time.Now().Add(10 * time.Second).Unix())
				validAfter := big.NewInt(time.Now().Add(-time.Second).Unix())
				timeData, _ := arg.Pack(validUntil, validAfter)
				paymasterAndData = append(paymasterAndData, timeData...)
				testNVM.Call(verifyingPaymaster, ethcommon.Address{}, func() {
					hash := verifyingPaymaster.getHash(*userOp, validUntil, validAfter)
					// error signature format
					ethHash := accounts.TextHash(hash[:])
					sig, _ := crypto.Sign(ethHash, sk)
					userOp.PaymasterAndData = append(paymasterAndData, sig...)
				})
			},
			ErrMsg: "AA34 signature error",
		},
		{
			// not enough token balance to pay gas
			BeforeRun: func(userOp *interfaces.UserOperation) {
				// deploy oracle
				oracleContractAddr := ethcommon.HexToAddress("0x3EBD66861C1d8F298c20ED56506b063206103227")
				contractAccount := testNVM.StateLedger.GetOrCreateAccount(types.NewAddress(oracleContractAddr.Bytes()))
				contractAccount.SetCodeAndHash(ethcommon.Hex2Bytes(oracleBytecode))
				// deploy erc20
				erc20ContractAddr := ethcommon.HexToAddress("0x82C6D3ed4cD33d8EC1E51d0B5Cc1d822Eaa0c3dC")
				contractAccount = testNVM.StateLedger.GetOrCreateAccount(types.NewAddress(erc20ContractAddr.Bytes()))
				contractAccount.SetCodeAndHash(ethcommon.Hex2Bytes(erc20bytecode))

				userOp.VerificationGasLimit = big.NewInt(90000)
				userOp.InitCode = nil
				userOp.Nonce, err = entrypoint.GetNonce(userOp.Sender, big.NewInt(salt))
				assert.Nil(t, err)

				testNVM.RunSingleTX(tokenPaymaster, ethcommon.HexToAddress(testNVM.Rep.GenesisConfig.SmartAccountAdmin), func() error {
					err := tokenPaymaster.AddToken(erc20ContractAddr, oracleContractAddr)
					assert.Nil(t, err)
					return err
				})

				// reset caller
				paymasterAndData := ethcommon.HexToAddress(common.TokenPaymasterContractAddr).Bytes()
				tokenAddr := erc20ContractAddr.Bytes()
				userOp.PaymasterAndData = append(paymasterAndData, tokenAddr...)
			},
			ErrMsg: "not enough token balance",
		},
	}

	for i, testcase := range testcases {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			tmpUserOp := *userOp
			testcase.BeforeRun(&tmpUserOp)

			testNVM.RunSingleTX(entrypoint, ethcommon.Address{}, func() error {
				entrypoint.Ctx.StateLedger.SetBalance(types.NewAddress(tmpUserOp.Sender.Bytes()), types.CoinNumberByAxc(3).ToBigInt())
				entrypoint.Ctx.StateLedger.SetBalance(entrypoint.Address, types.CoinNumberByAxc(3).ToBigInt())
				entrypoint.Ctx.StateLedger.SetBalance(verifyingPaymaster.Address, types.CoinNumberByAxc(3).ToBigInt())
				entrypoint.Ctx.StateLedger.SetBalance(tokenPaymaster.Address, types.CoinNumberByAxc(3).ToBigInt())

				if !testcase.IsSkipSign {
					userOpHash := entrypoint.GetUserOpHash(tmpUserOp)
					hash := accounts.TextHash(userOpHash.Bytes())
					sig, err := crypto.Sign(hash, sk)
					assert.Nil(t, err)
					tmpUserOp.Signature = sig
				} else {
					tmpUserOp.Signature = nil
				}

				err := entrypoint.HandleOps([]interfaces.UserOperation{tmpUserOp}, ethcommon.Address{})
				executeResultErr, ok := err.(*packer.RevertError)
				assert.True(t, ok, "can't convert err to RevertError")
				assert.Equal(t, executeResultErr.Err, vm.ErrExecutionReverted)
				assert.Contains(t, executeResultErr.Error(), testcase.ErrMsg)
				return err
			})
		})
	}
}

func TestEntryPoint_HandleOp_UserOperationRevert(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	entrypoint := EntryPointBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	accountFactory := SmartAccountFactoryBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	tokenPaymaster := TokenPaymasterBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	verifyingPaymaster := VerifyingPaymasterBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(tokenPaymaster, verifyingPaymaster, accountFactory, entrypoint)

	var err error
	userOp, sk := buildUserOp(t, testNVM, entrypoint, accountFactory)

	testcases := []struct {
		BeforeRun func(userOp *interfaces.UserOperation)
		ErrMsg    string
	}{
		{
			// session key validAfter > validUntil
			BeforeRun: func(userOp *interfaces.UserOperation) {
				userOp.VerificationGasLimit = big.NewInt(90000)
				userOp.Nonce, err = entrypoint.GetNonce(userOp.Sender, big.NewInt(salt))
				assert.Nil(t, err)

				sessionPriv, _ := crypto.GenerateKey()
				owner := crypto.PubkeyToAddress(sessionPriv.PublicKey)
				userOp.CallData = append([]byte{}, setSessionSig...)
				userOp.CallData = append(userOp.CallData, ethcommon.LeftPadBytes(owner.Bytes(), 32)...)
				userOp.CallData = append(userOp.CallData, ethcommon.LeftPadBytes(big.NewInt(3000000000000000000).Bytes(), 32)...)
				validUntil := big.NewInt(time.Now().Add(-time.Second).Unix())
				validAfter := big.NewInt(time.Now().Add(-time.Second).Unix())
				userOp.CallData = append(userOp.CallData, ethcommon.LeftPadBytes(validAfter.Bytes(), 32)...)
				userOp.CallData = append(userOp.CallData, ethcommon.LeftPadBytes(validUntil.Bytes(), 32)...)
			},
			ErrMsg: "session key validAfter must less than validUntil",
		},
		{
			// execute batch error
			BeforeRun: func(userOp *interfaces.UserOperation) {
				userOp.VerificationGasLimit = big.NewInt(90000)
				userOp.InitCode = nil
				userOp.Nonce, err = entrypoint.GetNonce(userOp.Sender, big.NewInt(salt))
				assert.Nil(t, err)
				sessionPriv, _ := crypto.GenerateKey()
				owner := crypto.PubkeyToAddress(sessionPriv.PublicKey)
				userOp.CallData = append([]byte{}, executeBatchSig...)
				// error arguments
				userOp.CallData = append(userOp.CallData, ethcommon.LeftPadBytes(owner.Bytes(), 32)...)
				userOp.CallData = append(userOp.CallData, ethcommon.LeftPadBytes(big.NewInt(100000).Bytes(), 32)...)
				userOp.CallGasLimit = big.NewInt(499923)
			},
			ErrMsg: "unpack error",
		},
	}

	beneficiaryPriv, _ := crypto.GenerateKey()
	beneficiary := crypto.PubkeyToAddress(beneficiaryPriv.PublicKey)
	for i, testcase := range testcases {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			tmpUserOp := *userOp
			testcase.BeforeRun(&tmpUserOp)

			testNVM.RunSingleTX(entrypoint, ethcommon.Address{}, func() error {
				entrypoint.Ctx.StateLedger.SetBalance(types.NewAddress(tmpUserOp.Sender.Bytes()), types.CoinNumberByAxc(3).ToBigInt())
				entrypoint.Ctx.StateLedger.SetBalance(entrypoint.Address, types.CoinNumberByAxc(3).ToBigInt())
				entrypoint.Ctx.StateLedger.SetBalance(verifyingPaymaster.Address, types.CoinNumberByAxc(3).ToBigInt())
				entrypoint.Ctx.StateLedger.SetBalance(tokenPaymaster.Address, types.CoinNumberByAxc(3).ToBigInt())

				userOpHash := entrypoint.GetUserOpHash(tmpUserOp)

				hash := accounts.TextHash(userOpHash.Bytes())
				sig, err := crypto.Sign(hash, sk)
				assert.Nil(t, err)
				tmpUserOp.Signature = sig

				err = entrypoint.HandleOps([]interfaces.UserOperation{tmpUserOp}, beneficiary)
				assert.Nil(t, err)

				logs := entrypoint.Ctx.GetLogs()
				assert.Greater(t, len(logs), 1)
				// get log
				userOpRevertEvent := abi.NewEvent("UserOperationRevertReason", "UserOperationRevertReason", false, abi.Arguments{
					{Name: "userOpHash", Type: common.Bytes32Type, Indexed: true},
					{Name: "sender", Type: common.AddressType, Indexed: true},
					{Name: "nonce", Type: common.BigIntType},
					{Name: "revertReason", Type: common.BytesType},
				})
				hasUserOpRevertEvent := false
				for _, currentLog := range logs {
					topicOne := *currentLog.Topics[0]
					if topicOne.ETHHash() == userOpRevertEvent.ID {
						revertReason := string(currentLog.Data[32:])
						if strings.Contains(revertReason, testcase.ErrMsg) {
							hasUserOpRevertEvent = true
						}
					}
				}
				assert.True(t, hasUserOpRevertEvent, "must have UserOperationRevertReason event")
				return err
			})
		})
	}
}

func TestEntryPoint_HandleOp(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	entrypoint := EntryPointBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	accountFactory := SmartAccountFactoryBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	tokenPaymaster := TokenPaymasterBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	verifyingPaymaster := VerifyingPaymasterBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(tokenPaymaster, verifyingPaymaster, accountFactory, entrypoint)

	paymasterPriv, err := crypto.HexToECDSA(repo.MockOperatorKeys[0])
	assert.Nil(t, err)

	userOp, sk := buildUserOp(t, testNVM, entrypoint, accountFactory)

	guardianPriv, _ := crypto.GenerateKey()
	guardian := crypto.PubkeyToAddress(guardianPriv.PublicKey)
	newOwnerPriv, _ := crypto.GenerateKey()
	newOwner := crypto.PubkeyToAddress(newOwnerPriv.PublicKey)

	testcases := []struct {
		BeforeRun func(userOp *interfaces.UserOperation) *ecdsa.PrivateKey
		AfterRun  func(userOp *interfaces.UserOperation)
	}{
		{
			BeforeRun: func(userOp *interfaces.UserOperation) *ecdsa.PrivateKey {
				userOp.VerificationGasLimit = big.NewInt(90000)
				return nil
			},
			AfterRun: func(userOp *interfaces.UserOperation) {
				balance := testNVM.StateLedger.GetBalance(types.NewAddress(userOp.Sender.Bytes()))
				assert.Greater(t, balance.Uint64(), uint64(0))
				assert.Greater(t, uint64(accountBalance), balance.Uint64())
			},
		},
		{
			// set verifying paymaster
			BeforeRun: func(userOp *interfaces.UserOperation) *ecdsa.PrivateKey {
				userOp.VerificationGasLimit = big.NewInt(90000)
				sk, _ := crypto.GenerateKey()
				owner := crypto.PubkeyToAddress(sk.PublicKey)
				initCode := ethcommon.Hex2Bytes("00000000000000000000000000000000000010095fbfb9cf0000000000000000000000009965507d1a55bcc2695c58ba16fb37d819b0a4dc0000000000000000000000000000000000000000000000000000000000015b43")
				// replace owner
				copy(initCode[20+4+(32-20):], owner.Bytes())
				userOp.InitCode = initCode
				userOp.Sender = accountFactory.GetAddress(owner, big.NewInt(salt))
				userOp.Nonce, err = entrypoint.GetNonce(userOp.Sender, big.NewInt(salt))
				assert.Nil(t, err)

				paymasterAndData := ethcommon.HexToAddress(common.VerifyingPaymasterContractAddr).Bytes()
				arg := abi.Arguments{
					{Name: "validUntil", Type: common.UInt48Type},
					{Name: "validAfter", Type: common.UInt48Type},
				}
				validUntil := big.NewInt(time.Now().Add(10 * time.Second).Unix())
				validAfter := big.NewInt(time.Now().Add(-time.Second).Unix())
				timeData, _ := arg.Pack(validUntil, validAfter)
				paymasterAndData = append(paymasterAndData, timeData...)
				var hash [32]byte
				testNVM.Call(verifyingPaymaster, owner, func() {
					hash = verifyingPaymaster.getHash(*userOp, validUntil, validAfter)
				})
				ethHash := accounts.TextHash(hash[:])
				sig, _ := crypto.Sign(ethHash, paymasterPriv)
				userOp.PaymasterAndData = append(paymasterAndData, sig...)

				return sk
			},
			AfterRun: func(userOp *interfaces.UserOperation) {
				balance := testNVM.StateLedger.GetBalance(types.NewAddress(userOp.Sender.Bytes()))
				assert.Equal(t, balance.Uint64(), uint64(accountBalance))
				t.Logf("after run, sender %s balance: %s", userOp.Sender.Hex(), balance.Text(10))

				balance = testNVM.StateLedger.GetBalance(types.NewAddressByStr(common.VerifyingPaymasterContractAddr))
				assert.Greater(t, balance.Uint64(), uint64(0))
				assert.Greater(t, uint64(accountBalance), balance.Uint64())
				t.Logf("after run, verifying paymaster balance: %s", balance.Text(10))
			},
		},
		{
			// set session key
			BeforeRun: func(userOp *interfaces.UserOperation) *ecdsa.PrivateKey {
				userOp.VerificationGasLimit = big.NewInt(90000)
				sk, _ := crypto.GenerateKey()
				owner := crypto.PubkeyToAddress(sk.PublicKey)
				initCode := ethcommon.Hex2Bytes("00000000000000000000000000000000000010095fbfb9cf0000000000000000000000009965507d1a55bcc2695c58ba16fb37d819b0a4dc0000000000000000000000000000000000000000000000000000000000015b43")
				// replace owner
				copy(initCode[20+4+(32-20):], owner.Bytes())
				userOp.InitCode = initCode
				userOp.Sender = accountFactory.GetAddress(owner, big.NewInt(salt))
				userOp.Nonce, err = entrypoint.GetNonce(userOp.Sender, big.NewInt(salt))
				assert.Nil(t, err)

				sessionPriv, _ := crypto.GenerateKey()
				owner = crypto.PubkeyToAddress(sessionPriv.PublicKey)
				userOp.CallData = append([]byte{}, setSessionSig...)
				userOp.CallData = append(userOp.CallData, ethcommon.LeftPadBytes(owner.Bytes(), 32)...)
				userOp.CallData = append(userOp.CallData, ethcommon.LeftPadBytes(big.NewInt(1000000000000000000).Bytes(), 32)...)
				validUntil := big.NewInt(time.Now().Add(10 * time.Second).Unix())
				validAfter := big.NewInt(time.Now().Add(-time.Second).Unix())
				userOp.CallData = append(userOp.CallData, ethcommon.LeftPadBytes(validAfter.Bytes(), 32)...)
				userOp.CallData = append(userOp.CallData, ethcommon.LeftPadBytes(validUntil.Bytes(), 32)...)

				return sk
			},
			AfterRun: func(userOp *interfaces.UserOperation) {
				balance := testNVM.StateLedger.GetBalance(types.NewAddress(userOp.Sender.Bytes()))
				assert.Greater(t, balance.Uint64(), uint64(0))
				assert.Greater(t, uint64(accountBalance), balance.Uint64())

				// get session
				sa := SmartAccountBuildConfig.BuildWithAddress(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}), userOp.Sender).SetRemainingGas(big.NewInt(MaxCallGasLimit))
				session := sa.getSession()
				t.Logf("after run, session: %v", session)
				assert.NotNil(t, session)
				assert.Equal(t, session.SpentAmount.Uint64(), uint64(0))
				assert.Equal(t, session.SpendingLimit.Uint64(), uint64(1000000000000000000))
			},
		},
		{
			// initialize smart account with guardian
			BeforeRun: func(userOp *interfaces.UserOperation) *ecdsa.PrivateKey {
				userOp.VerificationGasLimit = big.NewInt(90000)
				sk, _ := crypto.GenerateKey()
				owner := crypto.PubkeyToAddress(sk.PublicKey)
				initCode := ethcommon.Hex2Bytes("00000000000000000000000000000000000010095fbfb9cf0000000000000000000000009965507d1a55bcc2695c58ba16fb37d819b0a4dc0000000000000000000000000000000000000000000000000000000000015b43")
				// replace owner
				copy(initCode[20+4+(32-20):], owner.Bytes())
				// append guardian
				initCode = append(initCode, ethcommon.LeftPadBytes(guardian.Bytes(), 32)...)
				userOp.InitCode = initCode
				senderAddr := accountFactory.GetAddress(owner, big.NewInt(salt))
				userOp.Sender = senderAddr
				userOp.Nonce, err = entrypoint.GetNonce(userOp.Sender, big.NewInt(salt))
				assert.Nil(t, err)
				return sk
			},
			AfterRun: func(userOp *interfaces.UserOperation) {
				sa := SmartAccountBuildConfig.BuildWithAddress(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}), userOp.Sender).SetRemainingGas(big.NewInt(MaxCallGasLimit))
				g, err := sa.getGuardian()
				assert.Nil(t, err)
				assert.Equal(t, g, guardian)
			},
		},
		{
			// set guardian
			BeforeRun: func(userOp *interfaces.UserOperation) *ecdsa.PrivateKey {
				userOp.VerificationGasLimit = big.NewInt(90000)
				sk, _ := crypto.GenerateKey()
				owner := crypto.PubkeyToAddress(sk.PublicKey)
				initCode := ethcommon.Hex2Bytes("00000000000000000000000000000000000010095fbfb9cf0000000000000000000000009965507d1a55bcc2695c58ba16fb37d819b0a4dc0000000000000000000000000000000000000000000000000000000000015b43")
				// replace owner
				copy(initCode[20+4+(32-20):], owner.Bytes())
				userOp.InitCode = initCode
				senderAddr := accountFactory.GetAddress(owner, big.NewInt(salt))
				userOp.Sender = senderAddr
				userOp.Nonce, err = entrypoint.GetNonce(userOp.Sender, big.NewInt(salt))
				assert.Nil(t, err)

				userOp.CallData = append([]byte{}, setGuardianSig...)
				userOp.CallData = append(userOp.CallData, ethcommon.LeftPadBytes(userOp.Sender.Bytes(), 32)...)
				return sk
			},
			AfterRun: func(userOp *interfaces.UserOperation) {
				sa := SmartAccountBuildConfig.BuildWithAddress(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}), userOp.Sender).SetRemainingGas(big.NewInt(MaxCallGasLimit))
				g, err := sa.getGuardian()
				assert.Nil(t, err)
				assert.Equal(t, g, userOp.Sender)
			},
		},
		{
			// reset owner
			BeforeRun: func(userOp *interfaces.UserOperation) *ecdsa.PrivateKey {
				userOp.VerificationGasLimit = big.NewInt(90000)
				sk, _ := crypto.GenerateKey()
				owner := crypto.PubkeyToAddress(sk.PublicKey)
				initCode := ethcommon.Hex2Bytes("00000000000000000000000000000000000010095fbfb9cf0000000000000000000000009965507d1a55bcc2695c58ba16fb37d819b0a4dc0000000000000000000000000000000000000000000000000000000000015b43")
				// replace owner
				copy(initCode[20+4+(32-20):], owner.Bytes())
				userOp.InitCode = initCode
				senderAddr := accountFactory.GetAddress(owner, big.NewInt(salt))
				userOp.Sender = senderAddr
				userOp.Nonce, err = entrypoint.GetNonce(userOp.Sender, big.NewInt(salt))
				assert.Nil(t, err)
				userOp.CallData = append([]byte{}, resetOwnerSig...)
				userOp.CallData = append(userOp.CallData, ethcommon.LeftPadBytes(newOwner.Bytes(), 32)...)
				return sk
			},
			AfterRun: func(userOp *interfaces.UserOperation) {
				sa := SmartAccountBuildConfig.BuildWithAddress(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}), userOp.Sender).SetRemainingGas(big.NewInt(MaxCallGasLimit))
				owner, err := sa.getOwner()
				assert.Nil(t, err)
				assert.Equal(t, owner, newOwner)
			},
		},
		{
			// call smart account execute
			BeforeRun: func(userOp *interfaces.UserOperation) *ecdsa.PrivateKey {
				userOp.VerificationGasLimit = big.NewInt(90000)
				sk, _ := crypto.GenerateKey()
				owner := crypto.PubkeyToAddress(sk.PublicKey)
				initCode := ethcommon.Hex2Bytes("00000000000000000000000000000000000010095fbfb9cf0000000000000000000000009965507d1a55bcc2695c58ba16fb37d819b0a4dc0000000000000000000000000000000000000000000000000000000000015b43")
				// replace owner
				copy(initCode[20+4+(32-20):], owner.Bytes())
				userOp.InitCode = initCode
				senderAddr := accountFactory.GetAddress(owner, big.NewInt(salt))
				userOp.Sender = senderAddr
				userOp.Nonce, err = entrypoint.GetNonce(userOp.Sender, big.NewInt(salt))
				assert.Nil(t, err)

				userOp.CallData = append([]byte{}, executeSig...)
				data, err := executeMethod.Pack(entrypoint.EthAddress, big.NewInt(100000), []byte{})
				assert.Nil(t, err)
				userOp.CallData = append(userOp.CallData, data...)
				return sk
			},
			AfterRun: func(userOp *interfaces.UserOperation) {
				balance := testNVM.StateLedger.GetBalance(entrypoint.Address)
				fmt.Println("balance", balance)
				assert.Greater(t, balance.Uint64(), uint64(100000))
			},
		},
		{
			// call smart account execute batch
			BeforeRun: func(userOp *interfaces.UserOperation) *ecdsa.PrivateKey {
				userOp.VerificationGasLimit = big.NewInt(90000)
				sk, _ := crypto.GenerateKey()
				owner := crypto.PubkeyToAddress(sk.PublicKey)
				initCode := ethcommon.Hex2Bytes("00000000000000000000000000000000000010095fbfb9cf0000000000000000000000009965507d1a55bcc2695c58ba16fb37d819b0a4dc0000000000000000000000000000000000000000000000000000000000015b43")
				// replace owner
				copy(initCode[20+4+(32-20):], owner.Bytes())
				userOp.InitCode = initCode
				senderAddr := accountFactory.GetAddress(owner, big.NewInt(salt))
				userOp.Sender = senderAddr
				userOp.Nonce, err = entrypoint.GetNonce(userOp.Sender, big.NewInt(salt))
				assert.Nil(t, err)

				userOp.CallData = append([]byte{}, executeBatchSig...)
				data, err := executeBatchMethod.Pack([]ethcommon.Address{guardian, newOwner}, [][]byte{nil, nil})
				assert.Nil(t, err)
				userOp.CallData = append(userOp.CallData, data...)
				return sk
			},
			AfterRun: func(userOp *interfaces.UserOperation) {},
		},
	}

	beneficiaryPriv, _ := crypto.GenerateKey()
	beneficiary := crypto.PubkeyToAddress(beneficiaryPriv.PublicKey)
	for i, testcase := range testcases {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			tmpUserOp := *userOp

			priv := testcase.BeforeRun(&tmpUserOp)
			if priv == nil {
				priv = sk
			}

			testNVM.RunSingleTX(entrypoint, ethcommon.Address{}, func() error {
				entrypoint.Ctx.StateLedger.SetBalance(types.NewAddress(tmpUserOp.Sender.Bytes()), types.CoinNumberByAxc(3).ToBigInt())
				entrypoint.Ctx.StateLedger.SetBalance(entrypoint.Address, types.CoinNumberByAxc(3).ToBigInt())
				entrypoint.Ctx.StateLedger.SetBalance(verifyingPaymaster.Address, types.CoinNumberByAxc(3).ToBigInt())
				entrypoint.Ctx.StateLedger.SetBalance(tokenPaymaster.Address, types.CoinNumberByAxc(3).ToBigInt())

				userOpHash := entrypoint.GetUserOpHash(tmpUserOp)
				hash := accounts.TextHash(userOpHash.Bytes())
				sig, err := crypto.Sign(hash, priv)
				assert.Nil(t, err)
				tmpUserOp.Signature = sig

				err = entrypoint.HandleOps([]interfaces.UserOperation{tmpUserOp}, beneficiary)
				assert.Nil(t, err)
				testcase.AfterRun(&tmpUserOp)
				return err
			})
		})
	}
}

func TestEntryPoint_HandleOps_SessionKey(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	entrypoint := EntryPointBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	accountFactory := SmartAccountFactoryBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	tokenPaymaster := TokenPaymasterBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	verifyingPaymaster := VerifyingPaymasterBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(tokenPaymaster, verifyingPaymaster, accountFactory, entrypoint)

	entrypoint.Ctx.StateLedger.SetBalance(entrypoint.Address, types.CoinNumberByAxc(3).ToBigInt())
	entrypoint.Ctx.StateLedger.SetBalance(verifyingPaymaster.Address, types.CoinNumberByAxc(3).ToBigInt())
	entrypoint.Ctx.StateLedger.SetBalance(tokenPaymaster.Address, types.CoinNumberByAxc(3).ToBigInt())

	var err error
	userOp, _ := buildUserOp(t, testNVM, entrypoint, accountFactory)
	beneficiaryPriv, _ := crypto.GenerateKey()
	beneficiary := crypto.PubkeyToAddress(beneficiaryPriv.PublicKey)

	userOp.VerificationGasLimit = big.NewInt(90000)
	sk, _ := crypto.GenerateKey()
	owner := crypto.PubkeyToAddress(sk.PublicKey)
	initCode := ethcommon.Hex2Bytes("00000000000000000000000000000000000010095fbfb9cf0000000000000000000000009965507d1a55bcc2695c58ba16fb37d819b0a4dc0000000000000000000000000000000000000000000000000000000000015b43")
	// replace owner
	copy(initCode[20+4+(32-20):], owner.Bytes())
	userOp.InitCode = initCode
	userOp.Sender = accountFactory.GetAddress(owner, big.NewInt(salt))
	userOp.Nonce, err = entrypoint.GetNonce(userOp.Sender, big.NewInt(salt))
	assert.Nil(t, err)

	sessionPriv, _ := crypto.GenerateKey()
	sessionOwner := crypto.PubkeyToAddress(sessionPriv.PublicKey)
	userOp.CallData = append([]byte{}, setSessionSig...)
	sessionKeyLimit := big.NewInt(1000000000000000000)
	userOp.CallData = append(userOp.CallData, ethcommon.LeftPadBytes(sessionOwner.Bytes(), 32)...)
	userOp.CallData = append(userOp.CallData, ethcommon.LeftPadBytes(sessionKeyLimit.Bytes(), 32)...)
	validUntil := big.NewInt(time.Now().Add(10 * time.Second).Unix())
	validAfter := big.NewInt(time.Now().Add(-time.Second).Unix())
	userOp.CallData = append(userOp.CallData, ethcommon.LeftPadBytes(validAfter.Bytes(), 32)...)
	userOp.CallData = append(userOp.CallData, ethcommon.LeftPadBytes(validUntil.Bytes(), 32)...)

	userOpHash := entrypoint.GetUserOpHash(*userOp)
	hash := accounts.TextHash(userOpHash.Bytes())
	sig, err := crypto.Sign(hash, sk)
	assert.Nil(t, err)
	userOp.Signature = sig

	// use session key to sign
	sessionUserOp := *userOp
	sessionUserOp.VerificationGasLimit = big.NewInt(90000)
	sessionUserOp.InitCode = nil
	sessionUserOp.Nonce = new(big.Int).Add(userOp.Nonce, big.NewInt(1))
	sessionUserOp.CallData = append([]byte{}, executeSig...)
	transferAmount := big.NewInt(100000)
	data, err := executeMethod.Pack(entrypoint.EthAddress, transferAmount, []byte{})
	assert.Nil(t, err)
	sessionUserOp.CallData = append(sessionUserOp.CallData, data...)
	userOpHash = entrypoint.GetUserOpHash(sessionUserOp)
	hash = accounts.TextHash(userOpHash.Bytes())
	sessionKeySig, err := crypto.Sign(hash, sessionPriv)
	assert.Nil(t, err)
	sessionUserOp.Signature = sessionKeySig

	testNVM.RunSingleTX(entrypoint, ethcommon.Address{}, func() error {
		entrypoint.Ctx.StateLedger.SetBalance(types.NewAddress(userOp.Sender.Bytes()), types.CoinNumberByAxc(3).ToBigInt())

		err = entrypoint.HandleOps([]interfaces.UserOperation{*userOp}, beneficiary)
		assert.Nil(t, err)
		return err
	})

	// after set session key effect, use session key to sign
	testNVM.RunSingleTX(entrypoint, ethcommon.Address{}, func() error {
		entrypoint.Ctx.StateLedger.SetBalance(types.NewAddress(userOp.Sender.Bytes()), types.CoinNumberByAxc(3).ToBigInt())

		err = entrypoint.HandleOps([]interfaces.UserOperation{sessionUserOp}, beneficiary)
		assert.Nil(t, err)
		return err
	})

	// get session key spend limit
	sa := SmartAccountBuildConfig.BuildWithAddress(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}), sessionUserOp.Sender).SetRemainingGas(big.NewInt(MaxCallGasLimit))
	testNVM.RunSingleTX(sa, entrypoint.EthAddress, func() error {
		entrypoint.Ctx.StateLedger.SetBalance(types.NewAddress(userOp.Sender.Bytes()), types.CoinNumberByAxc(3).ToBigInt())

		err := sa.Initialize(ethcommon.Address{}, ethcommon.Address{})
		assert.Nil(t, err)
		return err
	})

	session := sa.getSession()
	assert.Greater(t, session.SpentAmount.Uint64(), transferAmount.Uint64())
	assert.EqualValues(t, sessionKeyLimit.Uint64(), session.SpendingLimit.Uint64())
}

func TestEntryPoint_HandleOps_Recovery(t *testing.T) {
	// use guardian to recovery owner
	testNVM := common.NewTestNVM(t)
	entrypoint := EntryPointBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	accountFactory := SmartAccountFactoryBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	tokenPaymaster := TokenPaymasterBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	verifyingPaymaster := VerifyingPaymasterBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(tokenPaymaster, verifyingPaymaster, accountFactory, entrypoint)

	var err error
	userOp, _ := buildUserOp(t, testNVM, entrypoint, accountFactory)
	beneficiaryPriv, _ := crypto.GenerateKey()
	beneficiary := crypto.PubkeyToAddress(beneficiaryPriv.PublicKey)

	userOp.VerificationGasLimit = big.NewInt(90000)
	sk, _ := crypto.GenerateKey()
	owner := crypto.PubkeyToAddress(sk.PublicKey)
	initCode := ethcommon.Hex2Bytes("00000000000000000000000000000000000010095fbfb9cf0000000000000000000000009965507d1a55bcc2695c58ba16fb37d819b0a4dc0000000000000000000000000000000000000000000000000000000000015b43")
	// replace owner
	copy(initCode[20+4+(32-20):], owner.Bytes())
	userOp.InitCode = initCode
	userOp.Sender = accountFactory.GetAddress(owner, big.NewInt(salt))
	userOp.Nonce, err = entrypoint.GetNonce(userOp.Sender, big.NewInt(salt))
	assert.Nil(t, err)

	// set guardian
	guardianPriv, _ := crypto.GenerateKey()
	guardianOwner := crypto.PubkeyToAddress(guardianPriv.PublicKey)
	userOp.CallData = append([]byte{}, setGuardianSig...)
	userOp.CallData = append(userOp.CallData, ethcommon.LeftPadBytes(guardianOwner.Bytes(), 32)...)

	userOpHash := entrypoint.GetUserOpHash(*userOp)
	hash := accounts.TextHash(userOpHash.Bytes())
	sig, err := crypto.Sign(hash, sk)
	assert.Nil(t, err)
	userOp.Signature = sig

	testNVM.RunSingleTX(entrypoint, ethcommon.Address{}, func() error {
		entrypoint.Ctx.StateLedger.SetBalance(types.NewAddress(userOp.Sender.Bytes()), types.CoinNumberByAxc(3).ToBigInt())

		err = entrypoint.HandleOps([]interfaces.UserOperation{*userOp}, beneficiary)
		assert.Nil(t, err)
		return err
	})

	// use guardian to sign to recovery owner
	userOp.InitCode = nil
	userOp.Nonce, err = entrypoint.GetNonce(userOp.Sender, big.NewInt(salt))
	assert.Nil(t, err)

	newOwnerPriv, _ := crypto.GenerateKey()
	newOwner := crypto.PubkeyToAddress(newOwnerPriv.PublicKey)
	// call HandleAccountRecovery, directly set new owner
	userOp.CallData = append([]byte{}, newOwner.Bytes()...)

	// guardian sign
	userOpHash = entrypoint.GetUserOpHash(*userOp)
	hash = accounts.TextHash(userOpHash.Bytes())
	sig, err = crypto.Sign(hash, guardianPriv)
	assert.Nil(t, err)
	userOp.Signature = sig

	// tmp set lock time
	LockedTime = 2 * time.Second
	testNVM.RunSingleTX(entrypoint, ethcommon.Address{}, func() error {
		entrypoint.Ctx.StateLedger.SetBalance(types.NewAddress(userOp.Sender.Bytes()), types.CoinNumberByAxc(3).ToBigInt())

		err = entrypoint.HandleAccountRecovery([]interfaces.UserOperation{*userOp}, beneficiary)
		assert.Nil(t, err)
		return err
	})

	// use new owner would failed
	userOp.InitCode = nil
	userOp.Nonce, err = entrypoint.GetNonce(userOp.Sender, big.NewInt(salt))
	assert.Nil(t, err)
	guardianPriv, _ = crypto.GenerateKey()
	guardianOwner = crypto.PubkeyToAddress(guardianPriv.PublicKey)
	userOp.CallData = append([]byte{}, setGuardianSig...)
	userOp.CallData = append(userOp.CallData, ethcommon.LeftPadBytes(guardianOwner.Bytes(), 32)...)
	userOpHash = entrypoint.GetUserOpHash(*userOp)
	hash = accounts.TextHash(userOpHash.Bytes())
	sig, err = crypto.Sign(hash, newOwnerPriv)
	assert.Nil(t, err)
	userOp.Signature = sig

	testNVM.RunSingleTX(entrypoint, ethcommon.Address{}, func() error {
		entrypoint.Ctx.StateLedger.SetBalance(types.NewAddress(userOp.Sender.Bytes()), types.CoinNumberByAxc(3).ToBigInt())

		err = entrypoint.HandleOps([]interfaces.UserOperation{*userOp}, beneficiary)
		assert.Contains(t, err.Error(), "signature error")
		return err
	})

	// after lock time, new owner would be success
	time.Sleep(LockedTime + time.Second)
	// update evm block time
	testNVM.RunSingleTX(entrypoint, ethcommon.Address{}, func() error {
		entrypoint.Ctx.StateLedger.SetBalance(types.NewAddress(userOp.Sender.Bytes()), types.CoinNumberByAxc(3).ToBigInt())

		entrypoint.Ctx.CurrentEVM.Context.Time = uint64(time.Now().Unix())
		userOp.Nonce, err = entrypoint.GetNonce(userOp.Sender, big.NewInt(salt))
		userOpHash = entrypoint.GetUserOpHash(*userOp)
		hash = accounts.TextHash(userOpHash.Bytes())
		sig, err = crypto.Sign(hash, newOwnerPriv)
		assert.Nil(t, err)
		userOp.Signature = sig
		err = entrypoint.HandleOps([]interfaces.UserOperation{*userOp}, beneficiary)
		assert.Nil(t, err)
		return err
	})
}

func TestEntryPoint_call(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	entrypoint := EntryPointBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	accountFactory := SmartAccountFactoryBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	tokenPaymaster := TokenPaymasterBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	verifyingPaymaster := VerifyingPaymasterBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(tokenPaymaster, verifyingPaymaster, accountFactory, entrypoint)

	// test call method
	testNVM.RunSingleTX(entrypoint, ethcommon.Address{}, func() error {
		entrypoint.Ctx.StateLedger.SetBalance(entrypoint.Address, types.CoinNumberByAxc(3).ToBigInt())

		res, usedGas, err := entrypoint.CrossCallEVMContract(big.NewInt(300000), accountFactory.EthAddress, []byte(""))
		assert.Nil(t, err)
		assert.Nil(t, res)
		assert.EqualValues(t, usedGas, 300000)
		return err
	})

	testNVM.RunSingleTX(entrypoint, ethcommon.Address{}, func() error {
		entrypoint.Ctx.StateLedger.SetBalance(entrypoint.Address, types.CoinNumberByAxc(3).ToBigInt())

		_, usedGas, err := entrypoint.CrossCallEVMContractWithValue(big.NewInt(300000), big.NewInt(10000), accountFactory.EthAddress, []byte(""))
		assert.Nil(t, err)
		assert.EqualValues(t, usedGas, 300000)
		return err
	})
}

func TestEntryPoint_GetUserOpHash(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	entrypoint := EntryPointBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	accountFactory := SmartAccountFactoryBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	tokenPaymaster := TokenPaymasterBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	verifyingPaymaster := VerifyingPaymasterBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(tokenPaymaster, verifyingPaymaster, accountFactory, entrypoint)

	dstUserOpHash := "8ecc047965a0ba0f814465f3152cbba2420b6d3fb97fea3d9e1332fc47040b54"
	sk, err := crypto.HexToECDSA("8b3a350cf5c34c9194ca85829a2df0ec3153be0318b5e2d3348e872092edffba")
	assert.Nil(t, err)
	owner := crypto.PubkeyToAddress(sk.PublicKey)

	t.Logf("owner addr: %s", owner.String())

	initCode := ethcommon.Hex2Bytes("00000000000000000000000000000000000010095fbfb9cf0000000000000000000000009965507d1a55bcc2695c58ba16fb37d819b0a4dc0000000000000000000000000000000000000000000000000000000000015b43")
	// replace owner
	copy(initCode[20+4+(32-20):], owner.Bytes())
	t.Logf("initCode addr: %v", initCode)
	callData := ethcommon.Hex2Bytes("b61d27f60000000000000000000000008464135c8f25da09e49bc8782676a84730c318bc000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000044a9059cbb000000000000000000000000c7f999b83af6df9e67d0a37ee7e900bf38b3d013000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000")

	var senderAddr ethcommon.Address
	testNVM.RunSingleTX(entrypoint, ethcommon.Address{}, func() error {
		senderAddr, _, err = entrypoint.createSender(big.NewInt(10000), initCode)
		assert.Nil(t, err)
		t.Logf("sender addr: %s", senderAddr.String())
		return err
	})

	userOp := &interfaces.UserOperation{
		Sender:               senderAddr,
		Nonce:                big.NewInt(0),
		InitCode:             initCode,
		CallData:             callData,
		CallGasLimit:         big.NewInt(12100),
		VerificationGasLimit: big.NewInt(21971),
		PreVerificationGas:   big.NewInt(50579),
		MaxFeePerGas:         big.NewInt(5875226332498),
		MaxPriorityFeePerGas: big.NewInt(0),
		PaymasterAndData:     nil,
		Signature:            ethcommon.Hex2Bytes("e00bd84bf950f07cfde52f7726bc33b8e1afd5bc709f00a58c0e1fa4b19cd0c21c4784569452603a608ebc338a99a9428cb25922acc71936befaf4d7b90925311b"),
	}

	var userOpHash ethcommon.Hash
	testNVM.RunSingleTX(entrypoint, ethcommon.Address{}, func() error {
		userOpHash = entrypoint.GetUserOpHash(*userOp)
		t.Logf("op hash: %x", userOpHash)
		assert.Equal(t, ethcommon.Hex2Bytes(dstUserOpHash), userOpHash.Bytes())
		return err
	})

	hash := accounts.TextHash(userOpHash.Bytes())
	// 0xc911ab31b05daa531d9ef77fc2670e246a9489610e48e88d8296604c26dcd5c4
	t.Logf("eth hash: %x", hash)
	sig, err := crypto.Sign(hash, sk)
	assert.Nil(t, err)
	// js return r|s|v, v only 1 byte
	// golang return rid, v = rid +27
	msgSig := make([]byte, 65)
	copy(msgSig, sig)
	msgSig[64] += 27
	t.Logf("msg sig: %x", msgSig)
	assert.Equal(t, msgSig, userOp.Signature)

	recoveredPub, err := crypto.Ecrecover(hash, sig)
	assert.Nil(t, err)
	pubKey, _ := crypto.UnmarshalPubkey(recoveredPub)
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	assert.Equal(t, owner, recoveredAddr)
}

func TestEntryPoint_createSender(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	entrypoint := EntryPointBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	accountFactory := SmartAccountFactoryBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	tokenPaymaster := TokenPaymasterBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	verifyingPaymaster := VerifyingPaymasterBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(tokenPaymaster, verifyingPaymaster, accountFactory, entrypoint)

	initCode := ethcommon.Hex2Bytes("00000000000000000000000000000000000010095fbfb9cf0000000000000000000000009965507d1a55bcc2695c58ba16fb37d819b0a4dc0000000000000000000000000000000000000000000000000000000000015b43")

	var addr ethcommon.Address
	var err error
	testNVM.RunSingleTX(entrypoint, ethcommon.Address{}, func() error {
		addr, _, err = entrypoint.createSender(big.NewInt(10000), initCode)
		assert.Nil(t, err)
		assert.NotNil(t, addr)
		assert.NotEqual(t, addr, ethcommon.Address{})
		return err
	})

	testNVM.RunSingleTX(entrypoint, ethcommon.Address{}, func() error {
		addr2, _, err := entrypoint.createSender(big.NewInt(10000), initCode)
		assert.Nil(t, err)
		assert.Equal(t, addr, addr2)
		return err
	})
}
