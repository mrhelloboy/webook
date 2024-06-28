// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

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

// Injectors from wire.go:

func InitAPP() *App {
	logger := ioc.InitLogger()
	srcDB := ioc.InitSRC(logger)
	dstDB := ioc.InitDST(logger)
	doubleWritePool := ioc.InitDoubleWritePool(srcDB, dstDB)
	db := ioc.InitBizDB(doubleWritePool)
	interactiveDAO := dao.NewGormInteractiveDAO(db)
	cmdable := ioc.InitRedis()
	interactiveCache := cache.NewRedisInteractiveCache(cmdable)
	interactiveRepository := repository.NewCachedInteractiveRepo(interactiveDAO, interactiveCache, logger)
	interactiveService := service.NewInteractiveService(interactiveRepository, logger)
	interactiveServiceServer := grpc.NewInteractiveServiceServer(interactiveService)
	server := ioc.InitGRPCxServer(interactiveServiceServer)
	client := ioc.InitKafka()
	interactiveReadEventConsumer := events.NewInteractiveReadEventConsumer(client, interactiveRepository, logger)
	consumer := ioc.InitFixDataConsumer(logger, srcDB, dstDB, client)
	v := ioc.NewConsumers(interactiveReadEventConsumer, consumer)
	syncProducer := ioc.InitSyncProducer(client)
	producer := ioc.InitMigradatorProducer(syncProducer)
	ginxServer := ioc.InitMigratorWeb(logger, srcDB, dstDB, doubleWritePool, producer)
	app := &App{
		server:    server,
		consumers: v,
		webAdmin:  ginxServer,
	}
	return app
}

// wire.go:

var thirdPartySet = wire.NewSet(ioc.InitDST, ioc.InitSRC, ioc.InitBizDB, ioc.InitDoubleWritePool, ioc.InitLogger, ioc.InitKafka, ioc.InitSyncProducer, ioc.InitRedis)

var interactiveSvcProvider = wire.NewSet(service.NewInteractiveService, repository.NewCachedInteractiveRepo, dao.NewGormInteractiveDAO, cache.NewRedisInteractiveCache)

var migratorProvider = wire.NewSet(ioc.InitMigratorWeb, ioc.InitFixDataConsumer, ioc.InitMigradatorProducer)
