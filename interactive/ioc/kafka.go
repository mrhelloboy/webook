package ioc

import (
	"github.com/IBM/sarama"
	"github.com/mrhelloboy/wehook/interactive/events"
	"github.com/mrhelloboy/wehook/interactive/repository/dao"
	"github.com/mrhelloboy/wehook/pkg/migrator/events/fixer"
	"github.com/mrhelloboy/wehook/pkg/saramax"

	"github.com/spf13/viper"
)

func InitKafka() sarama.Client {
	type Config struct {
		Addrs []string `yaml:"addrs"`
	}
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	client, err := sarama.NewClient(cfg.Addrs, saramaCfg)
	if err != nil {
		panic(err)
	}
	return client
}

func InitSyncProducer(client sarama.Client) sarama.SyncProducer {
	res, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}
	return res
}

// 规避 wire 的问题
type fixerInteractive *fixer.Consumer[dao.Interactive]

func NewConsumers(intr *events.InteractiveReadEventConsumer, fix *fixer.Consumer[dao.Interactive]) []saramax.Consumer {
	return []saramax.Consumer{
		intr,
		fix,
	}
}
