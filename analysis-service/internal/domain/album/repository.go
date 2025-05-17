package album

type AlbumRepository interface {
	Create(album *Album) error
}
