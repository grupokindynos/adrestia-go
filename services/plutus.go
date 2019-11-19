package services

import (
	"encoding/json"
	"fmt"
	"github.com/gookit/color"
	"github.com/grupokindynos/adrestia-go/models/balance"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/jwt"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/common/plutus"
	"github.com/grupokindynos/common/tokens/mrt"
	"github.com/grupokindynos/common/tokens/mvt"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var HTTPClient = http.Client{
	Timeout: time.Second * 15,
}

func GetWalletBalances() []balance.Balance {
	flagAllRates := false
	log.Println("Retrieving Wallet Balances...")
	var rawBalances []balance.Balance
	availableCoins := coinfactory.Coins
	for _, coin := range availableCoins {
		res, err := plutus.GetWalletBalance(os.Getenv("PLUTUS_URL"), strings.ToLower(coin.Tag), os.Getenv("ADRESTIA_PRIV_KEY"), "adrestia", os.Getenv("PLUTUS_AUTH_USERNAME"), os.Getenv("PLUTUS_AUTH_PASSWORD"), os.Getenv("PLUTUS_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
		if err != nil {
			fmt.Println(fmt.Sprintf("Plutus Service Error for %s: %v", coin.Tag, err))
		} else {
			// Create Balance Object
			b := balance.Balance{}
			b.ConfirmedBalance = res.Confirmed
			b.UnconfirmedBalance = res.Unconfirmed
			b.Ticker = coin.Tag
			rawBalances = append(rawBalances, b)
			fmt.Println(fmt.Sprintf("%.8f %s\t of a total of %.8f\t%.2f%%", b.ConfirmedBalance, b.Ticker, b.ConfirmedBalance + b.UnconfirmedBalance, b.GetConfirmedProportion()))
		}
	}
	log.Println("Finished Retrieving Balances")

	var errRates []string

	var updatedBalances []balance.Balance
	log.Println("Retrieving Wallet Rates...")
	for _, coin := range rawBalances {
		var currentBalance = coin
		rate, err := obol.GetCoin2CoinRates("https://obol-rates.herokuapp.com/", "btc", currentBalance.Ticker)
		if err != nil{
			flagAllRates = true
			errRates = append(errRates, coin.Ticker)
		} else {
			fmt.Println("Rate for ", coin.Ticker, " is ", rate)
			currentBalance.SetRate(rate)
			updatedBalances = append(updatedBalances, currentBalance)
		}
	}
	if flagAllRates {
		color.Error.Tips("Not all rates could be retrieved. Balancing the rest of them. Missing rates for %s", errRates)
	}
	return updatedBalances
}

func GetBtcAddress() (string, error){
	address, err := plutus.GetWalletAddress(os.Getenv("PLUTUS_URL"), "btc", os.Getenv("ADRESTIA_PRIV_KEY"), "adrestia", os.Getenv("PLUTUS_AUTH_USERNAME"), os.Getenv("PLUTUS_AUTH_PASSWORD"), os.Getenv("PLUTUS_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if err != nil {
		return "", err
	}
	return address, nil
}

func GetAddress(coin string) (string, error){
	address, err := plutus.GetWalletAddress(os.Getenv("PLUTUS_URL"), strings.ToLower(coin), os.Getenv("ADRESTIA_PRIV_KEY"), "adrestia", os.Getenv("PLUTUS_AUTH_USERNAME"), os.Getenv("PLUTUS_AUTH_PASSWORD"), os.Getenv("PLUTUS_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if err != nil {
		return "", err
	}
	return address, nil
}

// Extracted from tyche/services/plutus.go
func WithdrawToAddress(body plutus.SendAddressBodyReq) (txId string, err error) {
	req, err := mvt.CreateMVTToken("POST", plutus.ProductionURL+"/send/address", "tyche", os.Getenv("MASTER_PASSWORD"), body, os.Getenv("PLUTUS_AUTH_USERNAME"), os.Getenv("PLUTUS_AUTH_PASSWORD"), os.Getenv("TYCHE_PRIV_KEY"))
	if err != nil {
		return txId, err
	}
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return txId, err
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return txId, err
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		return txId, err
	}
	headerSignature := res.Header.Get("service")
	if headerSignature == "" {
		return txId, err
	}
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("PLUTUS_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return txId, err
	}
	var response string
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return txId, err
	}
	return response, nil
}
