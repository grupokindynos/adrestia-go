package services

import (
	"github.com/grupokindynos/common/hestia"
)

type HestiaService interface {
	GetAdrestiaCoins() ([]hestia.Coin, error)
}