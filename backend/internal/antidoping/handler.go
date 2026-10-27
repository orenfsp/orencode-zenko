package antidoping

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	filedomain "github.com/orenfsp/orencode-zenko.git/internal/file"
	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
	"github.com/orenfsp/orencode-zenko.git/pkg/logger"
	"go.uber.org/zap"
)

type Handler struct {
	service *Service
	files   *filedomain.Service
}

func NewHandler(service *Service, files *filedomain.Service) *Handler {
	return &Handler{service: service, files: files}
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 11<<20)
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			writeError(w, http.StatusRequestEntityTooLarge, "файл превышает максимальный размер 10 МБ")
			return
		}
		writeError(w, http.StatusBadRequest, "ожидается multipart/form-data")
		return
	}
	defer func() { _ = r.MultipartForm.RemoveAll() }()

	if len(r.MultipartForm.Value) != 0 || len(r.MultipartForm.File) != 1 {
		writeError(w, http.StatusBadRequest, "разрешено только поле certificate_file")
		return
	}
	files := r.MultipartForm.File["certificate_file"]
	if len(files) != 1 {
		writeError(w, http.StatusBadRequest, "certificate_file должен быть передан один раз")
		return
	}
	opened, err := files[0].Open()
	if err != nil {
		writeError(w, http.StatusBadRequest, "не удалось прочитать certificate_file")
		return
	}
	uploaded, err := h.files.Upload(r.Context(), filedomain.UploadInput{
		Purpose:  filedomain.PurposeAntiDopingCertificate,
		Filename: files[0].Filename,
		Size:     files[0].Size,
		Content:  opened,
	})
	_ = opened.Close()
	if err != nil {
		handleFileError(w, err)
		return
	}

	certificate, err := h.service.Create(r.Context(), identity.UserID, uploaded.ID)
	if err != nil {
		h.cleanupUploadedFile(r.Context(), uploaded.ID)
		handleError(w, err)
		return
	}
	writeSuccess(w, http.StatusCreated, certificate)
}

func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	certificates, err := h.service.ListByUser(r.Context(), identity.UserID)
	if err != nil {
		handleError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, certificates)
}

func (h *Handler) Review(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	certificateID, err := strconv.ParseInt(r.PathValue("certificateID"), 10, 64)
	if err != nil || certificateID <= 0 {
		writeError(w, http.StatusBadRequest, "некорректный certificateID")
		return
	}
	var request ReviewRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	certificate, err := h.service.Review(r.Context(), identity.UserID, certificateID, request)
	if err != nil {
		handleError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, certificate)
}

func (h *Handler) cleanupUploadedFile(ctx context.Context, fileID int64) {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
	defer cancel()
	if err := h.files.Discard(cleanupCtx, fileID); err != nil && !errors.Is(err, filedomain.ErrFileAttached) {
		logger.Log.Error("cleanup anti-doping file failed", zap.Int64("file_id", fileID), zap.Error(err))
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one JSON object")
	}
	return nil
}

func handleFileError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, filedomain.ErrFileTooLarge):
		writeError(w, http.StatusRequestEntityTooLarge, "файл превышает максимальный размер 10 МБ")
	case errors.Is(err, filedomain.ErrUnsupportedType):
		writeError(w, http.StatusBadRequest, "допустимы только PDF, JPEG и PNG")
	case errors.Is(err, filedomain.ErrEmptyFile), errors.Is(err, filedomain.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "некорректный certificate_file")
	default:
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
	}
}

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "некорректные данные сертификата")
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "сертификат не найден")
	case errors.Is(err, ErrConflict):
		writeError(w, http.StatusConflict, "сертификат уже рассмотрен")
	default:
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
	}
}

func writeSuccess(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
}
