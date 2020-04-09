package exchanges

import (
	"errors"
	"fmt"
	"github.com/Toorop/go-bittrex"
	"github.com/grupokindynos/adrestia-go/models"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/hestia"
	"github.com/shopspring/decimal"
	"strings"
)

type Bittrex struct {
	exchangeInfo hestia.ExchangeInfo
	exchange *bittrex.Bittrex
	minConfs map[string]int
}

func NewBittrex(exchange hestia.ExchangeInfo) (*Bittrex, error) {

	b := bittrex.New(exchange.ApiPublicKey, exchange.ApiPrivateKey)

	currencies, err := b.GetCurrencies()
	if err != nil {
		return nil, err
	}

	minConfs := make(map[string]int)

	for _, c := range currencies {
		minConfs[strings.ToLower(c.Currency)] = c.MinConfirmation
	}

	return &Bittrex{
		exchangeInfo: exchange,
		exchange: b,
		minConfs: minConfs,
	}, nil
}

func (b *Bittrex) GetName() (string, error) {
	return b.exchangeInfo.Name, nil
}

func (b *Bittrex) GetAddress(coin string) (string, error) {
	addr, err := b.exchange.GetDepositAddress(strings.ToLower(coin))
	if err != nil {
		return "", err
	}

	return addr.Address, nil
}

func (b *Bittrex) GetDepositStatus(addr string, txId string, asset string) (orderStatus hestia.ExchangeOrderInfo, err error) {
	coinInfo, _ := coinfactory.GetCoin(asset)
	if coinInfo.Info.Token {
		if val, err := blockbookConfirmed(addr, txId, b.minConfs[asset]); err == nil {
			return hestia.ExchangeOrderInfo{
				Status: hestia.ExchangeOrderStatusCompleted,
				ReceivedAmount: val,
			}, nil
		} else {
			return hestia.ExchangeOrderInfo{}, err
		}
	}
	orderStatus.Status = hestia.ExchangeOrderStatusOpen

	deposits, err := b.exchange.GetDepositHistory(asset)
	if err != nil {
		orderStatus.Status = hestia.ExchangeOrderStatusError
		return
	}

	for _, d := range deposits {
		if d.TxId == txId {
			if b.minConfs[strings.ToLower(d.Currency)] >= d.Confirmations {
				orderStatus.Status = hestia.ExchangeOrderStatusCompleted
				orderStatus.ReceivedAmount, _ = d.Amount.Float64()
			}
			return
		}
	}
	return
}

func (b *Bittrex) GetBalance(coin string) (float64, error) {
	balances, err := b.exchange.GetBalances()
	if err != nil {
		return 0, err
	}

	for _, asset := range balances {
		if coin == asset.Currency {
			value, _ := asset.Available.Float64()
			return value, nil
		}
	}

	return 0, errors.New("coin not found")
}

func getMarketName(base string, market string) string {
	return fmt.Sprintf("%s-%s", strings.ToUpper(base), strings.ToUpper(market))
}

func (b *Bittrex) SellAtMarketPrice(order hestia.Trade) (string, error) {
	market, base := order.GetTradingPair()
	marketName := getMarketName(base, market)

	if order.Side == "buy" {
		summary, err := b.exchange.GetMarketSummary(marketName)
		if err != nil {
			return "", err
		}
		price := summary[0].Ask
		buyAmount := decimal.NewFromFloat(order.Amount).Div(price)
		return b.exchange.BuyLimit(marketName, buyAmount, price)
	} else {
		return b.exchange.SellLimit(marketName, decimal.NewFromFloat(order.Amount), decimal.Zero)
	}
}

func (b *Bittrex) Withdraw(coin string, address string, amount float64) (string, error) {
	return b.exchange.Withdraw(address, strings.ToUpper(coin), decimal.NewFromFloat(amount))
}

func (b *Bittrex) GetOrderStatus(order hestia.Trade) (hestia.ExchangeOrderInfo, error) {
	status := hestia.ExchangeOrderInfo{}
	o, err := b.exchange.GetOrder(order.OrderId)
	if err != nil {
		return status, err
	}

	status.Status = hestia.ExchangeOrderStatusOpen

	amountExecuted, _ := o.Quantity.Sub(o.QuantityRemaining).Float64()
	status.ReceivedAmount = amountExecuted

	if o.QuantityRemaining.Equals(decimal.Zero) {
		status.Status = hestia.ExchangeOrderStatusCompleted
	}

	return status, nil
}

func (b *Bittrex) GetPair(fromCoin string, toCoin string) (models.TradeInfo, error) {
	var side models.TradeInfo

	markets, err := b.exchange.GetMarkets()
	if err != nil {
		return side, err
	}

	fromLower := strings.ToLower(fromCoin)
	toLower := strings.ToLower(toCoin)

	var book *bittrex.Market
	for _, m := range markets {
		marketLower := strings.ToLower(m.MarketCurrency)
		baseLower := strings.ToLower(m.BaseCurrency)
		if (marketLower == fromLower && baseLower == toLower) || (marketLower == toLower && baseLower == fromLower) {
			book = &m
			break
		}
	}

	if book == nil {
		return side, fmt.Errorf("could not find market for currencies %s and %s", fromCoin, toCoin)
	}

	side.Book = book.MarketName
	if strings.ToLower(book.BaseCurrency) == fromLower {
		side.Type = "buy"
	} else {
		side.Type = "sell"
	}

	return side, nil
}

func (b *Bittrex) GetWithdrawalTxHash(txId string, asset string) (string, error) {
	withdraws, err := b.exchange.GetWithdrawalHistory(strings.ToUpper(asset))
	if err != nil {
		return "", err
	}

	for _, w := range withdraws {
		if w.PaymentUuid == txId {
			return w.TxId, nil
		}
	}

	return "", fmt.Errorf("no withdrawal matching id: %s", txId)
}