package competition

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
	competitionerrs "github.com/orenfsp/orencode-zenko.git/internal/competition/errs"
)

type sqlRepository struct {
	queries *db.Queries
}

func NewRepository(queries *db.Queries) Repository {
	return &sqlRepository{queries: queries}
}

func (r *sqlRepository) CountPublic(ctx context.Context) (int64, error) {
	return r.queries.CountPublicCompetitions(ctx, publicCompetitionStatuses)
}

func (r *sqlRepository) ListPublic(ctx context.Context, limit, offset int32) ([]Competition, error) {
	rows, err := r.queries.ListPublicCompetitions(ctx, db.ListPublicCompetitionsParams{
		Limit:           limit,
		Offset:          offset,
		PPublicStatuses: publicCompetitionStatuses,
	})
	if err != nil {
		return nil, err
	}

	result := make([]Competition, 0, len(rows))
	for _, row := range rows {
		result = append(result, fromListRow(row))
	}
	return result, nil
}

func (r *sqlRepository) ListCalendar(ctx context.Context, filter CalendarFilter) ([]Competition, error) {
	rows, err := r.queries.ListCalendarCompetitions(ctx, db.ListCalendarCompetitionsParams{
		PPublicStatuses: publicCompetitionStatuses,
		PFrom:           timestamptz(filter.From),
		PTo:             timestamptz(filter.To),
		PDisciplineID:   nullableInt64(filter.DisciplineID),
		PLevelID:        nullableInt64(filter.LevelID),
		PLimit:          filter.Limit,
		POffset:         filter.Offset,
	})
	if err != nil {
		return nil, err
	}

	result := make([]Competition, 0, len(rows))
	for _, row := range rows {
		result = append(result, fromCalendarRow(row))
	}
	return result, nil
}

func (r *sqlRepository) GetPublic(ctx context.Context, id int64) (Competition, error) {
	row, err := r.queries.GetPublicCompetition(ctx, db.GetPublicCompetitionParams{
		ID: id, PPublicStatuses: publicCompetitionStatuses,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Competition{}, competitionerrs.ErrNotFound
		}
		return Competition{}, err
	}
	return fromGetRow(row), nil
}

func (r *sqlRepository) ValidateReferences(ctx context.Context, references CompetitionReferences) error {
	row, err := r.queries.ValidateCompetitionReferences(ctx, db.ValidateCompetitionReferencesParams{
		PDisciplineID:       references.DisciplineID,
		PCompetitionTypeID:  references.CompetitionTypeID,
		PCompetitionLevelID: references.CompetitionLevelID,
		PCalendarPlanID:     nullableInt64(references.CalendarPlanID),
	})
	if err != nil {
		return err
	}
	if !row.DisciplineExists || !row.CompetitionTypeExists || !row.CompetitionLevelExists ||
		!row.CalendarPlanExists.Valid || !row.CalendarPlanExists.Bool {
		return competitionerrs.ErrInvalidReference
	}
	return nil
}

func (r *sqlRepository) Create(ctx context.Context, input CreateCompetition) (Competition, error) {
	row, err := r.queries.CreateCompetition(ctx, db.CreateCompetitionParams{
		Name: input.Name, Description: nullableText(input.Description),
		DisciplineID: input.DisciplineID, CompetitionTypeID: input.CompetitionTypeID,
		CompetitionLevelID: input.CompetitionLevelID, Format: input.Format,
		StartAt: timestamptz(input.StartAt), EndAt: timestamptz(input.EndAt),
		RegistrationStartsAt: timestamptz(input.RegistrationStartsAt),
		RegistrationEndsAt:   timestamptz(input.RegistrationEndsAt),
		VenueAddress:         nullableText(input.VenueAddress), OnlineUrl: nullableText(input.OnlineURL),
		CreatedBy: input.CreatedBy, CalendarPlanID: nullableInt64(input.CalendarPlanID),
		CalendarPlanNumber: nullableText(input.CalendarPlanNumber), RegulationUrl: nullableText(input.RegulationURL),
	})
	if err != nil {
		return Competition{}, mapWriteError(err)
	}
	return fromCreateRow(row), nil
}

func (r *sqlRepository) GetByID(ctx context.Context, id int64) (Competition, error) {
	row, err := r.queries.GetCompetition(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Competition{}, competitionerrs.ErrNotFound
		}
		return Competition{}, err
	}
	return fromGetCompetitionRow(row), nil
}

