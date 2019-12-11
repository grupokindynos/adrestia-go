package exchanges

import (
	"context"
	"fmt"
	"github.com/grupokindynos/adrestia-go/models/balance"
	config2 "github.com/grupokindynos/adrestia-go/models/exchanges/config"
	"github.com/grupokindynos/adrestia-go/models/transaction"
	"github.com/grupokindynos/adrestia-go/utils"
	"github.com/joho/godotenv"
	l "log"
	"os"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/go-binance"
)

type Binance struct {
	Exchange
	AccountName  string
	binanceApi   binance.Binance
}

var BinanceInstance = NewBinance()

func NewBinance() *Binance {
	c := new(Binance)
	c.Name = "Binance"
	c.BaseUrl = ""
	data := GetSettings()
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "time", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	hmacSigner := &binance.HmacSigner{
		Key: []byte(data.PrivateApi),
	}
	ctx, _ := context.WithCancel(context.Background())
	// use second return value for cancelling request when shutting down the app

	// l.Println("Binance Service Building...")
	binanceService := binance.NewAPIService(
		"https://www.binance.com",
		data.PublicApi,
		hmacSigner,
		logger,
		ctx,
	)
	c.binanceApi = binance.NewBinance(binanceService)
	return c
}

func (b Binance) GetName() (string, error) {
	return b.Name, nil
}

func (b Binance) GetAddress(coin coins.Coin) (string, error) {
	/*var addresses = make(map[string]string)
	addresses["DASH"] = "XuVmLDmUHZCjaSjm8KfXkGVhRG8fVC3Jis"
	addresses["XZC"] = "aJUE5rLmGvSu9ThnWzUu4TpYgKPPgfbCAy"
	addresses["LTC"] = "LPZom4L6oTJ3JkRDJz6EYkdg9Bga9VrFFL"
	addresses["GRS"] = "FjC2vAtjhdPeWfjsKGxoxrfxEJw5KWNNmR"
	addresses["BTC"] = "157kMZrgThAmHrvinRLP4RKPC5AU4KdYKt"

	if val, ok := addresses[strings.ToUpper(coin.Tag)]; ok {
		return val, nil
	}*/
	address, err := b.binanceApi.DepositAddress(binance.AddressRequest{
		Asset: "btc",
		//RecvWindow: 5 * time.Second,
		//Status: true,
		Timestamp: time.Now(),
	})
	if err != nil {
		fmt.Println("binance exchange: ", err)
		return "", err
	}
	fmt.Println(*address)
	return address.Address, nil
}

func (b Binance) GetBalances() ([]balance.Balance, error) {
	s := fmt.Sprintf("[GetBalances] Retrieving Balances for coins at %s", b.Name)
	l.Println(s)
	var balances []balance.Balance
	res, err := b.binanceApi.Account(binance.AccountRequest{
		RecvWindow: 5 * time.Second,
		Timestamp:  time.Now(),
	})

	if err != nil {
		return balances, err
	}

	for _, asset := range res.Balances {
		rate, _ := obol.GetCoin2CoinRates("https://obol-rates.herokuapp.com/", "BTC", asset.Asset)
		var b = balance.Balance{
			Ticker:     		asset.Asset,
			ConfirmedBalance:   asset.Free,
			UnconfirmedBalance: asset.Locked,
			RateBTC:    		rate,
			DiffBTC:    		0,
			IsBalanced: 		false,
		}
		if b.GetTotalBalance() > 0.0 {
			balances = append(balances, b)
		}

	}
	s = utils.GetBalanceLog(balances, b.Name)
	l.Println(s)
	return balances, nil
}

