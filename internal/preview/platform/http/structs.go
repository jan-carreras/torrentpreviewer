package http

type Image struct {
	Src     string `json:"source"`
	Length  int    `json:"length"`
	IsValid bool   `json:"is_valid"`
}

type File struct {
	ID          int     `json:"id"`
	Length      int     `json:"length"`
	IsSupported bool    `json:"is_supported"`
	Name        string  `json:"name"`
	Images      []Image `json:"images"`
}

type Torrent struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Length int    `json:"length"`
	Files  []File `json:"files"`
}
