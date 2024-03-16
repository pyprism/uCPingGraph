package controllers

import (
	"errors"
	"github.com/pyprism/uCPingGraph/models"
)

func SaveStats(token string, latency float32) error {
	device := models.Device{}
	stat := models.Stat{}

	// get device id and network id from token
	deviceID, networkID, err := device.GetDeviceByToken(token)
	if err != nil {
		return errors.New("authorization header is invalid")
	}

	// create new stat
	err = stat.CreateStat(networkID, int(deviceID), latency)
	if err != nil {
		return errors.New("db error: " + err.Error())
	}
	return nil
}
