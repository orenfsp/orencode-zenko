package auth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	autherrs "github.com/orenfsp/orencode-zenko.git/internal/auth/errs"
	filedomain "github.com/orenfsp/orencode-zenko.git/internal/file"
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

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 22<<20)
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			writeError(w, http.StatusRequestEntityTooLarge, "регистрационная форма слишком большая")
			return
		}
		writeError(w, http.StatusBadRequest, "ожидается multipart/form-data")
		return
	}
	defer func() { _ = r.MultipartForm.RemoveAll() }()

	req, err := registrationRequestFromMultipart(r.MultipartForm)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	minor, err := h.service.ValidateRegistration(req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	personalHeader, err := singleMultipartFile(r.MultipartForm, "personal_data_consent_file", true)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	parentalHeader, err := singleMultipartFile(r.MultipartForm, "parental_consent_document_file", minor)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !minor && parentalHeader != nil {
		writeError(w, http.StatusBadRequest, "родительское согласие не требуется совершеннолетнему пользователю")
		return
	}

	personalFile, err := h.uploadMultipartFile(
		r.Context(),
		filedomain.PurposePersonalDataConsent,
		personalHeader,
	)
	if err != nil {
		handleFileError(w, err)
		return
	}
	uploadedFiles := []filedomain.UploadedFile{personalFile}

	var parentalFileID *int64
	if parentalHeader != nil {
		parentalFile, uploadErr := h.uploadMultipartFile(
			r.Context(),
			filedomain.PurposeParentalConsent,
			parentalHeader,
		)
		if uploadErr != nil {
			h.cleanupUploadedFiles(r.Context(), uploadedFiles)
			handleFileError(w, uploadErr)
			return
		}
		uploadedFiles = append(uploadedFiles, parentalFile)
		parentalFileID = &parentalFile.ID
	}

	response, err := h.service.Register(
		r.Context(),
		req,
		personalFile.ID,
		parentalFileID,
	)
	if err != nil {
		h.cleanupUploadedFiles(r.Context(), uploadedFiles)
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusCreated, response)
}

func (h *Handler) uploadMultipartFile(
	ctx context.Context,
	purpose filedomain.Purpose,
	header *multipart.FileHeader,
) (filedomain.UploadedFile, error) {
	opened, err := header.Open()
	if err != nil {
		return filedomain.UploadedFile{}, err
	}
	defer opened.Close()
	return h.files.Upload(ctx, filedomain.UploadInput{
		Purpose:  purpose,
		Filename: header.Filename,
		Size:     header.Size,
		Content:  opened,
	})
}

func (h *Handler) cleanupUploadedFiles(ctx context.Context, uploaded []filedomain.UploadedFile) {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
	defer cancel()
	for index := len(uploaded) - 1; index >= 0; index-- {
		if err := h.files.Discard(cleanupCtx, uploaded[index].ID); err != nil &&
			!errors.Is(err, filedomain.ErrFileAttached) {
			logger.Log.Error("cleanup registration file failed",
				zap.Int64("file_id", uploaded[index].ID),
				zap.Error(err),
			)
		}
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	response, err := h.service.Login(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, response)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := decodeJSON(w, r, &req); err != nil || strings.TrimSpace(req.RefreshToken) == "" {
		writeError(w, http.StatusBadRequest, "refresh_token обязателен")
		return
	}
	response, err := h.service.Refresh(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, response)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := decodeJSON(w, r, &req); err != nil || strings.TrimSpace(req.RefreshToken) == "" {
		writeError(w, http.StatusBadRequest, "refresh_token обязателен")
		return
	}
	if err := h.service.Logout(r.Context(), req.RefreshToken); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) PatchProfile(w http.ResponseWriter, r *http.Request) {
	accessToken, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется access-токен")
		return
	}
	var req PatchProfileRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	if req.TelegramUsername == nil && req.RoleID == nil {
		writeError(w, http.StatusBadRequest, "не переданы изменения профиля")
		return
	}
	response, err := h.service.PatchProfile(r.Context(), accessToken, req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, response)
}

