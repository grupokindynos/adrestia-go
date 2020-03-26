package main

import (
	"flag"
	"github.com/grupokindynos/adrestia-go/balancer"
	"github.com/grupokindynos/adrestia-go/processor"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/obol"
	"os"
	"time"
)

var (
	exchangesProcessor processor.ExchangesProcessor
	depositProcessor processor.DepositProcessor
	hwProcessor processor.HwProcessor
)

func runExchangesProcessor() {
	ticker := time.NewTicker(2 * time.Hour)
	for _ = range ticker.C {
		exchangesProcessor.Start()
	}
}

func runDepositProcessor() {
	ticker := time.NewTicker(5 * time.Minute)
	for _ = range ticker.C {
		depositProcessor.Start()
	}
}

func runHwProcessor() {
	ticker := time.NewTicker(10 * time.Minute)
	for _ = range ticker.C {
		hwProcessor.Start()
	}
}

func main() {
	var hestiaUrl string
	var plutusUrl string

	// Read input flag
	localRun := flag.Bool("local", false, "set this flag to run adrestia with local db")
	flag.Parse()

	// If flag was set, change the hestia request url to be local
	if *localRun {
		hestiaUrl = "HESTIA_LOCAL_URL"
		plutusUrl = "PLUTUS_LOCAL_URL"
	} else {
		hestiaUrl = "HESTIA_PRODUCTION_URL"
		plutusUrl = "PLUTUS_PRODUCTION_URL"
	}

	Obol := obol.ObolRequest{ObolURL: os.Getenv("OBOL_PRODUCTION_URL")}
	Hestia := services.HestiaInstance{HestiaURL: os.Getenv(hestiaUrl)}
	Plutus := services.PlutusInstance{PlutusURL: os.Getenv(plutusUrl), Obol: &Obol}

	Balancer := balancer.Balancer{
		Hestia: &Hestia,
		Plutus: &Plutus,
	}
	exchangesProcessor = processor.ExchangesProcessor{
		Obol:   &Obol,
		Hestia: &Hestia,
		Plutus: &Plutus,
	}
	depositProcessor = processor.DepositProcessor{
		Hestia: &Hestia,
		Plutus: &Plutus,
		Obol:   &Obol,
	}
	hwProcessor = processor.HwProcessor{
		Hestia:   &Hestia,
		Plutus:   &Plutus,
		Obol:     &Obol,
		Balancer: Balancer,
	}

	go runExchangesProcessor()
	go runDepositProcessor()
	go runHwProcessor()

	select {}
}
