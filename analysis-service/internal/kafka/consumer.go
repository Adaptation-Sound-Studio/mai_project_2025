package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/segmentio/kafka-go"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/fact"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/service"
)

type Consumer struct {
	reader  *kafka.Reader
	service *service.FactService
}

func NewConsumer(brokers []string, topic string, service *service.FactService) *Consumer {
	// Загрузка корневого CA-сертификата
	caCert, err := ioutil.ReadFile("/certs/ca.crt")
	if err != nil {
		log.Fatal("Ошибка чтения ca.crt:", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		log.Fatal("Ошибка добавления ca.crt в пул доверенных")
	}

	// TLS-конфигурация
	tlsConfig := &tls.Config{
		RootCAs:            caPool,
		InsecureSkipVerify: false, // проверяет, что CN=localhost
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: "analytics-consumer",
		Dialer: &kafka.Dialer{
			TLS: tlsConfig,
		},
		MinBytes: 1,
		MaxBytes: 10e6,
	})

	return &Consumer{
		reader:  reader,
		service: service,
	}
}

func (c *Consumer) Start(ctxt context.Context) {
	log.Println("Kafka консьюмер запущен...")
	defer c.reader.Close()

	for {
		m, err := c.reader.ReadMessage(ctxt)
		if err != nil {
			log.Printf("Ошибка чтения сообщения: %v", err)
			continue
		}

		var event fact.ListenFact
		if err := json.Unmarshal(m.Value, &event); err != nil {
			log.Printf("Ошибка парсинга сообщения: %v", err)
			continue
		}

		log.Printf("Получено: %+v", event)

		if err := c.service.Store(&event); err != nil {
			log.Printf("Ошибка записи факта: %v", err)
		}
	}
}
