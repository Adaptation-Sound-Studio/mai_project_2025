package event

type SongCreatedEvent struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
