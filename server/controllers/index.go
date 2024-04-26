package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type IndexController struct{}

func (n *IndexController) Home(c *gin.Context) {

	c.HTML(http.StatusOK, "index.html", gin.H{
		"htmlTitle": "μCPingGraph: Home",
	})
}
