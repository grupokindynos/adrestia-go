package models

import "github.com/grupokindynos/common/hestia"

type GetFilters struct {
	Id              string `json:"id"`
	IncludeComplete bool   `json:"include_complete"`
	AddedSince      int64  `json:"added_since"`
	BalancerId      string `json:"balancer_id"`
}

type TradeInfo struct {
	Book string `json:"book"`
	Type string `json:"type"`
}

type ExchangeTradeOrder struct {
	Symbol string  `json:"symbol"`
	Side   string  `json:"side"`
	Amount float64 `json:"amount"`
}

type Params struct {
	Coin string `json:"coin"`
	Service string
}

type ParamsV2 struct {
	Coin string `json:"coin"`
	Exchange string `json:"exchange"`
	Service hestia.ServiceAccount `json:"service"`
}

type AddressResponse struct {
	Coin            string          `json:"coin"`
	ExchangeAddress ExchangeAddress `json:"address"`
}

type ExchangeAddress struct {
	Address  string `json:"address"`
	Exchange string `json:"exchange"`
}

type PathParams struct {
	FromCoin string `json:"from_coin"`
	ToCoin   string `json:"to_coin"`
}

type VoucherPathParams struct {
	FromCoin string `json:"from_coin"`
}

type VoucherPathParamsV2 struct {
	FromCoin string `json:"from_coin"`
	AmountEuro float64 `json:"amount_euro"`
}

type WithdrawParams struct {
	Asset   string  `json:"asset"`
	Address string  `json:"address"`
	Amount  float64 `json:"amount"`
}

type WithdrawParamsV2 struct {
	Asset   string  `json:"asset"`
	Address string  `json:"address"`
	Amount  float64 `json:"amount"`
	Exchange string `json:"exchange"`
}

type DepositParams struct {
	Asset   string `json:"asset"`
	TxId    string `json:"txid"`
	Address string `json:"address"`
}

type WithdrawInfo struct {
	Exchange string `json:"exchange"`
	Asset    string `json:"asset"`
	TxId     string `json:"txid"`
}

type DepositInfo struct {
	Exchange    string                   `json:"exchange"`
	DepositInfo hestia.ExchangeOrderInfo `json:"deposit_info"`
}

type BalanceResponse struct {
	Exchange string  `json:"exchange"`
	Balance  float64 `json:"balance"`
	Asset    string  `json:"asset"`
}

type ExchangeTrade struct {
	FromCoin string    `json:"from_coin"`
	ToCoin   string    `json:"to_coin"`
	Exchange string    `json:"exchange"`
	Trade    TradeInfo `json:"trade"`
}

type PathResponse struct {
	InwardOrder  []ExchangeTrade `json:"in_order"`
	OutwardOrder []ExchangeTrade `json:"out_order"`
	Trade        bool            `json:"trade"`
}

type VoucherPathResponse struct {
	InwardOrder      []ExchangeTrade `json:"in_order"`
	Trade            bool            `json:"trade"`
	TargetStableCoin string          `json:"target"`
	Address          string          `json:"address"`
}

type GlobalBalanceResponse struct {
	Assets []AssetGlobalBalance `json:"assets"`
}

type AssetGlobalBalance struct {
	Asset string `json:"asset"`
	Balances []BalanceResponse `json:"balances"`
	Total float64 `json:"total"`
}

type ExchangeParams struct {
	Name string
	Keys hestia.ApiKeys
}
