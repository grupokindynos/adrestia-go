package controllers

import (
	"encoding/json"
	"errors"
	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/models"
	"github.com/grupokindynos/adrestia-go/services"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	cerror "github.com/grupokindynos/common/errors"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	"log"
	"sync"
)

type AdrestiaController struct {
	//PrepareShifts map[string]models.PrepareShiftInfo
	mapLock       sync.RWMutex
	Hestia        services.HestiaRequests
	Plutus        services.PlutusService
	Obol          obol.ObolService
	DevMode       bool
	ExFactory     *exchanges.ExchangeFactory
	ExInfo 		  []hestia.ExchangeInfo
}

func (a *AdrestiaController) Withdraw(_ string, body []byte, params models.Params) (interface{}, error) {
	var withdrawParams models.WithdrawParams
	err := json.Unmarshal(body, &withdrawParams)
	if err != nil {
		return nil, err
	}
	coinInfo, err := coinfactory.GetCoin(withdrawParams.Asset)
	if err != nil {
		return nil, err
	}
	ex, err := a.ExFactory.GetExchangeByCoin(*coinInfo)
	if err != nil {
		return nil, err
	}
	txid, err := ex.Withdraw(withdrawParams.Asset, withdrawParams.Address, withdrawParams.Amount)
	if err != nil {
		return nil, err
	}
	exName, err := ex.GetName()
	response := models.WithdrawInfo{
		Exchange: exName,
		Asset:    withdrawParams.Asset,
		TxId:     txid,
	}
	return response, nil
}

func (a *AdrestiaController) WithdrawV2(_ string, body []byte, params models.Params) (interface{}, error) {
	var withdrawParams models.WithdrawParamsV2
	var exchange exchanges.Exchange
	var err error

	err = json.Unmarshal(body, &withdrawParams)
	if err != nil {
		return nil, err
	}
	if withdrawParams.Exchange == "" {
		coinInfo, err := coinfactory.GetCoin(withdrawParams.Asset)
		if err != nil {
			return nil, err
		}
		exchange, err = a.ExFactory.GetExchangeByCoin(*coinInfo)
		if err != nil {
			return nil, err
		}
	} else {
		exchange, err = a.ExFactory.GetExchangeByName(withdrawParams.Exchange)
		if err != nil {
			return nil, err
		}
	}
	txid, err := exchange.Withdraw(withdrawParams.Asset, withdrawParams.Address, withdrawParams.Amount)
	if err != nil {
		return nil, err
	}
	exName, err := exchange.GetName()
	response := models.WithdrawInfo{
		Exchange: exName,
		Asset:    withdrawParams.Asset,
		TxId:     txid,
	}
	return response, nil
}


func (a *AdrestiaController) GetTradeStatus(_ string, body []byte, _ models.Params) (interface{}, error) {
	var trade hestia.Trade
	err := json.Unmarshal(body, &trade)
	if err != nil {
		return nil, err
	}
	exchange, err := a.ExFactory.GetExchangeByName(trade.Exchange)
	if err != nil {
		return nil, err
	}
	return exchange.GetOrderStatus(trade)
}

func (a *AdrestiaController) GetWithdrawalTxHash(_ string, body []byte, _ models.Params) (interface{}, error) {
	var withdrawInfo models.WithdrawInfo
	err := json.Unmarshal(body, &withdrawInfo)
	if err != nil {
		return "", err
	}
	exchange, err := a.ExFactory.GetExchangeByName(withdrawInfo.Exchange)
	if err != nil {
		return "", err
	}
	return exchange.GetWithdrawalTxHash(withdrawInfo.TxId, withdrawInfo.Asset)
}

func (a *AdrestiaController) GetAddress(_ string, _ []byte, params models.Params) (interface{}, error) {
	coinInfo, err := coinfactory.GetCoin(params.Coin)
	if err != nil {
		return nil, err
	}
	ex, err := a.ExFactory.GetExchangeByCoin(*coinInfo)
	if err != nil {
		return nil, err
	}
	exName, err := ex.GetName()
	if err != nil {
		return nil, err
	}
	address, err := ex.GetAddress(params.Coin)
	if err != nil {
		return nil, err
	}
	response := models.AddressResponse{
		Coin:     params.Coin,
		ExchangeAddress: models.ExchangeAddress{
			Address:  address,
			Exchange: exName,
		},
	}
	return response, nil
}

