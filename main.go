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
	"github.com/grupokindynos/adrestia-go/logger"
	"github.com/joho/godotenv"
)

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
var fileLog logger.FileLogger

func init() {
	var err error
	fileLog, err = logger.NewLogger("main_log", "main")
	if err != nil {
		log.Println("Couldn't initialize logger")
	}
	if err := godotenv.Load(); err != nil {
		fileLog.Println("No .env file found")
	}
}

func main() {
	fileLog.Println("Adrestia Started")

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
	fileLog.EndLogger()
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
	// Balancer is going to run everyday at 5am.
	// Digital Ocean time is upfront by 6 hours of our time, that's why it is going to run
	// every day at 11am, to compensate that difference.
	go runCronHour(11, b.StartBalancer, &wg)
	go runCronMinutes(1, processor.Start, &wg) // 10 minutes
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

func runCronHour(schedule int, function func(), wg *sync.WaitGroup) {
	go func() {
		defer func() {
			wg.Done()
		}()
		if currTime.Hour == schedule && currTime.Minute == 0 {
			function()
		}
		return
	}()
}
