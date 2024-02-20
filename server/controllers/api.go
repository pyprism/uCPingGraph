package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/pyprism/uCPingGraph/utils"
)

type APIController struct{}

type postStats struct {
	Latency float64 `json:"latency" binding:"required"`
}

func (n *APIController) PostStats(c *gin.Context) {
	// get token from header
	token := c.Request.Header.Get("Authorization")
	if token == "" {
		c.JSON(400, gin.H{"error": "Authorization header is missing"})
		return
	}

	// get ping info from body
	var stats postStats
	bindErr := c.BindJSON(&stats)

	if bindErr != nil {
		c.JSON(400, gin.H{"POST body parse error": bindErr.Error()})
		return
	}

	err := utils.SaveStats(token, float32(stats.Latency))
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	} else {
		c.JSON(200, gin.H{"success": "ok"})
		return
	}
}
