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
	router := gin.New()
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*.html")

	if os.Getenv("DEBUG") != "True" {
		gin.SetMode(gin.ReleaseMode)
		logger, _ := zap.NewProduction()
		router.Use(ginzap.Ginzap(logger, time.RFC3339, true))
		router.Use(ginzap.RecoveryWithZap(logger, true))
	} else {
		router.Use(gin.Logger())
		router.Use(gin.Recovery())
	}

	_ = router.SetTrustedProxies(nil)
	router.Use(cors.Default())
	router.Use(gzip.Gzip(gzip.BestCompression))
	router.Use(limits.RequestSizeLimiter(32 * 1024))

	api := new(controllers.APIController)
	index := new(controllers.IndexController)

	router.GET("/", index.Home)
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	apiGroup := router.Group("/api")
	{
		apiGroup.POST("/stats", api.PostStats)
		apiGroup.POST("/stats/", api.PostStats)
		apiGroup.GET("/networks", api.Networks)
		apiGroup.GET("/networks/:network/devices", api.DevicesByNetwork)
		apiGroup.GET("/series", api.Series)
	}

	return router
}
