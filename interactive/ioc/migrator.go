package ioc

import (
	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/mrhelloboy/wehook/interactive/repository/dao"
	"github.com/mrhelloboy/wehook/pkg/ginx"
	"github.com/mrhelloboy/wehook/pkg/gormx/connpool"
	"github.com/mrhelloboy/wehook/pkg/logger"
	"github.com/mrhelloboy/wehook/pkg/migrator/events"
	"github.com/mrhelloboy/wehook/pkg/migrator/events/fixer"
	"github.com/mrhelloboy/wehook/pkg/migrator/scheduler"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
)

const topic = "migrator_interactives"

func InitFixDataConsumer(l logger.Logger, src SrcDB, dst DstDB, client sarama.Client) *fixer.Consumer[dao.Interactive] {
	res, err := fixer.NewConsumer[dao.Interactive](client, l, topic, src, dst)
	if err != nil {
		panic(err)
	}
	return res
}

func InitMigradatorProducer(p sarama.SyncProducer) events.Producer {
	return events.NewSaramaProducer(p, topic)
}

func InitMigratorWeb(l logger.Logger, src SrcDB, dst DstDB, pool *connpool.DoubleWritePool, producer events.Producer) *ginx.Server {
	intrSch := scheduler.NewScheduler[dao.Interactive](src, dst, l, pool, producer)
	engine := gin.Default()
	ginx.InitCounter(prometheus.CounterOpts{
		Namespace: "geekbang_daming",
		Subsystem: "webook_intr_admin",
		Name:      "http_biz_code",
		Help:      "HTTP 的业务错误码",
	})
	intrSch.RegisterRoutes(engine.Group("/migrator"))
	addr := viper.GetString("migrator.web.addr")
	return &ginx.Server{
		Addr:   addr,
		Engine: engine,
	}
}
