package events

import (
	"context"
	"time"

	"github.com/mrhelloboy/wehook/interactive/repository"
	"github.com/mrhelloboy/wehook/pkg/saramax"

	"github.com/IBM/sarama"
	"github.com/mrhelloboy/wehook/pkg/logger"
)

type InteractiveReadEventBatchConsumer struct {
	client sarama.Client
	repo   repository.InteractiveRepository
	l      logger.Logger
}

func NewInteractiveReadEventBatchConsumer(client sarama.Client, repo repository.InteractiveRepository, l logger.Logger) *InteractiveReadEventBatchConsumer {
	return &InteractiveReadEventBatchConsumer{client: client, repo: repo, l: l}
}

func (i *InteractiveReadEventBatchConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", i.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(), []string{"read_article"}, saramax.NewBatchHandler[ReadEvent](i.l, i.Consume))
		if err != nil {
			i.l.Error("消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (i *InteractiveReadEventBatchConsumer) Consume(msg []*sarama.ConsumerMessage, ts []ReadEvent) error {
	ids := make([]int64, 0, len(ts))
	bizs := make([]string, 0, len(ts))
	for _, evt := range ts {
		ids = append(ids, evt.Aid)
		bizs = append(bizs, "article")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := i.repo.BatchIncrReadCnt(ctx, bizs, ids)
	if err != nil {
		i.l.Error("批量增加阅读计数失败", logger.Field{Key: "ids", Value: ids}, logger.Error(err))
	}
	return nil
}
