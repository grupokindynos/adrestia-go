package exchanges

import (
	"errors"
	"github.com/grupokindynos/adrestia-go/models"
	"github.com/shopspring/decimal"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/grupokindynos/common/hestia"
	south "github.com/oedipusK/go-southxchange"
)

type SouthXchange struct {
	exchangeInfo hestia.ExchangeInfo
	southClient  south.SouthXchange
}

func NewSouthXchange(exchange hestia.ExchangeInfo) *SouthXchange {
	s := new(SouthXchange)
	s.exchangeInfo = exchange
	s.southClient = *south.New(exchange.ApiPublicKey, exchange.ApiPrivateKey, "user-agent")
	return s
}

func (s *SouthXchange) GetName() (string, error) {
	return s.exchangeInfo.Name, nil
}

func (s *SouthXchange) GetAddress(coin string) (string, error) {
	// southxchange doesn't allow to have more than one address of tusd
	if strings.ToLower(coin) == "tusd" {
		return os.Getenv("SOUTH_TUSD_ADDRESS"), nil
	}
	var address string
	var err error

	for i := 0; i < 3; i++ { // try to get an address just 3 times
		address, err = s.southClient.GetDepositAddress(strings.ToLower(coin))
		envTag := "SOUTH_ADDRESS_" + coin
		if err != nil {
			log.Println("South GetAddress error " + err.Error())
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

func (s *SouthXchange) GetBalance(coin string) (float64, error) {  // tal vez la modifque para que solo regrese la que queremos
	res, err := s.southClient.GetBalances()
	if err != nil {
		log.Println("south - GetBalances - GetBalances() - ", err.Error())
		return 0, err
	}
	for _, asset := range res {
		if coin == asset.Currency {
			return asset.Available, nil
		}
	}
	return 0, errors.New("coin not found")
}

func (s *SouthXchange) getBestPrice(amount decimal.Decimal, listing string, reference string, side string) (float64, error) {
	book, err := s.southClient.GetBookOrders(listing, reference)
	if err != nil {
		return 0, err
	}

	var orders []south.BookOrder

	if side == "buy" {
		orders = book.SellOrders
	} else {
		orders = book.BuyOrders
	}

	cumulativeAmount := decimal.NewFromFloat(0)

	for _, order := range orders {
		if side == "buy" {
			cumulativeAmount = cumulativeAmount.Add(decimal.NewFromFloat(order.Price).Mul(decimal.NewFromFloat(order.Amount)))
		} else {
			cumulativeAmount = cumulativeAmount.Add(decimal.NewFromFloat(order.Amount))
		}
		if cumulativeAmount.GreaterThan(amount) {
			return order.Price, nil
		}
	}

	return 0, nil
}


func (s *SouthXchange) SellAtMarketPrice(order hestia.Trade) (string, error) {
	l, r := order.GetTradingPair()
	var res string
	var err error

	var orderType south.OrderType

	if order.Side == "buy" {
		orderType = south.Buy
		bestPrice, err := s.getBestPrice(decimal.NewFromFloat(order.Amount), l, r, "buy")
		if err != nil {
			return "", err
		}
		buyAmount, _ := decimal.NewFromFloat(order.Amount).Div(decimal.NewFromFloat(bestPrice)).Float64()
		res, err = s.southClient.PlaceOrder(l, r, orderType, buyAmount, bestPrice, false)
	} else {
		orderType = south.Sell
		bestPrice, err := s.getBestPrice(decimal.NewFromFloat(order.Amount), l, r, "sell")
		if err != nil {
			return "", err
		}
		res, err = s.southClient.PlaceOrder(l, r, orderType, order.Amount, bestPrice, false)
	}

	if err != nil {
		log.Println("south - SellAtMarketPrice - PlaceOrder - " + err.Error())
		return "", err
	}

	res = strings.ReplaceAll(res, "\"", "")
	return res, nil
}

func (s *SouthXchange) Withdraw(coin string, address string, amount float64) (string, error) {
	info, err := s.southClient.Withdraw(address, strings.ToUpper(coin), amount)
	if err != nil {
		log.Println("south - Withdraw - Withdraw() - ", err.Error())
		return "", err
	}
	id := strconv.FormatInt(info.MovementId, 10)
	return id, err
}

func (s *SouthXchange) GetDepositStatus(_ string, txId string, _ string) (hestia.ExchangeOrderInfo, error) {
	var status hestia.ExchangeOrderInfo
	txs, err := s.southClient.GetTransactions("deposits", 0, 1000, "", false)
	if err != nil {
		log.Println("south - GetDepositStatus - GetTransactions() - ", err.Error())
		return status, err
	}
	status.Status = hestia.ExchangeOrderStatusError
	for _, tx := range txs {
		if tx.Hash == txId {
			log.Println(tx)
			if tx.Status == "confirmed" || tx.Status == "executed" {
				status.Status = hestia.ExchangeOrderStatusCompleted
				status.ReceivedAmount = tx.Amount
				return status, nil
			} else if tx.Status == "pending" || tx.Status == "booked" {
				status.Status = hestia.ExchangeOrderStatusOpen
				return status, nil
			} else {
				return status, errors.New("south - GetDepositStatus - unknown status " + tx.Status)
			}
		}
	}
	return status, errors.New("south - transaction not found")
}

func (s *SouthXchange) GetOrderStatus(order hestia.Trade) (hestia.ExchangeOrderInfo, error) {
	var status hestia.ExchangeOrderInfo
	southOrder, err := s.southClient.GetOrder(order.OrderId)
	if err != nil {
		log.Println("south - GetOrderStatus - GetOrder() - ", err.Error())
		return status, err
	}
	if southOrder.Status == "executed" || southOrder.Status == "confirmed" {
		status.Status = hestia.ExchangeOrderStatusCompleted
		if southOrder.Type == "buy" {
			status.ReceivedAmount = southOrder.Amount * (1.0 - 0.003) // maker fee
		} else {
			status.ReceivedAmount = southOrder.Amount * southOrder.LimitPrice * (1.0 - 0.003)
		}
	} else if southOrder.Status == "pending" || southOrder.Status == "booked" || southOrder.Status == "partiallyexecuted"{
		status.Status = hestia.ExchangeOrderStatusOpen
	} else {
		status.Status = hestia.ExchangeOrderStatusError
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

func (s *SouthXchange) GetPair(fromCoin string, toCoin string) (models.TradeInfo, error) {
	var orderSide models.TradeInfo
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
	} else {
		orderSide.Type = "buy"
	}

	return orderSide, nil
}