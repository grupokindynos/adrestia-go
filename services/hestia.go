package services

import (
	"encoding/json"
	"errors"
	"github.com/google/go-querystring/query"
	"github.com/grupokindynos/adrestia-go/models/adrestia"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/tokens/mrt"
	"github.com/grupokindynos/common/tokens/mvt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type HestiaRequests struct {
	HestiaURL string
}

func (h *HestiaRequests) GetAdrestiaCoins() (availableCoins []hestia.Coin, err error) {
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
	// fmt.Println("Hestia Conf: ", response)
	for _, coin := range response {
		if coin.Adrestia {
			availableCoins = append(availableCoins, coin)
		}
	}
	return availableCoins, nil
}

func (h *HestiaRequests) GetBalancingOrders() ([]hestia.AdrestiaOrder, error) {
	req, err := mvt.CreateMVTToken("GET", h.HestiaURL+"/adrestia/orders", "adrestia", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"))
	if err != nil {
		return []hestia.AdrestiaOrder{}, err
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return []hestia.AdrestiaOrder{}, err
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []hestia.AdrestiaOrder{}, err
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		return []hestia.AdrestiaOrder{}, err
	}
	headerSignature := res.Header.Get("service")
	if headerSignature == "" {
		return []hestia.AdrestiaOrder{}, errors.New("no header signature")
	}
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("HESTIA_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return []hestia.AdrestiaOrder{}, err
	}
	var response []hestia.AdrestiaOrder
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return []hestia.AdrestiaOrder{}, err
	}
	// fmt.Println("Hestia Conf: ", response)
	return response, nil
}

func (h *HestiaRequests) CreateAdrestiaOrder(orderData hestia.AdrestiaOrder) (string, error) {
	req, err := mvt.CreateMVTToken("POST", h.HestiaURL+"/adrestia/new", "adrestia", os.Getenv("MASTER_PASSWORD"), orderData, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"))
	if err != nil {
		return "", err
	}
	client := http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       time.Second * 30,
	}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		return "", err
	}
	headerSignature := res.Header.Get("service")
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("HESTIA_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return "", err
	}
	var response string
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return "", err
	}
	return response, nil
}

func (h *HestiaRequests) UpdateAdrestiaOrder(orderData hestia.AdrestiaOrder) (string, error) {
	req, err := mvt.CreateMVTToken("PUT", h.HestiaURL+"/adrestia/update", "adrestia", os.Getenv("MASTER_PASSWORD"), orderData, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"))
	if err != nil {
		return "", err
	}
	client := http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       time.Second * 30,
	}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		log.Println("UpdateAdrestiaOrder:: error unmarshalling received", string(tokenResponse))
		return "", err
	}
	headerSignature := res.Header.Get("service")
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("HESTIA_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return "", err
	}
	var response string
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return "", err
	}
	return response, nil
}

func (h *HestiaRequests) UpdateAdrestiaOrderStatus(orderData hestia.AdrestiaOrderUpdate) (string, error) {
	req, err := mvt.CreateMVTToken("PUT", h.HestiaURL+"/adrestia/update/status", "adrestia", os.Getenv("MASTER_PASSWORD"), orderData, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"))
	if err != nil {
		return "", err
	}
	client := http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       time.Second * 30,
	}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		return "", err
	}
	headerSignature := res.Header.Get("service")
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("HESTIA_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return "", err
	}
	var response string
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return "", err
	}
	return response, nil
}

func (h *HestiaRequests) GetAllOrders(adrestiaOrderParams adrestia.OrderParams) ([]hestia.AdrestiaOrder, error) {
	req, err := mvt.CreateMVTToken(http.MethodGet, h.HestiaURL+"/adrestia/orders", "adrestia", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"))
	if err != nil {
		return nil, err
	}
	client := http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       time.Second * 30,
	}
	q := req.URL.RawQuery
	val, err := query.Values(adrestiaOrderParams)
	if err != nil {
		return nil, errors.New("problem with query parameters")
	}
	req.URL.RawQuery = q + val.Encode() // add encoded values
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		return nil, err
	}
	headerSignature := res.Header.Get("service")
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("HESTIA_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return nil, err
	}
	var response []hestia.AdrestiaOrder
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}
