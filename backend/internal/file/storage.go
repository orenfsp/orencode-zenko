package file

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var ErrInvalidStorageKey = errors.New("invalid storage key")

type Storage interface {
	Put(ctx context.Context, key string, src io.Reader) (int64, error)
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}

type LocalStorage struct {
	root string
}

func NewLocalStorage(root string) (*LocalStorage, error) {
	if strings.TrimSpace(root) == "" {
		return nil, errors.New("storage root is required")
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(absRoot, 0o700); err != nil {
		return nil, err
	}
	if err := os.Chmod(absRoot, 0o700); err != nil {
		return nil, err
	}
	return &LocalStorage{root: absRoot}, nil
}

func (s *LocalStorage) Put(ctx context.Context, key string, src io.Reader) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	target, err := s.resolve(key)
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return 0, err
	}
	temporary, err := os.CreateTemp(filepath.Dir(target), ".upload-*")
	if err != nil {
		return 0, err
	}
	temporaryName := temporary.Name()
	committed := false
	defer func() {
		_ = temporary.Close()
		if !committed {
			_ = os.Remove(temporaryName)
		}
	}()

	size, err := io.Copy(temporary, &contextReader{ctx: ctx, reader: src})
	if err != nil {
		return size, err
	}
	if err := temporary.Sync(); err != nil {
		return size, err
	}
	if err := temporary.Close(); err != nil {
		return size, err
	}
	if err := ctx.Err(); err != nil {
		return size, err
	}
	if err := os.Rename(temporaryName, target); err != nil {
		return size, err
	}
	committed = true
	return size, nil
}

func (s *LocalStorage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	target, err := s.resolve(key)
	if err != nil {
		return nil, err
	}
	opened, err := os.Open(target)
	if os.IsNotExist(err) {
		return nil, ErrFileNotFound
	}
	return opened, err
}

func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	target, err := s.resolve(key)
	if err != nil {
		return err
	}
	err = os.Remove(target)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (s *LocalStorage) resolve(key string) (string, error) {
	if key == "" || strings.Contains(key, `\`) || strings.HasPrefix(key, "/") || filepath.IsAbs(key) {
		return "", ErrInvalidStorageKey
	}
	cleaned := filepath.Clean(filepath.FromSlash(key))
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(os.PathSeparator)) {
		return "", ErrInvalidStorageKey
	}
	target := filepath.Join(s.root, cleaned)
	relative, err := filepath.Rel(s.root, target)
	if err != nil {
		return "", err
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return "", ErrInvalidStorageKey
	}
	return target, nil
}
