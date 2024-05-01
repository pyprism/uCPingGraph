package routers

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	limits "github.com/gin-contrib/size"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/pyprism/uCPingGraph/controllers"
	"go.uber.org/zap"
	"os"
	"time"
)

func NewRouter() *gin.Engine {
	var logger *zap.Logger

	router := gin.New()
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*.html")

	if os.Getenv("DEBUG") != "True" {
		gin.SetMode(gin.ReleaseMode)
		logger, _ = zap.NewProduction()
		router.Use(ginzap.Ginzap(logger, time.RFC3339, true))
		router.Use(ginzap.RecoveryWithZap(logger, true))
	}

	router.SetTrustedProxies([]string{"127.0.0.1"})
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors.Default())
	router.Use(gzip.Gzip(gzip.BestCompression))
	router.Use(limits.RequestSizeLimiter(10000)) // 10KB

	api := new(controllers.APIController)
	index := new(controllers.IndexController)
	other := new(controllers.CommonController)

	router.GET("/", index.Home)
	router.POST("/device/", index.GetDeviceList)
	router.POST("/chart/", index.Chart)
	router.GET("/:static", other.StaticFile)
	router.POST("/api/stats/", api.PostStats)

	// for debug
	router.GET("/status/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"hi": "hiren",
		})
	})

	return router
}
