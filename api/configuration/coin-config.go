package configuration

type Coin interface {
	GetExhange() string
}


type ExchangeBehaviour interface {
	SellAtMarketPrice() bool
}

type Exchange struct {

}
