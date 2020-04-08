package controllers

import (
	"encoding/json"
	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/models"
	"github.com/grupokindynos/adrestia-go/services"
	coinfactory "github.com/grupokindynos/common/coin-factory"
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
		Address:  address,
		Exchange: exName,
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
	var inPath []models.MockTrade
	var outPath []models.MockTrade


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
			inPath = append(inPath, models.MockTrade{
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
		inPath = append(inPath, models.MockTrade{
			FromCoin: pathParams.FromCoin,
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
	var exOutwardInfo hestia.ExchangeInfo
	for _, ex := range a.ExInfo {
		if ex.Name == exNameTarget {
			exOutwardInfo = ex
			break
		}
	}

	if targetCoinInfo.Info.StableCoin && pathParams.ToCoin == exOutwardInfo.StockCurrency {
		// target coin is the exchange's stock coin
	} else {
		outPath = append(outPath, models.MockTrade{
			FromCoin: exOutwardInfo.StockCurrency,
			ToCoin:   "BTC",
			Exchange: exNameTarget,
		})
		if pathParams.ToCoin != "BTC" {
			outPath = append(outPath, models.MockTrade{
				FromCoin: "BTC",
				ToCoin: pathParams.ToCoin,
				Exchange: exNameTarget,
			})
		}
	}
	// If origin coin is not BTC Convert first


	/*
	pairInfo, err := ex.GetPair(pathParams.FromCoin, pathParams.ToCoin)
	// TODO Sanity check
	 */
	/*if err != nil {
		// TODO pair does not exist
	}
	address, err := ex.GetAddress(pathParams.FromCoin)
	if err != nil {
		return nil, err
	}
	response := models.AddressResponse{
		Coin:     params.Coin,
		Address:  address,
		Exchange: exName,
	} */
	path.InwardOrder = inPath
	path.OutwardOrder = outPath
	return path, nil
}