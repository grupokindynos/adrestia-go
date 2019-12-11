package config

type CBAuth struct{
	MasterPassword string 	`json:"master_password"`
	AccountName string 		`json:"account_name"`
	BaseUrl string			`json:"base_url"`
	BitSharesUrl string		`json:"bitshares_url"`
}