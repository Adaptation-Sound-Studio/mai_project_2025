package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

type ProducerIface interface {
	SendWrappedEvent(eventType string, payload interface{}) error
}

var _ ProducerIface = (*Producer)(nil)

func NewProducer(brokers []string, topic string) (*Producer, error) {
	caCert, err := os.ReadFile("/certs/ca.crt")
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ca.crt: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("ошибка добавления ca.crt в пул доверенных")
	}

	tlsConfig := &tls.Config{
		RootCAs:            caPool,
		InsecureSkipVerify: false,
	}

	dialer := &kafka.Dialer{
		TLS: tlsConfig,
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
		Dialer:   dialer,
	})

	return &Producer{writer: writer}, nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

func (p *Producer) SendWrappedEvent(eventType string, payload interface{}) error {
	wrapped := struct {
		Type    string      `json:"type"`
		Payload interface{} `json:"payload"`
	}{
		Type:    eventType,
		Payload: payload,
	}

	data, err := json.Marshal(wrapped)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(eventType),
		Value: data,
	}

	return p.writer.WriteMessages(context.Background(), msg)
}
