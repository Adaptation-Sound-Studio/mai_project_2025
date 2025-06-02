package router

import (
	"net/http"

	"upload-service/internal/config"
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
	cfg *config.Config,
) *mux.Router {

	r := mux.NewRouter()

	r.Handle("/genres/search", middleware.RequireAnyRole(redisClient)(http.HandlerFunc(genreHandler.SearchGenres))).Methods("GET")
	r.Handle("/genres", middleware.RequireAnyRole(redisClient)(http.HandlerFunc(genreHandler.GetAllGenres))).Methods("GET")
	r.Handle("/genres", middleware.RequireAnyRole(redisClient, "admin")(http.HandlerFunc(genreHandler.CreateGenre))).Methods("POST")
	r.Handle("/genres/{genre_id}", middleware.RequireAnyRole(redisClient, "admin")(http.HandlerFunc(genreHandler.UpdateGenre))).Methods("PUT")

	r.Handle("/artists/search", middleware.RequireAnyRole(redisClient)(http.HandlerFunc(artistHandler.SearchArtists))).Methods("GET")
	r.Handle("/artists/user/{user_id}", middleware.RequireServiceAuth(cfg.APIKey)(http.HandlerFunc(artistHandler.GetArtistIDByUserID))).Methods("GET")
	r.Handle("/artists", middleware.RequireAnyRole(redisClient)(http.HandlerFunc(artistHandler.GetAllArtists))).Methods("GET")
	r.Handle("/artists/{artist_id}", middleware.RequireAnyRole(redisClient)(http.HandlerFunc(artistHandler.GetArtistByID))).Methods("GET")
	r.Handle("/artists/register/me", middleware.RequireAnyRole(redisClient)(http.HandlerFunc(artistHandler.RegisterArtist))).Methods("POST")
	r.Handle("/artists/{artist_id}", middleware.RequireAnyRole(redisClient, "artist", "admin")(http.HandlerFunc(artistHandler.UpdateArtist))).Methods("PUT")

	r.Handle("/songs/search", middleware.RequireAnyRole(redisClient)(http.HandlerFunc(songHandler.SearchSongs))).Methods("GET")
	r.Handle("/songs", middleware.RequireAnyRole(redisClient)(http.HandlerFunc(songHandler.GetAllSongs))).Methods("GET")
	r.Handle("/songs/{song_id}", middleware.RequireAnyRole(redisClient)(http.HandlerFunc(songHandler.GetSongByID))).Methods("GET")
	r.Handle("/songs", middleware.RequireAnyRole(redisClient, "artist", "admin")(http.HandlerFunc(songHandler.CreateSong))).Methods("POST")
	r.Handle("/songs/{song_id}", middleware.RequireAnyRole(redisClient, "artist", "admin")(http.HandlerFunc(songHandler.UpdateSong))).Methods("PUT")
	r.Handle("/songs/{song_id}/stream", middleware.RequireAnyRole(redisClient)(http.HandlerFunc(songHandler.StreamSongByID))).Methods("GET")

	r.Handle("/albums", middleware.RequireAnyRole(redisClient)(http.HandlerFunc(albumHandler.GetAllAlbums))).Methods("GET")
	r.Handle("/albums/{album_id}", middleware.RequireAnyRole(redisClient)(http.HandlerFunc(albumHandler.GetAlbumByID))).Methods("GET")
	r.Handle("/albums", middleware.RequireAnyRole(redisClient, "artist")(http.HandlerFunc(albumHandler.CreateAlbum))).Methods("POST")
	r.Handle("/albums/{album_id}", middleware.RequireAnyRole(redisClient, "artist", "admin")(http.HandlerFunc(albumHandler.UpdateAlbum))).Methods("PUT")

	return r
}
