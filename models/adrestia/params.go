package adrestia

type OrderParams struct {
	IncludeComplete		bool	`url:"include_complete"`
	AddedSince			int64	`url:"added_since"`
}