func (r *sqlRepository) ListAdmin(ctx context.Context, createdBy int64, status *string) ([]Competition, error) {
	rows, err := r.queries.ListAdminCompetitions(ctx, db.ListAdminCompetitionsParams{CreatedBy: createdBy, Status: nullableText(status)})
	if err != nil {
		return nil, err
	}
	result := make([]Competition, 0, len(rows))
	for _, row := range rows {
		result = append(result, fromAdminRow(row))
	}
	return result, nil
}

func (r *sqlRepository) Transition(ctx context.Context, id, actorID int64, fromStatus, toStatus string) error {
	_, err := r.queries.TransitionCompetition(ctx, db.TransitionCompetitionParams{ID: id, ChangedBy: actorID, FromStatus: fromStatus, ToStatus: toStatus})
	if errors.Is(err, pgx.ErrNoRows) {
		return competitionerrs.ErrInvalidTransition
	}
	return err
}

func (r *sqlRepository) ListHistory(ctx context.Context, id int64) ([]StatusHistory, error) {
	rows, err := r.queries.ListCompetitionStatusHistory(ctx, id)
	if err != nil {
		return nil, err
	}
	result := make([]StatusHistory, 0, len(rows))
	for _, row := range rows {
		result = append(result, StatusHistory{ID: row.ID, CompetitionID: row.CompetitionID, FromStatus: textPointer(row.FromStatus), ToStatus: row.ToStatus, ChangedBy: row.ChangedBy, ChangedAt: row.ChangedAt.Time})
	}
	return result, nil
}

func (r *sqlRepository) UpdateDraft(ctx context.Context, input UpdateCompetition) (Competition, error) {
	row, err := r.queries.UpdateDraftCompetition(ctx, db.UpdateDraftCompetitionParams{
		PName: input.Name, PDescription: nullableText(input.Description),
		PDisciplineID: input.DisciplineID, PCompetitionTypeID: input.CompetitionTypeID,
		PCompetitionLevelID: input.CompetitionLevelID, PFormat: input.Format,
		PStartAt: timestamptz(input.StartAt), PEndAt: timestamptz(input.EndAt),
		PRegistrationStartsAt: timestamptz(input.RegistrationStartsAt),
		PRegistrationEndsAt:   timestamptz(input.RegistrationEndsAt),
		PVenueAddress:         nullableText(input.VenueAddress), POnlineUrl: nullableText(input.OnlineURL),
		PCalendarPlanID:     nullableInt64(input.CalendarPlanID),
		PCalendarPlanNumber: nullableText(input.CalendarPlanNumber),
		PRegulationUrl:      nullableText(input.RegulationURL), PID: input.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			status, statusErr := r.queries.GetCompetitionStatus(ctx, input.ID)
			if errors.Is(statusErr, pgx.ErrNoRows) {
				return Competition{}, competitionerrs.ErrNotFound
			}
			if statusErr != nil {
				return Competition{}, statusErr
			}
			if status != "draft" {
				return Competition{}, competitionerrs.ErrNotDraft
			}
		}
		return Competition{}, mapWriteError(err)
	}
	return fromUpdateRow(row), nil
}

func mapWriteError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	switch pgErr.Code {
	case "23505":
		if pgErr.ConstraintName == "competition_name_key" {
			return competitionerrs.ErrNameConflict
		}
	case "23503":
		return competitionerrs.ErrInvalidReference
	}
	return err
}

func fromListRow(row db.ListPublicCompetitionsRow) Competition {
	return Competition{
		ID: row.ID, Name: row.Name, Description: textPointer(row.Description),
		DisciplineID: row.DisciplineID, DisciplineName: row.DisciplineName,
		CompetitionTypeID: row.CompetitionTypeID, CompetitionTypeName: row.CompetitionTypeName,
		CompetitionLevelID: row.CompetitionLevelID, CompetitionLevelCode: textPointer(row.CompetitionLevelCode),
		CompetitionLevelName: textPointer(row.CompetitionLevelName), Format: row.Format,
		StartAt: row.StartAt.Time, EndAt: row.EndAt.Time,
		RegistrationStartsAt: row.RegistrationStartsAt.Time, RegistrationEndsAt: row.RegistrationEndsAt.Time,
		VenueAddress: textPointer(row.VenueAddress), OnlineURL: textPointer(row.OnlineUrl), Status: row.Status,
		CalendarPlanID: int64Pointer(row.CalendarPlanID), CalendarPlanName: textPointer(row.CalendarPlanName),
		CalendarPlanYear: int16Pointer(row.CalendarPlanYear), CalendarPlanNumber: textPointer(row.CalendarPlanNumber),
		RegulationURL: textPointer(row.RegulationUrl),
	}
}

