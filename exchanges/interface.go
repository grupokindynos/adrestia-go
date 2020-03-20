package exchanges

type Exchange interface {
	GetAddress()
	GetBalance(asset string) (float64, error)
	SellAtMarketPrice()
	Withdraw()
	GetOrderStatus()
	GetPair()
	GetWithdrawalTxHash()
	GetDepositStatus()
}