func (a *AdrestiaController) GetConversionPath(_ string, body []byte, _ models.Params) (interface{}, error) {
	var pathParams models.PathParams
	err := json.Unmarshal(body, &pathParams)
	if err != nil {
		return nil, err
	}
	
	// Response Object
	var path models.PathResponse
	var inPath []models.ExchangeTrade
	var outPath []models.ExchangeTrade


	coinInfo, err := coinfactory.GetCoin(pathParams.FromCoin)
	if err != nil {
		return nil, err
	}
	ex, err := a.ExFactory.GetExchangeByCoin(*coinInfo)
	if err != nil {
		return nil, err
	}
	exName, err := ex.GetName()
	if err != nil {
		return nil, err
	}

	if coinInfo.Info.StableCoin {
		log.Println("payment already in stable coin")
	} else {
		if pathParams.FromCoin != "BTC" {
			log.Println("requires btc conversion")
			inPath = append(inPath, models.ExchangeTrade{
				FromCoin: pathParams.FromCoin,
				ToCoin:   "BTC",
				Exchange: exName,
			})
		}
		log.Println("requires btc to stable coin conversion")
		var exInwardInfo hestia.ExchangeInfo
		for _, ex := range a.ExInfo {
			if ex.Name == exName {
				exInwardInfo = ex
				break
			}
		}
		inPath = append(inPath, models.ExchangeTrade{
			FromCoin: "BTC",
			ToCoin:   exInwardInfo.StockCurrency,
			Exchange: exName,
		})
	}

	targetCoinInfo, err := coinfactory.GetCoin(pathParams.ToCoin)
	if err != nil {
		return nil, err
	}
	exTarget, err := a.ExFactory.GetExchangeByCoin(*targetCoinInfo)
	if err != nil {
		return nil, err
	}
	exNameTarget, err := exTarget.GetName()
	if err != nil {
		return nil, err
	}
	exOutwardInfo, err := a.getExchangeInfo(exNameTarget)

	if targetCoinInfo.Info.StableCoin && pathParams.ToCoin == exOutwardInfo.StockCurrency {
		// target coin is the exchange's stock coin
	} else {
		outPath = append(outPath, models.ExchangeTrade{
			FromCoin: exOutwardInfo.StockCurrency,
			ToCoin:   "BTC",
			Exchange: exNameTarget,
		})
		if pathParams.ToCoin != "BTC" {
			outPath = append(outPath, models.ExchangeTrade{
				FromCoin: "BTC",
				ToCoin: pathParams.ToCoin,
				Exchange: exNameTarget,
			})
		}
	}
	// If origin coin is not BTC Convert first
	tradeFlag := true;
	log.Println("CHECKING INPUT ORDER")
	for i, trade := range inPath {
		pairInfo, err := ex.GetPair(trade.FromCoin, trade.ToCoin)
		if err != nil{
			log.Println("could not find the desired trading pair for ", trade)
			tradeFlag = false
		} else {
			inPath[i].Trade = pairInfo
		}
	}

	log.Println("CHECKING OUTPUT ORDER")
	for i, trade := range outPath {
		pairInfo, err := exTarget.GetPair(trade.FromCoin, trade.ToCoin)
		if err != nil{
			log.Println("could not find the desired trading pair for ", trade)
			tradeFlag = false
		} else {
			outPath[i].Trade = pairInfo
		}
	}

	path.InwardOrder = inPath
	path.OutwardOrder = outPath
	path.Trade = tradeFlag
	return path, nil
}

