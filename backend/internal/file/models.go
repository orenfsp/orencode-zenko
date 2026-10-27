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
	ID               int64     `json:"id"`
	Purpose          Purpose   `json:"purpose"`
	StorageKey       string    `json:"storage_key"`
	OriginalFilename string    `json:"original_filename"`
	ContentType      string    `json:"content_type"`
	SizeBytes        int64     `json:"size_bytes"`
	SHA256           string    `json:"sha256"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type DownloadableFile struct {
	UploadedFile
	OwnerUserID int64
}
