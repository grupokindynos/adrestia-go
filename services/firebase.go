package services

import (
	"context"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"firebase.google.com/go/db"
	"github.com/grupokindynos/adrestia-go/models/balance"
	"google.golang.org/api/option"
)

type Firebase struct {
	DB   *db.Client
	Auth *auth.Client
}

func InitFirebase() *Firebase {
	// service account credentials
	opt := option.WithCredentialsFile("./fb_conf.json")
	ctx := context.Background()

	config := &firebase.Config{
		DatabaseURL: "https://polispay-copay.firebaseio.com",
	}
	firebaseApp, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		panic(err)
	}

	dbClient, err := firebaseApp.Database(ctx)
	if err != nil {
		panic(err)
	}
	authClient, err := firebaseApp.Auth(ctx)
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
