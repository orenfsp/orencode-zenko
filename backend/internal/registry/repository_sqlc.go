package registry

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
)

type sqlRepository struct{ queries *db.Queries }

func NewRepository(queries *db.Queries) Repository { return &sqlRepository{queries: queries} }

func (r *sqlRepository) ListUsers(ctx context.Context, filter UserFilter) (Page[User], error) {
	rows, err := r.queries.ListRegistryUsers(ctx, db.ListRegistryUsersParams{
		PSearch: nullableText(filter.Search), POrganizationID: nullableInt64(filter.OrganizationID),
		PMunicipalityID: nullableInt64(filter.MunicipalityID), PRoleCode: nullableText(filter.RoleCode),
		PDisciplineID: nullableInt64(filter.DisciplineID), PRankID: nullableInt64(filter.RankID),
		PAgeGroupID: nullableInt64(filter.AgeGroupID), POffset: filter.Offset, PLimit: filter.Limit,
	})
	if err != nil {
		return Page[User]{}, err
	}
	page := Page[User]{Items: make([]User, 0, len(rows)), Limit: filter.Limit, Offset: filter.Offset}
	for _, row := range rows {
		user := User{
			ID: row.ID, LastName: row.LastName, FirstName: row.FirstName,
			MiddleName: textPointer(row.MiddleName), Birth: row.Birth.Time,
			Organization: referencePointer(row.OrganizationID, row.OrganizationName),
			Municipality: referencePointer(row.MunicipalityID, row.MunicipalityName),
			Roles:        make([]Role, 0), Disciplines: make([]Reference, 0), Ranks: make([]Rank, 0),
		}
		if err := decodeJSONValue(row.Roles, &user.Roles); err != nil {
			return Page[User]{}, fmt.Errorf("decode roles: %w", err)
		}
		if err := decodeJSONValue(row.Disciplines, &user.Disciplines); err != nil {
			return Page[User]{}, fmt.Errorf("decode disciplines: %w", err)
		}
		if err := decodeJSONValue(row.Ranks, &user.Ranks); err != nil {
			return Page[User]{}, fmt.Errorf("decode ranks: %w", err)
		}
		if row.JudgeCategory != nil {
			var category JudgeCategory
			if err := decodeJSONValue(row.JudgeCategory, &category); err != nil {
				return Page[User]{}, fmt.Errorf("decode judge category: %w", err)
			}
			user.JudgeCategory = &category
		}
		page.Items = append(page.Items, user)
		page.Total = row.TotalCount
	}
	return page, nil
}

func (r *sqlRepository) ListOrganizations(ctx context.Context, filter OrganizationFilter) (Page[Organization], error) {
	rows, err := r.queries.ListRegistryOrganizations(ctx, db.ListRegistryOrganizationsParams{
		PSearch: nullableText(filter.Search), PMunicipalityID: nullableInt64(filter.MunicipalityID),
		POffset: filter.Offset, PLimit: filter.Limit,
	})
	if err != nil {
		return Page[Organization]{}, err
	}
	page := Page[Organization]{Items: make([]Organization, 0, len(rows)), Limit: filter.Limit, Offset: filter.Offset}
	for _, row := range rows {
		page.Items = append(page.Items, Organization{
			ID: row.ID, Name: row.Name, OrganizationType: row.OrganizationType,
			Municipality: referencePointer(row.MunicipalityID, row.MunicipalityName),
		})
		page.Total = row.TotalCount
	}
	return page, nil
}

func decodeJSONValue(value any, destination any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, destination)
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

func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

func referencePointer(id pgtype.Int8, name pgtype.Text) *Reference {
	if !id.Valid || !name.Valid {
		return nil
	}
	return &Reference{ID: id.Int64, Name: name.String}
}
