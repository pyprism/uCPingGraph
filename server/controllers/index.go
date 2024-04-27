package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/pyprism/uCPingGraph/models"
	"log"
	"net/http"
)

type IndexController struct{}

func (n *IndexController) Home(c *gin.Context) {
	networkModel := models.Network{}
	deviceModel := models.Device{}
	networks, err := networkModel.GetAllNetworkName()
	devices, err := deviceModel.Get
	if err != nil {
		log.Println(err.Error())
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"htmlTitle": "μCPingGraph: Home",
		"networks":  networks,
	})
}
