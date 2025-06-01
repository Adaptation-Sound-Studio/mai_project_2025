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

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Песня успешно создана"})
	}
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
		http.Error(w, "Некорректный song_id", http.StatusBadRequest)
		return
	}

	song, err := h.Service.GetSongByID(ctx, songID)
	if err != nil {
		http.Error(w, "Песня не найдена", http.StatusNotFound)
		return
	}

	object, err := h.MinioClient.GetObject(
		ctx,
		h.BucketName,
		song.NameOfMinio,
		minio.GetObjectOptions{},
	)
	if err != nil {
		http.Error(w, "Ошибка при получении файла", http.StatusInternalServerError)
		return
	}
	defer object.Close()

	// Получим размер файла, чтобы выставить заголовок Content-Length
	stat, err := object.Stat()
	if err == nil {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size))
	}

	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Accept-Ranges", "bytes")
	w.WriteHeader(http.StatusOK)

	const listenThreshold = 128 * 1024
	buffer := make([]byte, 32*1024)
	var total int64
	var counted bool

	for {
		n, err := object.Read(buffer)
		if n > 0 {
			total += int64(n)
			if _, writeErr := w.Write(buffer[:n]); writeErr != nil {
				break
			}

			if !counted && total >= listenThreshold {
				userID, _, err := auth.GetUserID(r, h.RedisClient)
				if err != nil {
					userID = 0
				}

				artists, err := h.Service.GetArtistsBySongID(ctx, songID)
				if err != nil || len(artists) == 0 {
					continue
				}
				artistID := artists[0].ArtistID

				currentArtistID, err := auth.GetCurrentArtistID(r, h.RedisClient)
				if err == nil && currentArtistID == artistID {
					counted = true
					continue
				}

				album, err := h.Service.GetAlbumBySongID(ctx, songID)
				if err != nil {
					continue
				}

				fact := model.ListenFact{
					UserID:     userID,
					SongID:     songID,
					ArtistID:   artistID,
					AlbumID:    album.AlbumID,
					GenreID:    song.GenreID,
					ListenedAt: time.Now(),
				}

				if err := h.KafkaProducer.SendListenFact(fact); err != nil {
					log.Printf("Kafka send error: %v", err)
				}
				counted = true
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Ошибка чтения mp3: %v", err)
			break
		}
	}
}
