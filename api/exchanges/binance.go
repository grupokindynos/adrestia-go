package exchanges

type Binance struct {
	Exchange
	AccountName string
	BitSharesUrl string
}

func NewBinance() *Binance {
	c := new(Binance)
	c.Name = "Binance"
	c.BaseUrl = "https://api.crypto-bridge.org/"
	return c
}
