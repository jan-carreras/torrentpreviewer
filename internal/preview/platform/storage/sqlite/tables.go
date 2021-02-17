package sqlite

const (
	sqlTorrentTable = "torrents"
	sqlFileTable    = "files"
	sqlMediaTable   = "media"
)

type torrent struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	Length      int    `db:"length"`
	PieceLength int    `db:"pieceLength"`
	Raw         []byte `db:"raw"`
}

type file struct {
	TorrentID string `db:"torrent_id"`
	ID        int    `db:"id"`
	Name      string `db:"name"`
	Length    int    `db:"length"`
}

type media struct {
	TorrentID string `db:"torrent_id"`
	FileID    int    `db:"file_id"`
	Name      string `db:"name"`
	Length    int    `db:"length"`
	Source    string `db:"source"`
}
