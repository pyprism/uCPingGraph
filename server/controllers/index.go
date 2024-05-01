package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/pyprism/uCPingGraph/models"
	"log"
	"net/http"
)

type IndexController struct{}

type network struct {
	Name string `json:"name" binding:"required"`
}

type statsPost struct {
	NetworkName string `json:"network_name" binging:"required"`
	DeviceName  string `json:"device_name" binging:"required"`
}

// Home route /
func (n *IndexController) Home(c *gin.Context) {
	networkModel := models.Network{}
	networks, err := networkModel.GetAllNetworkName()
	if err != nil {
		log.Println(err.Error())
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"htmlTitle": "μCPingGraph: Home",
		"networks":  networks,
	})
}

// GetDeviceList route /device/
func (n *IndexController) GetDeviceList(c *gin.Context) {
	var network network
	err := c.BindJSON(&network)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	deviceModel := models.Device{}
	networkModel := models.Network{}

	networkId, err := networkModel.GetNetworkIdByName(network.Name)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	devices, err := deviceModel.GetDevicesByNetwork(int(networkId))

	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, devices)

}

// Chart returns json for chart generation; route /chart/
func (n *IndexController) Chart(c *gin.Context) {
	var statusPost statsPost
	err := c.BindJSON(&statusPost)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	deviceModel := models.Device{}
	networkModel := models.Network{}
	statModel := models.Stat{}

	networkId, err := networkModel.GetNetworkIdByName(statusPost.NetworkName)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	deviceId, err := deviceModel.GetDeviceIdByName(statusPost.DeviceName, networkId)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	data, err := statModel.GetStats(networkId, deviceId)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, data)
}
