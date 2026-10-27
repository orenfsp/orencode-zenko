package auth

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	filedomain "github.com/orenfsp/orencode-zenko.git/internal/file"
)

type registrationFileRepositoryStub struct {
	nextID int64
}

func (r *registrationFileRepositoryStub) Create(_ context.Context, input filedomain.NewUploadedFile) (filedomain.UploadedFile, error) {
	r.nextID++
	return filedomain.UploadedFile{
		ID:               r.nextID,
		Purpose:          input.Purpose,
		StorageKey:       input.StorageKey,
		OriginalFilename: input.OriginalFilename,
		ContentType:      input.ContentType,
		SizeBytes:        input.SizeBytes,
		SHA256:           input.SHA256,
		Status:           filedomain.StatusUploaded,
	}, nil
}

func (r *registrationFileRepositoryStub) GetByID(context.Context, int64) (filedomain.UploadedFile, error) {
	return filedomain.UploadedFile{}, filedomain.ErrFileNotFound
}

func (r *registrationFileRepositoryStub) GetAttachedForDownload(context.Context, int64) (filedomain.DownloadableFile, error) {
	return filedomain.DownloadableFile{}, filedomain.ErrFileNotFound
}

func (r *registrationFileRepositoryStub) MarkDeletePending(context.Context, int64) (bool, error) {
	return false, nil
}

func (r *registrationFileRepositoryStub) MarkDeletePendingByStorageKey(context.Context, string) error {
	return nil
}

func (r *registrationFileRepositoryStub) DeleteDiscardable(context.Context, int64) error {
	return nil
}

func (r *registrationFileRepositoryStub) DeleteUploadedByStorageKey(context.Context, string) error {
	return nil
}

func (r *registrationFileRepositoryStub) ListPendingCleanup(context.Context, time.Time, int32) ([]filedomain.UploadedFile, error) {
	return nil, nil
}

func TestRegisterAcceptsAdultMultipartWithPersonalConsent(t *testing.T) {
	users := &userProviderStub{user: User{ID: 42, Email: "user@example.com", AccessStatus: true}}
	authService := newTestService(users, &sessionProviderStub{})
	storage, err := filedomain.NewLocalStorage(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocalStorage() error = %v", err)
	}
	fileService, err := filedomain.NewService(
		&registrationFileRepositoryStub{},
		storage,
		filedomain.DefaultPolicies(),
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	body, contentType := registrationMultipartBody(t, "2000-05-10", true, false)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", body)
	request.Header.Set("Content-Type", contentType)
	response := httptest.NewRecorder()

	NewHandler(authService, fileService).Register(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", response.Code, response.Body.String())
	}
	if users.createData.PersonalDataConsentFileID != 1 {
		t.Fatalf("personal consent file id = %d, want 1", users.createData.PersonalDataConsentFileID)
	}
	if users.createData.ParentalConsentFileID != nil {
		t.Fatalf("parental consent file id = %v, want nil", users.createData.ParentalConsentFileID)
	}
}

func TestRegisterMinorRequiresParentalFileBeforeUpload(t *testing.T) {
	users := &userProviderStub{user: User{ID: 42}}
	authService := newTestService(users, &sessionProviderStub{})
	body, contentType := registrationMultipartBody(t, "2010-01-01", true, false)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", body)
	request.Header.Set("Content-Type", contentType)
	response := httptest.NewRecorder()

	NewHandler(authService, nil).Register(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("parental_consent_document_file")) {
		t.Fatalf("response does not name missing parental file: %s", response.Body.String())
	}
}

func registrationMultipartBody(
	t *testing.T,
	birthDate string,
	withPersonalFile bool,
	withParentalFile bool,
) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fields := map[string]string{
		"first_name": "Иван",
		"last_name":  "Иванов",
		"birth_date": birthDate,
		"email":      "user@example.com",
		"password":   "password123",
	}
	for name, value := range fields {
		if err := writer.WriteField(name, value); err != nil {
			t.Fatalf("WriteField(%s) error = %v", name, err)
		}
	}
	writePNG := func(field, filename string) {
		part, err := writer.CreateFormFile(field, filename)
		if err != nil {
			t.Fatalf("CreateFormFile(%s) error = %v", field, err)
		}
		if _, err := io.Copy(part, bytes.NewReader([]byte("\x89PNG\r\n\x1a\ncontent"))); err != nil {
			t.Fatalf("copy %s error = %v", field, err)
		}
	}
	if withPersonalFile {
		writePNG("personal_data_consent_file", "personal.png")
	}
	if withParentalFile {
		writePNG("parental_consent_document_file", "parental.png")
	}
	if err := writer.Close(); err != nil && !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("multipart close error = %v", err)
	}
	return body, writer.FormDataContentType()
}
