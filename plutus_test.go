package main

import (
	"fmt"
	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/blockbook"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	plutus2 "github.com/grupokindynos/common/plutus"
	"github.com/joho/godotenv"
	"log"
	"os"
	"testing"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println(err)
	}
}

func TestSendToExchange(t *testing.T) {
	oboli := obol.ObolRequest{ObolURL: os.Getenv("OBOL_PRODUCTION_URL")}
	plutus := services.PlutusRequests{Obol: &oboli, PlutusURL: os.Getenv("PLUTUS_LOCAL_URL")}
	res, err := plutus.WithdrawToAddress(plutus2.SendAddressBodyReq{
		Address: "0x82ed4a3CAA3A67F56FdbB225002E204D332457CF",
		Coin:    "USDT",
		Amount:  2500,
	})
	if err != nil {
		fmt.Println("error", err)
		return
	}
	fmt.Println(res)
}


func TestGetBalance(t *testing.T) {
	oboli := obol.ObolRequest{ObolURL: os.Getenv("OBOL_PRODUCTION_URL")}
	plutus := services.PlutusRequests{Obol: &oboli, PlutusURL: os.Getenv("PLUTUS_PRODUCTION_URL")}
	bal, err := plutus.GetWalletBalance("dash")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(bal)
}

func TestGetAddress(t *testing.T) {
	oboli := obol.ObolRequest{ObolURL: os.Getenv("OBOL_PRODUCTION_URL")}
	plutus := services.PlutusRequests{Obol: &oboli, PlutusURL: os.Getenv("PLUTUS_PRODUCTION_URL")}
	fmt.Println(plutus.GetAddress("polis"))
}

func TestBlockbook(t *testing.T) {
	coin, _ := coinfactory.GetCoin("ETH")
	var blockExplorer blockbook.BlockBook
	blockExplorer.Url = "https://" + coin.BlockchainInfo.ExternalSource
	res, err := blockExplorer.GetEthAddress("0xaDaE31C0b1857A5c4B1b0e48128A22a6b11d8bdB")
	log.Println(err)
	//res, _ := blockExplorer.GetTx("0x475f5d6f71aec76c4f112a0902c7da506f0324504cf521c4fe00dbbabdac2a16")
	fmt.Printf("%+v\n", res)
}

func TestExchange(t *testing.T) {
	hr := services.HestiaRequests{HestiaURL:os.Getenv("HESTIA_LOCAL_URL")}
	exchange, _ := hr.GetExchange("southxchange")
	ex := exchanges.NewSouthXchange(exchange)
	buy := hestia.Trade{
		OrderId:        "",
		Amount:         18,
		FromCoin:       "TUSD",
		ToCoin:         "BTC",
		Symbol:         "BTCTUSD",
		Side:           "buy",
	}

	sell := hestia.Trade{
		OrderId:        "",
		Amount:         0.002,
		FromCoin:       "BTC",
		ToCoin:         "TUSD",
		Symbol:         "BTCTUSD",
		Side:           "sell",
	}
	log.Println(buy.Amount)
	txId, err := ex.SellAtMarketPrice(sell)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(txId)
}