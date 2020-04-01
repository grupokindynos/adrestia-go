package main

import (
	"encoding/json"
	"flag"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/grupokindynos/adrestia-go/balancer"
	"github.com/grupokindynos/adrestia-go/controllers"
	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/models"
	"github.com/grupokindynos/adrestia-go/processor"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/jwt"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/common/responses"
	"github.com/grupokindynos/common/tokens/ppat"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"time"
)


func init() {
	_ = godotenv.Load()
}

var (
	exchangesProcessor 	processor.ExchangesProcessor
	depositProcessor 	processor.DepositProcessor
	hwProcessor 		processor.HwProcessor

	// Flags
	devMode				bool

	// Urls
	hestiaUrl string
 	plutusUrl string
)

func runExchangesProcessor() {
	ticker := time.NewTicker(2 * time.Hour)
	for _ = range ticker.C {
		exchangesProcessor.Start()
	}
}

func runDepositProcessor() {
	ticker := time.NewTicker(5 * time.Minute)
	for _ = range ticker.C {
		depositProcessor.Start()
	}
}

func runHwProcessor() {
	ticker := time.NewTicker(10 * time.Minute)
	for _ = range ticker.C {
		hwProcessor.Start()
	}
}

func main() {
	// Read input flag
	localRun := flag.Bool("local", false, "set this flag to run adrestia with local db")
	port := flag.String("port", os.Getenv("PORT"), "set different port for local run")
	stopProcessor := flag.Bool("stop-proc", false, "set this flag to stop the automatic run of processor")
	dev := flag.Bool("dev", false, "return shift status as always available")
	flag.Parse()

	// If flag was set, change the hestia request url to be local
	if *localRun {
		hestiaUrl = "HESTIA_LOCAL_URL"
		plutusUrl = "PLUTUS_LOCAL_URL"
	} else {
		hestiaUrl = "HESTIA_PRODUCTION_URL"
		plutusUrl = "PLUTUS_PRODUCTION_URL"
	}
	devMode = *dev

	Obol := obol.ObolRequest{ObolURL: os.Getenv("OBOL_PRODUCTION_URL")}
	Hestia := services.HestiaRequests{HestiaURL: os.Getenv(hestiaUrl)}
	Plutus := services.PlutusRequests{PlutusURL: os.Getenv(plutusUrl), Obol: &Obol}

	Balancer := balancer.Balancer{
		Hestia: &Hestia,
		Plutus: &Plutus,
	}
	exchangesProcessor = processor.ExchangesProcessor{
		Obol:   &Obol,
		Hestia: &Hestia,
		Plutus: &Plutus,
	}
	depositProcessor = processor.DepositProcessor{
		Hestia: &Hestia,
		Plutus: &Plutus,
		Obol:   &Obol,
	}
	hwProcessor = processor.HwProcessor{
		Hestia:   &Hestia,
		Plutus:   &Plutus,
		Obol:     &Obol,
		Balancer: Balancer,
	}

	App := GetApp()
	_ = App.Run(":" + *port)

	if !*stopProcessor {
		log.Println("Starting processors")
		go runExchangesProcessor()
		go runDepositProcessor()
		go runHwProcessor()

		select {}
	}

}

func GetApp() *gin.Engine {
	App := gin.Default()
	corsConf := cors.DefaultConfig()
	corsConf.AllowAllOrigins = true
	corsConf.AllowHeaders = []string{"token", "service", "content-type"}
	App.Use(cors.New(corsConf))
	ApplyRoutes(App)
	return App
}

func ApplyRoutes(r *gin.Engine) {
	adrestiaCtrl := &controllers.AdrestiaController{
		Hestia:        services.HestiaRequests{HestiaURL: hestiaUrl},
		Plutus:        &services.PlutusRequests{},
		Obol:          &obol.ObolRequest{ObolURL: os.Getenv("OBOL_PRODUCTION_URL")},
		DevMode:	   devMode,
		ExFactory:     exchanges.NewExchangeFactory(&obol.ObolRequest{ObolURL: os.Getenv("OBOL_PRODUCTION_URL")}, &services.HestiaRequests{HestiaURL: hestiaUrl}),
	}

	api := r.Group("/")
	{
		api.GET("address/:coin", func(context *gin.Context) { ValidateRequest(context, adrestiaCtrl.GetAddress) })
	}
	r.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "Not Found")
	})


	username := os.Getenv("TEST_API_USER")
	password := os.Getenv("TEST_API_PASS")
	openApi := r.Group("/test/", gin.BasicAuth(gin.Accounts{
		username: password,
	}))
	{
		openApi.GET("address/:coin", func(context *gin.Context) {ValidateOpenRequest(context, adrestiaCtrl.GetAddress)})
	}
}

func ValidateRequest(c *gin.Context, method func(uid string, payload []byte, params models.Params) (interface{}, error)) {
	fbToken := c.GetHeader("token")
	if fbToken == "" {
		responses.GlobalResponseNoAuth(c)
		return
	}
	params := models.Params{
		Coin: c.Param("coin"),
	}
	tokenBytes, _ := c.GetRawData()
	var ReqBody hestia.BodyReq
	if len(tokenBytes) > 0 {
		err := json.Unmarshal(tokenBytes, &ReqBody)
		if err != nil {
			responses.GlobalResponseError(nil, err, c)
			return
		}
	}
	valid, payload, uid, err := ppat.VerifyPPATToken(hestiaUrl, "adrestia", os.Getenv("MASTER_PASSWORD"), fbToken, ReqBody.Payload, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("ADRESTIA_PRIV_KEY"), os.Getenv("HESTIA_PUBLIC_KEY"))
	if !valid {
		responses.GlobalResponseNoAuth(c)
		return
	}
	response, err := method(uid, payload, params)
	if err != nil {
		responses.GlobalResponseError(nil, err, c)
		return
	}
	token, err := jwt.EncryptJWE(uid, response)
	responses.GlobalResponseError(token, err, c)
	return
}

func ValidateOpenRequest(c *gin.Context, method func(uid string, payload []byte, params models.Params) (interface{}, error)) {
	uid := c.MustGet(gin.AuthUserKey).(string)
	if uid == "" {
		responses.GlobalOpenNoAuth(c)
	}
	params := models.Params{
		Coin: c.Param("coin"),
	}
	payload, err := c.GetRawData()
	response, err := method(uid, payload, params)
	if err != nil {
		responses.GlobalOpenError(nil, err, c)
		return
	}
	responses.GlobalResponse(response, c)
	return
}
