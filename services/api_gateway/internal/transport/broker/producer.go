package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type EventProducer interface {
	PublishPredictionEvent(ctx context.Context, event PredictionEvent) error
	Close()
}

type KafkaProducer struct {
	writer *kafka.Writer
	lg     *zap.Logger
}

var _ EventProducer = (*KafkaProducer)(nil)

func NewProducer(brokers []string, topic string, lg *zap.Logger) *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
			ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
				lg.Error(fmt.Sprintf(msg, args...))
			}),
			WriteTimeout: 10 * time.Second,
			ReadTimeout:  10 * time.Second,
		},
		lg: lg,
	}
}

func (p *KafkaProducer) PublishPredictionEvent(ctx context.Context, event PredictionEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		p.lg.Error("failed to marshal prediction event",
			zap.Error(err),
			zap.String("user_id", event.UserID),
		)
		return err
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.UserID),
		Value: payload,
	})

	if err != nil {
		p.lg.Error("failed to publish message to kafka",
			zap.Error(err),
			zap.String("topic", p.writer.Topic),
			zap.String("user_id", event.UserID),
		)
		return err
	}

	return nil
}

func (p *KafkaProducer) Close() {
	if err := p.writer.Close(); err != nil {
		p.lg.Error("failed to close kafka writer", zap.Error(err))
	}
}

type PredictionEvent struct {
	UserID       string    `json:"user_id"`
	ExperimentID string    `json:"experiment_id"`
	VariantID    string    `json:"variant_id"`
	IsControl    bool      `json:"is_control"`
	Timestamp    time.Time `json:"timestamp"`
}
