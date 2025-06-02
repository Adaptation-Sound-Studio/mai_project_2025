package artist

type ArtistRepository interface {
	Create(artist *Artist) error
	GetTopArtistsForUser(userID int, limit int) ([]Artist, error)
	GetMostPopularArtists(limit int) ([]Artist, error)
}
