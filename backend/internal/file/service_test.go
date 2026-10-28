package file

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

type repositoryStub struct {
	created            NewUploadedFile
	file               UploadedFile
	downloadable       DownloadableFile
	markDeletePending  bool
	deleteDiscardable  bool
	markedByStorageKey bool
	all                []UploadedFile
	listLimit          int32
	listOffset         int32
}

func (r *repositoryStub) Create(_ context.Context, input NewUploadedFile) (UploadedFile, error) {
	r.created = input
	r.file = UploadedFile{
		ID:               1,
		Purpose:          input.Purpose,
		StorageKey:       input.StorageKey,
		OriginalFilename: input.OriginalFilename,
		ContentType:      input.ContentType,
		SizeBytes:        input.SizeBytes,
		SHA256:           input.SHA256,
		Status:           StatusUploaded,
	}
	return r.file, nil
}

func (r *repositoryStub) GetByID(context.Context, int64) (UploadedFile, error) {
	if r.file.ID == 0 {
		return UploadedFile{}, ErrFileNotFound
	}
	return r.file, nil
}

func (r *repositoryStub) GetAttachedForDownload(context.Context, int64) (DownloadableFile, error) {
	if r.downloadable.ID == 0 {
		return DownloadableFile{}, ErrFileNotFound
	}
	return r.downloadable, nil
}

func (r *repositoryStub) MarkDeletePending(context.Context, int64) (bool, error) {
	if r.markDeletePending {
		r.file.Status = StatusDeletePending
	}
	return r.markDeletePending, nil
}

func (r *repositoryStub) MarkDeletePendingByStorageKey(context.Context, string) error {
	r.markedByStorageKey = true
	return nil
}

func (r *repositoryStub) DeleteDiscardable(context.Context, int64) error {
	r.deleteDiscardable = true
	return nil
}

func (*repositoryStub) DeleteUploadedByStorageKey(context.Context, string) error { return nil }

func (*repositoryStub) ListPendingCleanup(context.Context, time.Time, int32) ([]UploadedFile, error) {
	return nil, nil
}

func (r *repositoryStub) ListAll(_ context.Context, limit, offset int32) ([]UploadedFile, error) {
	r.listLimit = limit
	r.listOffset = offset
	return r.all, nil
}

type storageStub struct {
	values    map[string][]byte
	deleteErr error
	opened    bool
}

func (s *storageStub) Put(_ context.Context, key string, source io.Reader) (int64, error) {
	value, err := io.ReadAll(source)
	if err != nil {
		return 0, err
	}
	if s.values == nil {
		s.values = make(map[string][]byte)
	}
	s.values[key] = value
	return int64(len(value)), nil
}

func (s *storageStub) Open(_ context.Context, key string) (io.ReadCloser, error) {
	value, ok := s.values[key]
	if !ok {
		return nil, ErrFileNotFound
	}
	s.opened = true
	return io.NopCloser(bytes.NewReader(value)), nil
}

func (s *storageStub) Delete(_ context.Context, key string) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	delete(s.values, key)
	return nil
}

