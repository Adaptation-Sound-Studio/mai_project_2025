package album

type Album struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name"`
}
