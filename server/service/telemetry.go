package service

import (
	"errors"
	"strings"

	"github.com/pyprism/uCPingGraph/models"
	"gorm.io/gorm"
)

var (
	ErrInvalidToken     = errors.New("authorization token is invalid")
	ErrInvalidTelemetry = errors.New("telemetry payload is invalid")
)

type IngestTelemetry struct {
	LatencyMs         float64
	SentPackets       int
	ReceivedPackets   int
	PacketLossPercent float64
	Target            string
	Platform          string
	RSSI              int
}

func SaveStats(token string, telemetry IngestTelemetry) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return ErrInvalidToken
	}

	if telemetry.SentPackets <= 0 ||
		telemetry.ReceivedPackets < 0 ||
		telemetry.ReceivedPackets > telemetry.SentPackets ||
		telemetry.PacketLossPercent < 0 ||
		telemetry.PacketLossPercent > 100 {
		return ErrInvalidTelemetry
	}

	device := models.Device{}
	deviceID, networkID, err := device.GetDeviceByToken(token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidToken
		}
		return err
	}

	record := models.TelemetryRecord{
		LatencyMs:         telemetry.LatencyMs,
		SentPackets:       telemetry.SentPackets,
		ReceivedPackets:   telemetry.ReceivedPackets,
		PacketLossPercent: telemetry.PacketLossPercent,
		Target:            strings.TrimSpace(telemetry.Target),
		Platform:          strings.TrimSpace(telemetry.Platform),
		RSSI:              telemetry.RSSI,
	}

	stat := models.Stat{}
	return stat.CreateStat(networkID, int(deviceID), record)
}

func GetSeries(networkName, deviceName string, minute int) (*models.MetricsResponse, error) {
	network := models.Network{}
	device := models.Device{}
	stat := models.Stat{}

	networkID, err := network.GetNetworkIdByName(networkName)
	if err != nil {
		return nil, err
	}

	deviceID, err := device.GetDeviceIdByName(deviceName, networkID)
	if err != nil {
		return nil, err
	}

	return stat.GetStats(networkID, deviceID, minute)
}
