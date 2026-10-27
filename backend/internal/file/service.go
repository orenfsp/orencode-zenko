package file

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

var (
	ErrUnknownPurpose  = errors.New("unknown file purpose")
	ErrInvalidInput    = errors.New("invalid file input")
	ErrEmptyFile       = errors.New("file is empty")
	ErrFileTooLarge    = errors.New("file exceeds maximum size")
	ErrUnsupportedType = errors.New("unsupported file type")
	ErrFileNotFound    = errors.New("file not found")
	ErrFileForbidden   = errors.New("file access denied")
	ErrFileAttached    = errors.New("attached file cannot be deleted")
)

var validPurpose = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)

type Service struct {
	repo     Repository
	storage  Storage
	policies map[Purpose]Policy
}

func NewService(repo Repository, storage Storage, policies map[Purpose]Policy) (*Service, error) {
	if repo == nil || storage == nil {
		return nil, errors.New("file repository and storage are required")
	}
	if len(policies) == 0 {
		return nil, errors.New("at least one file policy is required")
	}
	copiedPolicies := make(map[Purpose]Policy, len(policies))
	for purpose, policy := range policies {
		if !validPurpose.MatchString(string(purpose)) {
			return nil, fmt.Errorf("invalid file purpose %q", purpose)
		}
		if policy.MaxBytes <= 0 || len(policy.AllowedMIME) == 0 {
			return nil, fmt.Errorf("invalid policy for purpose %q", purpose)
		}
		copiedMIME := make(map[string]string, len(policy.AllowedMIME))
		for contentType, extension := range policy.AllowedMIME {
			if contentType == "" || !strings.HasPrefix(extension, ".") || strings.ContainsAny(extension, `/\`) {
				return nil, fmt.Errorf("invalid MIME configuration for purpose %q", purpose)
			}
			copiedMIME[contentType] = extension
		}
		copiedPolicies[purpose] = Policy{MaxBytes: policy.MaxBytes, AllowedMIME: copiedMIME}
	}
	return &Service{repo: repo, storage: storage, policies: copiedPolicies}, nil
}

func (s *Service) Upload(ctx context.Context, input UploadInput) (UploadedFile, error) {
	if err := ctx.Err(); err != nil {
		return UploadedFile{}, err
	}
	if input.Content == nil || input.Size < -1 {
		return UploadedFile{}, ErrInvalidInput
	}
	policy, ok := s.policies[input.Purpose]
	if !ok {
		return UploadedFile{}, ErrUnknownPurpose
	}
	if input.Size > policy.MaxBytes {
		return UploadedFile{}, ErrFileTooLarge
	}

	content, err := io.ReadAll(io.LimitReader(
		&contextReader{ctx: ctx, reader: input.Content},
		policy.MaxBytes+1,
	))
	if err != nil {
		return UploadedFile{}, fmt.Errorf("read file: %w", err)
	}
	if len(content) == 0 {
		return UploadedFile{}, ErrEmptyFile
	}
	if int64(len(content)) > policy.MaxBytes {
		return UploadedFile{}, ErrFileTooLarge
	}

	contentType := detectContentType(content)
	extension, ok := policy.AllowedMIME[contentType]
	if !ok {
		return UploadedFile{}, fmt.Errorf("%w: %s", ErrUnsupportedType, contentType)
	}
	storageKey, err := generateStorageKey(input.Purpose, extension)
	if err != nil {
		return UploadedFile{}, err
	}

	size, err := s.storage.Put(ctx, storageKey, bytes.NewReader(content))
	if err != nil {
		return UploadedFile{}, errors.Join(
			fmt.Errorf("store file: %w", err),
			s.cleanupStorage(ctx, storageKey),
		)
	}
	if size != int64(len(content)) {
		return UploadedFile{}, errors.Join(
			fmt.Errorf("stored file size mismatch: got %d, want %d", size, len(content)),
			s.cleanupStorage(ctx, storageKey),
		)
	}

	digest := sha256.Sum256(content)
	record, err := s.repo.Create(ctx, NewUploadedFile{
		Purpose:          input.Purpose,
		StorageKey:       storageKey,
		OriginalFilename: sanitizeFilename(input.Filename),
		ContentType:      contentType,
		SizeBytes:        size,
		SHA256:           hex.EncodeToString(digest[:]),
	})
	if err != nil {
		return UploadedFile{}, errors.Join(
			fmt.Errorf("save file metadata: %w", err),
			s.cleanupAfterMetadataError(ctx, storageKey),
		)
	}
	return record, nil
}

func (s *Service) Download(ctx context.Context, fileID, requesterUserID int64, canReadAny bool) (Download, error) {
	if fileID <= 0 || requesterUserID <= 0 {
		return Download{}, ErrInvalidInput
	}
	record, err := s.repo.GetAttachedForDownload(ctx, fileID)
	if err != nil {
		return Download{}, err
	}
	if record.OwnerUserID != requesterUserID && !canReadAny {
		return Download{}, ErrFileForbidden
	}
	content, err := s.storage.Open(ctx, record.StorageKey)
	if err != nil {
		return Download{}, fmt.Errorf("open stored file: %w", err)
	}
	return Download{File: record.UploadedFile, Content: content}, nil
}

func (s *Service) Discard(ctx context.Context, fileID int64) error {
	if fileID <= 0 {
		return ErrInvalidInput
	}
	claimed, err := s.repo.MarkDeletePending(ctx, fileID)
	if err != nil {
		return err
	}
	if !claimed {
		record, getErr := s.repo.GetByID(ctx, fileID)
		if errors.Is(getErr, ErrFileNotFound) {
			return nil
		}
		if getErr != nil {
			return getErr
		}
		if record.Status == StatusAttached {
			return ErrFileAttached
		}
		return fmt.Errorf("file %d cannot be discarded in status %q", fileID, record.Status)
	}

	record, err := s.repo.GetByID(ctx, fileID)
	if err != nil {
		return err
	}
	if err := s.storage.Delete(ctx, record.StorageKey); err != nil {
		return fmt.Errorf("delete stored file: %w", err)
	}
	return s.repo.DeleteDiscardable(ctx, fileID)
}

func (s *Service) CleanupPending(
	ctx context.Context,
	uploadedBefore time.Time,
	batchSize int32,
) (CleanupResult, error) {
	result := CleanupResult{StartedAt: time.Now().UTC()}
	if batchSize <= 0 || batchSize > 1000 {
		return result, ErrInvalidInput
	}
	files, err := s.repo.ListPendingCleanup(ctx, uploadedBefore, batchSize)
	if err != nil {
		return result, err
	}
	var cleanupErrors []error
	for _, uploaded := range files {
		result.Processed++
		if err := s.Discard(ctx, uploaded.ID); err != nil {
			result.Failed++
			cleanupErrors = append(cleanupErrors, fmt.Errorf("discard file %d: %w", uploaded.ID, err))
		}
	}
	return result, errors.Join(cleanupErrors...)
}

func (s *Service) ListAll(ctx context.Context, limit, offset int32) ([]UploadedFile, error) {
	if limit < 1 || limit > 100 || offset < 0 {
		return nil, ErrInvalidInput
	}
	return s.repo.ListAll(ctx, limit, offset)
}

func (s *Service) cleanupStorage(ctx context.Context, storageKey string) error {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := s.storage.Delete(cleanupCtx, storageKey); err != nil {
		return fmt.Errorf("cleanup stored file: %w", err)
	}
	return nil
}

func (s *Service) cleanupAfterMetadataError(ctx context.Context, storageKey string) error {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := s.storage.Delete(cleanupCtx, storageKey); err != nil {
		markErr := s.repo.MarkDeletePendingByStorageKey(cleanupCtx, storageKey)
		return errors.Join(fmt.Errorf("cleanup stored file: %w", err), markErr)
	}
	return s.repo.DeleteUploadedByStorageKey(cleanupCtx, storageKey)
}

func detectContentType(content []byte) string {
	if isDOCX(content) {
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}
	if isXML(content) {
		return "application/xml"
	}
	header := content
	if len(header) > 512 {
		header = header[:512]
	}
	detected := http.DetectContentType(header)
	mediaType, _, err := mime.ParseMediaType(detected)
	if err == nil {
		return mediaType
	}
	return detected
}

func isDOCX(content []byte) bool {
	archive, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return false
	}
	var hasContentTypes, hasDocument bool
	for _, entry := range archive.File {
		switch strings.TrimPrefix(entry.Name, "/") {
		case "[Content_Types].xml":
			hasContentTypes = true
		case "word/document.xml":
			hasDocument = true
		}
	}
	return hasContentTypes && hasDocument
}

func isXML(content []byte) bool {
	decoder := xml.NewDecoder(bytes.NewReader(content))
	foundRoot := false
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			return foundRoot
		}
		if err != nil {
			return false
		}
		if _, ok := token.(xml.StartElement); ok {
			foundRoot = true
		}
	}
}

func generateStorageKey(purpose Purpose, extension string) (string, error) {
	var randomKey [32]byte
	if _, err := rand.Read(randomKey[:]); err != nil {
		return "", fmt.Errorf("generate storage key: %w", err)
	}
	return path.Join(string(purpose), hex.EncodeToString(randomKey[:])+extension), nil
}

func sanitizeFilename(filename string) string {
	filename = strings.ReplaceAll(filename, `\`, "/")
	filename = strings.TrimSpace(path.Base(filename))
	var result strings.Builder
	for _, character := range filename {
		if unicode.IsControl(character) {
			continue
		}
		runeSize := utf8.RuneLen(character)
		if runeSize < 0 || result.Len()+runeSize > 255 {
			break
		}
		result.WriteRune(character)
	}
	filename = strings.TrimSpace(result.String())
	if filename == "" || filename == "." || filename == ".." {
		return "file"
	}
	return filename
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *contextReader) Read(buffer []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(buffer)
}
