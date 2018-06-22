package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	// "strings"
)

type IAccount interface {
	Info()
	GetPubKey() string
	GetID() string
	GetChainType()(int)
	Sign(hexStr string) (string, error)
}

type EthAccount struct {
	Key *keystore.Key
	ChainType int
}

type Wallet struct {
	ID string  // wallet ID
	Accounts   []IAccount // all accounts in the wallet
	CurAccount IAccount   // current using account
}

func (a *EthAccount) GetChainType() int{
	return a.ChainType
}

func (a *EthAccount) GetID() string{
	return a.Key.Id.String()
}

func (a *EthAccount) GetPubKey() string{
	pubKey := a.Key.PrivateKey.PublicKey
	return base64.StdEncoding.EncodeToString(crypto.FromECDSAPub(&pubKey))
}

func (a *EthAccount) Info() {
	// key := a.Key
	fmt.Println("Id:", a.GetID())
	// fmt.Println("addr:\n", strings.ToLower(key.Address.Hex()))
	// fmt.Println("PublicKey base64: \n", a.GetPubKey())
}

func (a *EthAccount) Sign(hexStr string) (string, error) {
	hashData, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	sig, err := crypto.Sign(hashData, a.Key.PrivateKey)
	str := hex.EncodeToString(sig)
	fmt.Println("sig:\n", str)
	return str, nil
}
