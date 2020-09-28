package exchanges

import (
	"net/http"

	"github.com/grupokindynos/adrestia-go/models"
	"github.com/grupokindynos/common/hestia"
)

//Bithumb The attributes needed for the Bithumb exchanges
type Bithumb struct {
	Name          string
	user          string
	authorization string
	client        *http.Client
}

//NewBithumb Creates a new instance of Bithumb
func NewBithumb(params models.ExchangeParams) *Bithumb {
	b := new(Bithumb)
	b.Name = params.Name
	b.user = params.Keys.PublicKey
	b.authorization = params.Keys.PrivateKey
	b.client = &http.Client{}
	return b
}

// GetName Gets the exchange name
func (b *Bithumb) GetName() (string, error) {
	return b.Name, nil
}

func (b *Bithumb) GetAddress(asset string) (string, error) {
	return "", nil
}

func (b *Bithumb) GetBalance(asset string) (float64, error) {
	return 0, nil
}

func (b *Bithumb) SellAtMarketPrice(order hestia.Trade) (string, error) {
	return "", nil
}

func (b *Bithumb) Withdraw(asset string, address string, amount float64) (string, error) {
	return "", nil
}

func (b *Bithumb) GetOrderStatus(order hestia.Trade) (hestia.ExchangeOrderInfo, error) {
	h := hestia.ExchangeOrderInfo{}
	return h, nil
}

func (b *Bithumb) GetPair(fromCoin string, toCoin string) (models.TradeInfo, error) {
	t := models.TradeInfo{}
	return t, nil
}

func (b *Bithumb) GetWithdrawalTxHash(txId string, asset string) (string, error) {
	return "", nil
}

//  Gets the deposit status from an asset's exchange.
func (b *Bithumb) GetDepositStatus(addr string, txId string, asset string) (hestia.ExchangeOrderInfo, error) {
	e := hestia.ExchangeOrderInfo{}
	return e, nil
}
