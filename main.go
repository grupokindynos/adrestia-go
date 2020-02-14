package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/processor"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/obol"
	"github.com/joho/godotenv"
)

const fiatThreshold = 2.00 // USD // 2.0 for Testing, 10 USD for production
const orderTimeOut = 2 * time.Hour
const exConfirmationThreshold = 10
const walletConfirmationThreshold = 3
const testingAmount = 0.00001

type CurrentTime struct {
	Hour   int
	Day    int
	Minute int
	Second int
}

var currTime CurrentTime
var hestiaEnv string
var plutusEnv string

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	// Read input flag
	localRun := flag.Bool("local", false, "set this flag to run adrestia with local db")
	flag.Parse()

	// If flag was set, change the hestia request url to be local
	if *localRun {
		hestiaEnv = "HESTIA_LOCAL_URL"
		plutusEnv = "PLUTUS_LOCAL_URL"
	} else {
		hestiaEnv = "HESTIA_PRODUCTION_URL"
		plutusEnv = "PLUTUS_PRODUCTION_URL"
	}

	obolService := obol.ObolRequest{ObolURL: os.Getenv("OBOL_URL")}
	factoryParams := exchanges.Params{
		Obol: &obolService,
	}
	params := exchanges.Params{
		Plutus:          &services.PlutusRequests{Obol: &obolService, PlutusURL: os.Getenv(plutusEnv)},
		Hestia:          &services.HestiaRequests{HestiaURL: os.Getenv(hestiaEnv)},
		Obol:            &obolService,
		ExchangeFactory: exchanges.NewExchangeFactory(factoryParams),
	}
	processor.InitProcessor(params)
	processor.Start()

	timer()
}

func timer() {
	//ticker := time.NewTicker(10*time.Second)
	ticker := time.NewTicker(2 * time.Minute)
	//ticker := time.NewTicker(10 * time.Second)
	go func() {
		for _ = range ticker.C {
			processor.Start()
		}
	}()

	forever()
}

func forever() {
	for {
		time.Sleep(24 * time.Hour)
	}
}
