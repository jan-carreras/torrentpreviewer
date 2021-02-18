package http

type Image struct {
	Src string `json:"source"`
}

type File struct {
	ID     int     `json:"id"`
	Length int     `json:"length"`
	Name   string  `json:"name"`
	Images []Image `json:"images"`
}

type Torrent struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Length int    `json:"length"`
	Files  []File `json:"files"`
}
