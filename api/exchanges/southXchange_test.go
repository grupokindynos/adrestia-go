package exchanges

import (
	"fmt"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)
func init() {
	if err := godotenv.Load(); err != nil {
		log.Println(err)
	}
}

func TestAddress(t *testing.T){
	cb := NewSouthXchange()
	data, err := cb.GetAddress(coins.Polis)
	fmt.Println(data)
	fmt.Println(err)
	assert.Nil(t, err)
}
