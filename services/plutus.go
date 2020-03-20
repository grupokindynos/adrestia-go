package services

import (
	"encoding/json"
	"fmt"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/common/plutus"
	"github.com/grupokindynos/common/tokens/mrt"
	"github.com/grupokindynos/common/tokens/mvt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type PlutusInstance struct {
	PlutusURL string
	Obol obol.ObolService
}

func (p *PlutusInstance) GetWalletBalance(ticker string) (balance plutus.Balance, err error) {
	log.Println("Retrieving Wallet Balances...")
	coinInfo, err := coinfactory.GetCoin(ticker)
	if err != nil {
		fmt.Println("error jasdbsaisd")
		return
	}
	balance, err = plutus.GetWalletBalance(p.PlutusURL, strings.ToLower(coinInfo.Info.Tag), os.Getenv("ADRESTIA_PRIV_KEY"), "adrestia", os.Getenv("PLUTUS_AUTH_USERNAME"), os.Getenv("PLUTUS_AUTH_PASSWORD"), os.Getenv("PLUTUS_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if err != nil {
		fmt.Println(fmt.Sprintf("Plutus Service Error for %s: %v", coinInfo.Info.Tag, err))
	}
	return
}

func (p *PlutusInstance) WithdrawToAddress(body plutus.SendAddressBodyReq) (txId string, err error) {
	fmt.Printf("%+v\n", body)
	req, err := mvt.CreateMVTToken("POST", p.PlutusURL+"/send/address", "adrestia", os.Getenv("MASTER_PASSWORD"), body, os.Getenv("PLUTUS_AUTH_USERNAME"), os.Getenv("PLUTUS_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"))
	if err != nil {
		return txId, err
	}
	client := http.Client{
		Timeout: 30 * time.Second,
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
	log.Println(string(tokenResponse))
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		log.Println("WithdrawToAddress:: unmarshal error: received", string(tokenResponse))
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
