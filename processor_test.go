package main

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/grupokindynos/adrestia-go/mocks"
	"github.com/grupokindynos/adrestia-go/processor"
	"github.com/grupokindynos/common/hestia"
	obolMocks "github.com/grupokindynos/common/obol/mocks"
)

var adrestiaOrders = []hestia.AdrestiaOrder{
	hestia.AdrestiaOrder{
		ID:              "123",
		DualExchange:    true,
		Time:            time.Now().Unix(),
		Status:          hestia.AdrestiaStatusCreated,
		Amount:          15.32343345,
		BtcRate:         0.0034,
		FromCoin:        "POLIS",
		ToCoin:          "BTC",
		FirstExAddress:  "123456payhere1",
		SecondExAddress: "123456payhere2",
		WithdrawAddress: "123345payhere3",
	},
	hestia.AdrestiaOrder{
		ID:              "12345",
		DualExchange:    false,
		Time:            time.Now().Unix(),
		Status:          hestia.AdrestiaStatusFirstExchange,
		Amount:          12.32343345,
		BtcRate:         0.0034,
		FromCoin:        "POLIS",
		ToCoin:          "BTC",
		FirstExAddress:  "123456payhere1",
		SecondExAddress: "123456payhere2",
		WithdrawAddress: "123345payhere3",
	},
	hestia.AdrestiaOrder{
		ID:              "111111",
		DualExchange:    true,
		Time:            time.Now().Unix(),
		Status:          hestia.AdrestiaStatusSecondExchange,
		Amount:          13.32343345,
		BtcRate:         0.0034,
		FromCoin:        "POLIS",
		ToCoin:          "BTC",
		FirstExAddress:  "123456payhere1",
		SecondExAddress: "123456payhere2",
		WithdrawAddress: "123345payhere3",
	},
}

func TestHandleCreatedOrders(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockHestia := mocks.NewMockHestiaService(mockCtrl)
	mockPlutus := mocks.NewMockPlutusService(mockCtrl)
	mockExchange := mocks.NewMockIExchange(mockCtrl)
	mockExchangeFactory := mocks.NewMockIExchangeFactory(mockCtrl)
	mockObol := obolMocks.NewMockObolService(mockCtrl)

	testProcessor := &processor.Processor{
		Plutus:          mockPlutus,
		Hestia:          mockHestia,
		Obol:            mockObol,
		ExchangeFactory: mockExchangeFactory,
	}

	gomock.InOrder(
		mockHestia.EXPECT().GetAllOrders(gomock.Any()).Return(adrestiaOrders),
		mockPlutus.EXPECT().WithdrawToAddress(gomock.Any()).AnyTimes().Return("123txid", nil),
	)

}
