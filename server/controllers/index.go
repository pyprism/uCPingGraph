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

func (n *IndexController) Home(c *gin.Context) {
	networkModel := models.Network{}
	// deviceModel := models.Device{}
	networks, err := networkModel.GetAllNetworkName()
	// devices, err := deviceModel.Get
	if err != nil {
		log.Println(err.Error())
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"htmlTitle": "μCPingGraph: Home",
		"networks":  networks,
	})
}

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
