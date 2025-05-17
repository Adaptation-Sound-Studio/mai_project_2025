package artist

type ArtistRepository interface {
	Create(artist *Artist) error
}
