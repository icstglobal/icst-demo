package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"strings"
)

type IAccount interface {
	Info()
	GetPubKey() string
	GetID() string
	GetAddr() string
	GetChainType()(int)
	Sign(hexStr string) (string, error)
	SetID(accountID string)
}

type EthAccount struct {
	ID string
	Key *keystore.Key
	ChainType int
}

type RemoteAccount struct {
	ID string
	PubKey string
	ChainType int
}


type Wallet struct {
	ID string  // wallet ID
	Accounts   []IAccount // all accounts in the wallet
	CurAccount IAccount   // current using account
	Login bool
}


func (w *Wallet) GetAccount(accountID string) IAccount{
	for _, account :=range(w.Accounts){
		if account.GetID()== accountID{
			return account
		}
	}
	return nil
}

func (w *Wallet) UpdateAccount(a *RemoteAccount) {
	for _, account :=range(w.Accounts){
		if account.GetPubKey()== a.PubKey{
			account.SetID(a.ID)
		}
	}
}


func (a *EthAccount) GetChainType() int{
	return a.ChainType
}

func (a *EthAccount) GetID() string{
	return a.ID
}

func (a *EthAccount) SetID(accountID string) {
	a.ID = accountID
}

func (a *EthAccount) GetPubKey() string{
	pubKey := a.Key.PrivateKey.PublicKey
	// fmt.Println(base64.StdEncoding.EncodeToString(crypto.FromECDSAPub(&pubKey)))
	// fmt.Println(hex.EncodeToString(crypto.FromECDSAPub(&pubKey)))
	// fmt.Println(a.Key.PrivateKey)

	return base64.StdEncoding.EncodeToString(crypto.FromECDSAPub(&pubKey))
}

func (a *EthAccount) GetAddr() string{
	// key := a.Key
	// fmt.Println("Id:", a.GetID())
	return strings.ToLower(a.Key.Address.Hex())
}

func (a *EthAccount) Info() {
	// key := a.Key
	// fmt.Println("Id:", a.GetID())
	fmt.Println("addr:", strings.ToLower(a.Key.Address.Hex()))
	// fmt.Println("PublicKey base64: \n", a.GetPubKey())
}

func (a *EthAccount) Sign(hexStr string) (string, error) {
	hashData, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	sig, err := crypto.Sign(hashData, a.Key.PrivateKey)
	str := hex.EncodeToString(sig)
	// fmt.Println("sig:\n", str)
	return str, nil
}
