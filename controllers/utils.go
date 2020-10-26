package controllers


func HasDirectConversionToStableCoin(exchange string, coin string) bool {
	whiteList := getWhiteListMarketsMap(exchange)
	if whiteList == nil {
		return false
	}

	return whiteList[coin]
}

func StayInBTC(exchange string, _ string) bool {
	return exchange == "southxchange"
}

func getWhiteListMarketsMap(exchange string) map[string]bool {
	whiteList := map[string]map[string]bool{
		"binance": {
			"BTC": true,
			"DASH": true,
			"ETH":  true,
			"LTC":  true,
			"XZC":  true,
			"BAT":  true,
			"LINK": true,
			"NULS": true,
		},
		"bittrex": {
			"DGB": true,
		},
		"stex": {
			"DIVI": true,
		},
		"bithumb": {
			"GTH": true,
		},
	}

	return whiteList[exchange]
}
