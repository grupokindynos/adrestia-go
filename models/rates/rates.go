package rates


type Rates struct {
	Data []struct {
		Code string  `json:"code"`
		Name string  `json:"name"`
		Rate float64 `json:"rate"`
	} `json:"data"`
	Status int `json:"status"`
}
