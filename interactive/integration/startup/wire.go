//go:build wireinject

package startup

import (
	"github.com/google/wire"
	"github.com/mrhelloboy/wehook/interactive/repository"
	"github.com/mrhelloboy/wehook/interactive/repository/cache"
	"github.com/mrhelloboy/wehook/interactive/repository/dao"
	"github.com/mrhelloboy/wehook/interactive/service"
)

var thirdProvider = wire.NewSet(
	InitRedis, InitLog, InitTestDB,
)

var interactiveSvcProvider = wire.NewSet(
	service.NewInteractiveService,
	repository.NewCachedInteractiveRepo,
	dao.NewGormInteractiveDAO,
	cache.NewRedisInteractiveCache,
)

func InitInteractiveService() service.InteractiveService {
	wire.Build(thirdProvider, interactiveSvcProvider)
	return service.NewInteractiveService(nil, nil)
}
