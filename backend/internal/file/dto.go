package file

import (
	"io"
	"time"
)

type UploadInput struct {
	Purpose  Purpose
	Filename string
	Size     int64
	Content  io.Reader
}

type Download struct {
	File    UploadedFile
	Content io.ReadCloser
}

type CleanupResult struct {
	Processed int
	Failed    int
	StartedAt time.Time
}
