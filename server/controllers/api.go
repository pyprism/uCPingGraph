package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/pyprism/uCPingGraph/models"
)

type APIController struct{}

type postStats struct {
	Latency float64 `json:"latency" `
}

func (n *APIController) PostStats(c *gin.Context) {
	device := models.Device{}

	// get token from header
	token := c.Request.Header.Get("Authorization")
	if token == "" {
		c.JSON(400, gin.H{"error": "Authorization header is missing"})
		return
	}

	// get device id and network id from token
	deviceID, networkID, err := device.GetDeviceByToken(token)
	if err != nil {
		c.JSON(403, gin.H{"error": "Authorization header is invalid"})
		return
	}

	// get ping info from body
	var stats postStats
	err = c.BindJSON(&stats)

	if err != nil {
		c.JSON(400, gin.H{"POST body parse error": err.Error()})
		return
	}

	// create new stat
	stat := models.Stat{}
	err = stat.CreateStat(networkID, int(deviceID), float32(stats.Latency))
	if err != nil {
		c.JSON(400, gin.H{"DB error": err.Error()})
		return
	} else {
		c.JSON(200, gin.H{"success": "ok"})
		return
	}
}
