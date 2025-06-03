package http

import (
	"database/sql"
	"net/http"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/repository/postgres"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/service"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/transport/http/handler"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/transport/http/middleware"
	"github.com/gorilla/mux"
)

func NewRouter(db *sql.DB, adminSecret string, expectedApiKey string) *mux.Router {
	artistrepo := postgres.NewArtistRepo(db)
	artistService := service.NewArtistService(artistrepo)
	artistHandler := handler.NewArtistHandler(artistService, adminSecret)

	songRepo := postgres.NewSongRepo(db)
	songService := service.NewSongService(songRepo)
	songHandler := handler.NewSongHandler(songService, adminSecret)

	genreRepo := postgres.NewGenreRepo(db)
	genreService := service.NewGenreService(genreRepo)
	genreHandler := handler.NewGenreHandler(genreService, adminSecret)

	factRepo := postgres.NewFactRepo(db)
	factService := service.NewFactService(factRepo)
	factHandler := handler.NewFactHandler(factService, adminSecret)

	apiKeyMiddleware := middleware.ApiKeyMiddleware(expectedApiKey)

	router := mux.NewRouter()
	router.HandleFunc("/artists", artistHandler.CreateArtist).Methods("POST")
	router.HandleFunc("/songs", songHandler.CreateSong).Methods("POST")
	router.HandleFunc("/genres", genreHandler.CreateGenre).Methods("POST")
	router.HandleFunc("/facts", factHandler.CreateFact).Methods("POST")
	router.HandleFunc("/songs/admin-popular", songHandler.GetPopularSongs).Methods("GET")

	router.Handle("/users/{user_id:[0-9]+}/top-artists", apiKeyMiddleware(http.HandlerFunc(artistHandler.GetTopArtistsForUser))).Methods("GET")
	router.Handle("/users/{user_id:[0-9]+}/top-genres", apiKeyMiddleware(http.HandlerFunc(genreHandler.GetTopGenresForUser))).Methods("GET")
	router.Handle("/users/{user_id:[0-9]+}/top-songs", apiKeyMiddleware(http.HandlerFunc(songHandler.GetTopSongsForUser))).Methods("GET")

	router.Handle("/artists/popular", apiKeyMiddleware(http.HandlerFunc(artistHandler.GetMostPopularArtists))).Methods("GET")
	router.Handle("/songs/popular", apiKeyMiddleware(http.HandlerFunc(songHandler.GetMostPopularSongs))).Methods("GET")
	router.Handle("/genres/popular", apiKeyMiddleware(http.HandlerFunc(genreHandler.GetMostPopularGenres))).Methods("GET")

	router.Handle("/genres/top/night", apiKeyMiddleware(http.HandlerFunc(genreHandler.GetTopGenresAtNight))).Methods("GET")
	router.Handle("/genres/top/morning", apiKeyMiddleware(http.HandlerFunc(genreHandler.GetTopGenresInMorning))).Methods("GET")
	router.Handle("/genres/top/day", apiKeyMiddleware(http.HandlerFunc(genreHandler.GetTopGenresInDay))).Methods("GET")
	router.Handle("/genres/top/evening", apiKeyMiddleware(http.HandlerFunc(genreHandler.GetTopGenresInEvening))).Methods("GET")

	return router
}
