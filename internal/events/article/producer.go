package article

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"
)

type Producer interface {
	ProduceReadEvent(ctx context.Context, evt ReadEvent) error
	ProduceReadEventV1(ctx context.Context, evts ReadEventV1)
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

// ProduceReadEventV1 批量阅读文章事件(批量方式)
func (k *kafkaProducer) ProduceReadEventV1(ctx context.Context, evts ReadEventV1) {
	// TODO implement me
	panic("implement me")
}

type ReadEvent struct {
	Uid int64
	Aid int64
}

type ReadEventV1 struct {
	Uids []int64
	Aids []int64
}
