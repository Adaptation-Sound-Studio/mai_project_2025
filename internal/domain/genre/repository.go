package genre

type GenreRepository interface {
	Create(genre *Genre) error
}
