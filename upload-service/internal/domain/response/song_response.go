package response

import (
	"time"
	"upload-service/internal/domain/model"
)

type SongResponse struct {
	SongID    int64          `json:"song_id"`
	Name      string         `json:"name"`
	URL       string         `json:"url"`
	Auditions int64          `json:"auditions"`
	GenreID   int64          `json:"genre_id"`
	Date      time.Time      `json:"date"`
	Artists   []model.Artist `json:"artists"`
	Album     *model.Album   `json:"album,omitempty"`
}

func MapSongToResponse(song *model.Song, artists []model.Artist, album *model.Album) SongResponse {
	return SongResponse{
		SongID:    song.SongID,
		Name:      song.Name,
		URL:       song.URL,
		Auditions: song.Auditions,
		GenreID:   song.GenreID,
		Date:      song.Date,
		Artists:   artists,
		Album:     album,
	}
}
