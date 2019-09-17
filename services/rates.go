package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	rates "github.com/grupokindynos/adrestia-go/models/rates"
)

const ratesUrl = "https://obol-rates.herokuapp.com/"

type RateProvider struct {
	url string
}

func (r RateProvider) GetRate(coin string) float64 {
	// fmt.Println("\tRetrieving Rates for ", coin)

	var rates rates.Rates

	response, err := http.Get(ratesUrl + "simple/" + coin)

	if err != nil {
		fmt.Print(err)
	}

	defer response.Body.Close()

	if err != nil {
		fmt.Println(err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		err := json.Unmarshal(data, &rates)
		if err != nil {
			fmt.Println(err)
		}
	}

	for _, rate := range rates.Data {
		if rate.Code == "BTC" {
			return rate.Rate
		}
	}
	return 0.00
}
