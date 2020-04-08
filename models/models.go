package models

type GetFilters struct {
	Id string `json:"id"`
	IncludeComplete bool `json:"include_complete"`
	AddedSince int64 `json:"added_since"`
}

type TradeInfo struct {
	Book string `json:"book"`
	Type string `json:"type"`
}

type ExchangeTradeOrder struct {
	Symbol string `json:"symbol"`
	Side string `json:"side"`
	Amount float64 `json:"amount"`
}

type Params struct {
	Coin string `json:"coin"`
}

type AddressResponse struct {
	Coin string `json:"coin"`
	Address string `json:"address"`
	Exchange string `json:"exchange"`
}

type PathParams struct {
	FromCoin string `json:"from_coin"`
	ToCoin string `json:"to_coin"`
	FirstExchange string `json:"exchange"`
}

type MockTrade struct {
	FromCoin string `json:"from_coin"`
	ToCoin string `json:"to_coin"`
	Exchange string `json:"exchange"`
}

type PathResponse struct {
	InwardOrder []MockTrade `json:"in_order"`
	OutwardOrder []MockTrade `json:"out_order"`
}
