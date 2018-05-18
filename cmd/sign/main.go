package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/ethereum/go-ethereum/accounts/keystore"

	"github.com/ethereum/go-ethereum/crypto"
)

var input = flag.String("input", "", "string to be signed, must be exactly 32 bytes")
var keyfile = flag.String("keyfile", "", "file path of the key, can not be empty")
var pwd = flag.String("pwd", "", "password to decrypt the key file, can not be empty")

//to sign an input string with given key file
func main() {
	flag.Parse()

	if len(*input) == 0 || len(*keyfile) == 0 || len(*pwd) == 0 {
		flag.Usage()
		return
	}

	buf, err := ioutil.ReadFile(*keyfile)
	if err != nil {
		fmt.Println("failed to read key file", err)
		return
	}
	key, err := keystore.DecryptKey(buf, *pwd)
	if err != nil {
		fmt.Println("failed to decrypt key file, please check the password", err)
		return
	}
	sig, err := crypto.Sign([]byte(*input), key.PrivateKey)
	if err != nil {
		fmt.Println("failed to sign input", err)
		return
	}
	fmt.Println("sig is:", hex.EncodeToString(sig))
}
