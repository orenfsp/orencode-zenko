package file

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"

	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
)

type PermissionChecker interface {
	HasPermission(ctx context.Context, userID int64, permissionCode string) (bool, error)
}

type Handler struct {
	service     *Service
	permissions PermissionChecker
}

func NewHandler(service *Service, permissions PermissionChecker) (*Handler, error) {
	if service == nil || permissions == nil {
		return nil, errors.New("file service and permission checker are required")
	}
	return &Handler{service: service, permissions: permissions}, nil
}

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	fileID, err := strconv.ParseInt(r.PathValue("fileID"), 10, 64)
	if err != nil || fileID <= 0 {
		writeError(w, http.StatusBadRequest, "некорректный fileID")
		return
	}
	canReadAny, err := h.permissions.HasPermission(r.Context(), identity.UserID, rbac.PermissionFilesRead)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}
	download, err := h.service.Download(r.Context(), fileID, identity.UserID, canReadAny)
	if err != nil {
		switch {
		case errors.Is(err, ErrFileNotFound):
			writeError(w, http.StatusNotFound, "файл не найден")
		case errors.Is(err, ErrFileForbidden):
			writeError(w, http.StatusForbidden, "недостаточно прав")
		case errors.Is(err, ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "некорректные данные")
		default:
			writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		}
		return
	}
	defer func() { _ = download.Content.Close() }()

	disposition := mime.FormatMediaType("attachment", map[string]string{
		"filename": download.File.OriginalFilename,
	})
	w.Header().Set("Content-Type", download.File.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(download.File.SizeBytes, 10))
	w.Header().Set("Content-Disposition", disposition)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-File-Path", download.File.StorageKey)
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, download.Content)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
}
