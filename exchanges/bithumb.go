package exchanges

import (
	"errors"
	"fmt"
	"github.com/PrettyBoyHelios/go-bithumb"
	"github.com/grupokindynos/adrestia-go/models"
	"github.com/grupokindynos/common/hestia"
	"github.com/shopspring/decimal"
	"net/http"
	"strings"
)

//Bithumb The attributes needed for the Bithumb exchanges
type Bithumb struct {
	Name          string
	user          string
	authorization string
	url           string
	addresses     map[string]string
	client        *http.Client
	bithumbClient *go_bithumb.Bithumb
}

//NewBithumb Creates a new instance of Bithumb
func NewBithumb(params models.ExchangeParams) *Bithumb {
	b := new(Bithumb)
	b.Name = params.Name
	b.user = params.Keys.PublicKey
	b.authorization = params.Keys.PrivateKey
	b.client = &http.Client{}
	b.addresses = map[string]string{
		"BTC":  "1L9jPKCbUbK9aKgn5miwRmUy51Pm64SeW6",
		"GTH":  "0xb9cc9b046a901cff4a6943e26158c6de415a8b32",
		"USDT": "0xb9cc9b046a901cff4a6943e26158c6de415a8b32",
	}
	b.bithumbClient = go_bithumb.NewBithhumbAuth(params.Keys.PublicKey, params.Keys.PrivateKey)
	b.url = "https://global-openapi.bithumb.pro/openapi/v1"
	return b
}

// GetName Gets the exchange name
func (b *Bithumb) GetName() (string, error) {
	return b.Name, nil
}

// GetAddress Gets address from Bithumb WIP: Add error message
func (b *Bithumb) GetAddress(asset string) (string, error) {
	if val, ok := b.addresses[asset]; ok {
		return val, nil
	}
	return "", errors.New("address not found for " + asset)
}

// GetBalance Gets the balance for a given asset
func (b *Bithumb) GetBalance(asset string) (float64, error) {
	assetInfo, err := b.bithumbClient.Assets(asset)
	if err != nil {
		return 0, err
	}
	if len(assetInfo.Data) >= 1 {
		balance, _ := assetInfo.Data[0].Count.Float64()
		return balance, nil
	}
	return 0, errors.New("asset not found")
}

func (b *Bithumb) SellAtMarketPrice(order hestia.Trade) (string, error) {
	orderInfo, err := b.bithumbClient.CreateOrder(order.Symbol, strings.ToLower(order.Side), decimal.NewFromFloat(order.Amount), decimal.NewFromFloat(0), strings.ToLower("market"))
	if err != nil {
		return "", err
	}
	return orderInfo.Data.OrderId, nil
}

func (b *Bithumb) Withdraw(asset string, address string, amount float64) (string, error) {
	_, err := b.bithumbClient.Withdraw(asset, address, decimal.NewFromFloat(amount), "")
	if err != nil {
		return "", err
	}
	return "", nil
}

func (b *Bithumb) GetOrderStatus(order hestia.Trade) (hestia.ExchangeOrderInfo, error) {
	orderStatus, err := b.bithumbClient.OrderDetail(order.Symbol, order.OrderId)
	if err != nil {
		return hestia.ExchangeOrderInfo{
			Status:         hestia.ExchangeOrderStatusOpen,
			ReceivedAmount: 0,
		}, err
	}
	switch orderStatus.Data.Status {
	case "send":
		return hestia.ExchangeOrderInfo{
			Status:         hestia.ExchangeOrderStatusOpen,
			ReceivedAmount: 0,
		}, nil
	case "pending":
		return hestia.ExchangeOrderInfo{
			Status:         hestia.ExchangeOrderStatusOpen,
			ReceivedAmount: 0,
		}, nil
	case "success":
		tradedValue, _ := orderStatus.Data.TradedNum.Float64()
		return hestia.ExchangeOrderInfo{
			Status:         hestia.ExchangeOrderStatusCompleted,
			ReceivedAmount: tradedValue,
		}, nil
	case "cancel":
		return hestia.ExchangeOrderInfo{
			Status:         hestia.ExchangeOrderStatusError,
			ReceivedAmount: 0,
		}, nil
	default:
		return hestia.ExchangeOrderInfo{
			Status:         hestia.ExchangeOrderStatusOpen,
			ReceivedAmount: 0,
		}, nil
	}
}

func (b *Bithumb) GetPair(fromCoin string, toCoin string) (models.TradeInfo, error) {
	markets, err := b.bithumbClient.GetConfig()
	if err != nil {
		return models.TradeInfo{}, err
	}
	symbolBuy := fmt.Sprintf("%s-%s", strings.ToUpper(fromCoin), strings.ToUpper(toCoin))
	symbolSell := fmt.Sprintf("%s-%s", strings.ToUpper(toCoin), strings.ToUpper(fromCoin))
	for _, spotData := range markets.Data.SpotConfig {
		if spotData.Symbol == symbolBuy {
			return models.TradeInfo{
				Book: spotData.Symbol,
				Type: "sell",
			}, nil
		} else if spotData.Symbol == symbolSell {
			return models.TradeInfo{
				Book: spotData.Symbol,
				Type: "buy",
			}, nil
		}
	}
	t := models.TradeInfo{}
	return t, errors.New("pair not found")
}

func (b *Bithumb) GetWithdrawalTxHash(txId string, asset string) (string, error) {
	withdrawals, err := b.bithumbClient.WithdrawalHistory(asset)
	if err != nil {
		return "", err
	}
	for _, w := range withdrawals.Data {
		if w.Txid == txId {
			return w.Txid, nil
		}
	}
	// No way of knowing a withdrawal tx hash
	return "", errors.New(fmt.Sprintf("no withdrawal matching id: %s", txId))
}

//  Gets the deposit status from an asset's exchange.
func (b *Bithumb) GetDepositStatus(addr string, txId string, asset string) (hestia.ExchangeOrderInfo, error) {
	// bithumb does not provide a way of searching for this, maybe we can use block explorer to determine XY network confirmations
	depositInfo := hestia.ExchangeOrderInfo{}
	deposits, err := b.bithumbClient.DepositHistory(asset)
	if err != nil {
		return depositInfo, err
	}
	var orderStatus hestia.ExchangeOrderStatus
	var amount float64
	for _, deposit := range deposits.Data {
		if deposit.Txid == txId && deposit.Address == addr {
			amount, _ = deposit.Quantity.Float64()
			switch deposit.Status {
			case "0":
				orderStatus = hestia.ExchangeOrderStatusOpen
			case "1":
				orderStatus = hestia.ExchangeOrderStatusCompleted
			}
			return hestia.ExchangeOrderInfo{
				Status: orderStatus,
				ReceivedAmount: amount,
			}, nil
		}
	}
	return depositInfo, errors.New("deposit not found")
}
