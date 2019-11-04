package exchanges

import (
	"github.com/grupokindynos/adrestia-go/api/exchanges/config"
	"github.com/grupokindynos/common/coin-factory/coins"
	bitso "github.com/grupokindynos/gobitso"
	"os"
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

func (b BitsoI) GetSettings() config.BitsoAuth{
	var data config.BitsoAuth
	data.ApiKey = os.Getenv("BITSO_API_KEY")
	data.ApiSecret = os.Getenv("BITSO_API_SECRET")
	data.Url = os.Getenv("BITSO_URL")
	return data
}
