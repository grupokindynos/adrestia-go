package services

import (
	"context"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"firebase.google.com/go/db"
	"github.com/grupokindynos/adrestia-go/models/balance"
)

type Firebase struct {
	DB   *db.Client
	Auth *auth.Client
}

func InitFirebase(fbApp *firebase.App) *Firebase {
	ctx := context.Background()
	dbClient, err := fbApp.Database(ctx)
	if err != nil {
		panic(err)
	}
	authClient, err := fbApp.Auth(ctx)
	if err != nil {
		panic(err)
	}

	fb := &Firebase{
		DB:   dbClient,
		Auth: authClient,
	}
	return fb
}

func (fb *Firebase) GetConf() (conf balance.MinBalanceConfResponse, err error) {
	ref := fb.DB.NewRef("/conf").Child("BalanceConf")
	err = ref.Get(context.Background(), &conf)
	if err != nil {
		return conf, err
	}
	return conf, nil
}


