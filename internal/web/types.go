package web

import "github.com/gin-gonic/gin"

type handler interface {
	RegisterRouters(server *gin.Engine)
}