func (b Binance) SellAtMarketPrice(SellOrder transaction.ExchangeSell) (bool, error) {
	l.Println(fmt.Sprintf("[SellAtMarketPrice] Selling %.8f %s for %s on %s", SellOrder.Amount, SellOrder.FromCoin.Name, SellOrder.ToCoin.Name, b.Name))
	// Gets price from Obol considering the amount to sell
	rate, err := obol.GetCoin2CoinRatesWithAmount("https://obol-rates.herokuapp.com/", SellOrder.FromCoin.Tag, SellOrder.ToCoin.Tag, fmt.Sprintf("%f", SellOrder.Amount))
	if err != nil {
		return false, err
	}

	// Order creation an Post
	symbol := SellOrder.FromCoin.Tag + SellOrder.ToCoin.Tag
	fmt.Println(symbol)
	fmt.Println(rate)
	// TODO Update to work with new Rate Response models rate.AveragePrice
	// TODO Log for order info
	// TODO Test Order Post for Binance
	/*newOrder, err := b.binanceApi.NewOrder(binance.NewOrderRequest{
		Symbol:      symbol,
		Quantity:    SellOrder.Amount,
		Price:       1/rate,
		Side:        binance.SideSell,
		TimeInForce: binance.IOC, // Immediate OR Cancel - orders fills all or part of an order immediately and cancels the remaining part of the order.
		Type:        binance.TypeLimit,
		Timestamp:   time.Now(),
	})
	if err != nil {
		panic(err)
		// TODO Save failed order to Hestia DB
	}
	fmt.Println(newOrder)*/

	return true, nil
}

func (b Binance) Withdraw(coin coins.Coin, address string, amount float64) (bool, error) {
	// l.Println(fmt.Sprintf("[Withdraw] Retrieving Account Info for %s", b.Name))
	/*res, _ := b.binanceApi.Account(binance.AccountRequest{
		RecvWindow: 5 * time.Second,
		Timestamp:  time.Now(),
	})*/

	l.Println(fmt.Sprintf("[Withdraw] Performing withdraw request on %s for %s", b.Name, coin.Tag))
	withdrawal, err := b.binanceApi.Withdraw(binance.WithdrawRequest{
		Asset:      strings.ToLower(coin.Tag),
		Address:    address,
		Amount:     amount,
		Name:       "Adrestia-go Withdrawal",
		RecvWindow: 5 * time.Second,
		Timestamp:  time.Now(),
	})

	if err != nil {
		l.Println(fmt.Sprintf("[Withdraw] Binance failed to withdraw %s", err))
		return false, err
	}
	// TODO Binance go library has an issue signing withdrawals
	// fmt.Println(withdrawal)
	// fmt.Println(err)

	return withdrawal.Success, nil

}

// TODO Missing
func (b Binance) OneCoinToBtc(coin coins.Coin) (float64, error) {
	l.Println(fmt.Sprintf("[OneCoinToBtc] Calculating for %s using %s", coin.Name, b.Name))
	if coin.Tag == "BTC" {
		return 1.0, nil
	}
	// TODO Missing update on method, not strictly needed though
	rate, err := obol.GetCoin2CoinRatesWithAmount("https://obol-rates.herokuapp.com/", "btc", coin.Tag, fmt.Sprintf("%f", 1.0))
	if err != nil {
		return 0.0, err
	}
	l.Println(fmt.Sprintf("[OneCoinToBtc] Calculated rate for %s as %.8f BTC per Coin", coin.Name, rate))
	return rate.AveragePrice, nil
}

func GetSettings() config2.BinanceAuth {
	if err := godotenv.Load(); err != nil {
		l.Println(err)
	}
	// l.Println(fmt.Sprintf("[GetSettings] Retrieving settings for Binance"))
	var data config2.BinanceAuth
	data.PublicApi = os.Getenv("BINANCE_PUB_API")
	data.PrivateApi = os.Getenv("BINANCE_PRIV_API")
	data.PublicWithdrawKey = os.Getenv("BINANCE_PUB_WITHDRAW")
	data.PrivateWithdrawKey = os.Getenv("BINANCE_PRIV_WITHDRAW")
	return data
}
