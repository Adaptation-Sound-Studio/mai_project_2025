package song

type Song struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name"`
}

type PopularSong struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Artist  string `json:"artist"`
	Album   string `json:"album"`
	Genre   string `json:"genre"`
	Listens int64  `json:"listens"`
}
