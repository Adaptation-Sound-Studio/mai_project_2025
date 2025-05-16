package response

import (
	"time"
	"upload-service/internal/domain"
)

type SongResponse struct {
	SongID    int64           `json:"song_id"`
	Name      string          `json:"name"`
	Auditions int64           `json:"auditions"`
	GenreID   int64           `json:"genre_id"`
	Date      time.Time       `json:"date"`
	Link      string          `json:"link"`
	Artists   []domain.Artist `json:"artists"`
	Album     *domain.Album   `json:"album,omitempty"`
}

func MapSongToResponse(song *domain.Song, artists []domain.Artist, album *domain.Album) SongResponse {
	return SongResponse{
		SongID:    song.SongID,
		Name:      song.Name,
		Auditions: song.Auditions,
		GenreID:   song.GenreID,
		Date:      song.Date,
		Link:      song.Link,
		Artists:   artists,
		Album:     album,
	}
}
