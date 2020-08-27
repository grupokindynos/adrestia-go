package main

import (
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/explorer"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println(err)
	}
}

/* func TestGetAddress(t *testing.T) {
	oboli := obol.ObolRequest{ObolURL: os.Getenv("OBOL_PRODUCTION_URL")}
	plutus := services.PlutusRequests{Obol: &oboli, PlutusURL: os.Getenv("PLUTUS_PRODUCTION_URL")}
	fmt.Println(plutus.GetAddress("dash"))
}

func TestSendToExchange(t *testing.T) {
	oboli := obol.ObolRequest{ObolURL: os.Getenv("OBOL_PRODUCTION_URL")}
	plutus := services.PlutusRequests{Obol: &oboli, PlutusURL: os.Getenv("PLUTUS_PRODUCTION_URL")}
	res, err := plutus.WithdrawToAddress(plutus2.SendAddressBodyReq{
		Address: "XuXVd7D4Ef8ZHEVKWmGhvMDvuXP4GEPkyM",
		Coin:    "DASH",
		Amount:  0.2,
	})
	if err != nil {
		fmt.Println("error", err)
		return
	}
	fmt.Println(res)
} */

/* func TestGetBalance(t *testing.T) {
	oboli := obol.ObolRequest{ObolURL: os.Getenv("OBOL_PRODUCTION_URL")}
	plutus := services.PlutusRequests{Obol: &oboli, PlutusURL: os.Getenv("PLUTUS_PRODUCTION_URL")}
	bal, err := plutus.GetWalletBalance("dash")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Balance", bal)
} */

/* func TestSendAllBalanceToExchange(t *testing.T) {
	address := "PHg666Ef8Zz32y8V2i4essNMBSsDwXfr1q"
	asset := "POLIS"

	oboli := obol.ObolRequest{ObolURL: os.Getenv("OBOL_PRODUCTION_URL")}
	plutus := services.PlutusRequests{Obol: &oboli, PlutusURL: os.Getenv("PLUTUS_PRODUCTION_URL")}
	bal, err := plutus.GetWalletBalance(asset)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Balance: ", bal)

	res, err := plutus.WithdrawToAddress(plutus2.SendAddressBodyReq{
		Address: address,
		Coin:    asset,
		Amount:  bal.Confirmed * 0.9999,
	})
	if err != nil {
		fmt.Println("error", err)
		return
	}
	fmt.Println(res)
}*/

func TestBlockbook(t *testing.T) {
	coin, _ := coinfactory.GetCoin("ETH")
	var blockExplorer explorer.BlockBook
	blockExplorer.Url = "https://" + coin.BlockchainInfo.ExternalSource
	res, _ := blockExplorer.GetEthAddress("0xaDaE31C0b1857A5c4B1b0e48128A22a6b11d8bdB")
	assert.Equal(t, res.Address, "0xaDaE31C0b1857A5c4B1b0e48128A22a6b11d8bdB")
}

/* func TestExchange(t *testing.T) {
	hr := services.HestiaRequests{HestiaURL: os.Getenv("HESTIA_LOCAL_URL")}
	exchange, _ := hr.GetExchange("southxchange")
	ex := exchanges.NewSouthXchange(models.ExchangeParams{
		Name: "southxchange",
		Keys: hestia.ApiKeys{
			PublicKey: exchange.Accounts[0].PublicKey,
			PrivateKey: exchange.Accounts[0].PrivateKey,
		},
	})
	buy := hestia.Trade{
		OrderId:  "",
		Amount:   18,
		FromCoin: "TUSD",
		ToCoin:   "BTC",
		Symbol:   "BTCTUSD",
		Side:     "buy",
	}

	sell := hestia.Trade{
		OrderId:  "",
		Amount:   0.002,
		FromCoin: "BTC",
		ToCoin:   "TUSD",
		Symbol:   "BTCTUSD",
		Side:     "sell",
	}
	log.Println(buy.Amount)
	txId, err := ex.SellAtMarketPrice(sell)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(txId)
} */
