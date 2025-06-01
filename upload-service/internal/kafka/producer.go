package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"

	"upload-service/internal/domain/model"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

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

func (p *Producer) SendListenFact(fact model.ListenFact) error {
	data, err := json.Marshal(fact)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(fmt.Sprintf("%d", fact.UserID)),
		Value: data,
	}

	return p.writer.WriteMessages(context.Background(), msg)
}
