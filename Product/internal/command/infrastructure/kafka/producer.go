// internal/command/infrastructure/kafka/producer.go
package kafka

import (
	"Taller2/Product/internal/command/domain/events"
	"context"
	_ "errors"

	"github.com/segmentio/kafka-go"
)

// Renombra la interfaz para que coincida con lo que espera command
type EventProducer interface {
	PublishEvent(ctx context.Context, event events.ProductEvent) error
}

// Implementation
type Producer struct {
	writer *kafka.Writer
}

// Asegura que Producer implementa la interfaz
var _ EventProducer = (*Producer)(nil)

func NewProducer(brokers []string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    "product_events",
			Balancer: &kafka.Hash{},
		},
	}
}

func (p *Producer) PublishEvent(ctx context.Context, event events.ProductEvent) error {
	eventBytes, err := event.ToJSON()
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Value: eventBytes,
		Key:   []byte(event.EventType()), // Usamos el nuevo método
	})
}
