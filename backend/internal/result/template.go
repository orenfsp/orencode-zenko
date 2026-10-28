package result

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
)

// TemplateField describes one ordered, user-defined result column.
type TemplateField struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

func (s *Service) CreateTemplate(ctx context.Context, request CreateTemplateRequest) (ResultTemplate, error) {
	if request.DisciplineID <= 0 || validateTemplate(request.Name, request.Fields) != nil {
		return ResultTemplate{}, ErrInvalidRequest
	}
	return s.repository.CreateTemplate(ctx, request.DisciplineID, strings.TrimSpace(request.Name), request.Fields)
}

func (s *Service) ListTemplates(ctx context.Context, disciplineID int64) ([]ResultTemplate, error) {
	if disciplineID <= 0 {
		return nil, ErrInvalidRequest
	}
	return s.repository.ListTemplates(ctx, disciplineID)
}

func (s *Service) GetTemplate(ctx context.Context, id int64) (ResultTemplate, error) {
	if id <= 0 {
		return ResultTemplate{}, ErrInvalidRequest
	}
	return s.repository.GetTemplate(ctx, id)
}

func (s *Service) UpdateTemplate(ctx context.Context, id int64, request UpdateTemplateRequest) (ResultTemplate, error) {
	if request.Name == nil && request.Fields == nil && request.IsActive == nil {
		return ResultTemplate{}, ErrInvalidRequest
	}
	current, err := s.GetTemplate(ctx, id)
	if err != nil {
		return ResultTemplate{}, err
	}
	if request.Name != nil {
		current.Name = strings.TrimSpace(*request.Name)
	}
	if request.Fields != nil {
		current.Fields = *request.Fields
	}
	if request.IsActive != nil {
		current.IsActive = *request.IsActive
	}
	if err := validateTemplate(current.Name, current.Fields); err != nil {
		return ResultTemplate{}, err
	}
	return s.repository.UpdateTemplate(ctx, current)
}

func templateFromDB(row db.DisciplineResultTemplate) (ResultTemplate, error) {
	var fields []TemplateField
	if err := json.Unmarshal(row.Fields, &fields); err != nil {
		return ResultTemplate{}, err
	}
	return ResultTemplate{ID: row.ID, DisciplineID: row.DisciplineID, Name: row.Name, Fields: fields,
		IsActive: row.IsActive, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time}, nil
}

func (r *SQLCRepository) CreateTemplate(ctx context.Context, disciplineID int64, name string, fields []TemplateField) (ResultTemplate, error) {
	var exists bool
	if err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM discipline WHERE id=$1)`, disciplineID).Scan(&exists); err != nil {
		return ResultTemplate{}, err
	}
	if !exists {
		return ResultTemplate{}, ErrInvalidRequest
	}
	raw, err := json.Marshal(fields)
	if err != nil {
		return ResultTemplate{}, err
	}
	row, err := r.queries.CreateResultTemplate(ctx, db.CreateResultTemplateParams{DisciplineID: disciplineID, Name: name, Fields: raw})
	if err != nil {
		return ResultTemplate{}, err
	}
	return templateFromDB(row)
}

func (r *SQLCRepository) GetTemplate(ctx context.Context, id int64) (ResultTemplate, error) {
	row, err := r.queries.GetResultTemplate(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return ResultTemplate{}, ErrTemplateNotFound
	}
	if err != nil {
		return ResultTemplate{}, err
	}
	return templateFromDB(row)
}

func (r *SQLCRepository) ListTemplates(ctx context.Context, disciplineID int64) ([]ResultTemplate, error) {
	rows, err := r.queries.ListResultTemplates(ctx, disciplineID)
	if err != nil {
		return nil, err
	}
	items := make([]ResultTemplate, 0, len(rows))
	for _, row := range rows {
		item, err := templateFromDB(row)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *SQLCRepository) UpdateTemplate(ctx context.Context, template ResultTemplate) (ResultTemplate, error) {
	raw, err := json.Marshal(template.Fields)
	if err != nil {
		return ResultTemplate{}, err
	}
	row, err := r.queries.UpdateResultTemplate(ctx, db.UpdateResultTemplateParams{ID: template.ID, Name: template.Name, Fields: raw, IsActive: template.IsActive})
	if errors.Is(err, pgx.ErrNoRows) {
		return ResultTemplate{}, ErrTemplateNotFound
	}
	if err != nil {
		return ResultTemplate{}, err
	}
	return templateFromDB(row)
}

var templateKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

func validateTemplate(name string, fields []TemplateField) error {
	if strings.TrimSpace(name) == "" || utf8.RuneCountInString(name) > 200 || len(fields) == 0 || len(fields) > 12 {
		return ErrInvalidRequest
	}
	keys := make(map[string]bool, len(fields))
	for _, field := range fields {
		if !templateKeyPattern.MatchString(field.Key) || keys[field.Key] || strings.TrimSpace(field.Label) == "" ||
			utf8.RuneCountInString(field.Label) > 200 || (field.Type != "text" && field.Type != "number") {
			return ErrInvalidRequest
		}
		keys[field.Key] = true
	}
	return nil
}
