package exchanges

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	south "github.com/oedipusK/go-southxchange"
)

type SouthXchange struct {
	Exchange
	apiKey          string
	apiSecret       string
	southClient     south.SouthXchange
	Obol            obol.ObolService
}

func NewSouthXchange(params Params) *SouthXchange {
	s := new(SouthXchange)
	s.Name = "southxchange"
	data := s.getSettings()
	s.apiKey = data.ApiKey
	s.apiSecret = data.ApiSecret
	s.southClient = *south.New(s.apiKey, s.apiSecret, "user-agent")
	s.Obol = params.Obol
	return s
}

func (s *SouthXchange) GetName() (string, error) {
	return "southxchange", nil
}

func (s *SouthXchange) GetAddress(coin coins.Coin) (string, error) {
	var address string
	var err error

	for i := 0; i < 3; i++ { // try to get an address just 3 times
		address, err = s.southClient.GetDepositAddress(strings.ToLower(coin.Info.Tag))
		envTag := "SOUTH_ADDRESS_" + coin.Info.Tag
		if err != nil {
			if !strings.Contains(err.Error(), "400") {
				address = "southxchange address request error"
				break
			}

			if os.Getenv(envTag) != "" {
				address = os.Getenv(envTag)
				break
			}
			time.Sleep(90 * time.Second) // wait 90 seconds to generate a new address
		} else {
			os.Setenv(envTag, address)
			break
		}
	}
	str := string(address)
	str = strings.Replace(str, "\\", "", -1)
	str = strings.Replace(str, "\"", "", -1)
	str = strings.Replace(str, "/", "", -1)
	return str, nil
}

func (s *SouthXchange) GetBalances() ([]balance.Balance, error) {  // tal vez la modifque para que solo regrese la que queremos
	var balances []balance.Balance
	res, err := s.southClient.GetBalances()
	if err != nil {
		log.Println("south - GetBalances - GetBalances() - ", err.Error())
		return balances, err
	}
	var rate float64
	for _, asset := range res {
		if strings.ToLower(asset.Currency) != "btc" {
			rate, _ = s.Obol.GetCoin2CoinRates("BTC", asset.Currency)
		} else {
			rate = 1.0
		}
		var b = balance.Balance{
			Ticker:             asset.Currency,
			ConfirmedBalance:   asset.Available,
			UnconfirmedBalance: asset.Unconfirmed,
			RateBTC:            rate,
			DiffBTC:            0.0,
			IsBalanced:         false,
		}
		if b.GetTotalBalance() > 0.0 {
			balances = append(balances, b)
		}

	}
	return balances, nil
}

func (s *SouthXchange) SellAtMarketPrice(order hestia.ExchangeOrder) (string, error) {
	l, r := order.GetTradingPair()
	var res string
	var err error

	var orderType south.OrderType
	if order.Side == "buy" {
		orderType = south.Buy
		price, err := s.southClient.GetMarketPrice(l, r)
		if err != nil {
			log.Println("south - SellAtMarketPrice - GetMarketPrice - ", err.Error())
			return "", err
		}
		buyAmount := order.Amount / price.Bid
		res, err = s.southClient.PlaceOrder(l, r, orderType, buyAmount, price.Bid, true)
	} else {
		orderType = south.Sell
		res, err = s.southClient.PlaceOrder(l, r, orderType, order.Amount, 0.0, true)
	}

	if err != nil {
		log.Println("south - SellAtMarketPrice - PlaceOrder - " + err.Error())
		return "", err
	}

	res = strings.ReplaceAll(res, "\"", "")
	return res, nil
}

func (s *SouthXchange) Withdraw(coin coins.Coin, address string, amount float64) (string, error) {
	info, err := s.southClient.Withdraw(address, strings.ToUpper(coin.Info.Tag), amount)
	if err != nil {
		log.Println("south - Withdraw - Withdraw() - ", err.Error())
		return "", err
	}
	id := strconv.FormatInt(info.MovementId, 10)
	return id, err
}

