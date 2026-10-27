package file

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

func TestLocalStorageRejectsPathTraversal(t *testing.T) {
	storage, err := NewLocalStorage(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	_, err = storage.Put(context.Background(), "../escape.txt", bytes.NewReader([]byte("data")))
	if !errors.Is(err, ErrInvalidStorageKey) {
		t.Fatalf("Put() error = %v, want ErrInvalidStorageKey", err)
	}
}
