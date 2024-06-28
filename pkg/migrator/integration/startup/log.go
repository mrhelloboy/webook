package startup

import (
	"github.com/mrhelloboy/wehook/pkg/logger"
)

func InitLogger() logger.Logger {
	return logger.NewNopLogger()
}
