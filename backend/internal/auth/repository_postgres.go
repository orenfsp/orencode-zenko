package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	autherrs "github.com/orenfsp/orencode-zenko.git/internal/auth/errs"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) CreateUser(ctx context.Context, data CreateUser) (User, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return User{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var user User
	err = tx.QueryRow(ctx, `
		INSERT INTO "user" (
			first_name, last_name, middle_name, birth, email, password_hash
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, email, password_hash, tg_username, access_status
	`, data.FirstName, data.LastName, data.MiddleName, data.BirthDate, data.Email, data.PasswordHash).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.TelegramUsername,
		&user.AccessStatus,
	)
	if err != nil {
		return User{}, mapPostgresError(err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO consent_data (user_id, policy_version)
		VALUES ($1, $2)
	`, user.ID, data.ConsentPolicyVersion)
	if err != nil {
		return User{}, err
	}

	if data.ParentalConsentDocumentURL != nil {
		_, err = tx.Exec(ctx, `
			INSERT INTO parental_consent (user_id, document_url)
			VALUES ($1, $2)
		`, user.ID, *data.ParentalConsentDocumentURL)
		if err != nil {
			return User{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return User{}, err
	}
	return user, nil
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (User, error) {
	return r.getUser(ctx, `u.email = $1`, email)
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id int64) (User, error) {
	return r.getUser(ctx, `u.id = $1`, id)
}

func (r *PostgresRepository) getUser(ctx context.Context, condition string, value any) (User, error) {
	var user User
	var roleID pgtype.Int8
	query := `
		SELECT u.id, u.email, u.password_hash, u.tg_username, u.access_status, selected_role.role_id
		FROM "user" u
		LEFT JOIN LATERAL (
			SELECT ru.role_id
			FROM role_user ru
			WHERE ru.user_id = u.id
			ORDER BY ru.granted_at DESC, ru.id DESC
			LIMIT 1
		) selected_role ON true
		WHERE ` + condition
	err := r.pool.QueryRow(ctx, query, value).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.TelegramUsername,
		&user.AccessStatus,
		&roleID,
	)
	if err != nil {
		return User{}, err
	}
	if roleID.Valid {
		user.RoleID = &roleID.Int64
	}
	return user, nil
}

func (r *PostgresRepository) SetRole(ctx context.Context, userID, roleID int64) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var roleExists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM role WHERE id = $1)`, roleID).Scan(&roleExists); err != nil {
		return err
	}
	if !roleExists {
		return autherrs.ErrRoleNotFound
	}
	if _, err := tx.Exec(ctx, `DELETE FROM role_user WHERE user_id = $1`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO role_user (user_id, role_id) VALUES ($1, $2)`, userID, roleID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PostgresRepository) UpdateTelegramUsername(ctx context.Context, userID int64, username *string) error {
	command, err := r.pool.Exec(ctx, `
		UPDATE "user"
		SET tg_username = $2, updated_at = now()
		WHERE id = $1
	`, userID, username)
	if err != nil {
		return mapPostgresError(err)
	}
	if command.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *PostgresRepository) CreateSession(ctx context.Context, data CreateSession) (Session, error) {
	var session Session
	err := r.pool.QueryRow(ctx, `
		INSERT INTO user_session (user_id, refresh_token_hash, device_id, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, refresh_token_hash, device_id, expires_at
	`, data.UserID, data.RefreshTokenHash, data.DeviceID, data.ExpiresAt).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&session.DeviceID,
		&session.ExpiresAt,
	)
	return session, err
}

func (r *PostgresRepository) GetSessionByRefreshHash(ctx context.Context, hash string) (Session, error) {
	var session Session
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, refresh_token_hash, device_id, expires_at
		FROM user_session
		WHERE refresh_token_hash = $1
	`, hash).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&session.DeviceID,
		&session.ExpiresAt,
	)
	return session, err
}

func (r *PostgresRepository) RotateSession(ctx context.Context, id int64, refreshHash string, expiresAt time.Time) error {
	command, err := r.pool.Exec(ctx, `
		UPDATE user_session
		SET refresh_token_hash = $2, expires_at = $3, updated_at = now()
		WHERE id = $1
	`, id, refreshHash, expiresAt)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return autherrs.ErrInvalidSession
	}
	return nil
}

func (r *PostgresRepository) DeleteSession(ctx context.Context, id int64) error {
	command, err := r.pool.Exec(ctx, `DELETE FROM user_session WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return autherrs.ErrInvalidSession
	}
	return nil
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
