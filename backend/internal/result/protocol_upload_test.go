package result

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
)

type protocolTestStorage struct {
	files  map[string][]byte
	putErr error
}

func (s *protocolTestStorage) Put(_ context.Context, key string, src io.Reader) (int64, error) {
	if s.putErr != nil {
		return 0, s.putErr
	}
	data, err := io.ReadAll(src)
	if err == nil {
		s.files[key] = data
	}
	return int64(len(data)), err
}
func (s *protocolTestStorage) Open(_ context.Context, key string) (io.ReadCloser, error) {
	data, ok := s.files[key]
	if !ok {
		return nil, errors.New("missing")
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}
func (s *protocolTestStorage) Delete(_ context.Context, key string) error {
	delete(s.files, key)
	return nil
}

func TestProtocolReplacementAndDelete(t *testing.T) {
	old := "protocols/old.pdf"
	repo := &repositoryStub{contextData: algorithmicContext(), table: Table{Sheet: Sheet{Status: SheetApproved, ProtocolStorageKey: &old}}}
	storage := &protocolTestStorage{files: map[string][]byte{old: []byte("old")}}
	service := NewService(repo, storage, &protocolStub{})
	url, err := service.UploadProtocol(context.Background(), 1, 2, []byte("new"))
	if err != nil || url != protocolURL(1, 2) {
		t.Fatalf("upload: %q %v", url, err)
	}
	if _, exists := storage.files[old]; exists {
		t.Fatal("old file was not removed")
	}
	key := *repo.table.Sheet.ProtocolStorageKey
	if string(storage.files[key]) != "new" {
		t.Fatal("new file not saved")
	}
	if err := service.DeleteProtocol(context.Background(), 1, 2); err != nil {
		t.Fatal(err)
	}
	if repo.table.Sheet.ProtocolStorageKey != nil || len(storage.files) != 0 {
		t.Fatal("protocol not removed")
	}
	if err := service.DeleteProtocol(context.Background(), 1, 2); !errors.Is(err, ErrProtocolNotFound) {
		t.Fatalf("second delete: %v", err)
	}
}

func TestProtocolUploadFailurePreservesOldFile(t *testing.T) {
	old := "protocols/old.pdf"
	repo := &repositoryStub{contextData: algorithmicContext(), table: Table{Sheet: Sheet{Status: SheetApproved, ProtocolStorageKey: &old}}}
	storage := &protocolTestStorage{files: map[string][]byte{old: []byte("old")}}
	storage.putErr = errors.New("storage unavailable")
	service := NewService(repo, storage, &protocolStub{})
	if _, err := service.UploadProtocol(context.Background(), 1, 2, []byte("new")); err == nil {
		t.Fatal("expected storage error")
	}
	if string(storage.files[old]) != "old" || *repo.table.Sheet.ProtocolStorageKey != old {
		t.Fatal("old protocol changed")
	}
	storage.putErr = nil
	repo.table.Sheet.Status = SheetDraft
	if _, err := service.UploadProtocol(context.Background(), 1, 2, []byte("new")); !errors.Is(err, ErrProtocolNotApproved) {
		t.Fatalf("unapproved: %v", err)
	}
	if len(storage.files) != 1 || string(storage.files[old]) != "old" {
		t.Fatal("failed upload left or removed file")
	}
}