func (a *AdrestiaController) GetVoucherConversionPath(_ string, body []byte, _ models.Params) (interface{}, error) {
	var pathParams models.VoucherPathParams
	err := json.Unmarshal(body, &pathParams)
	if err != nil {
		log.Println("GetVoucherConversionPath::Unmarshal::", body)
		return nil, err
	}

	// Response Object
	var path models.VoucherPathResponse
	var inPath []models.ExchangeTrade
	var exInwardInfo hestia.ExchangeInfo

	coinInfo, err := coinfactory.GetCoin(pathParams.FromCoin)
	if err != nil {
		log.Println("GetVoucherConversionPath::GetCoin::", pathParams)
		return nil, err
	}
	ex, err := a.ExFactory.GetExchangeByCoin(*coinInfo)
	if err != nil {
		log.Println("GetVoucherConversionPath::GetExchangeByCoin::", coinInfo.Info.Name, "::", ex)
		return nil, err
	}
	exName, err := ex.GetName()
	if err != nil {
		log.Println("GetVoucherConversionPath::GetName::", err)
		return nil, err
	}

	address, err := ex.GetAddress(pathParams.FromCoin)
	if err != nil || address == "" {
		log.Println("GetVoucherConversionPath::GetAddress::", exName)
		if err != nil {
			log.Println(err)
		}
		return nil, errors.New("adrestia could not retrieve address")
	}
	if coinInfo.Info.StableCoin {
		log.Println("payment already in stable coin")
	} else {
		if pathParams.FromCoin != "BTC" {
			inPath = append(inPath, models.ExchangeTrade{
				FromCoin: pathParams.FromCoin,
				ToCoin:   "BTC",
				Exchange: exName,
			})
		}
		for _, ex := range a.ExInfo {
			if ex.Name == exName {
				exInwardInfo = ex
				break
			}
		}
		inPath = append(inPath, models.ExchangeTrade{
			FromCoin: "BTC",
			ToCoin:   exInwardInfo.StockCurrency,
			Exchange: exName,
		})
	}
	// If origin coin is not BTC Convert first
	tradeFlag := true
	for i, trade := range inPath {
		pairInfo, err := ex.GetPair(trade.FromCoin, trade.ToCoin)
		if err != nil{
			log.Println("could not find the desired trading pair for ", trade)
			tradeFlag = false
		} else {
			inPath[i].Trade = pairInfo
		}
	}

	path.InwardOrder = inPath
	path.Trade = tradeFlag
	path.TargetStableCoin = exInwardInfo.StockCurrency
	path.Address = address
	return path, nil
}


func (a *AdrestiaController) GetVoucherConversionPathV2(_ string, body []byte, _ models.Params) (interface{}, error) {
	var pathParams models.VoucherPathParamsV2
	err := json.Unmarshal(body, &pathParams)
	if err != nil {
		log.Println("GetVoucherConversionPath::Unmarshal::", body)
		return nil, err
	}

	// Response Object
	var path models.VoucherPathResponse
	var inPath []models.ExchangeTrade
	var exInwardInfo hestia.ExchangeInfo

	coinInfo, err := coinfactory.GetCoin(pathParams.FromCoin)
	if err != nil {
		log.Println("GetVoucherConversionPath::GetCoin::", pathParams)
		return nil, err
	}
	ex, err := a.ExFactory.GetExchangeByCoin(*coinInfo)
	if err != nil {
		log.Println("GetVoucherConversionPath::GetExchangeByCoin::", coinInfo.Info.Name, "::", ex)
		return nil, err
	}
	exName, err := ex.GetName()
	if err != nil {
		log.Println("GetVoucherConversionPath::GetName::", err)
		return nil, err
	}

	// Conversion of values less than 10 USDT is not possible on binance
	if exName == "binance" && pathParams.AmountEuro < 10.0 {
		if coinInfo.Rates.FallBackExchange == "" {
			return nil, cerror.ErrorNotSupportedAmount
		}

		ex, err = a.ExFactory.GetExchangeByName(coinInfo.Rates.FallBackExchange)
		if err != nil {
			return nil, err
		}
		exName, err = ex.GetName()
		if err != nil {
			return nil, err
		}
	}

	directConversion := HasDirectConversionToStableCoin(exName, pathParams.FromCoin)

	address, err := ex.GetAddress(pathParams.FromCoin)
	if err != nil || address == "" {
		log.Println("GetVoucherConversionPath::GetAddress::", exName)
		if err != nil {
			log.Println(err)
		}
		return nil, errors.New("adrestia could not retrieve address")
	}
	if coinInfo.Info.StableCoin {
		log.Println("payment already in stable coin")
	} else {
		for _, ex := range a.ExInfo {
			if ex.Name == exName {
				exInwardInfo = ex
				break
			}
		}

		if directConversion {
			inPath = append(inPath, models.ExchangeTrade{
				FromCoin: pathParams.FromCoin,
				ToCoin:   exInwardInfo.StockCurrency,
				Exchange: exName,
			})
		} else {
			if pathParams.FromCoin != "BTC" {
				inPath = append(inPath, models.ExchangeTrade{
					FromCoin: pathParams.FromCoin,
					ToCoin:   "BTC",
					Exchange: exName,
				})
			}
			inPath = append(inPath, models.ExchangeTrade{
				FromCoin: "BTC",
				ToCoin:   exInwardInfo.StockCurrency,
				Exchange: exName,
			})
		}
	}
	// If origin coin is not BTC Convert first
	tradeFlag := true
	for i, trade := range inPath {
		pairInfo, err := ex.GetPair(trade.FromCoin, trade.ToCoin)
		if err != nil{
			log.Println("could not find the desired trading pair for ", trade)
			tradeFlag = false
		} else {
			inPath[i].Trade = pairInfo
		}
	}

	path.InwardOrder = inPath
	path.Trade = tradeFlag
	path.TargetStableCoin = exInwardInfo.StockCurrency
	path.Address = address
	return path, nil
}


