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
	"go.uber.org/zap"
)

func NewRouter() *gin.Engine {
	if os.Getenv("DEBUG") != "True" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*.html")

	zapLogger := logger.Get()

	router.Use(requestLogger(zapLogger))
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

func requestLogger(zapLogger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now().UTC()
		fields := []zap.Field{
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.Float64("request_latency_seconds", end.Sub(start).Seconds()),
			zap.String("time", end.Format(time.RFC3339)),
		}

		if len(c.Errors) > 0 {
			for _, err := range c.Errors.Errors() {
				zapLogger.Error(err, fields...)
			}
			return
		}

		zapLogger.Info(path, fields...)
	}
}
