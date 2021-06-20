package services

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type BSCBase struct {
	Status int32 `json:"status"`
 	Error string `json:"error"`
	Data interface{} `json:"data"`
}

type BSCApiService struct {
	BSCApiUrl   string
	MasterPassword string
}

func New() *BSCApiService {
	service := &BSCApiService{
		BSCApiUrl:   os.Getenv("PANCAKE_URL"),
		MasterPassword: os.Getenv("PANCAKE_PASSWORD"),
	}
	return service
}

func (bsc *BSCApiService) GetAddress() (float64, error) {
	path := bsc.BSCApiUrl + "/api/v1/address"
	log.Println("getting address: ", path)
	// token := "Bearer " + os.Getenv("BITCOU_TOKEN_V2")
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return 0, err
	}
	// req.Header.Add("Authorization", token)
	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	contents, _ := ioutil.ReadAll(res.Body)
	var response BSCBase
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