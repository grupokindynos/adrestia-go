package exchanges

import (
	"github.com/grupokindynos/adrestia-go/api/exchanges/config"
	"os"
)

type Bitso struct {
	Exchange
	AccountName string
	BitSharesUrl string
}

func NewBitso() *Bitso {
	b := new(Bitso)
	b.Name = "Bitso"
	b.BaseUrl = "https://api.crypto-bridge.org/"
	return b
}

func (b Bitso) GetSettings() {
	var data config.CBAuth
	data.AccountName = os.Getenv("CB_ACCOUNT_NAME")
	data.BaseUrl = os.Getenv("CB_BASE_URL")
	data.MasterPassword = os.Getenv("CB_MASTER_PASSWORD")
	data.BitSharesUrl = os.Getenv("CB_BITSHARES_URL")
}
