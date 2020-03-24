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