package exchanges

import (
	"fmt"
	"github.com/Toorop/go-bittrex"
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	"github.com/shopspring/decimal"
	"os"
	"strings"
)

type Bittrex struct {
	exchange *bittrex.Bittrex
	minConfs map[string]int

	Obol obol.ObolService
}

func NewBittrex(params Params) (*Bittrex, error) {
	b := bittrex.New(os.Getenv("BITTREX_API_KEY"), os.Getenv("BITTREX_API_SECRET"))

	currencies, err := b.GetCurrencies()
	if err != nil {
		return nil, err
	}

	minConfs := make(map[string]int)

	for _, c := range currencies {
		minConfs[strings.ToLower(c.Currency)] = c.MinConfirmation
	}

	return &Bittrex{
		exchange: b,
		minConfs: minConfs,
		Obol:     params.Obol,
	}, nil
}

func (b *Bittrex) GetName() (string, error) {
	return "Bittrex", nil
}

func (b *Bittrex) GetAddress(coin coins.Coin) (string, error) {
	addr, err := b.exchange.GetDepositAddress(strings.ToLower(coin.Info.Tag))
	if err != nil {
		return "", err
	}

	return addr.Address, nil
}

func (b *Bittrex) GetDepositStatus(txid string, asset string) (orderStatus hestia.OrderStatus, err error) {
	orderStatus.AvailableAmount = 0
	orderStatus.Status = hestia.ExchangeStatusOpen

	deposits, err := b.exchange.GetDepositHistory(asset)
	if err != nil {
		orderStatus.Status = hestia.ExchangeStatusError
		return
	}

	for _, d := range deposits {
		if d.TxId == txid {
			if b.minConfs[strings.ToLower(d.Currency)] >= d.Confirmations {
				orderStatus.Status = hestia.ExchangeStatusCompleted
				orderStatus.AvailableAmount, _ = d.Amount.Float64()
			}
			return
		}
	}
	return
}

func (b *Bittrex) GetBalances() ([]balance.Balance, error) {
	balances, err := b.exchange.GetBalances()
	if err != nil {
		return nil, err
	}

	bs := make([]balance.Balance, 0, len(balances))
	var rate float64
	for _, asset := range balances {
		if strings.ToLower(asset.Currency) != "btc" {
			rate, _ = b.Obol.GetCoin2CoinRates("BTC", asset.Currency)
		} else {
			rate = 1.0
		}
		confirmedBalance, _ := asset.Available.Float64()
		unconfirmedBalance, _ := asset.Pending.Float64()
		var b = balance.Balance{
			Ticker:             asset.Currency,
			ConfirmedBalance:   confirmedBalance,
			UnconfirmedBalance: unconfirmedBalance,
			RateBTC:            rate,
			DiffBTC:            0,
			IsBalanced:         false,
		}
		if b.GetTotalBalance() > 0.0 {
			bs = append(bs, b)
		}
	}

	return bs, nil
}

func getMarketName(base string, market string) string {
	return fmt.Sprintf("%s-%s", strings.ToUpper(base), strings.ToUpper(market))
}

func (b *Bittrex) SellAtMarketPrice(order hestia.ExchangeOrder) (string, error) {
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

func (b *Bittrex) Withdraw(coin coins.Coin, address string, amount float64) (string, error) {
	return b.exchange.Withdraw(address, strings.ToUpper(coin.Info.Tag), decimal.NewFromFloat(amount))
}

func (b *Bittrex) GetOrderStatus(order hestia.ExchangeOrder) (hestia.OrderStatus, error) {
	status := hestia.OrderStatus{}
	o, err := b.exchange.GetOrder(order.OrderId)
	if err != nil {
		return status, err
	}

	status.Status = hestia.ExchangeStatusOpen

	amountExecuted, _ := o.Quantity.Sub(o.QuantityRemaining).Float64()
	status.AvailableAmount = amountExecuted

	if o.QuantityRemaining.Equals(decimal.Zero) {
		status.Status = hestia.ExchangeStatusCompleted
	}

	return status, nil
}

func (b *Bittrex) GetPair(fromCoin string, toCoin string) (OrderSide, error) {
	var side OrderSide

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
		side.ReceivedCurrency = book.MarketCurrency
		side.SoldCurrency = book.BaseCurrency
	} else {
		side.Type = "sell"
		side.ReceivedCurrency = book.BaseCurrency
		side.SoldCurrency = book.MarketCurrency
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

var _ IExchange = &Bittrex{}
