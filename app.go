package main

import (
	"github.com/gin-gonic/gin"
	"github.com/mrhelloboy/wehook/internal/events"
)

type App struct {
	web       *gin.Engine
	consumers []events.Consumer
}