func newTestService(t *testing.T, repository Repository, storage Storage) *Service {
	t.Helper()
	service, err := NewService(repository, storage, DefaultPolicies())
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func TestUploadXML(t *testing.T) {
	repository := &repositoryStub{}
	storage := &storageStub{}
	service := newTestService(t, repository, storage)
	content := []byte(`<?xml version="1.0"?><consent><accepted>true</accepted></consent>`)

	uploaded, err := service.Upload(context.Background(), UploadInput{
		Purpose:  PurposePersonalDataConsent,
		Filename: `../consent.xml`,
		Size:     int64(len(content)),
		Content:  bytes.NewReader(content),
	})
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if uploaded.ContentType != "application/xml" || !strings.HasSuffix(uploaded.StorageKey, ".xml") {
		t.Fatalf("uploaded = %#v", uploaded)
	}
	if uploaded.OriginalFilename != "consent.xml" {
		t.Fatalf("filename = %q", uploaded.OriginalFilename)
	}
	if len(uploaded.SHA256) != 64 {
		t.Fatalf("sha256 = %q", uploaded.SHA256)
	}
}

func TestUploadDOCX(t *testing.T) {
	repository := &repositoryStub{}
	service := newTestService(t, repository, &storageStub{})
	content := makeDOCX(t)

	uploaded, err := service.Upload(context.Background(), UploadInput{
		Purpose:  PurposeParentalConsent,
		Filename: "consent.docx",
		Size:     int64(len(content)),
		Content:  bytes.NewReader(content),
	})
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if uploaded.ContentType != "application/vnd.openxmlformats-officedocument.wordprocessingml.document" ||
		!strings.HasSuffix(uploaded.StorageKey, ".docx") {
		t.Fatalf("uploaded = %#v", uploaded)
	}
}

func TestUploadRejectsArbitraryZIP(t *testing.T) {
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	entry, err := archive.Create("data.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = entry.Write([]byte("not a docx"))
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	service := newTestService(t, &repositoryStub{}, &storageStub{})

	_, err = service.Upload(context.Background(), UploadInput{
		Purpose:  PurposePersonalDataConsent,
		Filename: "fake.docx",
		Size:     int64(buffer.Len()),
		Content:  bytes.NewReader(buffer.Bytes()),
	})
	if !errors.Is(err, ErrUnsupportedType) {
		t.Fatalf("Upload() error = %v, want ErrUnsupportedType", err)
	}
}

func TestUploadRejectsActualOversize(t *testing.T) {
	policies := DefaultPolicies()
	policy := policies[PurposePersonalDataConsent]
	policy.MaxBytes = 8
	policies[PurposePersonalDataConsent] = policy
	service, err := NewService(&repositoryStub{}, &storageStub{}, policies)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Upload(context.Background(), UploadInput{
		Purpose:  PurposePersonalDataConsent,
		Filename: "file.xml",
		Size:     -1,
		Content:  strings.NewReader("<document>too large</document>"),
	})
	if !errors.Is(err, ErrFileTooLarge) {
		t.Fatalf("Upload() error = %v, want ErrFileTooLarge", err)
	}
}

func TestDiscardRefusesAttachedFile(t *testing.T) {
	repository := &repositoryStub{
		file: UploadedFile{ID: 1, Status: StatusAttached, StorageKey: "attached/file.pdf"},
	}
	service := newTestService(t, repository, &storageStub{})

	err := service.Discard(context.Background(), 1)
	if !errors.Is(err, ErrFileAttached) {
		t.Fatalf("Discard() error = %v, want ErrFileAttached", err)
	}
}

func TestDownloadAllowsOwnerAndReturnsStoredContent(t *testing.T) {
	stored := []byte("stored content")
	storage := &storageStub{values: map[string][]byte{"parental/file.xml": stored}}
	repository := &repositoryStub{downloadable: DownloadableFile{
		UploadedFile: UploadedFile{
			ID: 5, StorageKey: "parental/file.xml", Status: StatusAttached,
			OriginalFilename: "consent.xml", ContentType: "application/xml", SizeBytes: int64(len(stored)),
		},
		OwnerUserID: 42,
	}}
	service := newTestService(t, repository, storage)

	download, err := service.Download(context.Background(), 5, 42, false)
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}
	defer download.Content.Close()
	content, _ := io.ReadAll(download.Content)
	if !bytes.Equal(content, stored) {
		t.Fatalf("content = %q", content)
	}
}

func TestDownloadRejectsOtherUserWithoutPermission(t *testing.T) {
	storage := &storageStub{values: map[string][]byte{"file.xml": []byte("content")}}
	repository := &repositoryStub{downloadable: DownloadableFile{
		UploadedFile: UploadedFile{ID: 5, StorageKey: "file.xml", Status: StatusAttached},
		OwnerUserID:  42,
	}}
	service := newTestService(t, repository, storage)

	_, err := service.Download(context.Background(), 5, 7, false)
	if !errors.Is(err, ErrFileForbidden) {
		t.Fatalf("Download() error = %v, want ErrFileForbidden", err)
	}
	if storage.opened {
		t.Fatal("storage must not be opened for unauthorized user")
	}
}

func TestDownloadAllowsReaderToOpenGeneratedFile(t *testing.T) {
	stored := []byte("generated report")
	storage := &storageStub{values: map[string][]byte{"reports/users.xlsx": stored}}
	repository := &repositoryStub{file: UploadedFile{
		ID: 6, StorageKey: "reports/users.xlsx", Status: StatusUploaded,
		OriginalFilename: "federation-users.xlsx", ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	}}
	service := newTestService(t, repository, storage)

	download, err := service.Download(context.Background(), 6, 1, true)
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}
	defer download.Content.Close()
	content, _ := io.ReadAll(download.Content)
	if !bytes.Equal(content, stored) {
		t.Fatalf("content = %q", content)
	}
}

func TestDownloadRejectsDeletePendingFileForReader(t *testing.T) {
	repository := &repositoryStub{file: UploadedFile{
		ID: 6, StorageKey: "reports/users.xlsx", Status: StatusDeletePending,
	}}
	storage := &storageStub{values: map[string][]byte{"reports/users.xlsx": []byte("generated report")}}
	service := newTestService(t, repository, storage)

	_, err := service.Download(context.Background(), 6, 1, true)
	if !errors.Is(err, ErrFileNotFound) {
		t.Fatalf("Download() error = %v, want ErrFileNotFound", err)
	}
	if storage.opened {
		t.Fatal("storage must not be opened for delete-pending file")
	}
}

func TestListAllPassesPagination(t *testing.T) {
	repo := &repositoryStub{all: []UploadedFile{{ID: 1, Status: StatusAttached}}}
	service, err := NewService(repo, &storageStub{}, DefaultPolicies())
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	files, err := service.ListAll(context.Background(), 20, 40)
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}
	if len(files) != 1 || repo.listLimit != 20 || repo.listOffset != 40 {
		t.Fatalf("files = %#v, pagination = %d/%d", files, repo.listLimit, repo.listOffset)
	}
}

func TestListAllRejectsInvalidPagination(t *testing.T) {
	service, err := NewService(&repositoryStub{}, &storageStub{}, DefaultPolicies())
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	_, err = service.ListAll(context.Background(), 0, 0)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("ListAll() error = %v, want ErrInvalidInput", err)
	}
}

func makeDOCX(t *testing.T) []byte {
	t.Helper()
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	for name, value := range map[string]string{
		"[Content_Types].xml": `<?xml version="1.0"?><Types/>`,
		"word/document.xml":   `<?xml version="1.0"?><document/>`,
	} {
		entry, err := archive.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(value)); err != nil {
			t.Fatal(err)
		}
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}
