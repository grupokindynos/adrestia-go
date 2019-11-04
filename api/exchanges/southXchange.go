package exchanges

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	south "github.com/bitbandi/go-southxchange"
	"github.com/grupokindynos/adrestia-go/api/exchanges/config"
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/utils"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/obol"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type SouthXchange struct {
	Exchange
	apiKey   	string
	apiSecret  	string
	southClient	south.SouthXchange
}

var SouthInstance = NewSouthXchange()

func NewSouthXchange() *SouthXchange {
	s := new(SouthXchange)
	s.Name = "SouthXchange"
	data := s.GetSettings()
	s.apiKey = data.ApiKey
	s.apiSecret = data.ApiSecret
	s.southClient = *south.New(s.apiKey, s.apiSecret, "user-agent")
	return s
}

func (s SouthXchange) GetSettings() config.SouthXchangeAuth {
	if err := godotenv.Load(); err != nil {
		log.Println(err)
	}
	log.Println(fmt.Sprintf("[GetSettings] Retrieving settings for Binance"))
	var data config.SouthXchangeAuth
	data.ApiKey = os.Getenv("SOUTH_API_KEY")
	data.ApiSecret = os.Getenv("BINANCE_PRIV_WITHDRAW")
	return data
}

func (s SouthXchange) GetBalances() ([]balance.Balance, error) {
	str := fmt.Sprintf("[GetBalances] Retrieving Balances for coins at %s", s.Name)
	log.Println(str)
	var balances []balance.Balance
	res, err := s.southClient.GetBalances()

	if err != nil {
		return balances, err
	}

	for _, asset := range res {
		rate, _ := obol.GetCoin2CoinRates("https://obol-rates.herokuapp.com/", "BTC", asset.Currency)
		var b = balance.Balance{
			Ticker:     asset.Currency,
			Balance:    asset.Available,
			RateBTC:    rate,
			DiffBTC:    0,
			IsBalanced: false,
		}
		if b.Balance > 0.0 {
			balances = append(balances, b)
		}

	}
	str = utils.GetBalanceLog(balances, s.Name)
	log.Println(str)
	return balances, nil
}

func (s *SouthXchange) GetAddress(coin coins.Coin) (string, error) {
	var client = http.Client{}
	payload := make(map[string]string)
	payload["key"] = s.apiKey
	payload["nonce"] = strconv.FormatInt(time.Now().UnixNano(), 10)
	formData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, "https://www.southxchange.com/api/generatenewaddress", strings.NewReader(string(formData)))
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/json;charset=utf-8")
	//req.Header.Add("Accept", "application/json") // cloudflare protected api doesnt accept this, i got captcha page
	req.Header.Add("Accept", "*/*")

	// Auth
	if len(s.apiKey) == 0 || len(s.apiSecret) == 0 {
		err = errors.New("you need to set API Key and API secret to call this method")
		return "", err
	}
	mac := hmac.New(sha512.New, []byte(s.apiSecret))
	_, err = mac.Write(formData)
	if err != nil {
		return "", err
	}
	sig := hex.EncodeToString(mac.Sum(nil))
	req.Header.Add("Hash", sig)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return string(response), err
	}
	if resp.StatusCode != 200 && resp.StatusCode != 401 {
		err = errors.New(resp.Status + ": "+strings.Trim(string(response), "\""))
	}
	return string(response), err

}