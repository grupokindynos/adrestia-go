package processor

import (
	"errors"
	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/services"
	cf "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/explorer"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	"math"
	"strconv"
)

type Params struct {
	Hestia services.HestiaService
	Plutus services.PlutusService
	Obol   obol.ObolService
}

func getPlutusReceivedAmount(addr string, txId string) (float64, error) {
	var blockExplorer explorer.BlockBook
	coin, err := cf.GetCoin("ETH")
	if err != nil {
		return 0.0, errors.New("Unable to get ETH coin")
	}
	blockExplorer.Url = "https://" + coin.BlockchainInfo.ExternalSource
	addrInfo, err := blockExplorer.GetEthAddress(addr)
	if err != nil {
		return 0.0, errors.New("Error while getting eth address from blockbook " + err.Error())
	}
	receivedAmount := 0.0
	found := false
	for _, txInfo := range addrInfo.Transactions {
		if txInfo.Txid == txId && txInfo.Confirmations > 0 {
			found = true
			for _, tokenTxInfo := range txInfo.TokenTransfers {
				val, _ := strconv.ParseFloat(tokenTxInfo.Value, 64)
				receivedAmount += val / (math.Pow10(tokenTxInfo.Decimals) * 1.0)
			}
			break
		}
	}
	if found {
		return receivedAmount, nil
	}

	return 0.0, errors.New("tx not found or still not confirmed")
}

func getBalance(exFac *exchanges.ExchangeFactory, exchangeName string, currency string) (float64, error) {
	exchange, err := exFac.GetExchangeByName(exchangeName, hestia.ShiftAccount)
	if err != nil {
		return 0, nil
	}
	bal, err := exchange.GetBalance(currency)
	if err != nil {
		return 0, nil
	}
	return bal, nil
}

// Returns stock balance of all exchanges without the pending amount of running shifts
func GetStockBalancesWithoutPendingShifts(h services.HestiaService, exInfo []hestia.ExchangeInfo, exFactory *exchanges.ExchangeFactory) (map[string]float64, error) {
	mp := make(map[string]float64)
	openShifts, err := h.GetOpenShifts("") // if empty returns open shifts from a day ago
	if err != nil {
		return nil, err
	}

	for _, shift := range openShifts {
		if len(shift.OutboundTrade.Conversions) == 0 { // should be in withdrawn status
			if shift.OutboundTrade.Status < hestia.ShiftV2TradeStatusWithdrawn {
				mp[shift.OutboundTrade.Exchange] += float64(shift.ToAmount) * 1e-8
			}
		} else if shift.OutboundTrade.Status == hestia.ShiftV2TradeStatusCreated {
			mp[shift.OutboundTrade.Exchange] += shift.OutboundTrade.Conversions[0].Amount
		}
	}

	for _, exchange := range exInfo {
		ex, err := exFactory.GetExchangeByName(exchange.Name, hestia.ShiftAccount)
		if err != nil {
			return nil, err
		}
		bal, err := ex.GetBalance(exchange.StockCurrency)
		if err != nil {
			return nil, err
		}
		mp[exchange.Name] = bal - mp[exchange.Name]
	}

	return mp, nil
}
