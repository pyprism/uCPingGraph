package controllers

import "github.com/gin-gonic/gin"

type CommonController struct{}

func (n *CommonController) StaticFile(context *gin.Context) {
	static := context.Param("static")
	context.File("./static/" + static)
}
