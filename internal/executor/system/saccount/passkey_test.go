package saccount

// import (
// 	"math/big"
// 	"testing"

// 	"github.com/axiomesh/axiom-kit/types"
// 	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
// 	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/interfaces"
// 	ethcommon "github.com/ethereum/go-ethereum/common"
// 	"github.com/go-webauthn/webauthn/protocol/webauthncbor"
// 	"github.com/go-webauthn/webauthn/protocol/webauthncose"
// 	"github.com/stretchr/testify/assert"
// )

// func Test_validatePasskey(t *testing.T) {
// 	sender := types.NewAddressByStr("A48A2bA5d7C23BeBDaCb77C6F79762d0f55D4781")
// 	callData := ethcommon.Hex2Bytes("b61d27f6000000000000000000000000c7f999b83af6df9e67d0a37ee7e900bf38b3d0130000000000000000000000000000000000000000000000000de0b6b3a764000000000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000000")
// 	userOp := &interfaces.UserOperation{
// 		Sender:               sender.ETHAddress(),
// 		Nonce:                big.NewInt(4),
// 		InitCode:             []byte{},
// 		CallData:             callData,
// 		CallGasLimit:         new(big.Int).SetBytes(ethcommon.Hex2Bytes("ec54")),
// 		VerificationGasLimit: new(big.Int).SetBytes(ethcommon.Hex2Bytes("55d3")),
// 		PreVerificationGas:   new(big.Int).SetBytes(ethcommon.Hex2Bytes("e685")),
// 		MaxFeePerGas:         big.NewInt(0),
// 		MaxPriorityFeePerGas: big.NewInt(0),
// 		PaymasterAndData:     []byte{},
// 		Signature:            ethcommon.Hex2Bytes("a9e93a16c3dc805c27a3041d29a9bc805dd78956aeeabeca2ead88400e83eda2d60e1f6ac4490e08dd72f7e0fa1365ae46be959978d3eadd740295547a5fe21d"),
// 		AuthData:             ethcommon.Hex2Bytes("49960de5880e8c687434170f6476605b8fe4aeb9a28632c7995cf3ba831d97630500000000"),
// 		ClientData:           ethcommon.Hex2Bytes("7b2274797065223a22776562617574686e2e676574222c226368616c6c656e6765223a22614d626c7255693236437871707842366a6d426d4d4a686275784b57547a415246496351784b484563354d222c226f726967696e223a22687474703a2f2f6c6f63616c686f73743a38303030222c2263726f73734f726967696e223a66616c73652c226f746865725f6b6579735f63616e5f62655f61646465645f68657265223a22646f206e6f7420636f6d7061726520636c69656e74446174614a534f4e20616761696e737420612074656d706c6174652e205365652068747470733a2f2f676f6f2e676c2f796162506578227d"),
// 	}
// 	entryPointAddr := types.NewAddressByStr(common.EntryPointContractAddr)
// 	userOpHash := interfaces.GetUserOpHash(userOp, entryPointAddr.ETHAddress(), big.NewInt(1356))
// 	t.Logf("user op hash: %x", userOpHash.Bytes())

// 	// Signature（secp256r1）
// 	t.Logf("userOp: %+v", userOp)

// 	pub := ethcommon.Hex2Bytes("a5010203262001215820d60c30d70d20f157137415303627fe1d0bbc5ccd13a4edf5a67151fc87b4f37522582086a8c83194750eb38ab4341b268958b574d62ea130bcd2f53868e844c5099b0e")

// 	var pk webauthncose.EC2PublicKeyData
// 	err := webauthncbor.Unmarshal(pub, &pk)
// 	assert.Nil(t, err)
// 	t.Logf("pk: %+v\n", pk)

// 	passkey := &Passkey{
// 		PubKeyX: new(big.Int).SetBytes(pk.XCoord),
// 		PubKeyY: new(big.Int).SetBytes(pk.YCoord),
// 		Algo:    AlgoSecp256R1,
// 	}

// 	validation, err := passkey.Validate(userOp, userOpHash)
// 	assert.Nil(t, err)
// 	assert.Equal(t, int(interfaces.SigValidationSucceeded), int(validation.SigValidation))
// }
