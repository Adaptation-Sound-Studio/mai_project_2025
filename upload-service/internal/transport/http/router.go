package router

import (
	"net/http"

	"upload-service/internal/transport/http/handler"
	"upload-service/internal/transport/http/middleware"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

func NewRouter(
	genreHandler *handler.GenreHandler,
	artistHandler *handler.ArtistHandler,
	songHandler *handler.SongHandler,
	albumHandler *handler.AlbumHandler,
	redisClient *redis.Client,
) *mux.Router {

	r := mux.NewRouter()

	r.HandleFunc("/genres", genreHandler.GetAllGenres).Methods("GET")
	r.Handle("/genres", middleware.RequireAnyRole(redisClient, "admin")(http.HandlerFunc(genreHandler.CreateGenre))).Methods("POST")
	r.Handle("/genres/{genre_id}", middleware.RequireAnyRole(redisClient, "admin")(http.HandlerFunc(genreHandler.UpdateGenre))).Methods("PUT")

	r.HandleFunc("/artists", artistHandler.GetAllArtists).Methods("GET")
	r.HandleFunc("/artists/{artist_id}", artistHandler.GetArtistByID).Methods("GET")
	r.Handle("/artists/register/me", middleware.RequireAnyRole(redisClient, "artist")(http.HandlerFunc(artistHandler.RegisterArtist))).Methods("POST")
	r.Handle("/artists/{artist_id}", middleware.RequireAnyRole(redisClient, "artist", "admin")(http.HandlerFunc(artistHandler.UpdateArtist))).Methods("PUT")

	r.HandleFunc("/songs", songHandler.GetAllSongs).Methods("GET")
	r.HandleFunc("/songs/{song_id}", songHandler.GetSongByID).Methods("GET")
	r.Handle("/songs", middleware.RequireAnyRole(redisClient, "artist")(http.HandlerFunc(songHandler.CreateSong))).Methods("POST")
	r.Handle("/songs/{song_id}", middleware.RequireAnyRole(redisClient, "artist", "admin")(http.HandlerFunc(songHandler.UpdateSong))).Methods("PUT")

	r.HandleFunc("/albums", albumHandler.GetAllAlbums).Methods("GET")
	r.HandleFunc("/albums/{album_id}", albumHandler.GetAlbumByID).Methods("GET")
	r.Handle("/albums", middleware.RequireAnyRole(redisClient, "artist")(http.HandlerFunc(albumHandler.CreateAlbum))).Methods("POST")
	r.Handle("/albums/{album_id}", middleware.RequireAnyRole(redisClient, "artist", "admin")(http.HandlerFunc(albumHandler.UpdateAlbum))).Methods("PUT")

	return r
}
