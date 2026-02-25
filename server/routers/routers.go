package routers

import (
	"os"
	"time"

	"github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	limits "github.com/gin-contrib/size"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/pyprism/uCPingGraph/controllers"
	"github.com/pyprism/uCPingGraph/logger"
)

func NewRouter() *gin.Engine {
	router := gin.New()
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*.html")

	zapLogger := logger.Get()

	if os.Getenv("DEBUG") != "True" {
		gin.SetMode(gin.ReleaseMode)
	}

	router.Use(ginzap.Ginzap(zapLogger, time.RFC3339, true))
	router.Use(ginzap.RecoveryWithZap(zapLogger, true))
	router.Use(sentrygin.New(sentrygin.Options{Repanic: true}))

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
