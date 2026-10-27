package file

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
)

type SQLCRepository struct {
	queries *db.Queries
}

func NewSQLCRepository(queries *db.Queries) (*SQLCRepository, error) {
	if queries == nil {
		return nil, errors.New("sqlc queries are required")
	}
	return &SQLCRepository{queries: queries}, nil
}

func (r *SQLCRepository) Create(ctx context.Context, input NewUploadedFile) (UploadedFile, error) {
	row, err := r.queries.CreateUploadedFile(ctx, db.CreateUploadedFileParams{
		PurposeCode:      string(input.Purpose),
		StorageKey:       input.StorageKey,
		OriginalFilename: input.OriginalFilename,
		ContentType:      input.ContentType,
		SizeBytes:        input.SizeBytes,
		Sha256:           input.SHA256,
	})
	if err != nil {
		return UploadedFile{}, fmt.Errorf("create uploaded file: %w", err)
	}
	return mapUploadedFile(row), nil
}

func (r *SQLCRepository) GetByID(ctx context.Context, fileID int64) (UploadedFile, error) {
	row, err := r.queries.GetUploadedFileByID(ctx, fileID)
	if errors.Is(err, pgx.ErrNoRows) {
		return UploadedFile{}, ErrFileNotFound
	}
	if err != nil {
		return UploadedFile{}, fmt.Errorf("get uploaded file: %w", err)
	}
	return mapUploadedFile(row), nil
}

func (r *SQLCRepository) GetAttachedForDownload(ctx context.Context, fileID int64) (DownloadableFile, error) {
	row, err := r.queries.GetAttachedFileForDownload(ctx, fileID)
	if errors.Is(err, pgx.ErrNoRows) {
		return DownloadableFile{}, ErrFileNotFound
	}
	if err != nil {
		return DownloadableFile{}, fmt.Errorf("get file for download: %w", err)
	}
	return DownloadableFile{
		UploadedFile: UploadedFile{
			ID:               row.ID,
			Purpose:          Purpose(row.PurposeCode),
			StorageKey:       row.StorageKey,
			OriginalFilename: row.OriginalFilename,
			ContentType:      row.ContentType,
			SizeBytes:        row.SizeBytes,
			SHA256:           row.Sha256,
			Status:           row.Status,
			CreatedAt:        row.CreatedAt.Time,
			UpdatedAt:        row.UpdatedAt.Time,
		},
		OwnerUserID: row.OwnerUserID,
	}, nil
}

func (r *SQLCRepository) MarkDeletePending(ctx context.Context, fileID int64) (bool, error) {
	rows, err := r.queries.MarkUploadedFileDeletePending(ctx, fileID)
	if err != nil {
		return false, fmt.Errorf("mark file delete pending: %w", err)
	}
	return rows > 0, nil
}

func (r *SQLCRepository) MarkDeletePendingByStorageKey(ctx context.Context, storageKey string) error {
	_, err := r.queries.MarkUploadedFileDeletePendingByStorageKey(ctx, storageKey)
	if err != nil {
		return fmt.Errorf("mark file delete pending by storage key: %w", err)
	}
	return nil
}

func (r *SQLCRepository) DeleteDiscardable(ctx context.Context, fileID int64) error {
	_, err := r.queries.DeleteDiscardableUploadedFile(ctx, fileID)
	if err != nil {
		return fmt.Errorf("delete discardable file: %w", err)
	}
	return nil
}

func (r *SQLCRepository) DeleteUploadedByStorageKey(ctx context.Context, storageKey string) error {
	_, err := r.queries.DeleteUploadedFileByStorageKey(ctx, storageKey)
	if err != nil {
		return fmt.Errorf("delete uploaded file by storage key: %w", err)
	}
	return nil
}

func (r *SQLCRepository) ListPendingCleanup(
	ctx context.Context,
	uploadedBefore time.Time,
	batchSize int32,
) ([]UploadedFile, error) {
	rows, err := r.queries.ListFilesPendingCleanup(ctx, db.ListFilesPendingCleanupParams{
		UploadedBefore: pgtype.Timestamptz{Time: uploadedBefore, Valid: true},
		BatchSize:      batchSize,
	})
	if err != nil {
		return nil, fmt.Errorf("list files pending cleanup: %w", err)
	}
	result := make([]UploadedFile, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapUploadedFile(row))
	}
	return result, nil
}

func mapUploadedFile(row db.UploadedFile) UploadedFile {
	return UploadedFile{
		ID:               row.ID,
		Purpose:          Purpose(row.PurposeCode),
		StorageKey:       row.StorageKey,
		OriginalFilename: row.OriginalFilename,
		ContentType:      row.ContentType,
		SizeBytes:        row.SizeBytes,
		SHA256:           row.Sha256,
		Status:           row.Status,
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}
