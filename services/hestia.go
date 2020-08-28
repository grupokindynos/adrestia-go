package services

import (
	"encoding/json"
	"errors"
	"fmt"
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

type HestiaRequests struct {
	HestiaURL string
}

func (h *HestiaRequests) GetAdrestiaCoins() (availableCoins []hestia.Coin, err error) {
	payload, err := h.get("/coins", models.GetFilters{})
	if err != nil {
		return nil, err
	}
	var response []hestia.Coin
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return nil, err
	}

	for _, coin := range response {
		if coin.Adrestia.Available {
			availableCoins = append(availableCoins, coin)
		}
	}
	return availableCoins, nil
}

func (h *HestiaRequests) GetExchange(exchange string) (hestia.ExchangeInfo, error) {
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

func (h *HestiaRequests) GetExchanges() (exchangesInfo []hestia.ExchangeInfo, err error) {
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

func (h *HestiaRequests) GetDeposits(includeComplete bool, sinceTimestamp int64) ([]hestia.SimpleTx, error) {
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

func (h *HestiaRequests) GetWithdrawals(includeComplete bool, sinceTimestamp int64, balancerId string) ([]hestia.SimpleTx, error) {
	payload, err := h.get("/adrestia/withdrawals", models.GetFilters{IncludeComplete:includeComplete, AddedSince:sinceTimestamp, BalancerId: balancerId})
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

func (h *HestiaRequests) GetBalanceOrders(includeComplete bool, sinceTimestamp int64) ([]hestia.BalancerOrder, error) {
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

func (h *HestiaRequests) GetBalancer() (hestia.Balancer, error) {
	payload, err := h.get("/adrestia/balancer", models.GetFilters{IncludeComplete: false})
	if err != nil {
		return hestia.Balancer{}, err
	}
	var response []hestia.Balancer
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return hestia.Balancer{}, err
	}
	if len(response) > 1 {
		return hestia.Balancer{}, errors.New("there's more than one balancer")
	}
	if len(response) == 0 {
		return hestia.Balancer{}, nil
	}

	return response[0], nil
}

func (h *HestiaRequests) CreateDeposit(simpleTx hestia.SimpleTx) (string, error) {
	req, err := mvt.CreateMVTToken("POST", h.HestiaURL+"/adrestia/new/deposit", "adrestia", os.Getenv("MASTER_PASSWORD"), simpleTx, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"))
	if err != nil {
		return "", err
	}
	return doResponseToString(h.do(req))
}

func (h *HestiaRequests) CreateWithdrawal(simpleTx hestia.SimpleTx) (string, error) {
	req, err := mvt.CreateMVTToken("POST", h.HestiaURL+"/adrestia/new/withdrawal", "adrestia", os.Getenv("MASTER_PASSWORD"), simpleTx, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"))
	if err != nil {
		return "", err
	}
	return doResponseToString(h.do(req))
}

func (h *HestiaRequests) CreateBalancerOrder(balancerOrder hestia.BalancerOrder) (string, error) {
	req, err := mvt.CreateMVTToken("POST", h.HestiaURL+"/adrestia/new/order", "adrestia", os.Getenv("MASTER_PASSWORD"), balancerOrder, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"))
	if err != nil {
		return "", err
	}
	return doResponseToString(h.do(req))
}

func (h *HestiaRequests) CreateBalancer(balancer hestia.Balancer) (string, error) {
	req, err := mvt.CreateMVTToken("POST", h.HestiaURL+"/adrestia/new/balancer", "adrestia", os.Getenv("MASTER_PASSWORD"), balancer, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"))
	if err != nil {
		return "", err
	}
	return doResponseToString(h.do(req))
}

func (h *HestiaRequests) UpdateDeposit(simpleTx hestia.SimpleTx) (string, error) {
	req, err := mvt.CreateMVTToken("PUT", h.HestiaURL+"/adrestia/update/deposit", "adrestia", os.Getenv("MASTER_PASSWORD"), simpleTx, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"))
	if err != nil {
		return "", err
	}
	return doResponseToString(h.do(req))
}

func (h *HestiaRequests) UpdateWithdrawal(simpleTx hestia.SimpleTx) (string, error) {
	req, err := mvt.CreateMVTToken("PUT", h.HestiaURL+"/adrestia/update/withdrawal", "adrestia", os.Getenv("MASTER_PASSWORD"), simpleTx, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"))
	if err != nil {
		return "", err
	}
	return doResponseToString(h.do(req))
}

func (h *HestiaRequests) UpdateBalancer(balancer hestia.Balancer) (string, error) {
	req, err := mvt.CreateMVTToken("PUT", h.HestiaURL+"/adrestia/update/balancer", "adrestia", os.Getenv("MASTER_PASSWORD"), balancer, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"))
	if err != nil {
		return "", err
	}
	return doResponseToString(h.do(req))
}

func (h *HestiaRequests) UpdateBalancerOrder(order hestia.BalancerOrder) (string, error) {
	req, err := mvt.CreateMVTToken("PUT", h.HestiaURL+"/adrestia/update/order", "adrestia", os.Getenv("MASTER_PASSWORD"), order, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"))
	if err != nil {
		return "", err
	}
	return doResponseToString(h.do(req))
}

// Bitcou Payment
func (h *HestiaRequests) GetVouchersByStatusV2(status hestia.VoucherStatusV2) ([]hestia.VoucherV2, error) {
	payload, err := h.get("/voucher2/all?filter="+fmt.Sprintf("%d", status), models.GetFilters{})
	if err != nil {
		return nil, err
	}
	var response []hestia.VoucherV2
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (h *HestiaRequests) get(url string, params models.GetFilters) ([]byte, error) {
	req, err := mvt.CreateMVTToken(http.MethodGet, h.HestiaURL+url, "adrestia", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"))
	if err != nil {
		return nil, err
	}
	q := req.URL.RawQuery
	val, err := query.Values(params)
	if err != nil {
		return nil, errors.New("problem with query parameters")
	}
	req.URL.RawQuery = q + val.Encode() // add encoded values
	return h.do(req)
}

func (h *HestiaRequests) do(req *http.Request) ([]byte, error) {
	client := http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       time.Second * 30,
	}
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

func doResponseToString(payload []byte, err error) (string, error) {
	if err != nil {
		return "", err
	}
	var response string
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return "", err
	}
	return response, nil
}

// tyche
func (h *HestiaRequests) GetOpenShifts(timestamp string) (shifts []hestia.ShiftV2, err error) {
	payload, err := h.get("/shift2/open/all?timestamp="+timestamp, models.GetFilters{})
	if err != nil {
		return nil, err
	}
	var response []hestia.ShiftV2
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (h *HestiaRequests) ChangeShiftProcessorStatus(status bool) error {
	payload, err := h.get("/config", models.GetFilters{})
	if err != nil {
		return err
	}
	var response hestia.Config
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return err
	}

	response.Shift.Processor = status

	req, err := mvt.CreateMVTToken("POST", h.HestiaURL+"/config/update", "adrestia", os.Getenv("MASTER_PASSWORD"), response, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"))
	if err != nil {
		return err
	}
	client := http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       time.Second * 30,
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		return err
	}
	headerSignature := res.Header.Get("service")
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("HESTIA_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return err
	}

	return nil
}



