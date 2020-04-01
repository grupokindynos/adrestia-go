package controllers

import (
	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/models"
	"github.com/grupokindynos/adrestia-go/services"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/obol"
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
	address, err := ex.GetAddress(params.Coin)
	if err != nil {
		return nil, err
	}
	return address, nil
}