func (a *AdrestiaController) Trade(_ string, body []byte, _ models.Params) (interface{}, error) {
	var trade hestia.Trade
	err := json.Unmarshal(body, &trade)
	if err != nil {
		return "", err
	}
	exchange, err := a.ExFactory.GetExchangeByName(trade.Exchange)
	if err != nil {
		return "", err
	}
	txId, err := exchange.SellAtMarketPrice(trade)
	if err != nil {
		return "", err
	}
	return txId, nil
}

/*
Returns the deposit status for a given txid.
*/
func (a *AdrestiaController) Deposit(_ string, body []byte, params models.Params) (interface{}, error) {
	var depositParams models.DepositParams
	err := json.Unmarshal(body, &depositParams)
	if err != nil {
		return nil, err
	}
	coinInfo, err := coinfactory.GetCoin(depositParams.Asset)
	if err != nil {
		return nil, err
	}
	ex, err := a.ExFactory.GetExchangeByCoin(*coinInfo)
	if err != nil {
		return nil, err
	}
	exOrderInfo, err := ex.GetDepositStatus(depositParams.Address, depositParams.TxId, depositParams.Asset)
	if err != nil {
		name, _ := ex.GetName()
		log.Println("Deposit::GetDepositStatus::", name, "::", err)
		return nil, err
	}
	exName, err := ex.GetName()
	response := models.DepositInfo{
		Exchange: exName,
		DepositInfo:    exOrderInfo,
	}
	return response, nil
}

func (a *AdrestiaController) getExchangeInfo(exchange string) (info hestia.ExchangeInfo, err error) {
	for _, ex := range a.ExInfo {
		if ex.Name == exchange {
			info = ex
			return
		}
	}
	return info, errors.New("exchange not found")
}

/*
Returns an exchange's stock balance, given an input coin.
*/
func (a *AdrestiaController) StockBalance(_ string, body []byte, params models.Params) (interface{}, error) {
	coinInfo, err := coinfactory.GetCoin(params.Coin)
	if err != nil {
		return nil, err
	}
	ex, err := a.ExFactory.GetExchangeByCoin(*coinInfo)
	if err != nil {
		return nil, err
	}
	exName, err := ex.GetName()
	if err != nil {
		return nil, err
	}
	exInfo, err := a.getExchangeInfo(exName)
	if err != nil {
		return nil, err
	}
	// stockCoinInfo, err := coinfactory.GetCoin(exInfo.StockCurrency)
	balance, err := ex.GetBalance(exInfo.StockCurrency)
	if err != nil {
		return nil, err
	}
	response := models.BalanceResponse{
		Exchange: exName,
		Balance:  balance,
		Asset: exInfo.StockCurrency,
	}
	return response, nil
}