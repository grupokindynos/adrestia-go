package ladon

// server uses UTC
var bitcouPaymentHourUTC = 14
var minimumPaymentAmount = 700.0
var minimumWithdrawalAmount = 50.0
var serverHourDifference = 5 // server is 5 hours forward to local hour
var BTCExchanges = map[string]bool{"southxchange": true}

var BotUsersWhiteList = map[int]bool{635587721:true, 529339513:true}

func GetBitcouPaymentHour() int {
	return bitcouPaymentHourUTC
}

func IsBotAuthorizedUser(id int) bool {
	_, ok := BotUsersWhiteList[id];
	return ok
}