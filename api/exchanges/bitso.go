package exchanges

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
