package file

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, file NewUploadedFile) (UploadedFile, error)
	GetByID(ctx context.Context, fileID int64) (UploadedFile, error)
	GetAttachedForDownload(ctx context.Context, fileID int64) (DownloadableFile, error)
	MarkDeletePending(ctx context.Context, fileID int64) (bool, error)
	MarkDeletePendingByStorageKey(ctx context.Context, storageKey string) error
	DeleteDiscardable(ctx context.Context, fileID int64) error
	DeleteUploadedByStorageKey(ctx context.Context, storageKey string) error
	ListPendingCleanup(ctx context.Context, uploadedBefore time.Time, batchSize int32) ([]UploadedFile, error)
	ListAll(ctx context.Context, limit, offset int32) ([]UploadedFile, error)
}
