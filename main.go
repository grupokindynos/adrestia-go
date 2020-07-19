package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/grupokindynos/adrestia-go/balancer"
	"github.com/grupokindynos/adrestia-go/controllers"
	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/models"
	"github.com/grupokindynos/adrestia-go/processor"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/common/responses"
	"github.com/grupokindynos/common/tokens/mrt"
	"github.com/grupokindynos/common/tokens/mvt"
	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}

var (
	exchangesProcessor processor.ExchangesProcessor
	depositProcessor   processor.DepositProcessor
	hwProcessor        processor.HwProcessor

	// Flags
	devMode bool

	// Urls
	hestiaUrl string
	plutusUrl string
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("Unable to initialize .env " + err.Error())
	}
}

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
	ticker := time.NewTicker(5 * time.Minute)
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
		log.Println("using local hestia and plutus")
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
		Obol:   &Obol,
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

	if !*stopProcessor {
		log.Println("Starting processors")
		//go runExchangesProcessor()
		//go runDepositProcessor()
		//go runHwProcessor()
	}

	App := GetApp()
	_ = App.Run(":" + *port)
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
	fmt.Println("PORT: ", os.Getenv("PORT"))
	auxHestia := services.HestiaRequests{HestiaURL: os.Getenv(hestiaUrl)}
	exchangeInfo, err := auxHestia.GetExchanges()

	if err != nil {
		log.Fatalln(err)
	}
	adrestiaCtrl := &controllers.AdrestiaController{
		Hestia:    services.HestiaRequests{HestiaURL: hestiaUrl},
		Plutus:    &services.PlutusRequests{PlutusURL: plutusUrl, Obol: &obol.ObolRequest{ObolURL: os.Getenv("OBOL_PRODUCTION_URL")}},
		Obol:      &obol.ObolRequest{ObolURL: os.Getenv("OBOL_PRODUCTION_URL")},
		DevMode:   devMode,
		ExFactory: exchanges.NewExchangeFactory(&obol.ObolRequest{ObolURL: os.Getenv("OBOL_PRODUCTION_URL")}, &services.HestiaRequests{HestiaURL: os.Getenv(hestiaUrl)}),
		ExInfo:    exchangeInfo,
	}
	authUser := os.Getenv("HESTIA_AUTH_USERNAME")
	authPassword := os.Getenv("HESTIA_AUTH_PASSWORD")
	api := r.Group("/", gin.BasicAuth(gin.Accounts{
		authUser: authPassword,
	}))
	{
		api.GET("address/:coin", func(context *gin.Context) { ValidateRequest(context, adrestiaCtrl.GetAddress) })
		api.POST("trade/status", func(context *gin.Context) {ValidateRequest(context, adrestiaCtrl.GetTradeStatus)})
		api.POST("withdraw/hash", func(context *gin.Context) {ValidateRequest(context, adrestiaCtrl.GetWithdrawalTxHash)})
		api.POST("path", func(context *gin.Context) { ValidateRequest(context, adrestiaCtrl.GetConversionPath) })
		api.POST("voucher/path", func(context *gin.Context) { ValidateRequest(context, adrestiaCtrl.GetVoucherConversionPath) })
		api.POST("trade", func(context *gin.Context) { ValidateRequest(context, adrestiaCtrl.Trade) })
		api.POST("withdraw", func(context *gin.Context) { ValidateRequest(context, adrestiaCtrl.Withdraw) })
		api.POST("deposit", func(context *gin.Context) { ValidateRequest(context, adrestiaCtrl.Deposit) })
		api.GET("stock/balance/:coin", func(context *gin.Context) { ValidateRequest(context, adrestiaCtrl.StockBalance) })
	}
	apiV2 := r.Group("/v2/", gin.BasicAuth(gin.Accounts{
		authUser: authPassword,
	}))
	{
		apiV2.POST("voucher/path", func(context *gin.Context) { ValidateRequest(context, adrestiaCtrl.GetVoucherConversionPathV2) })
		apiV2.POST("withdraw", func(context *gin.Context) { ValidateRequest(context, adrestiaCtrl.WithdrawV2) })
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
		openApi.GET("address/:coin", func(context *gin.Context) { ValidateOpenRequest(context, adrestiaCtrl.GetAddress) })
		openApi.POST("path", func(context *gin.Context) { ValidateOpenRequest(context, adrestiaCtrl.GetConversionPath) })
		openApi.GET("stock/balance/:coin", func(context *gin.Context) { ValidateOpenRequest(context, adrestiaCtrl.StockBalance) })
		openApi.POST("voucher/path", func(context *gin.Context) { ValidateOpenRequest(context, adrestiaCtrl.GetVoucherConversionPath) })
		openApi.GET("balance", func(context *gin.Context) { ValidateOpenRequest(context, adrestiaCtrl.Balances) })
		openApi.POST("voucher/path2", func(context *gin.Context) { ValidateOpenRequest(context, adrestiaCtrl.GetVoucherConversionPathV2) })
	}
}

func ValidateRequest(c *gin.Context, method func(uid string, payload []byte, params models.Params) (interface{}, error)) {
	uid := c.MustGet(gin.AuthUserKey).(string)
	if uid == "" {
		responses.GlobalOpenNoAuth(c)
	}
	params := models.Params{
		Coin: c.Param("coin"),
		Service: c.GetHeader("service"),
	}
	payload, err := mvt.VerifyRequest(c)
	response, err := method(uid, payload, params)
	if err != nil {
		responses.GlobalOpenError(nil, err, c)
		return
	}
	header, body, err := mrt.CreateMRTToken("adrestia", os.Getenv("MASTER_PASSWORD"), response, os.Getenv("ADRESTIA_PRIV_KEY"))
	responses.GlobalResponseMRT(header, body, c)
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
