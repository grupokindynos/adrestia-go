package services

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type BitcouV2Service struct {
	BitcouURL   string
	BitcouToken string
}

func InitBitcouV2Service() *BitcouV2Service {
	service := &BitcouV2Service{
		BitcouURL:   os.Getenv("BITCOU_URL_PROD_V2"),
		BitcouToken: os.Getenv("BITCOU_TOKEN_V2"),
	}
	return service
}

func (bs *BitcouV2Service) GetFloatingAccountInfo() (float64, error) {
	path := bs.BitcouURL + "/account/balance"
	log.Println("Getting floating account balance using url: ", path)
	token := "Bearer " + os.Getenv("BITCOU_TOKEN_V2")
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Add("Authorization", token)
	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	contents, _ := ioutil.ReadAll(res.Body)
	var response BaseResponse
	err = json.Unmarshal(contents, &response)
	if err != nil {
		return 0, err
	}
	var vouchersList []BalanceResponse
	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return 0, err
	}
	err = json.Unmarshal(dataBytes, &vouchersList)
	if err != nil {
		return 0, err
	}
	return vouchersList[0].Amount, nil
}

type BaseResponse struct {
	Data []interface{} `json:"data"`
	Meta MetaData      `json:"meta"`
}

type MetaData struct {
	Datetime string `json:"datetime"`
}

type BalanceResponse struct {
	Amount float64 `json:"amount"`
}