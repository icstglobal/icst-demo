package main

import (
	"fmt"
	"strconv"
	"io/ioutil"
	"log"
	"strings"
	"encoding/hex"
	"encoding/base64"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/common"
	// "github.com/ethereum/go-ethereum/cmd/geth/accountcmd"
)

func cmdList(){
	lst := []string{"accounts", "new account", "send tran from 0 to 1", "sign data"}
	for i, cmd := range lst {
		i += 1
		_cmd := strconv.Itoa(i) + ":" + cmd
		fmt.Println(_cmd)
	}
}

func accounts()([]*keystore.Key){
	print("exec accounts \n")
	// main.accountList(nil)
	keystoreDir := "./keystore"
	files, err := ioutil.ReadDir(keystoreDir)
	fmt.Println(files)
	if err != nil {
		log.Fatal(err)
	}

	var keys []*keystore.Key;

	for _, file := range files {
		fmt.Println(keystoreDir + file.Name())
		jsonStr, err := ioutil.ReadFile(keystoreDir + "/" + file.Name())
		// fmt.Println(jsonStr, "jsonStr")
		pwd := "wzp"
		key, err := keystore.DecryptKey(jsonStr, pwd)
		fmt.Println("")
		fmt.Println("Id:\n", key.Id)
		fmt.Println("addr:\n", strings.ToLower(key.Address.Hex()))
		fmt.Println("PrivateKey struct:\n", key.PrivateKey)
		fmt.Printf("PrivateKey hex: %x\n", crypto.FromECDSA(key.PrivateKey))
		fmt.Printf("PrivateKey hex: %x\n", key.PrivateKey.D.Bytes())
		fmt.Println("PrivateKey base64: \n", base64.StdEncoding.EncodeToString(crypto.FromECDSA(key.PrivateKey)))
		fmt.Println("PublicKey struct:\n", key.PrivateKey.PublicKey)
		fmt.Println("PublicKey hex: \n", common.ToHex(crypto.FromECDSAPub(&key.PrivateKey.PublicKey)))
		fmt.Println("PublicKey base64: \n", base64.StdEncoding.EncodeToString(crypto.FromECDSAPub(&key.PrivateKey.PublicKey)))
		if err != nil {
			log.Fatal(err)
		}
		// sig, err := crypto.Sign([]byte("19330186464252272190159761906888"), key.PrivateKey)
		// if err != nil {
			// log.Fatal(err)
		// }
		// str := hex.EncodeToString(sig)
		// fmt.Println("sig:\n", str)
		keys = append(keys, key)
	}
	return keys
}

func newAccount(){
	print("exec newAccount \n")

}

func tran0to1(){
	print("exec tran0to1 \n")
}

func sign1(){
	print("exec sign, please input data:\n")
	var hexStr string
	fmt.Scanln(&hexStr)
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		log.Fatal(err)
	}
	keys := accounts()
	sig, err := crypto.Sign(data, keys[0].PrivateKey)
	str := hex.EncodeToString(sig)
	fmt.Println("sig:\n", str)

}


func main1() {
	cmdList()
	m := map[int]interface{}{
		1: accounts,
		2: newAccount,
		3: tran0to1,
		4: sign1,
	}
	for {
		var input string
		fmt.Scanln(&input)
		intInput, err := strconv.Atoi(input)
		if err != nil{
			continue
		}
		// m[intInput].(func())()
		fun := m[intInput]
		switch fun.(type) {
			case func():
				fun.(func())()
			case func()([]*keystore.Key):
				fun.(func()[]*keystore.Key)()
		}

	}
}
