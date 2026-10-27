package file

import "time"

const (
	StatusUploaded      = "uploaded"
	StatusAttached      = "attached"
	StatusDeletePending = "delete_pending"
)

type NewUploadedFile struct {
	Purpose          Purpose
	StorageKey       string
	OriginalFilename string
	ContentType      string
	SizeBytes        int64
	SHA256           string
}

type UploadedFile struct {
	ID               int64
	Purpose          Purpose
	StorageKey       string
	OriginalFilename string
	ContentType      string
	SizeBytes        int64
	SHA256           string
	Status           string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type DownloadableFile struct {
	UploadedFile
	OwnerUserID int64
}
