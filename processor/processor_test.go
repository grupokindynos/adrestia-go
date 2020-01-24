package processor

import (
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/mocks"
	"github.com/grupokindynos/common/hestia"
	obolMocks "github.com/grupokindynos/common/obol/mocks"
)

var (
	testOrderExchange = []hestia.ExchangeOrder{
		hestia.ExchangeOrder{
			OrderId:          "133",
			Symbol:           "POLIS_BTC",
			Side:             "buy",
			Amount:           23.4,
			ReceivedAmount:   3.5,
			Exchange:         "testExchange",
			ReceivedCurrency: "POLIS",
			SoldCurrency:     "BTC",
		},
		hestia.ExchangeOrder{
			OrderId:          "133",
			Symbol:           "POLIS_BTC",
			Side:             "buy",
			Amount:           23.4,
			ReceivedAmount:   3.5,
			Exchange:         "testExchange",
			ReceivedCurrency: "POLIS",
			SoldCurrency:     "BTC",
		},
	}

	testOrders = []hestia.AdrestiaOrder{
		hestia.AdrestiaOrder{
			ID:              "123",
			DualExchange:    true,
			CreatedTime:     time.Now().Unix(),
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
			CreatedTime:     time.Now().Unix(),
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
			CreatedTime:     time.Now().Unix(),
			Status:          hestia.AdrestiaStatusSecondExchange,
			Amount:          13.32343345,
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
			CreatedTime:     time.Now().Unix(),
			Status:          hestia.AdrestiaStatusFirstConversion,
			Amount:          123.32343345,
			BtcRate:         0.0034,
			FromCoin:        "POLIS",
			ToCoin:          "BTC",
			FirstOrder:      testOrderExchange[0],
			FinalOrder:      testOrderExchange[1],
			FirstExAddress:  "123456payhere1",
			SecondExAddress: "123456payhere2",
			WithdrawAddress: "123345payhere3",
		},
		hestia.AdrestiaOrder{
			ID:              "111111",
			DualExchange:    true,
			CreatedTime:     time.Now().Unix(),
			Status:          hestia.AdrestiaStatusSecondConversion,
			Amount:          513.32343345,
			BtcRate:         0.0034,
			FromCoin:        "POLIS",
			ToCoin:          "BTC",
			FirstOrder:      testOrderExchange[1],
			FinalOrder:      testOrderExchange[0],
			FirstExAddress:  "123456payhere1",
			SecondExAddress: "123456payhere2",
			WithdrawAddress: "123345payhere3",
		},
	}

	mockHestia          *mocks.MockHestiaService
	mockPlutus          *mocks.MockPlutusService
	mockExchangeFactory *mocks.MockIExchangeFactory
	mockExchange        *mocks.MockIExchange
	mockObol            *obolMocks.MockObolService
	params              exchanges.Params
)

func InitParams(mockCtrl *gomock.Controller) {
	mockHestia = mocks.NewMockHestiaService(mockCtrl)
	mockPlutus = mocks.NewMockPlutusService(mockCtrl)
	mockExchangeFactory = mocks.NewMockIExchangeFactory(mockCtrl)
	mockExchange = mocks.NewMockIExchange(mockCtrl)
	mockObol = obolMocks.NewMockObolService(mockCtrl)

	params = exchanges.Params{
		Plutus:          mockPlutus,
		Hestia:          mockHestia,
		Obol:            mockObol,
		ExchangeFactory: mockExchangeFactory,
	}
}

func TestHandleCreatedOrders(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// withdrawError := errors.New("Not enough balance")

	InitParams(mockCtrl)
	InitProcessor(params)

	mockHestia.EXPECT().GetAllOrders(gomock.Any()).Return(testOrders, nil)
	mockPlutus.EXPECT().WithdrawToAddress(gomock.Any()).Return("123txid", nil)
	mockHestia.EXPECT().UpdateAdrestiaOrder(gomock.Any()).Return("ok", nil)

	var wg sync.WaitGroup
	wg.Add(1)
	go handleCreatedOrders(&wg)
	wg.Wait()
}

func TestHandleExchange(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	InitParams(mockCtrl)
	InitProcessor(params)

	mockHestia.EXPECT().GetAllOrders(gomock.Any()).Return(testOrders, nil)
	mockExchangeFactory.EXPECT().GetExchangeByCoin(gomock.Any()).Times(2).Return(mockExchange, nil)
	mockExchange.EXPECT().GetDepositStatus(gomock.Any(), gomock.Any()).Return(true, nil)
	mockExchange.EXPECT().GetDepositStatus(gomock.Any(), gomock.Any()).Return(false, nil)

	var wg sync.WaitGroup
	wg.Add(1)
	go handleExchange(&wg)
	wg.Wait()

	if adrestiaOrders[1].Status != hestia.AdrestiaStatusFirstConversion {
		t.Fatal("TestHandleExchange - Adrestia status didn't change")
	}

	if adrestiaOrders[2].Status != hestia.AdrestiaStatusSecondExchange {
		t.Fatal("TestHandleExchange - Adrestia status didn't change")
	}
}

func TestHandleConversion(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	InitParams(mockCtrl)
	InitProcessor(params)

}
