package ethnode

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ConnNode connects to a node, simple pimple. client can be used asyncronously.
func Connect(node string) (client *ethclient.Client, err error) {
	client, err = ethclient.Dial(node)
	if err != nil {
		fmt.Println("Could not connect to node:", node, err)
		return client, err
	}
	return client, nil
}

// AuthInfo holds everything needed to sign the first transaction. Be sure to add
// to the nonce in the main func rather than adding to this one
type AuthInfo struct {
	PubKey       *ecdsa.PublicKey
	PrivKey      *ecdsa.PrivateKey
	InitNonce    uint64
	InitGasPrice *big.Int
	Address      common.Address
}

func GenerateAuthInfo(client *ethclient.Client, keyName string) (authInfo *AuthInfo, err error) {
	emptyAuth := &AuthInfo{}
	privKey, err := crypto.HexToECDSA(os.Getenv(keyName))
	if err != nil {
		fmt.Println("Could not find privKey")
		return emptyAuth, err
	}
	publicKey := privKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		fmt.Println("couldn't get the public key from the private!")
		return emptyAuth, err
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return emptyAuth, err
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return emptyAuth, err
	}
	authInfo = &AuthInfo{
		PubKey:       publicKeyECDSA,
		PrivKey:      privKey,
		InitNonce:    nonce,
		InitGasPrice: gasPrice,
		Address:      fromAddress,
	}
	return authInfo, nil
}
