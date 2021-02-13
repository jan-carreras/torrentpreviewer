package sqlite

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/huandu/go-sqlbuilder"
	"prevtorrent/internal/preview"
	"strings"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

type TorrentRepository struct {
	db *sql.DB
}

func NewTorrentRepository(db *sql.DB) *TorrentRepository {
	return &TorrentRepository{db: db}
}

func (r *TorrentRepository) Persist(ctx context.Context, data []byte) error {
	// TODO:  We should pass a preview.Info to store, really. Not this raw []byte data crap

	// TODO: This block can be deleted
	buf := bytes.NewBuffer(data)
	d := bencode.NewDecoder(buf)

	metaInfo := new(metainfo.MetaInfo)
	if err := d.Decode(metaInfo); err != nil {
		return err
	}

	info, err := metaInfo.UnmarshalInfo()
	if err != nil {
		return err
	}
	// TODO: Up until here...

	torrentSQLStruct := sqlbuilder.NewStruct(new(torrent))
	query, args := torrentSQLStruct.InsertInto(sqlTorrentTable, torrent{
		ID:          strings.ToLower(metaInfo.HashInfoBytes().HexString()),
		Name:        info.Name,
		Length:      int(info.TotalLength()),
		PieceLength: int(info.PieceLength),
		Raw:         data,
	}).Build()

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error trying to persist torrent on database: %v", err)
	}

	return nil
}

func (r *TorrentRepository) Get(ctx context.Context, id string) (preview.Info, error) {
	id = strings.ToLower(id)

	torrentSQLStruct := sqlbuilder.NewStruct(new(torrent))
	query := torrentSQLStruct.SelectFrom(sqlTorrentTable)
	query.Where(query.Equal("id", id))

	sqlRaw, args := query.Build()
	rows, _ := r.db.Query(sqlRaw, args...)
	defer rows.Close()

	if !rows.Next() {
		return preview.Info{}, preview.ErrNotFound
	}

	var t torrent
	if err := rows.Scan(torrentSQLStruct.Addr(&t)...); err != nil {
		return preview.Info{}, err
	}

	// TODO: Store Files as well
	return preview.NewInfo(id, t.Name, t.PieceLength, t.PieceLength, nil, t.Raw)
}
