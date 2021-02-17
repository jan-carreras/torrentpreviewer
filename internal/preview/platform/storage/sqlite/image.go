package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/huandu/go-sqlbuilder"
	"prevtorrent/internal/preview"
)

type ImageRepository struct {
	db *sql.DB
}

func NewImageRepository(db *sql.DB) *ImageRepository {
	return &ImageRepository{db: db}
}

func (r *ImageRepository) ByMedia(ctx context.Context, id string) ([]preview.Image, error) {
	panic("implement me")
}

func (r *ImageRepository) Persist(ctx context.Context, img preview.Image) error {
	torrentSQLStruct := sqlbuilder.NewStruct(new(media))
	query, args := torrentSQLStruct.InsertInto(sqlMediaTable, media{
		TorrentID: img.TorrentID(),
		FileID:    img.FileID(),
		Name:      img.Name(),
		Length:    img.Length(),
		Source:    img.Source(),
	}).Build()

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error trying to persist a media on database: %v", err)
	}

	return err
}
