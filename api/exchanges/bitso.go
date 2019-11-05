package exchanges

import (
	"fmt"
	"github.com/grupokindynos/adrestia-go/api/exchanges/config"
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/obol"
	bitso "github.com/grupokindynos/gobitso"
	"os"
	"strconv"
	"strings"
)

var BitsoInstance = NewBitso()

type BitsoI struct {
	Exchange
	bitsoService bitso.Bitso
}

func NewBitso() *BitsoI {
	b := new(BitsoI)
	data := b.GetSettings()
	b.bitsoService = *bitso.NewBitso(data.Url)
	b.bitsoService.SetAuth(data.ApiKey, data.ApiSecret)
	return b
}

func (b BitsoI) GetName() (string, error){
	return "bitso", nil
}

func (b BitsoI) GetAddress(coin coins.Coin) (string, error) {
	if strings.ToLower(coin.Tag) == "btc" {
		return "btc address", nil
	}
	return "Missing Implementation", nil
}

func (b BitsoI) OneCoinToBtc(coin coins.Coin) (float64, error) {
	if coin.Tag == "BTC" {
		return 1.0, nil
	}
	rate, err := obol.GetCoin2CoinRatesWithAmmount("https://obol-rates.herokuapp.com/", "btc", coin.Tag, fmt.Sprintf("%f", 1.0))
	if err != nil {
		return 0.0, err
	}
	return rate, nil
}

func (b BitsoI) GetBalances() ([]balance.Balance, error) {
	bal, err := b.bitsoService.Balances()
	if err!=nil {
		return nil, err
	}
	var balances []balance.Balance

	for _, asset := range bal.Payload.Balances {
		rate, _ := obol.GetCoin2CoinRates("https://obol-rates.herokuapp.com/", "BTC", asset.Currency)
		totalAmount, err := strconv.ParseFloat(asset.Total, 64)
		if err != nil {
			return nil, err
		}
		var b = balance.Balance{
			Ticker:     asset.Currency,
			Balance:    totalAmount,
			RateBTC:    rate,
			DiffBTC:    0,
			IsBalanced: false,
		}
		if b.Balance > 0.0 {
			balances = append(balances, b)
		}

	}
	return balances, nil
}

func (b BitsoI) GetSettings() config.BitsoAuth{
	var data config.BitsoAuth
	data.ApiKey = os.Getenv("BITSO_API_KEY")
	data.ApiSecret = os.Getenv("BITSO_API_SECRET")
	data.Url = os.Getenv("BITSO_URL")
	return data
}
