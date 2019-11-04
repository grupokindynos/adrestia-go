package exchanges

import (
	"fmt"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddress(t *testing.T){
	cb := NewSouthXchange()
	data, err := cb.GetAddress(coins.Polis)
	fmt.Println(data)
	fmt.Println(err)
	assert.Nil(t, err)
}
