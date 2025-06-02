package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	minio "github.com/minio/minio-go/v7"

	"upload-service/internal/domain/event"
	"upload-service/internal/domain/model"
	"upload-service/internal/domain/request"
	"upload-service/internal/kafka"
	"upload-service/internal/service"
	auth "upload-service/internal/transport/http/helper"

	redislib "github.com/go-redis/redis/v8"
)

type SongHandler struct {
	Service       *service.SongService
	MinioClient   *minio.Client
	BucketName    string
	RedisClient   *redislib.Client
	KafkaProducer *kafka.Producer
}

func NewSongHandler(service *service.SongService, minioClient *minio.Client, bucketName string, redisClient *redislib.Client, producer *kafka.Producer) *SongHandler {
	return &SongHandler{
		Service:       service,
		MinioClient:   minioClient,
		BucketName:    bucketName,
		RedisClient:   redisClient,
		KafkaProducer: producer,
	}
}

func (h *SongHandler) GetAllSongs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		return
	}
	songs, err := h.Service.GetAllSongs(r.Context())
	if err != nil {
		http.Error(w, "Ошибка при получении песен", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(songs)
}

func (h *SongHandler) GetSongByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["song_id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Некорректный ID песни", http.StatusBadRequest)
		return
	}

	song, err := h.Service.GetSongByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Ошибка при получении песни", http.StatusInternalServerError)
		return
	}
	if song == nil {
		http.Error(w, "Песня не найдена", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(song); err != nil {
		http.Error(w, "Ошибка сериализации ответа", http.StatusInternalServerError)
	}
}

func (h *SongHandler) CreateSong(w http.ResponseWriter, r *http.Request) {
	currentArtistID, err := auth.GetCurrentArtistID(r, h.RedisClient)
	if err != nil {
		http.Error(w, err.Error(), err.(*auth.HttpError).Code)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 20<<20)

	if err := r.ParseMultipartForm(20 << 20); err != nil {
		http.Error(w, "Ошибка парсинга формы: "+err.Error(), http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	genreIDStr := r.FormValue("genre_id")
	genreID, err := strconv.ParseInt(genreIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Некорректный genre_id", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Не удалось прочитать файл: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	artistIDs := []int64{currentArtistID}

	song := &model.Song{
		Name:    name,
		GenreID: genreID,
	}

	err = h.Service.CreateSong(r.Context(), song, artistIDs, file, header.Filename)
	if err != nil {
		http.Error(w, "Ошибка при создании песни: "+err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Песня успешно создана"})
}

func (h *SongHandler) UpdateSong(w http.ResponseWriter, r *http.Request) {
	currentArtistID, err := auth.GetCurrentArtistID(r, h.RedisClient)
	if err != nil {
		http.Error(w, err.Error(), err.(*auth.HttpError).Code)
		return
	}

	idStr := mux.Vars(r)["song_id"]
	songID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || songID <= 0 {
		http.Error(w, "Некорректный ID песни", http.StatusBadRequest)
		return
	}

	artists, err := h.Service.GetSongArtists(r.Context(), songID)
	if err != nil {
		http.Error(w, "Ошибка при проверке прав", http.StatusInternalServerError)
		return
	}

	allowed := false
	for _, a := range artists {
		if a.ArtistID == currentArtistID {
			allowed = true
			break
		}
	}
	if !allowed {
		http.Error(w, "Вы не можете редактировать эту песню", http.StatusForbidden)
		return
	}

	var req request.UpdateSongRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	song := &model.Song{
		SongID:  songID,
		Name:    req.Name,
		GenreID: req.GenreID,
	}

	if err := h.Service.UpdateSong(r.Context(), song); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Песня успешно обновлена"})
}

func (h *SongHandler) StreamSongByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := mux.Vars(r)["song_id"]
	songID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || songID <= 0 {
		log.Printf("[Stream] Некорректный song_id: %s", idStr)
		http.Error(w, "Некорректный song_id", http.StatusBadRequest)
		return
	}
	log.Printf("[Stream] Запрос на песню с ID: %d", songID)

	song, err := h.Service.GetSongByID(ctx, songID)
	if err != nil {
		log.Printf("[Stream] Песня не найдена: %v", err)
		http.Error(w, "Песня не найдена", http.StatusNotFound)
		return
	}
	log.Printf("[Stream] Найдена песня: %+v", song)

	object, err := h.MinioClient.GetObject(
		ctx,
		h.BucketName,
		song.NameOfMinio,
		minio.GetObjectOptions{},
	)
	if err != nil {
		log.Printf("[Stream] Ошибка получения из MinIO: %v", err)
		http.Error(w, "Ошибка при получении файла", http.StatusInternalServerError)
		return
	}
	defer object.Close()

	stat, err := object.Stat()
	if err == nil {
		log.Printf("[Stream] Размер файла: %d байт", stat.Size)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size))
	} else {
		log.Printf("[Stream] Не удалось получить размер файла: %v", err)
	}

	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Accept-Ranges", "bytes")
	w.WriteHeader(http.StatusOK)

	const listenThreshold = 128 * 1024
	buffer := make([]byte, 32*1024)
	var total int64
	var counted bool

	log.Printf("[Stream] Начало стриминга...")

	for {
		n, err := object.Read(buffer)
		if n > 0 {
			total += int64(n)
			if _, writeErr := w.Write(buffer[:n]); writeErr != nil {
				log.Printf("[Stream] Ошибка при записи в ответ: %v", writeErr)
				break
			}

			if !counted && total >= listenThreshold {
				log.Printf("[Stream] Превышен порог в %d байт, пробуем отправить событие", listenThreshold)

				userID, _, err := auth.GetUserID(r, h.RedisClient)
				if err != nil {
					log.Printf("[Stream] Ошибка получения userID: %v", err)
					userID = 0
				} else {
					log.Printf("[Stream] userID = %d", userID)
				}

				artists, err := h.Service.GetArtistsBySongID(ctx, songID)
				if err != nil || len(artists) == 0 {
					log.Printf("[Stream] Не удалось получить артистов песни: %v", err)
					break
				}
				artistID := artists[0].ArtistID
				log.Printf("[Stream] Первый артист: artistID = %d", artistID)

				currentArtistID, err := auth.GetCurrentArtistID(r, h.RedisClient)
				if err == nil {
					log.Printf("[Stream] currentArtistID = %d", currentArtistID)
				} else {
					log.Printf("[Stream] Не удалось получить currentArtistID (может не артист): %v", err)
				}

				if err == nil && currentArtistID == artistID {
					log.Printf("[Stream] Артист слушает сам свою песню, событие не отправляется")
					counted = true
					continue
				}

				now := time.Now()
				fact := event.ListenFact{
					UserID:     userID,
					SongID:     songID,
					ArtistID:   artistID,
					GenreID:    song.GenreID,
					ListenedAt: &now,
				}

				log.Printf("[Stream] Отправка события в Kafka: %+v", fact)
				err = h.KafkaProducer.SendWrappedEvent("listen_fact", fact)
				if err != nil {
					log.Printf("[Stream] Ошибка при отправке события в Kafka: %v", err)
				} else {
					log.Printf("[Stream] Событие успешно отправлено в Kafka")
				}
				counted = true
			}
		}

		if err == io.EOF {
			log.Printf("[Stream] Конец файла")
			break
		}
		if err != nil {
			log.Printf("[Stream] Ошибка чтения файла: %v", err)
			break
		}
	}

	log.Printf("[Stream] Завершён стриминг песни ID: %d", songID)
}

func (h *SongHandler) SearchSongs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		return
	}

	filters := map[string]string{
		"name":   r.URL.Query().Get("name"),
		"genre":  r.URL.Query().Get("genre"),
		"artist": r.URL.Query().Get("artist"),
	}

	songs, err := h.Service.SearchSongs(r.Context(), filters)
	if err != nil {
		log.Printf("Ошибка при поиске песни: %v", err)
		http.Error(w, "Ошибка при поиске песни", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(songs)
}
