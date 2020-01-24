package main

import (
	"flag"
	"log"
	"os"
	"sync"
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
	} else {
		hestiaEnv = "HESTIA_PRODUCTION_URL"
	}

	obolService := obol.ObolRequest{ObolURL: os.Getenv("OBOL_URL")}
	factoryParams := exchanges.Params{
		Obol: &obolService,
	}
	params := exchanges.Params{
		Plutus:          &services.PlutusRequests{Obol: &obolService, PlutusURL: os.Getenv("PLUTUS_URL")},
		Hestia:          &services.HestiaRequests{HestiaURL: os.Getenv(hestiaEnv)},
		Obol:            &obolService,
		ExchangeFactory: exchanges.NewExchangeFactory(factoryParams),
	}
	processor.InitProcessor(params)
	processor.Start()
	// go timer()
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
	var wg sync.WaitGroup
	wg.Add(1)
	go runCronMinutes(1440, processor.Start, &wg) // 24 hrs
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
