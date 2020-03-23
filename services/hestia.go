package services

import (
	"encoding/json"
	"errors"
	"github.com/google/go-querystring/query"
	"github.com/grupokindynos/adrestia-go/models"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/tokens/mrt"
	"github.com/grupokindynos/common/tokens/mvt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type HestiaInstance struct {
	HestiaURL string
}

func (h *HestiaInstance) UpdateExchangeBalance(exchange string, amount float64) (string, error) {
	exchangeInfo, err := h.GetExchange(exchange)
	if err != nil{
		return "", err
	}
	exchangeInfo.StockAmount = amount
	req, err := mvt.CreateMVTToken("PUT", h.HestiaURL+"/exchanges/update", "adrestia", os.Getenv("MASTER_PASSWORD"), exchangeInfo, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"))
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

func (h *HestiaInstance) GetExchange(exchange string) (hestia.ExchangeInfo, error) {
	payload, err := h.get("/exchange", models.GetFilters{Id:exchange})
	if err != nil {
		return hestia.ExchangeInfo{}, err
	}
	var response hestia.ExchangeInfo
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return hestia.ExchangeInfo{}, err
	}
	return response, nil
}

func (h *HestiaInstance) GetExchanges() (exchangesInfo []hestia.ExchangeInfo, err error) {
	payload, err := h.get("/exchanges", models.GetFilters{})
	if err != nil {
		return nil, err
	}
	var response []hestia.ExchangeInfo
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (h *HestiaInstance) GetDeposits(includeComplete bool, sinceTimestamp int64) ([]hestia.SimpleTx, error) {
	payload, err := h.get("/adrestia/deposits", models.GetFilters{IncludeComplete:includeComplete, AddedSince:sinceTimestamp})
	if err != nil {
		return nil, err
	}
	var response []hestia.SimpleTx
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (h *HestiaInstance) GetWithdrawals(includeComplete bool, sinceTimestamp int64) ([]hestia.SimpleTx, error) {
	payload, err := h.get("/adrestia/withdrawals", models.GetFilters{IncludeComplete:includeComplete, AddedSince:sinceTimestamp})
	if err != nil {
		return nil, err
	}
	var response []hestia.SimpleTx
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (h *HestiaInstance) GetBalanceOrders(includeComplete bool, sinceTimestamp int64) ([]hestia.BalancerOrder, error) {
	payload, err := h.get("/adrestia/orders", models.GetFilters{IncludeComplete:includeComplete, AddedSince:sinceTimestamp})
	if err != nil {
		return nil, err
	}
	var response []hestia.BalancerOrder
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (h *HestiaInstance) GetBalancer() (hestia.Balancer, error) {
	payload, err := h.get("/adrestia/balancer", models.GetFilters{})
	if err != nil {
		return hestia.Balancer{}, err
	}
	var response []hestia.Balancer
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return hestia.Balancer{}, err
	}
	return response[0], nil
}

func (h *HestiaInstance) CreateDeposit(simpleTx hestia.SimpleTx) (string, error) {
	return h.doSimpleTx("POST", "/adrestia/new/deposit", simpleTx)
}

func (h *HestiaInstance) CreateWithdrawal(simpleTx hestia.SimpleTx) (string, error) {
	return h.doSimpleTx("POST", "/adrestia/new/withdrawal", simpleTx)
}

func (h *HestiaInstance) UpdateDeposit(simpleTx hestia.SimpleTx) (string, error) {
	return h.doSimpleTx("PUT", "/adrestia/update/deposit", simpleTx)
}

func (h *HestiaInstance) UpdateWithdrawal(simpleTx hestia.SimpleTx) (string, error) {
	return h.doSimpleTx("PUT", "/adrestia/update/withdrawal", simpleTx)
}

func (h *HestiaInstance) UpdateBalancer(balancer hestia.Balancer) (string, error) {
	req, err := mvt.CreateMVTToken("PUT", h.HestiaURL+"/adrestia/update/balancer", "adrestia", os.Getenv("MASTER_PASSWORD"), balancer, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"))
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

func (h *HestiaInstance) doSimpleTx(method string, url string, simpleTx hestia.SimpleTx) (string, error) {
	req, err := mvt.CreateMVTToken(method, h.HestiaURL+url, "adrestia", os.Getenv("MASTER_PASSWORD"), simpleTx, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"))
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

func (h *HestiaInstance) get(url string, params models.GetFilters) ([]byte, error) {
	req, err := mvt.CreateMVTToken(http.MethodGet, h.HestiaURL+url, "adrestia", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"))
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
	val, err := query.Values(params)
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

	return payload, nil
}
