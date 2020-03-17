package exchanges

type Exchange interface {
	GetAddress()
	GetBalance()
	SellAtMarketPrice()
	Withdraw()
	GetOrderStatus()
	GetPair()
	GetWithdrawalTxHash()
	GetDepositStatus()
}