func fromCalendarRow(row db.ListCalendarCompetitionsRow) Competition {
	return Competition{
		ID: row.ID, 
		Name: row.Name,
		DisciplineID: row.DisciplineID, 
		DisciplineName: row.DisciplineName,
		CompetitionTypeID: row.CompetitionTypeID, 
		CompetitionTypeName: row.CompetitionTypeName,
		CompetitionLevelID: row.CompetitionLevelID, 
		CompetitionLevelCode: textPointer(row.CompetitionLevelCode),
		CompetitionLevelName: textPointer(row.CompetitionLevelName), 
		Format: row.Format,
		StartAt: row.StartAt.Time,
		EndAt: row.EndAt.Time,
		RegistrationStartsAt: row.RegistrationStartsAt.Time,
        RegistrationEndsAt:   row.RegistrationEndsAt.Time,
		VenueAddress: textPointer(row.VenueAddress), 
		OnlineURL: textPointer(row.OnlineUrl),
		RegulationURL: textPointer(row.RegulationUrl), 
		Status: row.Status,
	}
}

func fromGetRow(row db.GetPublicCompetitionRow) Competition {
	return Competition{
		ID: row.ID, Name: row.Name, Description: textPointer(row.Description),
		DisciplineID: row.DisciplineID, DisciplineName: row.DisciplineName,
		CompetitionTypeID: row.CompetitionTypeID, CompetitionTypeName: row.CompetitionTypeName,
		CompetitionLevelID: row.CompetitionLevelID, CompetitionLevelCode: textPointer(row.CompetitionLevelCode),
		CompetitionLevelName: textPointer(row.CompetitionLevelName), Format: row.Format,
		StartAt: row.StartAt.Time, EndAt: row.EndAt.Time,
		RegistrationStartsAt: row.RegistrationStartsAt.Time, RegistrationEndsAt: row.RegistrationEndsAt.Time,
		VenueAddress: textPointer(row.VenueAddress), OnlineURL: textPointer(row.OnlineUrl), Status: row.Status,
		CalendarPlanID: int64Pointer(row.CalendarPlanID), CalendarPlanName: textPointer(row.CalendarPlanName),
		CalendarPlanYear: int16Pointer(row.CalendarPlanYear), CalendarPlanNumber: textPointer(row.CalendarPlanNumber),
		RegulationURL: textPointer(row.RegulationUrl),
	}
}

func fromGetCompetitionRow(row db.GetCompetitionRow) Competition {
	return Competition{
		ID: row.ID, Name: row.Name, Description: textPointer(row.Description),
		DisciplineID: row.DisciplineID, DisciplineName: row.DisciplineName,
		CompetitionTypeID: row.CompetitionTypeID, CompetitionTypeName: row.CompetitionTypeName,
		CompetitionLevelID: row.CompetitionLevelID, CompetitionLevelCode: textPointer(row.CompetitionLevelCode),
		CompetitionLevelName: textPointer(row.CompetitionLevelName), Format: row.Format,
		StartAt: row.StartAt.Time, EndAt: row.EndAt.Time,
		RegistrationStartsAt: row.RegistrationStartsAt.Time, RegistrationEndsAt: row.RegistrationEndsAt.Time,
		VenueAddress: textPointer(row.VenueAddress), OnlineURL: textPointer(row.OnlineUrl), Status: row.Status, CreatedBy: row.CreatedBy,
		CalendarPlanID: int64Pointer(row.CalendarPlanID), CalendarPlanName: textPointer(row.CalendarPlanName),
		CalendarPlanYear: int16Pointer(row.CalendarPlanYear), CalendarPlanNumber: textPointer(row.CalendarPlanNumber),
		RegulationURL: textPointer(row.RegulationUrl),
	}
}

