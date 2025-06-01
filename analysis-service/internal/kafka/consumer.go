package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/segmentio/kafka-go"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/artist"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/fact"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/genre"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/song"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/service"
)

type Consumer struct {
	reader        *kafka.Reader
	factService   *service.FactService
	artistService *service.ArtistService
	songService   *service.SongService
	genreService  *service.GenreService
}

func NewConsumer(
	brokers []string,
	topic string,
	factSvc *service.FactService,
	artistSvc *service.ArtistService,
	songSvc *service.SongService,
	genreSvc *service.GenreService,
) *Consumer {
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
		reader:        reader,
		factService:   factSvc,
		artistService: artistSvc,
		songService:   songSvc,
		genreService:  genreSvc,
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

		type EventWrapper struct {
			Type    string          `json:"type"`
			Payload json.RawMessage `json:"payload"`
		}

		var wrapper EventWrapper
		if err := json.Unmarshal(m.Value, &wrapper); err != nil {
			log.Printf("Ошибка парсинга обёртки события: %v", err)
			continue
		}

		switch wrapper.Type {
		case "listen_fact":
			var event fact.ListenFact
			if err := json.Unmarshal(wrapper.Payload, &event); err != nil {
				log.Printf("Ошибка парсинга ListenFact: %v", err)
				continue
			}
			log.Printf("Получено ListenFact: %+v", event)
			if err := c.factService.Store(&event); err != nil {
				log.Printf("Ошибка записи факта: %v", err)
			}

		case "artist_created":
			var a artist.Artist
			if err := json.Unmarshal(wrapper.Payload, &a); err != nil {
				log.Printf("Ошибка парсинга Artist: %v", err)
				continue
			}
			log.Printf("Получен Artist: %+v", a)
			if err := c.artistService.CreateArtist(&a); err != nil {
				log.Printf("Ошибка записи артиста: %v", err)
			}

		case "song_created":
			var s song.Song
			if err := json.Unmarshal(wrapper.Payload, &s); err != nil {
				log.Printf("Ошибка парсинга Song: %v", err)
				continue
			}
			log.Printf("Получено событие Song: %+v", s)
			if err := c.songService.CreateSong(&s); err != nil {
				log.Printf("Ошибка записи Song: %v", err)
			}

		case "genre_created":
			var g genre.Genre
			if err := json.Unmarshal(wrapper.Payload, &g); err != nil {
				log.Printf("Ошибка парсинга Genre: %v", err)
				continue
			}
			log.Printf("Получено событие Genre: %+v", g)
			if err := c.genreService.CreateGenre(&g); err != nil {
				log.Printf("Ошибка записи Genre: %v", err)
			}
		default:
			log.Printf("Неизвестный тип события: %s", wrapper.Type)
		}

	}
}
