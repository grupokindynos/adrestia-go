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
	c := new(Bitso)
	c.Name = "Bitso"
	c.BaseUrl = "https://api.crypto-bridge.org/"
	return c
}

func (b Bitso) GetSettings() {
	var data config.CBAuth
	data.AccountName = os.Getenv("CB_ACCOUNT_NAME")
	data.BaseUrl = os.Getenv("CB_BASE_URL")
	data.MasterPassword = os.Getenv("CB_MASTER_PASSWORD")
	data.BitSharesUrl = os.Getenv("CB_BITSHARES_URL")
}
