package services

import (
	"github.com/grupokindynos/hestia"
	"http"
)

type HestiaInstance struct {
	HestiaURL string
}

func (h *HestiaInstance) GetAdrestiaCoins() (availableCoins []hestia.Coin, err error) {
	req, err := mvt.CreateMVTToken("GET", h.HestiaURL+"/coins", "adrestia", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"))
	if err != nil {
		return []hestia.Coin{}, err
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return []hestia.Coin{}, err
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []hestia.Coin{}, err
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		return []hestia.Coin{}, err
	}
	headerSignature := res.Header.Get("service")
	if headerSignature == "" {
		return []hestia.Coin{}, errors.New("no header signature")
	}
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("HESTIA_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return []hestia.Coin{}, err
	}
	var response []hestia.Coin
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return []hestia.Coin{}, err
	}
	for _, coin := range response {
		if coin.Adrestia {
			availableCoins = append(availableCoins, coin)
		}
	}
	return availableCoins, nil
}