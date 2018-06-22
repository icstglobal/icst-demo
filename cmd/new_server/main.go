package main

import (
	"context"
	"strconv"
	"crypto/md5"
	"encoding/hex"
	// "fmt"

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
	url string = "http://119.28.248.204:8541"
)

func returnData(status string, data map[string]interface{}, msg string) map[string]interface{}{
	res := map[string]interface{}{
		"status": status,
		"msg": msg,
	}
	res["data"] = data
	return res
}


func main() {
	app := iris.New()
	app.Use(logger.New())
	app.Logger().SetLevel(golog.Levels[golog.DebugLevel].Name)
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
		Price, err := strconv.Atoi(ctx.PostValue("Platform"))
		Ratio, err := strconv.Atoi(ctx.PostValue("Ratio"))
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

		err := w.AfterSign(context.Background(), a, sigHex, trans)
		if err != nil{
			app.Logger().Error("failed to send transaction,", err)
			ctx.StatusCode(iris.StatusInternalServerError)
			return
		}
		data := map[string]interface{}{
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
		// ethAccountID := ctx.PostValue("ethAccountID")
		// ethPubKey := ctx.PostValue("ethPubKey")
		// ethChainType := chain.Eth

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
		err := w.Init(url, []int{int(chain.Eth), int(chain.EOS)})  
		if err != nil {
			app.Logger().Error("init wallet error:", err)
		}
		// w.SetAccount(context.Background(), ethAccountID, ethPubKey, ethChainType)

		ses := session.Start(ctx)
		ses.Set("wallet", w)

		data := map[string]interface{}{
			"ID": w.ID,
		}
		ctx.JSON(returnData("success", data, "Logged in."))
	})

	app.Post("/use", func(ctx iris.Context) {
		app.Logger().Debugf("use0")

		accountID := ctx.PostValue("accountID")
		pubKey := ctx.PostValue("pubKey")
		intChainType, err := strconv.Atoi(ctx.PostValue("chainType"))
		chainType := chain.ChainType(intChainType)

		res := session.Start(ctx).Get("wallet")
		app.Logger().Debugf("login", res)

		if res == nil{
			ctx.JSON(returnData("fail", nil, "not login"))
			return
		}
		w := res.(*wallets.Wallet)

		if !w.IsExistAccount(context.Background(), accountID){
			w.SetAccount(context.Background(), accountID, pubKey, chainType)
		}
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

	app.Run(iris.Addr(":8000"))
}
