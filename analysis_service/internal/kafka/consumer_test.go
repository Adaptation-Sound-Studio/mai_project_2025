package kafka

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/fact"
)

type Reader interface {
	ReadMessage(context.Context) (kafka.Message, error)
	Close() error
}

type FactService interface {
	Store(*fact.ListenFact) error
}

type MockReader struct {
	mock.Mock
	messages []kafka.Message
	i        int
}

func (m *MockReader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	if m.i >= len(m.messages) {
		return kafka.Message{}, context.Canceled
	}
	msg := m.messages[m.i]
	m.i++
	return msg, nil
}

func (m *MockReader) Close() error {
	return nil
}

type MockFactService struct {
	mock.Mock
}

func (m *MockFactService) Store(f *fact.ListenFact) error {
	args := m.Called(f)
	return args.Error(0)
}

type testConsumer struct {
	reader  Reader
	service FactService
}

func (c *testConsumer) Start(ctx context.Context) {
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			break
		}

		var event fact.ListenFact
		if err := json.Unmarshal(m.Value, &event); err != nil {
			continue
		}

		c.service.Store(&event)
	}
}

func TestConsumer_Start_HandlesValidAndInvalidJSON(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	validEvent := fact.ListenFact{
		UserID:   1,
		SongID:   2,
		ArtistID: 3,
	}
	validJSON, _ := json.Marshal(validEvent)

	messages := []kafka.Message{
		{Value: []byte("not-json")},
		{Value: validJSON},
	}

	mockReader := &MockReader{messages: messages}
	mockService := &MockFactService{}

	mockService.
		On("Store", mock.AnythingOfType("*fact.ListenFact")).
		Return(nil).
		Once()

	consumer := &testConsumer{
		reader:  mockReader,
		service: mockService,
	}

	go consumer.Start(ctx)

	time.Sleep(200 * time.Millisecond)

	mockService.AssertExpectations(t)
	require.Equal(t, 2, mockReader.i)
}
