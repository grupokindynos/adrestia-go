package models

type OrderParams struct {
	IncludeComplete bool `json:include_complete`
	AddedSince int64 `json:added_since`
}