func fromAdminRow(row db.ListAdminCompetitionsRow) Competition {
	return Competition{
		ID: row.ID, Name: row.Name, Description: textPointer(row.Description), CreatedBy: row.CreatedBy,
		DisciplineID: row.DisciplineID, DisciplineName: row.DisciplineName,
		CompetitionTypeID: row.CompetitionTypeID, CompetitionTypeName: row.CompetitionTypeName,
		CompetitionLevelID: row.CompetitionLevelID, CompetitionLevelCode: textPointer(row.CompetitionLevelCode),
		CompetitionLevelName: textPointer(row.CompetitionLevelName), Format: row.Format,
		StartAt: row.StartAt.Time, EndAt: row.EndAt.Time,
		RegistrationStartsAt: row.RegistrationStartsAt.Time, RegistrationEndsAt: row.RegistrationEndsAt.Time,
		VenueAddress: textPointer(row.VenueAddress), OnlineURL: textPointer(row.OnlineUrl), Status: row.Status,
		CalendarPlanID: int64Pointer(row.CalendarPlanID), CalendarPlanName: textPointer(row.CalendarPlanName),
		CalendarPlanYear: int16Pointer(row.CalendarPlanYear), CalendarPlanNumber: textPointer(row.CalendarPlanNumber),
		RegulationURL: textPointer(row.RegulationUrl),
	}
}

func fromCreateRow(row db.CreateCompetitionRow) Competition {
	return Competition{
		ID: row.ID, Name: row.Name, Description: textPointer(row.Description),
		DisciplineID: row.DisciplineID, DisciplineName: row.DisciplineName,
		CompetitionTypeID: row.CompetitionTypeID, CompetitionTypeName: row.CompetitionTypeName,
		CompetitionLevelID: row.CompetitionLevelID, CompetitionLevelCode: textPointer(row.CompetitionLevelCode),
		CompetitionLevelName: textPointer(row.CompetitionLevelName), Format: row.Format,
		StartAt: row.StartAt.Time, EndAt: row.EndAt.Time,
		RegistrationStartsAt: row.RegistrationStartsAt.Time, RegistrationEndsAt: row.RegistrationEndsAt.Time,
		VenueAddress: textPointer(row.VenueAddress), OnlineURL: textPointer(row.OnlineUrl), Status: row.Status,
		CalendarPlanID: int64Pointer(row.CalendarPlanID), CalendarPlanName: textPointer(row.CalendarPlanName),
		CalendarPlanYear: int16Pointer(row.CalendarPlanYear), CalendarPlanNumber: textPointer(row.CalendarPlanNumber),
		RegulationURL: textPointer(row.RegulationUrl),
	}
}

func fromUpdateRow(row db.UpdateDraftCompetitionRow) Competition {
	return Competition{
		ID: row.ID, Name: row.Name, Description: textPointer(row.Description),
		DisciplineID: row.DisciplineID, DisciplineName: row.DisciplineName,
		CompetitionTypeID: row.CompetitionTypeID, CompetitionTypeName: row.CompetitionTypeName,
		CompetitionLevelID: row.CompetitionLevelID, CompetitionLevelCode: textPointer(row.CompetitionLevelCode),
		CompetitionLevelName: textPointer(row.CompetitionLevelName), Format: row.Format,
		StartAt: row.StartAt.Time, EndAt: row.EndAt.Time,
		RegistrationStartsAt: row.RegistrationStartsAt.Time, RegistrationEndsAt: row.RegistrationEndsAt.Time,
		VenueAddress: textPointer(row.VenueAddress), OnlineURL: textPointer(row.OnlineUrl), Status: row.Status,
		CalendarPlanID: int64Pointer(row.CalendarPlanID), CalendarPlanName: textPointer(row.CalendarPlanName),
		CalendarPlanYear: int16Pointer(row.CalendarPlanYear), CalendarPlanNumber: textPointer(row.CalendarPlanNumber),
		RegulationURL: textPointer(row.RegulationUrl),
	}
}

func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func int64Pointer(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}
	return &value.Int64
}

func int16Pointer(value pgtype.Int2) *int16 {
	if !value.Valid {
		return nil
	}
	return &value.Int16
}

func nullableText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func nullableInt64(value *int64) pgtype.Int8 {
	if value == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *value, Valid: true}
}

func timestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}
