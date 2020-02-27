package main

import (
	"flag"
	"log"
	"os"
	"sync"
	"time"

	"github.com/grupokindynos/adrestia-go/balancer"
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
var mainHestiaEnv string
var mainPlutusEnv string
var globalParams exchanges.Params

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
		mainHestiaEnv = "HESTIA_LOCAL_URL"
		mainPlutusEnv = "PLUTUS_LOCAL_URL"
	} else {
		mainHestiaEnv = "HESTIA_PRODUCTION_URL"
		mainPlutusEnv = "PLUTUS_PRODUCTION_URL"
	}

	obolService := obol.ObolRequest{ObolURL: os.Getenv("OBOL_URL")}
	factoryParams := exchanges.Params{
		Obol: &obolService,
	}
	globalParams = exchanges.Params{
		Plutus:          &services.PlutusRequests{Obol: &obolService, PlutusURL: os.Getenv(mainPlutusEnv)},
		Hestia:          &services.HestiaRequests{HestiaURL: os.Getenv(mainHestiaEnv)},
		Obol:            &obolService,
		ExchangeFactory: exchanges.NewExchangeFactory(factoryParams),
	}
	timer()
}

func timer() {
	for {
		time.Sleep(1 * time.Second)
		currTime = CurrentTime{
			Hour:   time.Now().Hour(),
			Day:    time.Now().Day(),
			Minute: time.Now().Minute(),
			Second: time.Now().Second(),
		}
		if currTime.Second == 0 {
			var wg sync.WaitGroup
			wg.Add(1)
			runCrons(&wg)
			wg.Wait()
		}
	}
}

func runCrons(mainWg *sync.WaitGroup) {
	defer func() {
		mainWg.Done()
	}()
	b := balancer.NewBalancer(globalParams)
	processor.InitProcessor(globalParams)
	var wg sync.WaitGroup
	wg.Add(2)
	go runCronMinutes(2160, processor.Start, &wg) // 1 day and half
	go runCronMinutes(10, b.StartBalancer, &wg) // 10 minutes
	wg.Wait()
}

func runCronMinutes(schedule int, function func(), wg *sync.WaitGroup) {
	go func() {
		defer func() {
			wg.Done()
		}()
		remainder := currTime.Minute % schedule
		if remainder == 0 {
			function()
		}
		return
	}()
}
