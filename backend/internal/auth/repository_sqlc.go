package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
	autherrs "github.com/orenfsp/orencode-zenko.git/internal/auth/errs"
	filedomain "github.com/orenfsp/orencode-zenko.git/internal/file"
)

type sqlRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewRepository(queries *db.Queries, pool *pgxpool.Pool) Repository {
	return &sqlRepository{queries: queries, pool: pool}
}

func (r *sqlRepository) CreateUser(ctx context.Context, data CreateUser) (User, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return User{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)
	if data.PersonalDataConsentFileID <= 0 ||
		(data.ParentalConsentFileID != nil && *data.ParentalConsentFileID == data.PersonalDataConsentFileID) {
		return User{}, autherrs.ErrInvalidDocument
	}
	if err := validateRegistrationFile(
		ctx,
		queries,
		data.PersonalDataConsentFileID,
		filedomain.PurposePersonalDataConsent,
	); err != nil {
		return User{}, err
	}
	if data.ParentalConsentFileID != nil {
		if err := validateRegistrationFile(
			ctx,
			queries,
			*data.ParentalConsentFileID,
			filedomain.PurposeParentalConsent,
		); err != nil {
			return User{}, err
		}
	}

	row, err := queries.CreateAuthUser(ctx, db.CreateAuthUserParams{
		FirstName:    data.FirstName,
		LastName:     data.LastName,
		MiddleName:   optionalText(data.MiddleName),
		Birth:        pgtype.Date{Time: data.BirthDate, Valid: true},
		Email:        data.Email,
		PasswordHash: data.PasswordHash,
	})
	if err != nil {
		return User{}, mapPostgresError(err)
	}
	user := User{
		ID:               row.ID,
		Email:            row.Email,
		PasswordHash:     row.PasswordHash,
		TelegramUsername: optionalString(row.TgUsername),
		AccessStatus:     row.AccessStatus,
	}

	if err := queries.CreateConsentData(ctx, db.CreateConsentDataParams{
		UserID: user.ID,
		FileID: pgtype.Int8{Int64: data.PersonalDataConsentFileID, Valid: true},
	}); err != nil {
		return User{}, err
	}
	if data.ParentalConsentFileID != nil {
		if err := queries.CreateParentalConsent(ctx, db.CreateParentalConsentParams{
			UserID: user.ID,
			FileID: pgtype.Int8{Int64: *data.ParentalConsentFileID, Valid: true},
		}); err != nil {
			return User{}, err
		}
	}
	if rows, err := queries.MarkUploadedFileAttached(ctx, data.PersonalDataConsentFileID); err != nil {
		return User{}, err
	} else if rows != 1 {
		return User{}, autherrs.ErrInvalidDocument
	}
	if data.ParentalConsentFileID != nil {
		if rows, err := queries.MarkUploadedFileAttached(ctx, *data.ParentalConsentFileID); err != nil {
			return User{}, err
		} else if rows != 1 {
			return User{}, autherrs.ErrInvalidDocument
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return User{}, err
	}
	return user, nil
}

func (r *sqlRepository) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row, err := r.queries.GetAuthUserByEmail(ctx, email)
	if err != nil {
		return User{}, err
	}
	return User{
		ID:               row.ID,
		Email:            row.Email,
		PasswordHash:     row.PasswordHash,
		TelegramUsername: optionalString(row.TgUsername),
		RoleID:           optionalID(row.RoleID),
		AccessStatus:     row.AccessStatus,
	}, nil
}

func (r *sqlRepository) GetUserByID(ctx context.Context, id int64) (User, error) {
	row, err := r.queries.GetAuthUserByID(ctx, id)
	if err != nil {
		return User{}, err
	}
	return User{
		ID:               row.ID,
		Email:            row.Email,
		PasswordHash:     row.PasswordHash,
		TelegramUsername: optionalString(row.TgUsername),
		RoleID:           optionalID(row.RoleID),
		AccessStatus:     row.AccessStatus,
	}, nil
}

func (r *sqlRepository) SetRole(ctx context.Context, userID, roleID int64) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)

	assignable, err := queries.IsRoleSelfAssignable(ctx, roleID)
	if err != nil {
		return err
	}
	if !assignable {
		return autherrs.ErrRoleNotFound
	}
	if err := queries.DeleteUserRoles(ctx, userID); err != nil {
		return err
	}
	if err := queries.AssignRole(ctx, db.AssignRoleParams{UserID: userID, RoleID: roleID}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *sqlRepository) UpdateTelegramUsername(ctx context.Context, userID int64, username *string) error {
	rows, err := r.queries.UpdateTelegramUsername(ctx, db.UpdateTelegramUsernameParams{
		ID:         userID,
		TgUsername: optionalText(username),
	})
	if err != nil {
		return mapPostgresError(err)
	}
	if rows == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *sqlRepository) CreateSession(ctx context.Context, data CreateSession) (Session, error) {
	row, err := r.queries.CreateUserSession(ctx, db.CreateUserSessionParams{
		UserID:           data.UserID,
		RefreshTokenHash: data.RefreshTokenHash,
		ExpiresAt:        pgtype.Timestamptz{Time: data.ExpiresAt, Valid: true},
	})
	if err != nil {
		return Session{}, err
	}
	return Session{
		ID:               row.ID,
		UserID:           row.UserID,
		RefreshTokenHash: row.RefreshTokenHash,
		ExpiresAt:        row.ExpiresAt.Time,
	}, nil
}

func (r *sqlRepository) GetSessionByRefreshHash(ctx context.Context, hash string) (Session, error) {
	row, err := r.queries.GetUserSessionByRefreshHash(ctx, hash)
	if err != nil {
		return Session{}, err
	}
	return Session{
		ID:               row.ID,
		UserID:           row.UserID,
		RefreshTokenHash: row.RefreshTokenHash,
		ExpiresAt:        row.ExpiresAt.Time,
	}, nil
}

func validateRegistrationFile(
	ctx context.Context,
	queries *db.Queries,
	fileID int64,
	expectedPurpose filedomain.Purpose,
) error {
	row, err := queries.GetUploadedFileForAttachment(ctx, fileID)
	if errors.Is(err, pgx.ErrNoRows) {
		return autherrs.ErrInvalidDocument
	}
	if err != nil {
		return err
	}
	if row.Status != filedomain.StatusUploaded || row.PurposeCode != string(expectedPurpose) {
		return autherrs.ErrInvalidDocument
	}
	return nil
}

func (r *sqlRepository) RotateSession(ctx context.Context, id int64, refreshHash string, expiresAt time.Time) error {
	rows, err := r.queries.RotateUserSession(ctx, db.RotateUserSessionParams{
		ID:               id,
		RefreshTokenHash: refreshHash,
		ExpiresAt:        pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return autherrs.ErrInvalidSession
	}
	return nil
}

func (r *sqlRepository) DeleteSession(ctx context.Context, id int64) error {
	rows, err := r.queries.DeleteUserSession(ctx, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		return autherrs.ErrInvalidSession
	}
	return nil
}

func optionalText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func optionalString(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

func optionalID(value int64) *int64 {
	if value == 0 {
		return nil
	}
	result := value
	return &result
}

func mapPostgresError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return err
	}
	switch pgErr.ConstraintName {
	case "user_email_key":
		return autherrs.ErrEmailAlreadyExists
	case "user_tg_username_key":
		return autherrs.ErrTelegramUsernameTaken
	default:
		return err
	}
}
