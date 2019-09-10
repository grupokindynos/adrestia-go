package services

import (
	"encoding/json"
	"fmt"
	rates "github.com/grupokindynos/adrestia-go/models/rates"
	"io/ioutil"
	"net/http"
)

const ratesUrl  = "https://obol-rates.herokuapp.com/"

type RateProvider struct {
	url string
}

func (r RateProvider) GetRate(coin string) float64{
	fmt.Println("\tRetrieving Rates for ", coin)

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
		if err != nil{
			fmt.Println(err)
		}
	}
	return rates.Data.BTC
}
