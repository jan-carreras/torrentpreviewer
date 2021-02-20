package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"prevtorrent/internal/preview"

	"github.com/huandu/go-sqlbuilder"
)

type ImageRepository struct {
	db *sql.DB
}

func NewImageRepository(db *sql.DB) *ImageRepository {
	return &ImageRepository{db: db}
}

func (r *ImageRepository) ByTorrent(ctx context.Context, id string) (*preview.TorrentImages, error) {
	sqlStructure := sqlbuilder.NewStruct(new(media))
	query := sqlStructure.SelectFrom(sqlMediaTable)
	query.Where(query.Equal("torrent_id", id))
	query.OrderBy("id").Asc()

	sqlRaw, args := query.Build()
	rows, err := r.db.Query(sqlRaw, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []preview.Image
	for rows.Next() {
		var m media
		if err := rows.Scan(sqlStructure.Addr(&m)...); err != nil {
			return nil, err
		}
		images = append(images, preview.NewImage(m.TorrentID, m.FileID, m.Name, m.Length))
	}
	return preview.NewTorrentImages(images), nil

}

func (r *ImageRepository) Persist(ctx context.Context, img preview.Image) error {
	torrentSQLStruct := sqlbuilder.NewStruct(new(media))
	query, args := torrentSQLStruct.InsertInto(sqlMediaTable, media{
		TorrentID: img.TorrentID(),
		FileID:    img.FileID(),
		Name:      img.Name(),
		Length:    img.Length(),
	}).Build()

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error trying to persist a media on database: %v", err)
	}

	return err
}
