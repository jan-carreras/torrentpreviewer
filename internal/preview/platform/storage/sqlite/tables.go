package sqlite

const (
	sqlTorrentTable = "torrents"
)

type torrent struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	Length      int    `db:"length"`
	PieceLength int    `db:"pieceLength"`
	Raw         []byte `db:"raw"`
}
