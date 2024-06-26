package ioc

import (
	"github.com/mrhelloboy/wehook/pkg/logger"
	"go.uber.org/zap"
)

func InitLogger() logger.Logger {
	zLog, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return logger.NewZapLogger(zLog)
}
