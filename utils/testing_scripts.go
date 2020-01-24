package utils

import (
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/hestia"
	"log"
	"os"
)

func CreateTestOrder(order hestia.AdrestiaOrder) (id string, err error){
	Hestia := services.HestiaRequests{HestiaURL: os.Getenv("HESTIA_URL_DEV")} // Make sure Hestia runs locally with -local flag
	id, err = Hestia.CreateAdrestiaOrder(order)
	if err != nil {
		log.Println("error:", err)
		return
	}
	return
}