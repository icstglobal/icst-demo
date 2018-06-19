package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"github.com/kataras/golog"

	"github.com/icstglobal/go-icst/transaction"
	"github.com/icstglobal/icst-demo/domain"

	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/sessions"

	"github.com/icstglobal/go-icst/chain"
	"github.com/icstglobal/go-icst/chain/eth"
	"github.com/kataras/iris"
)

const url string = "http://119.28.248.204:8541"

var blc chain.Chain

func main() {
	app := iris.New()
	app.Use(logger.New())
	app.Logger().SetLevel(golog.Levels[golog.DebugLevel].Name)
	templateDir := "../../views"
	app.RegisterView(iris.HTML(templateDir, ".html").Reload(true))

	session := sessions.New(sessions.Config{})

	go func() {
		var err error
		blc, err = eth.DialEthereum(url)
		if err != nil {
			app.Logger().Error("connecto to blockchain failed:", err)
		}
	}()

	app.Get("/", func(ctx iris.Context) {
		ctx.View("index.html")
	})

	app.Get("/contract/create", func(ctx iris.Context) {
		ctx.View("create.html")
	})

	app.Post("/contract/create", func(ctx iris.Context) {
		sc := &domain.SkillContract{}
		if err := ctx.ReadForm(sc); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.Text(fmt.Sprintf("failed parse request form, err:%v", err))
			return
		}
		app.Logger().Debugf("SkillContract:%v", sc)

		producerAddr, _ := hex.DecodeString(sc.Producer)
		platformAddr, _ := hex.DecodeString(sc.Platform)
		consumerAddr, _ := hex.DecodeString(sc.Consumer)
		hash := md5.Sum([]byte(sc.Content))

		contractData := struct {
			PHash      string
			PPublisher []byte
			PPlatform  []byte
			PConsumer  []byte
			PPrice     uint32
			PRatio     uint8
		}{
			PHash:      string(hash[:]),
			PPublisher: producerAddr,
			PPlatform:  platformAddr,
			PConsumer:  consumerAddr,
			PPrice:     sc.Price,
			PRatio:     sc.Ratio,
		}
		app.Logger().Debugf("contractData:%v", contractData)
		trans, err := blc.NewContract(context.Background(), producerAddr, "Skill", contractData)
		if err != nil {
			app.Logger().Error("failed to create contract:", err)
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.Text("failed to create contract")
			return
		}
		//save transaction unconfimred into session
		ses := session.Start(ctx)
		ses.Set("trans", trans)

		ctx.ViewData("contract", sc)
		transHash := hex.EncodeToString(trans.Hash())
		ctx.ViewData("transHash", transHash)
		ctx.View("sign.html")
	})

	app.Post("/contract/sign", func(ctx iris.Context) {
		sigHex := ctx.PostValue("sig")
		sig, err := hex.DecodeString(sigHex)
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.Text("input sig is not a valid hex string")
			return
		}
		trans := session.Start(ctx).Get("trans").(*transaction.ContractTransaction)
		if err = blc.ConfirmTrans(context.TODO(), trans, sig); err != nil {
			app.Logger().Error("failed to send transaction,", err)
			ctx.StatusCode(iris.StatusInternalServerError)
			return
		}

		if err = blc.WaitMined(context.TODO(), trans); err != nil {
			app.Logger().Error(err)
			ctx.StatusCode(iris.StatusInternalServerError)
			return
		}

		ctx.Redirect("/contract/view?addr="+hex.EncodeToString(trans.ContractAddr), iris.StatusSeeOther)
	})

	app.Get("/contract/view", func(ctx iris.Context) {
		ctx.ViewData("addr", ctx.URLParam("addr"))
		ctx.View("view.html")
	})

	app.Run(iris.Addr(":8000"))
}
