package startup

import "github.com/mrhelloboy/wehook/pkg/logger"

func InitLog() logger.Logger {
	return logger.NewNopLogger()
}
