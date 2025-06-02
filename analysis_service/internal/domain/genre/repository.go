package genre

type GenreRepository interface {
	Create(genre *Genre) error
	GetTopGenresForUser(userID int, limit int) ([]Genre, error)
	GetMostPopularGenres(limit int) ([]Genre, error)
	GetTopGenresAtNight(limit int) ([]Genre, error)
	GetTopGenresInMorning(limit int) ([]Genre, error)
	GetTopGenresInDay(limit int) ([]Genre, error)
	GetTopGenresInEvening(limit int) ([]Genre, error)
}
