package repo

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/axiomesh/axiom-kit/fileutil"
)

func loadKey(repoRoot string, keyFileName string) (*ecdsa.PrivateKey, error) {
	keyPath := path.Join(repoRoot, keyFileName)
	if !fileutil.Exist(keyPath) {
		key, err := GenerateKey()
		if err != nil {
			return nil, err
		}
		if err := WriteKey(keyPath, key); err != nil {
			return nil, err
		}
		return key, nil
	}
	return ReadKey(keyPath)
}

// GenerateKey use secp256k1
func GenerateKey() (*ecdsa.PrivateKey, error) {
	return ethcrypto.GenerateKey()
}

func Libp2pKeyFromECDSAKey(sk *ecdsa.PrivateKey) (libp2pcrypto.PrivKey, error) {
	raw := ethcrypto.FromECDSA(sk)
	return libp2pcrypto.UnmarshalSecp256k1PrivateKey(raw)
}

func Libp2pIDToPubKey(id string) (libp2pcrypto.PubKey, error) {
	p2pID, err := peer.Decode(id)
	if err != nil {
		return nil, err
	}

	return p2pID.ExtractPublicKey()
}

func KeyToNodeID(sk *ecdsa.PrivateKey) (string, error) {
	pk, err := Libp2pKeyFromECDSAKey(sk)
	if err != nil {
		return "", err
	}

	id, err := peer.IDFromPublicKey(pk.GetPublic())
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func KeyString(key *ecdsa.PrivateKey) string {
	return hex.EncodeToString(ethcrypto.FromECDSA(key))
}

func WriteKey(keyPath string, key *ecdsa.PrivateKey) error {
	return os.WriteFile(keyPath, []byte(KeyString(key)), 0600)
}

func ParseKey(keyBytes []byte) (*ecdsa.PrivateKey, error) {
	return ethcrypto.HexToECDSA(string(keyBytes))
}

func ReadKey(keyPath string) (*ecdsa.PrivateKey, error) {
	keyFile, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	return ParseKey(keyFile)
}

// GenerateCustomKeyJson generates a JSON key for authentication and store it in the given directory.
//
// It returns a string which is the address of the generated key.
func GenerateCustomKeyJson(auth string, keyDir string, privateKey *ecdsa.PrivateKey) (string, error) {
	var err error
	// if key is not exist, generate it
	if privateKey == nil {
		privateKey, err = ethcrypto.GenerateKey()
		if err != nil {
			return "", err
		}
	}
	// generate KeyJson
	ks := keystore.NewKeyStore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)
	account, err := ks.ImportECDSA(privateKey, auth)
	if err != nil {
		fmt.Println("Error creating keystore:", err)
		return "", err
	}
	// add custom field, such as nodeId
	nodeID, err := KeyToNodeID(privateKey)
	if err != nil {
		return "", err
	}
	err = addCustomFieldToKeyJson(account, nodeID)
	if err != nil {
		return "", err
	}
	if DefaultKeyJsonPassword == auth {
		fmt.Println("Warning: Using default keystore password [", DefaultKeyJsonPassword, "], please change it")
	}
	fmt.Println("Account generated:", account.Address.Hex())
	return account.Address.Hex(), nil
}

func addCustomFieldToKeyJson(account accounts.Account, nodeId string) error {
	keyJsonPath := account.URL.Path
	keyJsonData, err := os.ReadFile(keyJsonPath)
	if err != nil {
		return err
	}
	var keyJson map[string]interface{}
	if err := json.Unmarshal(keyJsonData, &keyJson); err != nil {
		return err
	}
	// Add the custom NodeId field
	keyJson[NodeP2PIdName] = nodeId
	modifiedKeyJsonData, err := json.Marshal(keyJson)
	if err != nil {
		return err
	}
	// Write the modified key file
	return os.WriteFile(keyJsonPath, modifiedKeyJsonData, 0600)
}

// GenerateP2PKeyJson generates a P2P key JSON file for the given authentication string, key directory, and private key.
//
// Returns:
// - string: The address of the generated key JSON file.
// - error: An error if any occurred during the generation process.
func GenerateP2PKeyJson(auth string, keyDir string, privateKey *ecdsa.PrivateKey) (string, error) {
	// default auth for dev and test
	if auth == "" {
		auth = DefaultKeyJsonPassword
	}
	addr, err := GenerateCustomKeyJson(auth, keyDir, privateKey)
	if err != nil {
		return "", err
	}
	// rename keystore file
	err = renameKeystoreFile(keyDir, addr, P2PKeyFileName)
	if err != nil {
		return "", err
	}
	return addr, nil
}

func FromKeyJson(auth string, keyPath string) (*ecdsa.PrivateKey, error) {
	// if file is not exist
	if !fileutil.Exist(keyPath) {
		_, err := GenerateP2PKeyJson(auth, filepath.Dir(keyPath), nil)
		if err != nil {
			return nil, err
		}
	}
	keyJSON, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	key, err := keystore.DecryptKey(keyJSON, auth)
	if err != nil {
		return nil, err
	}
	return key.PrivateKey, nil
}

func ChangeAuthOfKeyJson(oldAuth string, newAuth string, keyPath string) error {
	keyJSON, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}
	err = os.Remove(keyPath)
	if err != nil {
		return err
	}
	key, err := keystore.DecryptKey(keyJSON, oldAuth)
	if err != nil {
		return err
	}
	_, err = GenerateP2PKeyJson(newAuth, filepath.Dir(keyPath), key.PrivateKey)
	return err
}

func fromP2PKeyJson(auth string, repoRoot string) (*ecdsa.PrivateKey, error) {
	if auth == "" {
		auth = DefaultKeyJsonPassword
	}
	return FromKeyJson(auth, path.Join(repoRoot, P2PKeyFileName))
}

// renameKeystoreFile renames the keystore file for the specified account.
func renameKeystoreFile(keystoreDir string, accountAddress string, newName string) error {
	if strings.HasPrefix(accountAddress, "0x") {
		accountAddress = strings.ToLower(accountAddress[2:])
	}
	searchPattern := path.Join(keystoreDir, fmt.Sprintf("UTC--*--%s", accountAddress))
	files, err := filepath.Glob(searchPattern)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("no keystore file found for account %s", accountAddress)
	}

	// maybe there are multiple files, but we only need the first one
	keystoreFile := files[0]
	return os.Rename(keystoreFile, path.Join(keystoreDir, newName))
}

type CustomKeyJson struct {
	Address   string `json:"address"`
	NodeP2PId string `json:"node_p2p_id"`
}

func GetCustomKeyJson(keyPath string) (*CustomKeyJson, error) {
	keyFile, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	var keyJson CustomKeyJson
	if err := json.Unmarshal(keyFile, &keyJson); err != nil {
		return nil, err
	}
	return &keyJson, nil
}
