package ioc

import (
	"github.com/mrhelloboy/wehook/internal/service/sms"
	"github.com/mrhelloboy/wehook/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	return memory.NewService()
}
