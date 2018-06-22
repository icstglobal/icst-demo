package main

import (
	"fmt"
	"strconv"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"log"
	"strings"
	"encoding/json"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	// "github.com/ethereum/go-ethereum/cmd/geth/accountcmd"
)


var m map[string]interface{}
var w Wallet
var URL string
var httpClient http.Client

func Init() {
	m = map[string]interface{}{
		"accounts": accounts,
		"login": login,
		"use": use,
		"create": create,
	}
	URL = "http://127.0.0.1:8000"

	jar, _ := cookiejar.New(nil)
    fmt.Println("Start Request Server")
    httpClient = http.Client{
        Jar: jar,
    }

	w = Wallet{ID: "12345", Accounts:[]IAccount{}}

	keystoreDir := "./keystore"
	files, err := ioutil.ReadDir(keystoreDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		path := keystoreDir + "/" + file.Name()
		fmt.Println(path)
		jsonStr, err := ioutil.ReadFile(path)
		pwd := "wzp"
		key, err := keystore.DecryptKey(jsonStr, pwd)

		if err != nil {
			log.Fatal(err)
		}
		a := &EthAccount{Key:key}
		w.Accounts = append(w.Accounts, a)
	}
}

func accounts(){
	for i, account := range(w.Accounts){
		fmt.Printf("%d>", i+1)
		account.Info()
	}
}


func login(){
	data := map[string]string{
		"ID": "12345",
	}
	res, err := doRequest(URL + "/login", data)
	if err != nil{
		fmt.Println(err.Error())
	}
	resData := res["data"].(map[string]interface{})
	fmt.Printf("user %s logged in.\n", resData["ID"])
}


func use(param string) string{
	index, err := strconv.Atoi(param)
	if err != nil || (index > len(w.Accounts) && index < 1){
		print("please input a correct index!\n")
		return ""
	}
	a := w.Accounts[index-1]

	data := map[string]string{
		"chainType": strconv.Itoa(a.GetChainType()),
		"accountID": a.GetID(),
		"pubKey": a.GetPubKey(),
	}
	res, err := doRequest(URL + "/use", data)
	if err != nil{
		fmt.Println(err.Error())
		return ""
	}
	w.CurAccount = a
	resData := res["data"].(map[string]interface{})
	return resData["accountID"].(string)
}


func checkCurAccount() bool{
	if w.CurAccount == nil{
		print("please choose a account.")
		return false
	}
	return true
}

func sign(hexStr string){
	str, err := w.CurAccount.Sign(hexStr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("sig:\n", str)

	data := map[string]string{
		"sig": str,
	}
	res, err := doRequest(URL + "/contract/sign", data)
	if err != nil{
		fmt.Println(err.Error())
		return
	}
	resData := res["data"].(map[string]interface{})
	fmt.Printf("sign res: contractAddr(%s)\n", resData["contractAddr"])
}

func create(){
	if !checkCurAccount(){
		return
	}
	data := map[string]string{
		"chainType": "0",
		"Producer": "0",
		"Consumer": "0",
		"Content": "0",
		"Platform": "0",
		"Price": "0",
		"Ratio": "0",
	}

	res, err := doRequest(URL + "/contract/create", data)
	if err != nil{
		fmt.Println(err.Error())
		return
	}

	resData := res["data"].(map[string]interface{})
	fmt.Printf("create res: transHash(%s)\n", resData["transHash"])

	transHash := resData["transHash"].(string)
	sign(transHash)
}


func doRequest(url string, data map[string]string)(map[string]interface{}, error){
	var r http.Request
	r.ParseForm()
	for k, v := range(data){
		r.Form.Add(k, v)
	}
	req, _ := http.NewRequest("POST", url, strings.NewReader(r.Form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, _ := httpClient.Do(req)
	// fmt.Printf("cookies %s \n jar %v\n", resp.Cookies(), httpClient.Jar)
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		var dat map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &dat); err != nil {
			panic(err)
		}
		if dat["status"] == "fail" || dat["status"] == "error"{
			return nil, fmt.Errorf(dat["msg"].(string))
		}
		return dat, err
	}
	return nil, fmt.Errorf("http request error code:%s", resp.StatusCode) 
}

