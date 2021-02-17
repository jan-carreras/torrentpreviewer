package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/huandu/go-sqlbuilder"
	"prevtorrent/internal/preview"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type TorrentRepository struct {
	db *sql.DB
}

func NewTorrentRepository(db *sql.DB) *TorrentRepository {
	return &TorrentRepository{db: db}
}

func (r *TorrentRepository) Persist(ctx context.Context, t preview.Info) error {
	torrentSQLStruct := sqlbuilder.NewStruct(new(torrent))
	query, args := torrentSQLStruct.InsertInto(sqlTorrentTable, torrent{
		ID:          t.ID(),
		Name:        t.Name(),
		Length:      t.TotalLength(),
		PieceLength: t.PieceLength(),
		Raw:         t.Raw(),
	}).Build()

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error trying to persist torrent on database: %v", err)
	}

	err = r.storeFiles(ctx, t)
	return err
}

func (r *TorrentRepository) storeFiles(ctx context.Context, t preview.Info) error {
	fileSQLStructure := sqlbuilder.NewStruct(new(file))
	for _, f := range t.Files() {
		newF := file{
			TorrentID: t.ID(),
			ID:        f.ID(),
			Name:      f.Name(),
			Length:    f.Length(),
		}
		query, args := fileSQLStructure.InsertInto(sqlFileTable, newF).Build()
		_, err := r.db.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *TorrentRepository) Get(ctx context.Context, id string) (preview.Info, error) {
	id = strings.ToLower(id)

	torrentSQLStruct := sqlbuilder.NewStruct(new(torrent))
	query := torrentSQLStruct.SelectFrom(sqlTorrentTable)
	query.Where(query.Equal("id", id))

	sqlRaw, args := query.Build()
	rows, err := r.db.Query(sqlRaw, args...)
	if err != nil {
		return preview.Info{}, err
	}
	defer rows.Close()

	if !rows.Next() {
		return preview.Info{}, preview.ErrNotFound
	}

	var t torrent
	if err := rows.Scan(torrentSQLStruct.Addr(&t)...); err != nil {
		return preview.Info{}, err
	}

	files, err := r.readFiles(ctx, id)
	if err != nil {
		return preview.Info{}, err
	}

	return preview.NewInfo(id, t.Name, t.PieceLength, files, t.Raw)
}

func (r *TorrentRepository) readFiles(ctx context.Context, id string) ([]preview.FileInfo, error) {
	fileSQLStructure := sqlbuilder.NewStruct(new(file))
	query := fileSQLStructure.SelectFrom(sqlFileTable)
	query.Where(query.Equal("torrent_id", id))
	query.OrderBy("id").Asc()

	sqlRaw, args := query.Build()
	rows, err := r.db.Query(sqlRaw, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []preview.FileInfo
	for rows.Next() {
		var f file
		if err := rows.Scan(fileSQLStructure.Addr(&f)...); err != nil {
			return nil, err
		}
		fi, err := preview.NewFileInfo(f.ID, f.Length, f.Name)
		if err != nil {
			return nil, err
		}
		files = append(files, fi)
	}
	return files, nil
}
