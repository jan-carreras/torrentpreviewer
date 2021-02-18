package http

type httpError struct {
	Message string `json:"message"`
}

type getTorrentResponse struct {
	Torrent Torrent `json:"torrent"`
}
