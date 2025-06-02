package song

type SongRepository interface {
	Create(song *Song) error
	GetPopularSongs(limit int) ([]PopularSong, error)
	GetTopSongsForUser(userID int, limit int) ([]Song, error)
	GetMostPopularSongs(limit int) ([]Song, error)
}
