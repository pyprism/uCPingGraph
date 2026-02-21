package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pyprism/uCPingGraph/models"
	"github.com/pyprism/uCPingGraph/service"
)

type APIController struct{}

type postStats struct {
	LatencyMs         *float64 `json:"latency_ms"`
	LatencyLegacy     *float64 `json:"latency"`
	SentPackets       int      `json:"sent_packets"`
	ReceivedPackets   int      `json:"received_packets"`
	PacketLossPercent *float64 `json:"packet_loss_percent"`
	Target            string   `json:"target"`
	Platform          string   `json:"platform"`
	RSSI              int      `json:"rssi"`
}

func (n *APIController) PostStats(c *gin.Context) {
	token := strings.TrimSpace(c.Request.Header.Get("Authorization"))
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
		return
	}

	var req postStats
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON payload"})
		return
	}

	latency := 0.0
	if req.LatencyMs != nil {
		latency = *req.LatencyMs
	} else if req.LatencyLegacy != nil {
		latency = *req.LatencyLegacy
	}
	if latency < 0 {
		latency = 0
	}

	sentPackets := req.SentPackets
	receivedPackets := req.ReceivedPackets
	packetLoss := 0.0

	if req.PacketLossPercent != nil {
		packetLoss = *req.PacketLossPercent
	}

	// Backward compatibility for old clients that only send latency.
	if sentPackets == 0 && receivedPackets == 0 && req.PacketLossPercent == nil {
		sentPackets = 1
		receivedPackets = 1
		packetLoss = 0
	}

	if sentPackets > 0 {
		calculatedLoss := (float64(sentPackets-receivedPackets) / float64(sentPackets)) * 100
		if calculatedLoss >= 0 && calculatedLoss <= 100 {
			packetLoss = calculatedLoss
		}
	}

	err := service.SaveStats(token, service.IngestTelemetry{
		LatencyMs:         latency,
		SentPackets:       sentPackets,
		ReceivedPackets:   receivedPackets,
		PacketLossPercent: packetLoss,
		Target:            req.Target,
		Platform:          req.Platform,
		RSSI:              req.RSSI,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidToken):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid device token"})
		case errors.Is(err, service.ErrInvalidTelemetry):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid telemetry payload"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist telemetry"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "ok"})
}

func (n *APIController) Networks(c *gin.Context) {
	networkModel := models.Network{}
	networks, err := networkModel.GetAllNetworkName()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch networks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": networks})
}

func (n *APIController) DevicesByNetwork(c *gin.Context) {
	networkName := c.Param("network")
	networkModel := models.Network{}
	deviceModel := models.Device{}

	networkID, err := networkModel.GetNetworkIdByName(networkName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "network not found"})
		return
	}

	devices, err := deviceModel.GetDevicesByNetwork(int(networkID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch devices"})
		return
	}

	items := make([]string, 0, len(devices))
	for _, d := range devices {
		items = append(items, d.Name)
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (n *APIController) Series(c *gin.Context) {
	network := strings.TrimSpace(c.Query("network"))
	device := strings.TrimSpace(c.Query("device"))
	minutesRaw := c.DefaultQuery("minutes", "60")

	minutes, err := strconv.Atoi(minutesRaw)
	if err != nil || minutes <= 0 || minutes > 10080 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "minutes must be between 1 and 10080"})
		return
	}

	if network == "" || device == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "network and device are required"})
		return
	}

	data, err := service.GetSeries(network, device, minutes)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "metric series not found"})
		return
	}

	c.JSON(http.StatusOK, data)
}
