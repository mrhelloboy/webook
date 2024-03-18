package article

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"
)

type Producer interface {
	ProduceReadEvent(ctx context.Context, evt ReadEvent) error
}

type kafkaProducer struct {
	producer sarama.SyncProducer
}

func NewKafkaProducer(producer sarama.SyncProducer) Producer {
	return &kafkaProducer{producer: producer}
}

// ProduceReadEvent 阅读文章事件
// 如果有复杂的重试逻辑，用装饰器封装处理
func (k *kafkaProducer) ProduceReadEvent(ctx context.Context, evt ReadEvent) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = k.producer.SendMessage(&sarama.ProducerMessage{
		Topic: "read_article",
		Value: sarama.ByteEncoder(data),
	})
	return err
}

type ReadEvent struct {
	Uid int64
	Aid int64
}
