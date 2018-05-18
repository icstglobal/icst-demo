package main

import (
	"context"
	"encoding/hex"

	"github.com/icstglobal/go-icst/transaction"
	"github.com/icstglobal/icst-demo/domain"

	"github.com/kataras/iris/sessions"

	"github.com/kataras/iris/middleware/logger"

	"github.com/icstglobal/go-icst/chain"
	"github.com/kataras/iris"
)

const url string = "http://:8545"

var blc chain.Chain

func main() {
	app := iris.New()
	app.Use(logger.New())
	app.RegisterView(iris.HTML("./views", ".html"))

	session := sessions.New(sessions.Config{})

	go func() {
		var err error
		blc, err = chain.DialEthereum(url)
		if err != nil {
			app.Logger().Error("connecto to blockchain failed:", err)
		}
	}()

	app.Get("/", func(ctx iris.Context) {
		ctx.View("index.html")
	})

	app.Post("/contract/create", func(ctx iris.Context) {
		sc := &domain.SkillContract{}
		if err := ctx.ReadForm(sc); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.Text("failed parse request form")
			return
		}

		contractData := make(map[string]interface{})
		consumer := []byte(sc.Consumer)
		trans, err := blc.NewContract(context.Background(), consumer, "Skill", contractData)
		if err != nil {
			app.Logger().Error("failed to create contract:", err)
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.Text("failed to create contract")
			return
		}
		//save transaction unconfimred into session
		ses := session.Start(ctx)
		ses.Set("trans", trans)

		//return the hash for signature to the client
		ctx.Write(trans.Hash())
	})

	app.Post("/contract/sign", func(ctx iris.Context) {
		sig := []byte(ctx.PostValue("sig"))
		trans := session.Start(ctx).Get("trans").(*transaction.ContractTransaction)
		var err error
		if err = blc.ConfirmTrans(context.TODO(), trans, sig); err != nil {
			app.Logger().Error("failed to send transaction,", err)
			ctx.StatusCode(iris.StatusInternalServerError)
			return
		}

		if err = blc.WaitMined(context.TODO(), trans.RawTx()); err != nil {
			app.Logger().Error(err)
			ctx.StatusCode(iris.StatusInternalServerError)
			return
		}

		//return contract address in hex string to the client
		ctx.Text(hex.EncodeToString(trans.ContractAddr))
		ctx.StatusCode(iris.StatusOK)
	})

	app.Run(iris.Addr(":8000"))
}
