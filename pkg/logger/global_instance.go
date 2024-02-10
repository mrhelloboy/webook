package logger

import "sync"

// 类似 zap 的用法
// 这样 logger 就可以避免在每个地方都进行依赖注入

var gl Logger
var lMutex sync.RWMutex

func SetGlobalLogger(logger Logger) {
	lMutex.Lock()
	defer lMutex.Unlock()
	gl = logger
}

func L() Logger {
	lMutex.RLock()
	g := gl
	lMutex.Unlock()
	return g
}
