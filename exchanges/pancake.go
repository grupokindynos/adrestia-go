package exchanges

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/grupokindynos/adrestia-go/models"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/hestia"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type Pancake struct {
	Name string
	password string
	url string
}

func NewPancake(params models.ExchangeParams) *Pancake {
	p := new(Pancake)
	p.Name = "pancake_swap"
	p.url = os.Getenv("PANCAKE_URL")
	p.password = os.Getenv("PANCAKE_PASSWORD")

	return p
}

func (p *Pancake) GetName() (string, error) {
	return p.Name, nil
}

func (p *Pancake) GetAddress(coin string) (string, error) {
	path := p.url + fmt.Sprintf("/api/v1/address")
	req, err := http.NewRequest("GET", path, nil)
	if req != nil {
		req.Header.Add("api-key", os.Getenv("PANCAKE_PASSWORD"))
	}
	if err != nil {
		return "", err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	data, _ := ioutil.ReadAll(res.Body)
	var address BSCAddressResponse
	err = json.Unmarshal(data, &address)
	if err != nil  || address.Status != 200{
		return "", err
	}
	fmt.Println(res)
	return address.Data.Address, nil
}

func (p *Pancake) GetBalance(coin string) (float64, error) {  // tal vez la modifque para que solo regrese la que queremos
	path := p.url + fmt.Sprintf("/api/v1/balance/%s", coin)
	req, err := http.NewRequest("GET", path, nil)
	if req != nil {
		req.Header.Add("api-key", os.Getenv("PANCAKE_PASSWORD"))
	}
	if err != nil {
		return 0, err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	data, _ := ioutil.ReadAll(res.Body)
	var balance BSCBalanceResponse
	err = json.Unmarshal(data, &balance)
	if err != nil || balance.Status != 200{
		return 0, err
	}
	return balance.Data.Balance, nil
}


func (p *Pancake) SellAtMarketPrice(order hestia.Trade) (string, error) {
	path := p.url + fmt.Sprintf("/api/v1/swap")
	swap := BSCSwapInfo{
		CoinFrom:  order.FromCoin,
		AmountIn:  int64(order.Amount) * 1e18,
		AmountOut: 0,
	}
	swapJSON, _ := json.Marshal(swap)
	req, err := http.NewRequest("POST", path, bytes.NewReader(swapJSON))
	if req != nil {
		req.Header.Add("api-key", os.Getenv("PANCAKE_PASSWORD"))
	}
	if err != nil {
		return "", err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	data, _ := ioutil.ReadAll(res.Body)
	var txInfo BSCSwapResponse
	err = json.Unmarshal(data, &txInfo)
	if err != nil || txInfo.Status != 200{
		return "", err
	}
	return txInfo.Data.TxID, nil
}

func (p *Pancake) Withdraw(coin string, address string, amount float64) (string, error) {
	path := p.url + fmt.Sprintf("/api/v1/withdraw")
	withdrawData, _ := json.Marshal(BSCWithdrawInput{
		Address: address,
		Asset:   coin,
		Amount:  amount,
	})
	req, err := http.NewRequest("GET", path, bytes.NewReader(withdrawData))
	if req != nil {
		req.Header.Add("api-key", os.Getenv("PANCAKE_PASSWORD"))
	}
	if err != nil {
		return "", err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	data, _ := ioutil.ReadAll(res.Body)
	var balance BSCWithdrawResponse
	err = json.Unmarshal(data, &balance)
	if err != nil || balance.Status != 200{
		return "", err
	}
	return balance.Data.TxID, nil
}

func (p *Pancake) GetDepositStatus(_ string, txId string, asset string) (hestia.ExchangeOrderInfo, error) {
	path := p.url + fmt.Sprintf("/api/v1/tx/%s", txId)
	coinInfo, _ := coinfactory.GetCoin(asset)

	req, err := http.NewRequest("POST", path, nil)
	if req != nil {
		req.Header.Add("api-key", os.Getenv("PANCAKE_PASSWORD"))
	}
	if err != nil {
		return hestia.ExchangeOrderInfo{
			Status: hestia.ExchangeOrderStatusError,
			ReceivedAmount: 0.0,
		}, err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return hestia.ExchangeOrderInfo{
			Status: hestia.ExchangeOrderStatusError,
			ReceivedAmount: 0.0,
		}, err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	data, _ := ioutil.ReadAll(res.Body)
	var txInfo BSCTxInfoResponse
	err = json.Unmarshal(data, &txInfo)
	if err != nil || txInfo.Status != 200{
		return hestia.ExchangeOrderInfo{
			Status: hestia.ExchangeOrderStatusError,
			ReceivedAmount: 0.0,
		}, err
	}
	if txInfo.Data.Confirmations > int64(coinInfo.BlockchainInfo.MinConfirmations) {
		return hestia.ExchangeOrderInfo{
			Status: hestia.ExchangeOrderStatusCompleted,
			ReceivedAmount: txInfo.Data.ReceivedAmount,
		}, nil
	} else {
		return hestia.ExchangeOrderInfo{
			Status: hestia.ExchangeOrderStatusOpen,
			ReceivedAmount: txInfo.Data.ReceivedAmount,
		}, nil
	}
}

func (p *Pancake) GetOrderStatus(order hestia.Trade) (hestia.ExchangeOrderInfo, error) {
	path := p.url + fmt.Sprintf("/api/v1/exchange/%s", order.OrderId)
	coinInfo, _ := coinfactory.GetCoin(order.ToCoin)

	req, err := http.NewRequest("POST", path, nil)
	if req != nil {
		req.Header.Add("api-key", os.Getenv("PANCAKE_PASSWORD"))
	}
	if err != nil {
		return hestia.ExchangeOrderInfo{
			Status: hestia.ExchangeOrderStatusError,
			ReceivedAmount: 0.0,
		}, err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return hestia.ExchangeOrderInfo{
			Status: hestia.ExchangeOrderStatusError,
			ReceivedAmount: 0.0,
		}, err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	data, _ := ioutil.ReadAll(res.Body)
	var txInfo BSCTxInfoResponse
	err = json.Unmarshal(data, &txInfo)
	if err != nil || txInfo.Status != 200{
		return hestia.ExchangeOrderInfo{
			Status: hestia.ExchangeOrderStatusError,
			ReceivedAmount: 0.0,
		}, err
	}
	if txInfo.Data.Confirmations > int64(coinInfo.BlockchainInfo.MinConfirmations) {
		return hestia.ExchangeOrderInfo{
			Status: hestia.ExchangeOrderStatusCompleted,
			ReceivedAmount: txInfo.Data.ReceivedAmount,
		}, nil
	} else {
		return hestia.ExchangeOrderInfo{
			Status: hestia.ExchangeOrderStatusOpen,
			ReceivedAmount: txInfo.Data.ReceivedAmount,
		}, nil
	}
}

func (p *Pancake) GetWithdrawalTxHash(txId string, _ string) (string, error) {
	panic("not implemented")

	//txs, err := s.southClient.GetTransactions("", 0, 500, "", true)
	//if err != nil {
	//	log.Println("south - GetWithdrawalTxHash - GetTransactions() - ", err.Error())
	//	return "", err
	//}
	//tradeId, err := strconv.ParseInt(txId, 10, 64)
	//if err != nil {
	//	log.Println("south - GetWithdrawalTxHash - ParseInt() - ", err.Error())
	//	return "", err
	//}
	//for _, tx := range txs {
	//	if tx.MovementId == tradeId && tx.Type == "withdraw" {
	//		return tx.Hash, nil
	//	}
	//}
	//
	//return "", errors.New("south - withdrawal not found")
}

func (p *Pancake) GetPair(fromCoin string, toCoin string) (models.TradeInfo, error) {
	panic("not implemented")

	//var orderSide models.TradeInfo
	//fromCoin = strings.ToUpper(fromCoin)
	//toCoin = strings.ToUpper(toCoin)
	//books, err := s.southClient.GetMarketSummaries()
	//if err != nil {
	//	log.Println("south - GetPair - GetMarketSummaries() - ", err.Error())
	//	return orderSide, err
	//}
	//var bookName south.MarketSummary
	//for _, book := range books {
	//	if (book.Coin == fromCoin || book.Base == fromCoin) && (book.Coin == toCoin || book.Base == toCoin) {
	//		bookName = book
	//		break
	//	}
	//}
	//
	//orderSide.Book = bookName.Coin + bookName.Base
	//if bookName.Coin == fromCoin {
	//	orderSide.Type = "sell"
	//} else {
	//	orderSide.Type = "buy"
	//}
	//
	//return orderSide, nil
}

type BSCBase struct {
	Status int32 `json:"status"`
	Error string `json:"error"`
}

type BSCAddressResponse struct {
	BSCBase
	Data struct{
		Address string `json:"address"`
	} `json:"data"`
}

type BSCBalanceResponse struct {
	BSCBase
	Data struct{
		Balance float64 `json:"balance"`
	} `json:"data"`
}

type BSCSwapInfo struct {
	CoinFrom string `json:"coin_from"`
	AmountIn int64 `json:"amount_in"`
	AmountOut int64 `json:"amount_out"`
}

type BSCSwapResponse struct {
	BSCBase
	Data struct{
		Path []string `json:"path"`
		TxID string `json:"tx_id"`
	} `json:"data"`
}

type BSCTxInfoResponse struct {
	BSCBase
	Data struct{
		Block int64 `json:"block"`
		Confirmations int64 `json:"confirmations"`
		IsServiceTx bool `json:"service_tx"`
		TransactionHash string `json:"tx_hash"`
		ReceivedAmount float64 `json:"received_amount"`
	} `json:"data"`
}

type BSCWithdrawResponse struct {
	BSCBase
	Data struct{
		TxID string `json:"tx_id"`
	} `json:"data"`
}

type BSCWithdrawInput struct {
	Address string `json:"address"`
	Asset string `json:"asset"`
	Amount float64 `json:"amount"`
}