package web

import "github.com/gin-gonic/gin"

type Handler interface {
	RegisterRouters(server *gin.Engine)
}
