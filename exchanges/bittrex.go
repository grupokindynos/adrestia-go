package exchanges

import (
	"errors"
	"fmt"
	"github.com/PrettyBoyHelios/go-bittrex"
	"github.com/grupokindynos/adrestia-go/models"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/hestia"
	"github.com/shopspring/decimal"
	"strings"
)

type bittrexCurrrencyInfo struct {
	minConfirms int
	txFee float64
}

type Bittrex struct {
	Name string
	exchange *bittrex.Bittrex
	minConfs map[string]bittrexCurrrencyInfo
}

func NewBittrex(params models.ExchangeParams) (*Bittrex, error) {
	b := bittrex.New(params.Keys.PublicKey, params.Keys.PrivateKey)

	currencies, err := b.GetCurrencies()
	if err != nil {
		return nil, err
	}

	minConfs := make(map[string]bittrexCurrrencyInfo)

	for _, c := range currencies {
		fee,_ :=  c.TxFee.Float64()
		minConfs[strings.ToLower(c.Symbol)] = bittrexCurrrencyInfo {
			minConfirms: c.MinConfirmations,
			txFee: fee,
		}
	}
	// bittrex returns 0 for USDT
	ci := minConfs["usdt"]
	ci.minConfirms = 36
	minConfs["usdt"] = ci

	return &Bittrex{
		Name: params.Name,
		exchange: b,
		minConfs: minConfs,
	}, nil
}

func (b *Bittrex) GetName() (string, error) {
	return b.Name, nil
}

func (b *Bittrex) GetAddress(coin string) (string, error) {
	addr, err := b.exchange.GetDepositAddress(strings.ToLower(coin))
	if err != nil {
		return "", err
	}

	return addr.CryptoAddress, nil
}

func (b *Bittrex) GetDepositStatus(addr string, txId string, asset string) (orderStatus hestia.ExchangeOrderInfo, err error) {
	coinInfo, _ := coinfactory.GetCoin(asset)
	if coinInfo.Info.Token {
		if val, err := blockbookConfirmed(addr, txId, b.minConfs[asset].minConfirms); err == nil {
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
			if d.Confirmations >= b.minConfs[strings.ToLower(d.Currency)].minConfirms {
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
		if coin == asset.CurrencySymbol {
			value, _ := asset.Available.Float64()
			return value, nil
		}
	}

	return 0, errors.New("coin not found")
}

// cantidad a comprar/vender
func (b *Bittrex) getBestPrice(amount decimal.Decimal, market string, side string) (decimal.Decimal, error) {
	var err error
	var orders []bittrex.OrderbV3
	if side == "buy" {
		orders, err = b.exchange.GetOrderBookBuySell(market, 50,"sell")
	} else {
		orders, err = b.exchange.GetOrderBookBuySell(market, 50,"buy")
	}

	if err != nil {
		return decimal.Decimal{}, err
	}
	cumulativeAmount := decimal.NewFromFloat(0)
	var price decimal.Decimal
	for _, order := range orders {
		cumulativeAmount = cumulativeAmount.Add(order.Quantity)
		if cumulativeAmount.GreaterThan(amount) {
			price = order.Rate
			break
		}
	}

	return price, nil
}

func getMarketName(base string, market string) string {
	return fmt.Sprintf("%s-%s", strings.ToUpper(base), strings.ToUpper(market))
}

func (b *Bittrex) SellAtMarketPrice(order hestia.Trade) (string, error) {
	market, base := order.GetTradingPair()
	marketName := getMarketName(base, market)

	summary, err := b.exchange.GetMarketSummary(marketName)
	if err != nil {
		return "", err
	}

	if order.Side == "buy" {
		bidPrice := summary[0].High
		buyAmount := decimal.NewFromFloat(order.Amount).Div(bidPrice)
		fee := decimal.NewFromFloat( b.minConfs[strings.ToLower(order.ToCoin)].txFee)
		buyAmount.Add(fee.Neg())

		bestPrice, err := b.getBestPrice(buyAmount, marketName, "buy")
		if err != nil {
			return "", err
		}
		return b.exchange.BuyLimit(marketName, buyAmount, bestPrice)
	} else {
		order.Amount -= b.minConfs[strings.ToLower(order.FromCoin)].txFee
		bestPrice, err := b.getBestPrice(decimal.NewFromFloat(order.Amount), marketName, "sell")
		if err != nil {
			return "", err
		}
		return b.exchange.SellLimit(marketName, decimal.NewFromFloat(order.Amount), bestPrice)
	}
}

func (b *Bittrex) Withdraw(coin string, address string, amount float64) (string, error) {
	withdrawRes, err := b.exchange.Withdraw(address, coin, decimal.NewFromFloat(amount), "")
	if err != nil {
		return "", err
	}
	return withdrawRes.TxID, nil
}

func (b *Bittrex) GetOrderStatus(order hestia.Trade) (hestia.ExchangeOrderInfo, error) {
	status := hestia.ExchangeOrderInfo{}
	o, err := b.exchange.GetOrder(order.OrderId)
	if err != nil {
		return status, err
	}

	status.Status = hestia.ExchangeOrderStatusOpen

	amountExecuted, _ := o.Quantity.Sub(o.QuantityRemaining).Float64()
	if o.Type == "LIMIT_BUY" {
		status.ReceivedAmount = amountExecuted
	} else {
		status.ReceivedAmount, _ = o.Limit.Mul(decimal.NewFromFloat(amountExecuted)).Float64()
	}

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
	// TODO Optimize with V3 Method
	withdraws, err := b.exchange.GetClosedWithdrawals(strings.ToUpper(asset), bittrex.ALL)
	if err != nil {
		return "", err
	}

	for _, w := range withdraws {
		if w.TxID == txId {
			return w.TxID, nil
		}
	}

	return "", fmt.Errorf("no withdrawal matching id: %s", txId)
}