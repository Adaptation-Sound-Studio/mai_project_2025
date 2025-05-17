package song

type SongRepository interface {
	Create(song *Song) error
	GetPopularSongs(limit int) ([]PopularSong, error)
}