func registrationRequestFromMultipart(form *multipart.Form) (RegisterRequest, error) {
	allowedValues := map[string]bool{
		"first_name":  true,
		"last_name":   true,
		"middle_name": true,
		"birth_date":  true,
		"email":       true,
		"password":    true,
	}
	for field := range form.Value {
		if !allowedValues[field] {
			return RegisterRequest{}, errors.New("неизвестное поле: " + field)
		}
	}
	allowedFiles := map[string]bool{
		"personal_data_consent_file":     true,
		"parental_consent_document_file": true,
	}
	for field := range form.File {
		if !allowedFiles[field] {
			return RegisterRequest{}, errors.New("неизвестное файловое поле: " + field)
		}
	}
	firstName, err := singleMultipartValue(form, "first_name", true)
	if err != nil {
		return RegisterRequest{}, err
	}
	lastName, err := singleMultipartValue(form, "last_name", true)
	if err != nil {
		return RegisterRequest{}, err
	}
	middleName, err := singleMultipartValue(form, "middle_name", false)
	if err != nil {
		return RegisterRequest{}, err
	}
	birthDate, err := singleMultipartValue(form, "birth_date", true)
	if err != nil {
		return RegisterRequest{}, err
	}
	email, err := singleMultipartValue(form, "email", true)
	if err != nil {
		return RegisterRequest{}, err
	}
	password, err := singleMultipartValue(form, "password", true)
	if err != nil {
		return RegisterRequest{}, err
	}
	return RegisterRequest{
		FirstName:  firstName,
		LastName:   lastName,
		MiddleName: middleName,
		BirthDate:  birthDate,
		Email:      email,
		Password:   password,
	}, nil
}

func singleMultipartValue(form *multipart.Form, field string, required bool) (string, error) {
	values := form.Value[field]
	if len(values) == 0 {
		if required {
			return "", errors.New("обязательное поле отсутствует: " + field)
		}
		return "", nil
	}
	if len(values) != 1 {
		return "", errors.New("поле должно быть передано один раз: " + field)
	}
	return values[0], nil
}

func singleMultipartFile(
	form *multipart.Form,
	field string,
	required bool,
) (*multipart.FileHeader, error) {
	files := form.File[field]
	if len(files) == 0 {
		if required {
			return nil, errors.New("обязательный файл отсутствует: " + field)
		}
		return nil, nil
	}
	if len(files) != 1 {
		return nil, errors.New("файл должен быть передан один раз: " + field)
	}
	return files[0], nil
}

func handleFileError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, filedomain.ErrFileTooLarge):
		writeError(w, http.StatusRequestEntityTooLarge, "файл превышает допустимый размер")
	case errors.Is(err, filedomain.ErrUnsupportedType):
		writeError(w, http.StatusUnsupportedMediaType, "неподдерживаемый формат файла")
	case errors.Is(err, filedomain.ErrEmptyFile), errors.Is(err, filedomain.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "некорректный файл")
	default:
		logger.Log.Error("registration file upload failed", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "ошибка сохранения файла")
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

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, autherrs.ErrConsentRequired):
		writeError(w, http.StatusBadRequest, "необходимо согласие на обработку персональных данных")
	case errors.Is(err, autherrs.ErrParentalConsentRequired):
		writeError(w, http.StatusBadRequest, "для пользователя младше 18 лет необходимо письменное разрешение родителей")
	case errors.Is(err, autherrs.ErrInvalidRegistrationData):
		writeError(w, http.StatusBadRequest, "некорректные данные")
	case errors.Is(err, autherrs.ErrInvalidDocument):
		writeError(w, http.StatusBadRequest, "регистрационный документ недействителен")
	case errors.Is(err, autherrs.ErrTelegramUsernameInvalid):
		writeError(w, http.StatusBadRequest, "некорректный Telegram username")
	case errors.Is(err, autherrs.ErrRoleNotFound):
		writeError(w, http.StatusBadRequest, "роль не найдена")
	case errors.Is(err, autherrs.ErrEmailAlreadyExists):
		writeError(w, http.StatusConflict, "пользователь с таким email уже существует")
	case errors.Is(err, autherrs.ErrTelegramUsernameTaken):
		writeError(w, http.StatusConflict, "этот Telegram username уже используется")
	case errors.Is(err, autherrs.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "неверный email или пароль")
	case errors.Is(err, autherrs.ErrInvalidSession), errors.Is(err, autherrs.ErrAccessDenied):
		writeError(w, http.StatusUnauthorized, "недействительная или истёкшая сессия")
	default:
		logger.Log.Error("auth request failed", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Message: message})
}

func writeSuccess(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(SuccessResponse{Data: data})
}
