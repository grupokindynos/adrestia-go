package main

import (
	"log"
	"os"
	"time"

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

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	hestiaService := services.HestiaRequests{}
	obolService := obol.ObolRequest{ObolURL: os.Getenv("OBOL_URL")}
	plutusService := services.PlutusRequests{Obol: &obolService}

	proc := processor.Processor{
		Hestia: &hestiaService,
		Plutus: plutusService,
		Obol: &obolService,
	}

	proc.Start()
}
