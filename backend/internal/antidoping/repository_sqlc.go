package antidoping

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
	filedomain "github.com/orenfsp/orencode-zenko.git/internal/file"
)

type SQLCRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewRepository(queries *db.Queries, pool *pgxpool.Pool) (*SQLCRepository, error) {
	if queries == nil || pool == nil {
		return nil, errors.New("anti-doping queries and pool are required")
	}
	return &SQLCRepository{queries: queries, pool: pool}, nil
}

func (r *SQLCRepository) Create(ctx context.Context, userID, fileID int64) (Certificate, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Certificate{}, fmt.Errorf("begin create anti-doping certificate: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)

	file, err := queries.GetAntiDopingFileForAttachment(ctx, fileID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Certificate{}, ErrInvalidInput
	}
	if err != nil {
		return Certificate{}, fmt.Errorf("get anti-doping file: %w", err)
	}
	if file.Status != filedomain.StatusUploaded || file.PurposeCode != string(filedomain.PurposeAntiDopingCertificate) {
		return Certificate{}, ErrInvalidInput
	}

	row, err := queries.CreateAntiDopingCertificate(ctx, db.CreateAntiDopingCertificateParams{
		UserID: userID,
		FileID: fileID,
	})
	if err != nil {
		return Certificate{}, fmt.Errorf("create anti-doping certificate: %w", err)
	}
	rows, err := queries.MarkUploadedFileAttached(ctx, fileID)
	if err != nil {
		return Certificate{}, fmt.Errorf("attach anti-doping file: %w", err)
	}
	if rows != 1 {
		return Certificate{}, ErrInvalidInput
	}
	if err := tx.Commit(ctx); err != nil {
		return Certificate{}, fmt.Errorf("commit anti-doping certificate: %w", err)
	}
	return mapCertificate(row), nil
}

func (r *SQLCRepository) ListByUser(ctx context.Context, userID int64) ([]Certificate, error) {
	rows, err := r.queries.ListAntiDopingCertificatesByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list anti-doping certificates: %w", err)
	}
	result := make([]Certificate, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapCertificate(row))
	}
	return result, nil
}

func (r *SQLCRepository) ListAll(ctx context.Context) ([]Certificate, error) {
	rows, err := r.queries.ListAllAntiDopingCertificates(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all anti-doping certificates: %w", err)
	}
	result := make([]Certificate, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapCertificate(row))
	}
	return result, nil
}

func (r *SQLCRepository) Review(
	ctx context.Context,
	reviewerUserID, certificateID int64,
	input ReviewInput,
) (Certificate, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Certificate{}, fmt.Errorf("begin anti-doping review: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)

	current, err := queries.GetAntiDopingCertificateForReview(ctx, certificateID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Certificate{}, ErrNotFound
	}
	if err != nil {
		return Certificate{}, fmt.Errorf("get anti-doping certificate for review: %w", err)
	}
	if current.Status != StatusManualReview {
		return Certificate{}, ErrConflict
	}

	var reviewed db.AntiDoping
	switch input.Status {
	case StatusValid:
		reviewed, err = queries.ApproveAntiDopingCertificate(ctx, db.ApproveAntiDopingCertificateParams{
			CompletedAt:      pgtype.Date{Time: input.CompletedAt, Valid: true},
			ReviewedByUserID: pgtype.Int8{Int64: reviewerUserID, Valid: true},
			CertificateID:    certificateID,
		})
	case StatusRejected:
		reviewed, err = queries.RejectAntiDopingCertificate(ctx, db.RejectAntiDopingCertificateParams{
			ReviewedByUserID: pgtype.Int8{Int64: reviewerUserID, Valid: true},
			RejectionReason:  pgtype.Text{String: input.RejectionReason, Valid: true},
			CertificateID:    certificateID,
		})
	default:
		return Certificate{}, ErrInvalidInput
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return Certificate{}, ErrInvalidInput
	}
	if err != nil {
		return Certificate{}, fmt.Errorf("review anti-doping certificate: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Certificate{}, fmt.Errorf("commit anti-doping review: %w", err)
	}
	return mapCertificate(reviewed), nil
}

func (r *SQLCRepository) Expire(ctx context.Context) (int64, error) {
	rows, err := r.queries.ExpireAntiDopingCertificates(ctx)
	if err != nil {
		return 0, fmt.Errorf("expire anti-doping certificates: %w", err)
	}
	return rows, nil
}

func mapCertificate(row db.AntiDoping) Certificate {
	result := Certificate{
		ID:        row.ID,
		UserID:    row.UserID,
		FileID:    row.FileID,
		Status:    row.Status,
		CreatedAt: row.CreatedAt.Time,
	}
	result.CompletedAt = dateString(row.CompletedAt)
	result.ValidUntil = dateString(row.ValidUntil)
	if row.ReviewedAt.Valid {
		reviewedAt := row.ReviewedAt.Time
		result.ReviewedAt = &reviewedAt
	}
	if row.RejectionReason.Valid {
		reason := row.RejectionReason.String
		result.RejectionReason = &reason
	}
	return result
}

func dateString(value pgtype.Date) *string {
	if !value.Valid {
		return nil
	}
	formatted := value.Time.Format(time.DateOnly)
	return &formatted
}
