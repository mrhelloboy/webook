//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/mrhelloboy/wehook/interactive/events"
	"github.com/mrhelloboy/wehook/interactive/grpc"
	"github.com/mrhelloboy/wehook/interactive/ioc"
	"github.com/mrhelloboy/wehook/interactive/repository"
	"github.com/mrhelloboy/wehook/interactive/repository/cache"
	"github.com/mrhelloboy/wehook/interactive/repository/dao"
	"github.com/mrhelloboy/wehook/interactive/service"
)

var thirdPartySet = wire.NewSet(
	ioc.InitDST,
	ioc.InitSRC,
	ioc.InitBizDB,
	ioc.InitDoubleWritePool,
	ioc.InitLogger,
	ioc.InitKafka,
	ioc.InitSyncProducer,
	ioc.InitRedis,
)

var interactiveSvcProvider = wire.NewSet(
	service.NewInteractiveService,
	repository.NewCachedInteractiveRepo,
	dao.NewGormInteractiveDAO,
	cache.NewRedisInteractiveCache,
)

var migratorProvider = wire.NewSet(
	ioc.InitMigratorWeb,
	ioc.InitFixDataConsumer,
	ioc.InitMigradatorProducer,
)

func InitAPP() *App {
	wire.Build(
		thirdPartySet,
		interactiveSvcProvider,
		migratorProvider,
		events.NewInteractiveReadEventConsumer,
		grpc.NewInteractiveServiceServer,
		ioc.NewConsumers,
		ioc.InitGRPCxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
