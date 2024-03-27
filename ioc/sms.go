package ioc

import (
	"github.com/mrhelloboy/wehook/internal/service/sms"
	"github.com/mrhelloboy/wehook/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	// return metrics.NewPrometheusDecorator(memory.NewService())
	return memory.NewService()
}
