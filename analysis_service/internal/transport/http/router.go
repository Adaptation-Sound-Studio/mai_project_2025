package http

import (
	"database/sql"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/repository/postgres"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/service"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/transport/http/handler"
	"github.com/gorilla/mux"
)

func NewRouter(db *sql.DB, adminSecret string) *mux.Router {
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

	router := mux.NewRouter()
	router.HandleFunc("/artists", artistHandler.CreateArtist).Methods("POST")
	router.HandleFunc("/songs", songHandler.CreateSong).Methods("POST")
	router.HandleFunc("/genres", genreHandler.CreateGenre).Methods("POST")
	router.HandleFunc("/facts", factHandler.CreateFact).Methods("POST")
	router.HandleFunc("/songs/popular", songHandler.GetPopularSongs).Methods("GET")

	return router
}
