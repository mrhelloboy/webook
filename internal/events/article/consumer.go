package article

import (
	"context"
	"time"

	"github.com/mrhelloboy/wehook/pkg/saramax"

	"github.com/IBM/sarama"
	"github.com/mrhelloboy/wehook/internal/repository"
	"github.com/mrhelloboy/wehook/pkg/logger"
)

type InteractiveReadEventConsumer struct {
	client sarama.Client
	repo   repository.InteractiveRepository
	l      logger.Logger
}

func (i *InteractiveReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", i.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{"article_read"},
			saramax.NewHandler[ReadEvent](i.l, i.Consume))
		if err != nil {
			i.l.Error("消费循环异常", logger.Error(err))
		}
	}()
	return err
}

// Consume 这个不是幂等的
func (i *InteractiveReadEventConsumer) Consume(msg *sarama.ConsumerMessage, t ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return i.repo.IncrReadCnt(ctx, "article", t.Aid)
}

func NewInteractiveReadEventConsumer(client sarama.Client, repo repository.InteractiveRepository, l logger.Logger) *InteractiveReadEventConsumer {
	return &InteractiveReadEventConsumer{
		client: client,
		repo:   repo,
		l:      l,
	}
}
