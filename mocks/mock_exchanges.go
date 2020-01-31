// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/grupokindynos/adrestia-go/exchanges (interfaces: IExchange,IExchangeFactory)

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	exchanges "github.com/grupokindynos/adrestia-go/exchanges"
	balance "github.com/grupokindynos/adrestia-go/models/balance"
	transaction "github.com/grupokindynos/adrestia-go/models/transaction"
	coins "github.com/grupokindynos/common/coin-factory/coins"
	hestia "github.com/grupokindynos/common/hestia"
	reflect "reflect"
)

// MockIExchange is a mock of IExchange interface
type MockIExchange struct {
	ctrl     *gomock.Controller
	recorder *MockIExchangeMockRecorder
}

// MockIExchangeMockRecorder is the mock recorder for MockIExchange
type MockIExchangeMockRecorder struct {
	mock *MockIExchange
}

// NewMockIExchange creates a new mock instance
func NewMockIExchange(ctrl *gomock.Controller) *MockIExchange {
	mock := &MockIExchange{ctrl: ctrl}
	mock.recorder = &MockIExchangeMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockIExchange) EXPECT() *MockIExchangeMockRecorder {
	return m.recorder
}

// GetAddress mocks base method
func (m *MockIExchange) GetAddress(arg0 coins.Coin) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAddress", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAddress indicates an expected call of GetAddress
func (mr *MockIExchangeMockRecorder) GetAddress(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAddress", reflect.TypeOf((*MockIExchange)(nil).GetAddress), arg0)
}

// GetBalances mocks base method
func (m *MockIExchange) GetBalances() ([]balance.Balance, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBalances")
	ret0, _ := ret[0].([]balance.Balance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBalances indicates an expected call of GetBalances
func (mr *MockIExchangeMockRecorder) GetBalances() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBalances", reflect.TypeOf((*MockIExchange)(nil).GetBalances))
}

// GetDepositStatus mocks base method
func (m *MockIExchange) GetDepositStatus(arg0, arg1 string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDepositStatus", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDepositStatus indicates an expected call of GetDepositStatus
func (mr *MockIExchangeMockRecorder) GetDepositStatus(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDepositStatus", reflect.TypeOf((*MockIExchange)(nil).GetDepositStatus), arg0, arg1)
}

// GetName mocks base method
func (m *MockIExchange) GetName() (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetName")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetName indicates an expected call of GetName
func (mr *MockIExchangeMockRecorder) GetName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetName", reflect.TypeOf((*MockIExchange)(nil).GetName))
}

// GetOrderStatus mocks base method
func (m *MockIExchange) GetOrderStatus(arg0 hestia.ExchangeOrder) (hestia.OrderStatus, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrderStatus", arg0)
	ret0, _ := ret[0].(hestia.OrderStatus)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrderStatus indicates an expected call of GetOrderStatus
func (mr *MockIExchangeMockRecorder) GetOrderStatus(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrderStatus", reflect.TypeOf((*MockIExchange)(nil).GetOrderStatus), arg0)
}

// GetRateByAmount mocks base method
func (m *MockIExchange) GetRateByAmount(arg0 transaction.ExchangeSell) (float64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRateByAmount", arg0)
	ret0, _ := ret[0].(float64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRateByAmount indicates an expected call of GetRateByAmount
func (mr *MockIExchangeMockRecorder) GetRateByAmount(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRateByAmount", reflect.TypeOf((*MockIExchange)(nil).GetRateByAmount), arg0)
}

// OneCoinToBtc mocks base method
func (m *MockIExchange) OneCoinToBtc(arg0 coins.Coin) (float64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OneCoinToBtc", arg0)
	ret0, _ := ret[0].(float64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OneCoinToBtc indicates an expected call of OneCoinToBtc
func (mr *MockIExchangeMockRecorder) OneCoinToBtc(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OneCoinToBtc", reflect.TypeOf((*MockIExchange)(nil).OneCoinToBtc), arg0)
}

// SellAtMarketPrice mocks base method
func (m *MockIExchange) SellAtMarketPrice(arg0 transaction.ExchangeSell) (bool, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SellAtMarketPrice", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// SellAtMarketPrice indicates an expected call of SellAtMarketPrice
func (mr *MockIExchangeMockRecorder) SellAtMarketPrice(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SellAtMarketPrice", reflect.TypeOf((*MockIExchange)(nil).SellAtMarketPrice), arg0)
}

// Withdraw mocks base method
func (m *MockIExchange) Withdraw(arg0 coins.Coin, arg1 string, arg2 float64) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Withdraw", arg0, arg1, arg2)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Withdraw indicates an expected call of Withdraw
func (mr *MockIExchangeMockRecorder) Withdraw(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Withdraw", reflect.TypeOf((*MockIExchange)(nil).Withdraw), arg0, arg1, arg2)
}

// MockIExchangeFactory is a mock of IExchangeFactory interface
type MockIExchangeFactory struct {
	ctrl     *gomock.Controller
	recorder *MockIExchangeFactoryMockRecorder
}

// MockIExchangeFactoryMockRecorder is the mock recorder for MockIExchangeFactory
type MockIExchangeFactoryMockRecorder struct {
	mock *MockIExchangeFactory
}

// NewMockIExchangeFactory creates a new mock instance
func NewMockIExchangeFactory(ctrl *gomock.Controller) *MockIExchangeFactory {
	mock := &MockIExchangeFactory{ctrl: ctrl}
	mock.recorder = &MockIExchangeFactoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockIExchangeFactory) EXPECT() *MockIExchangeFactoryMockRecorder {
	return m.recorder
}

// GetExchangeByCoin mocks base method
func (m *MockIExchangeFactory) GetExchangeByCoin(arg0 coins.Coin) (exchanges.IExchange, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetExchangeByCoin", arg0)
	ret0, _ := ret[0].(exchanges.IExchange)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetExchangeByCoin indicates an expected call of GetExchangeByCoin
func (mr *MockIExchangeFactoryMockRecorder) GetExchangeByCoin(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetExchangeByCoin", reflect.TypeOf((*MockIExchangeFactory)(nil).GetExchangeByCoin), arg0)
}

// GetExchangeByName mocks base method
func (m *MockIExchangeFactory) GetExchangeByName(arg0 string) (exchanges.IExchange, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetExchangeByName", arg0)
	ret0, _ := ret[0].(exchanges.IExchange)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetExchangeByName indicates an expected call of GetExchangeByName
func (mr *MockIExchangeFactoryMockRecorder) GetExchangeByName(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetExchangeByName", reflect.TypeOf((*MockIExchangeFactory)(nil).GetExchangeByName), arg0)
}