func (s *SouthXchange) GetDepositStatus(txid string, asset string) (hestia.OrderStatus, error) {
	var status hestia.OrderStatus
	txs, err := s.southClient.GetTransactions("deposits", 0, 1000, "", false)
	if err != nil {
		log.Println("south - GetDepositStatus - GetTransactions() - ", err.Error())
		return status, err
	}
	status.Status = hestia.ExchangeStatusError
	for _, tx := range txs {
		if tx.Hash == txid {
			log.Println(tx)
			if tx.Status == "confirmed" || tx.Status == "executed" {
				status.Status = hestia.ExchangeStatusCompleted
				status.AvailableAmount = tx.Amount
				return status, nil
			} else if tx.Status == "pending" || tx.Status == "booked" {
				status.Status = hestia.ExchangeStatusOpen
				return status, nil
			} else {
				return status, errors.New("south - GetDepositStatus - unknown status " + tx.Status)
			}
		}
	}
	return status, errors.New("south - transaction not found")
}

func (s *SouthXchange) GetOrderStatus(order hestia.ExchangeOrder) (hestia.OrderStatus, error) {
	var status hestia.OrderStatus
	southOrder, err := s.southClient.GetOrder(order.OrderId)
	if err != nil {
		log.Println("south - GetOrderStatus - GetOrder() - ", err.Error())
		return status, err
	}
	if southOrder.Status == "executed" || southOrder.Status == "confirmed" {
		status.Status = hestia.ExchangeStatusCompleted
		amount, err := s.getAvailableAmount(order)
		if err != nil {
			log.Println("south - GetOrderStatus - getAvailableAmount() - ", err.Error())
			return status, err
		}
		status.AvailableAmount = amount
	} else if southOrder.Status == "pending" || southOrder.Status == "booked" {
		status.Status = hestia.ExchangeStatusOpen
		status.AvailableAmount = 0.0
	} else {
		status.Status = hestia.ExchangeStatusError
		status.AvailableAmount = 0
		err = errors.New("south - unknown order status " + southOrder.Status)
	}
	return status, err
}

func (s *SouthXchange) GetWithdrawalTxHash(txId string, _ string) (string, error) {
	txs, err := s.southClient.GetTransactions("", 0, 500, "", true)
	if err != nil {
		log.Println("south - GetWithdrawalTxHash - GetTransactions() - ", err.Error())
		return "", err
	}
	tradeId, err := strconv.ParseInt(txId, 10, 64)
	if err != nil {
		log.Println("south - GetWithdrawalTxHash - ParseInt() - ", err.Error())
		return "", err
	}
	for _, tx := range txs {
		if tx.MovementId == tradeId && tx.Type == "withdraw" {
			return tx.Hash, nil
		}
	}

	return "", errors.New("south - withdrawal not found")
}

func (s *SouthXchange) GetPair(fromCoin string, toCoin string) (OrderSide, error) {
	var orderSide OrderSide
	fromCoin = strings.ToUpper(fromCoin)
	toCoin = strings.ToUpper(toCoin)
	books, err := s.southClient.GetMarketSummaries()
	if err != nil {
		log.Println("south - GetPair - GetMarketSummaries() - ", err.Error())
		return orderSide, err
	}
	var bookName south.MarketSummary
	for _, book := range books {
		if (book.Coin == fromCoin || book.Base == fromCoin) && (book.Coin == toCoin || book.Base == toCoin) {
			bookName = book
			break
		}
	}

	orderSide.Book = bookName.Coin + bookName.Base
	if bookName.Coin == fromCoin {
		orderSide.Type = "sell"
		orderSide.ReceivedCurrency = bookName.Base
		orderSide.SoldCurrency = bookName.Coin
	} else {
		orderSide.Type = "buy"
		orderSide.ReceivedCurrency = bookName.Coin
		orderSide.SoldCurrency = bookName.Base
	}

	return orderSide, nil
}

func (s *SouthXchange) getAvailableAmount(order hestia.ExchangeOrder) (float64, error) {
	txs, err := s.southClient.GetTransactions("", 0, 1000, "", true)
	if err != nil {
		return 0, err
	}

	availableAmount := 0.0
	set := make(map[int]bool)

	for _, tx := range txs {
		if tx.OrderCode == order.OrderId {
			set[tx.TradeId] = true
		}
	}

	for _, tx := range txs {
		if set[tx.TradeId] {
			availableAmount += tx.Amount
		}
	}

	if availableAmount == 0.0 {
		return 0.0, errors.New("tx not found")
	}

	if order.Side == "sell" {
		availableAmount += order.Amount
	}

	return availableAmount, nil
}

func (s *SouthXchange) getSettings() config.SouthXchangeAuth {
	var data config.SouthXchangeAuth
	data.ApiKey = os.Getenv("SOUTH_API_KEY")
	data.ApiSecret = os.Getenv("SOUTH_API_SECRET")
	return data
}