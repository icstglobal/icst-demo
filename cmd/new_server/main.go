package main

import (
	"flag"
	"context"
	"strconv"
	"crypto/md5"
	"encoding/hex"
	// "encoding/json"
	"fmt"

	"github.com/kataras/golog"

	"github.com/icstglobal/go-icst/transaction"
	// "github.com/icstglobal/icst-demo/domain"

	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/sessions"

	"github.com/icstglobal/go-icst/chain"

	"github.com/kataras/iris"
	"github.com/icstglobal/go-icst/wallets"
)

const (
	url string = "http://119.28.248.204:8549"
)


type Msg struct{
	Status string `json:"status"`
	MsgStr string `json:"msgstr"`
	Data interface{} `json:"data"`
}

func returnData(status string, data interface{}, msgstr string) interface{}{
	msg := Msg{
		Status: status,	
		MsgStr: msgstr,
		Data: data,
	}

	return msg
}

var helpInfo = "help\n  -h      		帮助\n  -c conf/conf.toml	配置文件路径(必选)，默认conf/conf.toml\n"
var cmdConf = flag.String("c", "", "配置文件路径")
var cmdHelp = flag.Bool("h", false, "帮助")


func main() {
	app := iris.New()
	app.Use(logger.New())
	app.Logger().SetLevel(golog.Levels[golog.DebugLevel].Name)
	log := app.Logger()
	// flog
	confFile := "" // 默认conf/conf.toml

	//解析命令行标志
	flag.Parse() // Scans the arg list and sets up flags
	log.Infof("server start begin")
	if *cmdConf != "" {
		confFile = *cmdConf
	} else if !*cmdHelp {
		fmt.Printf(helpInfo)
		return
	}
	if *cmdHelp {
		fmt.Printf(helpInfo)
		return
	}

	templateDir := "../../views"
	app.RegisterView(iris.HTML(templateDir, ".html").Reload(true))

	session := sessions.New(sessions.Config{})
	// init blc config

	app.Get("/", func(ctx iris.Context) {
		ctx.View("index.html")
	})

	app.Get("/contract/create", func(ctx iris.Context) {
		ctx.View("create.html")
	})

	app.Post("/contract/create", func(ctx iris.Context) {
		app.Logger().Debugf("create0")

		// chainType, err := strconv.Atoi(ctx.PostValue("chainType"))
		Producer := ctx.PostValue("Producer")
		// Consumer := ctx.PostValue("Consumer")
		Content := ctx.PostValue("Content")
		Platform := ctx.PostValue("Platform")
		Price, err := ctx.PostValueInt("Price")
		Ratio, err := ctx.PostValueInt("Ratio")
		if err != nil{
			app.Logger().Error("failed to get chainType,", err)
		}

		producerAddr, _ := hex.DecodeString(Producer)
		platformAddr, _ := hex.DecodeString(Platform)
		// consumerAddr, _ := hex.DecodeString(sc.Consumer)
		hash := md5.Sum([]byte(Content))

		contractData := make(map[string]interface{})
		contractData["PPublisher"] = producerAddr
		contractData["PPlatform"] = platformAddr
		contractData["PHash"] = string(hash[:])
		contractData["PPrice"] = uint32(Price)
		contractData["PRatio"] = uint8(Ratio)
		app.Logger().Debugf("contractData:%v", contractData)
		w := session.Start(ctx).Get("wallet").(*wallets.Wallet)
		a := session.Start(ctx).Get("account").(*wallets.Account)

		trans, err := w.CreateContentContractTrans(context.Background(), a, contractData)
		if err != nil {
			app.Logger().Error("failed to create contract:", err)
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.Text("failed to create contract")
			return
		}

		//save transaction unconfimred into session
		ses := session.Start(ctx)
		ses.Set("trans", trans)

		transHash := hex.EncodeToString(trans.Hash())
		data := map[string]interface{}{
			"transHash": transHash,
		}
		ctx.JSON(returnData("success", data, "create succeed."))
	})

	app.Post("/contract/sign", func(ctx iris.Context) {
		sigHex := ctx.PostValue("sig")
		trans := session.Start(ctx).Get("trans").(*transaction.ContractTransaction)
		a := session.Start(ctx).Get("account").(*wallets.Account)
		w := session.Start(ctx).Get("wallet").(*wallets.Wallet)

		app.Logger().Debug("sign trans.rawTrans before", trans.Hex())
		err := w.AfterSign(context.Background(), a, sigHex, trans)
		if err != nil{
			app.Logger().Error("failed to send transaction,", err)
			ctx.StatusCode(iris.StatusInternalServerError)
			return
		}
		app.Logger().Debug("sign trans.rawTrans after", trans.Hex())
		data := map[string]interface{}{
			"transHash": trans.Hex(),
			"contractAddr": hex.EncodeToString(trans.ContractAddr),
		}
		ctx.JSON(returnData("success", data, "sign succeed."))
	})

	app.Get("/contract/view", func(ctx iris.Context) {
		ctx.ViewData("addr", ctx.URLParam("addr"))
		ctx.View("view.html")
	})

	app.Post("/login", func(ctx iris.Context) {
		app.Logger().Debugf("login0", ctx.Request().Cookies())
		walletID := ctx.PostValue("ID")

		res := session.Start(ctx).Get("wallet")
		app.Logger().Debugf("login", res)

		if res != nil{
			w := res.(*wallets.Wallet)
			ctx.JSON(map[string]interface{}{
				"ID": w.ID,
				"msg": "user has logined in.",
			})
			return
		}
		w := &wallets.Wallet{ID:walletID}
		err := w.Init(url, []int{int(chain.Eth), int(chain.EOS)}, confFile)  
		if err != nil {
			app.Logger().Error("init wallet error:", err)
		}
		accounts, err := w.GetAccounts(context.Background(), w.ID)
		if err != nil {
			app.Logger().Error("get accounts error:", err)
		}

		ses := session.Start(ctx)
		ses.Set("wallet", w)

		data := map[string]interface{}{
			"ID": w.ID,
			"accounts": accounts,
		}
		ctx.JSON(returnData("success", data, "Logged in."))
	})

	app.Post("/use", func(ctx iris.Context) {
		app.Logger().Debugf("use0")

		accountID := ctx.PostValue("accountID")
		// pubKey := ctx.PostValue("pubKey")
		// intChainType, err := strconv.Atoi(ctx.PostValue("chainType"))
		// chainType := chain.ChainType(intChainType)

		res := session.Start(ctx).Get("wallet")
		app.Logger().Debugf("login", res)

		if res == nil{
			ctx.JSON(returnData("fail", nil, "not login"))
			return
		}
		w := res.(*wallets.Wallet)

		a, err := w.UseAccount(context.Background(), accountID)
		if err != nil{
			app.Logger().Error(err)
		}
		ses := session.Start(ctx)
		ses.Set("account", a)


		data := map[string]interface{}{
			"accountID": accountID,
			"msg": "use account:" + accountID,
		}
		ctx.JSON(returnData("success", data, ""))
	})

	app.Post("/import", func(ctx iris.Context) {
		app.Logger().Debugf("import0")

		pubKey := ctx.PostValue("pubKey")
		intChainType, err := strconv.Atoi(ctx.PostValue("chainType"))
		chainType := chain.ChainType(intChainType)

		res := session.Start(ctx).Get("wallet")
		app.Logger().Debugf("import", res)

		if res == nil{
			ctx.JSON(returnData("fail", nil, "not login"))
			return
		}
		w := res.(*wallets.Wallet)

		if w.IsExistAccount(context.Background(), pubKey, chainType){
		}

		data, err := w.SetAccount(context.Background(), w.ID, pubKey, chainType)
		if err != nil {
			ctx.JSON(returnData("fail", nil, err.Error()))
		}

		ctx.JSON(returnData("success", data, ""))
	})

	app.Run(iris.Addr(":8000"))
}
