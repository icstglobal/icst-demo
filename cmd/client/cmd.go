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
	"github.com/mitchellh/mapstructure"
	// "github.com/ethereum/go-ethereum/cmd/geth/accountcmd"
)


var m map[string]interface{}
var w Wallet
var URL string
var httpClient http.Client

func Init() {
	m = map[string]interface{}{
		"accounts": accounts,
		// "localAccounts": localAccounts,
		"login": login,
		"useAccount": useAccount,
		"importAccount": importAccount,
		"create": create,
	}
	URL = "http://127.0.0.1:8000"

	jar, _ := cookiejar.New(nil)
    fmt.Println("Start Request Server")
    httpClient = http.Client{
        Jar: jar,
    }

	w = Wallet{
		ID: "12345", 
		Accounts:[]IAccount{},
	}

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

func localAccounts(){
	i := 1
	for _, account := range(w.Accounts){
		fmt.Printf("%d>", i)
		account.Info()
		i = i + 1
	}
}

func accounts(){
	if !w.Login {
		print("not login!\n")
		return
	}
	i := 1
	for _, account := range(w.Accounts){
		id := account.GetID()
		if len(id) > 0 {
			fmt.Printf("%d> %s imported\n", i, id)
		} else {
			fmt.Printf("%d> %s\n", i, account.GetAddr())
		}
		i = i + 1
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
	w.Login = true

	fmt.Printf("user %s logged in. \n", resData["ID"])

	for _, account := range(resData["accounts"].([]interface{})){
		ma := account.(map[string]interface{})

		a := &RemoteAccount{}
		err = mapstructure.Decode(ma, &a)
		if err != nil {
			fmt.Println(err)
		}
		w.UpdateAccount(a)

		// fmt.Printf("account %s, %s \n", i, a)
	}
	accounts()
}


func useAccount(param string) string{
	if !w.Login {
		print("not login!\n")
		return ""
	}

	index, err := strconv.Atoi(param)
	if err != nil || (index > len(w.Accounts) && index < 1){
		print("please input a correct index!\n")
		return ""
	}

	a := w.Accounts[index-1]

	if len(a.GetID()) == 0 {
		print("it's not imported account\n")
		return ""
	}

	data := map[string]string{
		"accountID": a.GetID(),
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

func importAccount(param string){
	if !w.Login {
		print("not login!\n")
		return
	}
	index, err := strconv.Atoi(param)
	if err != nil || (index > len(w.Accounts) && index < 1){
		print("please input a correct index!\n")
		return
	}
	a := w.Accounts[index-1]

	data := map[string]string{
		"chainType": strconv.Itoa(a.GetChainType()),
		"pubKey": a.GetPubKey(),
	}
	res, err := doRequest(URL + "/import", data)
	if err != nil{
		fmt.Println(err.Error())
		return
	}
	// fmt.Println(res)
	ra := &RemoteAccount{}
	err = mapstructure.Decode(res["data"], &a)
	if err != nil {
		fmt.Println(err)
	}
	w.UpdateAccount(ra)
	accounts()
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
	fmt.Printf("sign res: contractAddr(%s)\n transHash(%s)\n", resData["contractAddr"], resData["transHash"])
}

func create(){
	if !checkCurAccount(){
		return
	}
	data := map[string]string{
		"chainType": "0",
		"Producer": "85d6e595a3e64d3353b888bc49ee27f1b9f2a656",
		"Consumer": "0",
		"Content": "0",
		"Platform": "571ee16fce50d1be5bf11e54cc2a2036f6c31046",
		"Price": "10000",
		"Ratio": "20",
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

