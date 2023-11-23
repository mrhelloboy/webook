package main

import (
	"github.com/gin-gonic/gin"
	"github.com/mrhelloboy/wehook/web"
)

func main() {
	server := gin.Default()
	userHandler := web.NewUserHandler()
	userHandler.RegisterRouters(server)
	err := server.Run(":8080")
	if err != nil {
		panic(err)
	}
}
