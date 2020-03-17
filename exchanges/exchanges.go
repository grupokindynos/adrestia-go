package exchanges

import (
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/hestia"
)

type IExchange interface {
	GetName() (string, error) // Returns the name of the exchange
	GetAddress(coin coins.Coin) (string, error) // Retrieves an address for the specified coin, must return an address every time.
	GetBalances() ([]balance.Balance, error) // Returns the balance of all of the assets in an exchange converting them to the type balance.Balance
	SellAtMarketPrice(sellOrder hestia.ExchangeOrder) (string, error) // Sells an order at market price and returns the order ID the exchange sets for the order.
	Withdraw(coin coins.Coin, address string, amount float64) (string, error) // Withdraws a determined amount of a coin to a specified address from an exchange. Must return the txid.
	GetOrderStatus(order hestia.ExchangeOrder) (hestia.OrderStatus, error) // Returns the status of an order posted to an exchange. The ExchangeOrder object contains the order ID and the exchange, returns the status as an enum
	GetPair(fromCoin string, toCoin string) (OrderSide, error) // Returns the book information and side for a desired trade in an exchange
	GetWithdrawalTxHash(txId string, asset string) (string, error) // Retrieves the TxHash from an Exchange Withdrawal
	GetDepositStatus(txid string, asset string) (hestia.OrderStatus, error) // Returns the state of a deposit of an exchange
}

type IExchangeFactory interface {
	GetExchangeByCoin(coin coins.Coin) (IExchange, error)
	GetExchangeByName(name string) (IExchange, error)
}

type Exchange struct {
	IExchange
	Name    string
	BaseUrl string
}
