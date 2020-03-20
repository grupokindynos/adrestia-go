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

func (h *HestiaInstance) GetExchanges() (exchangesInfo []hestia.ExchangeInfo, err error) {
	payload, err := h.get("/exchanges", models.OrderParams{})
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

func (h *HestiaInstance) GetDeposits(params models.OrderParams) ([]hestia.SimpleTx, error) {
	payload, err := h.get("/adrestia/deposits", params)
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

func (h *HestiaInstance) CreateDeposit(simpleTx hestia.SimpleTx) (string, error) {
	return h.doSimpleTx("POST", "adrestia/new/deposit", simpleTx)
}

func (h *HestiaInstance) UpdateDeposit(simpleTx hestia.SimpleTx) (string, error) {
	return h.doSimpleTx("PUT", "adrestia/update/deposit", simpleTx)
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

func (h *HestiaInstance) get(url string, params models.OrderParams) ([]byte, error) {
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

