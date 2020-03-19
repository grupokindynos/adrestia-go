package exchanges

import (
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"testing"
)

func TestBittrex_GetAddress(t *testing.T) {
	b := NewBittrex()

	c, err := coinfactory.GetCoin("BTC")
	if err != nil {
		t.Fatal(err)
	}

	addr, err := b.GetAddress(*c)
	if err != nil {
		t.Fatal(err)
	}

	if len(addr) == 0 {
		t.Fatal("expected address to be returned")
	}

	t.Fatal(addr)
}
