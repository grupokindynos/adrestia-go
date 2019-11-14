package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-querystring/query"
	"github.com/grupokindynos/adrestia-go/models/services"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/tokens/mrt"
	"github.com/grupokindynos/common/tokens/mvt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func GetCoinConfiguration() ([]hestia.Coin, error) {
	req, err := mvt.CreateMVTToken("GET", os.Getenv("HESTIA_URL")+"/coins", "adrestia", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"))
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
	fmt.Println("Hestia Conf: ", response)
	return response, nil
}

func GetBalancingOrders() ([]hestia.AdrestiaOrder, error) {
	req, err := mvt.CreateMVTToken("GET", os.Getenv("HESTIA_URL") + "/adrestia/orders", "adrestia", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"))
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

func CreateAdrestiaOrder(orderData hestia.AdrestiaOrder) (string, error) {
	req, err := mvt.CreateMVTToken("POST", hestia.ProductionURL + "/adrestia/new", "adrestia", os.Getenv("MASTER_PASSWORD"), orderData, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"))
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

func GetAllOrders(adrestiaOrderParams services.AdrestiaOrderParams) ([]hestia.AdrestiaOrder, error){
	fmt.Println(hestia.ProductionURL)
	req, err := mvt.CreateMVTToken(http.MethodGet, os.Getenv("HESTIA_URL") + "/adrestia/orders", "adrestia", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"))
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
	req.URL.RawQuery = q + val.Encode()
	fmt.Println(req.URL.RawQuery)

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