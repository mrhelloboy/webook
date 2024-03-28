package main

import (
	"github.com/gin-gonic/gin"
	"github.com/mrhelloboy/wehook/internal/events"
	"github.com/robfig/cron/v3"
)

type App struct {
	web       *gin.Engine
	consumers []events.Consumer
	cron      *cron.Cron
}
