package config

type BinanceAuth struct {
	PublicApi string 	`json:"publicApi"`
	PrivateApi string 		`json:"privateApi"`
	PublicWithdrawKey string			`json:"publicWithdrawKey"`
	PrivateWithdrawKey string		`json:"privateWithdrawKey"`
}
