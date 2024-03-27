package web

import (
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ObservabilityHandler struct{}

func (o *ObservabilityHandler) RegisterRouters(server *gin.Engine) {
	g := server.Group("test")
	g.GET("/metric", func(ctx *gin.Context) {
		sleep := rand.Int32N(1000)
		time.Sleep(time.Millisecond * time.Duration(sleep))
		ctx.String(http.StatusOK, "ok")
	})
}
