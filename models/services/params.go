package services

type AdrestiaOrderParams struct {
	IncludeComplete		bool	`url:"include_complete"`
	AddedSince			int64	`url:"added_since"`
}
