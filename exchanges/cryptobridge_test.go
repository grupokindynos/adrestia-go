package exchanges

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSettings(t *testing.T){
	cb := NewCryptobridge()
	assert.NotNil(t, cb.BitSharesUrl)
	assert.NotNil(t, cb.BaseUrl)
	assert.NotNil(t, cb.AccountName)
	assert.NotNil(t, cb.MasterPassword)
}

func TestBalancing(t *testing.T){

	cb := NewCryptobridge()
	fmt.Println(cb)
}

