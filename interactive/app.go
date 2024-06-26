package main

import (
	"github.com/mrhelloboy/wehook/pkg/grpcx"
	"github.com/mrhelloboy/wehook/pkg/saramax"
)

type App struct {
	server    *grpcx.Server
	consumers []saramax.Consumer
}
