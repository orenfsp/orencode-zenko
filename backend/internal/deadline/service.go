package deadline

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type Service struct {
	repository  Repository
	permissions PermissionChecker
}

func NewService(repository Repository, permissions PermissionChecker) *Service {
	return &Service{repository: repository, permissions: permissions}
}

func (s *Service) List(ctx context.Context, userID int64) ([]Item, error) {
	if userID <= 0 {
		return nil, ErrInvalidInput
	}
	isFederation, err := s.permissions.HasPermission(ctx, userID, PermissionManage)
	if err != nil {
		return nil, err
	}
	roles, err := s.repository.RoleCodes(ctx, userID)
	if err != nil {
		return nil, err
	}
	roleSet := make(map[string]bool, len(roles))
	for _, role := range roles {
		roleSet[role] = true
	}

	items := make([]Item, 0)
	appendItems := func(next []Item, err error) error {
		if err != nil {
			return err
		}
		items = append(items, next...)
		return nil
	}
	if isFederation {
		if err := appendItems(s.repository.Registration(ctx)); err != nil {
			return nil, err
		}
		if err := appendItems(s.repository.AllRanks(ctx)); err != nil {
			return nil, err
		}
		if err := appendItems(s.repository.AllJudgeCategories(ctx)); err != nil {
			return nil, err
		}
		if err := appendItems(s.repository.AllJudgeAssignments(ctx)); err != nil {
			return nil, err
		}
		reports, err := s.repository.ListReports(ctx)
		if err != nil {
			return nil, err
		}
		for _, report := range reports {
			items = append(items, Item{Type: "report", ID: report.ID, Title: report.Name, DueAt: report.DueAt, Status: report.Status})
		}
	} else {
		if roleSet["participant"] || roleSet["coach"] {
			if err := appendItems(s.repository.Registration(ctx)); err != nil {
				return nil, err
			}
		}
		if roleSet["participant"] {
			if err := appendItems(s.repository.AthleteRanks(ctx, userID)); err != nil {
				return nil, err
			}
		}
		if roleSet["coach"] {
			if err := appendItems(s.repository.CoachAthleteRanks(ctx, userID)); err != nil {
				return nil, err
			}
		}
		if roleSet["judge"] {
			if err := appendItems(s.repository.JudgeCategories(ctx, userID)); err != nil {
				return nil, err
			}
			if err := appendItems(s.repository.JudgeAssignments(ctx, userID)); err != nil {
				return nil, err
			}
		}
	}
	items = unique(items)
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].DueAt.Equal(items[j].DueAt) {
			if items[i].Type == items[j].Type {
				return items[i].ID < items[j].ID
			}
			return items[i].Type < items[j].Type
		}
		return items[i].DueAt.Before(items[j].DueAt)
	})
	return items, nil
}

func (s *Service) ListReports(ctx context.Context) ([]Report, error) {
	items, err := s.repository.ListReports(ctx)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = make([]Report, 0)
	}
	return items, nil
}

func (s *Service) CreateReport(ctx context.Context, actorID int64, request CreateReportRequest) (Report, error) {
	name := strings.TrimSpace(request.Name)
	status := strings.TrimSpace(request.Status)
	if status == "" {
		status = "planned"
	}
	deadlineType := strings.TrimSpace(request.Type)
	if deadlineType == "" {
		deadlineType = "report"
	}
	dueAt, err := parseTime(request.DueAt)
	if actorID <= 0 || err != nil || name == "" || utf8.RuneCountInString(name) > 250 || !validStatus(status) || !validType(deadlineType) || !validGovernorDate(deadlineType, dueAt) {
		return Report{}, ErrInvalidInput
	}
	return s.repository.CreateReport(ctx, actorID, name, dueAt, status, deadlineType)
}

func (s *Service) UpdateReport(ctx context.Context, id int64, request UpdateReportRequest) (Report, error) {
	if id <= 0 || (request.Name == nil && request.DueAt == nil && request.Status == nil && request.Type == nil) {
		return Report{}, ErrInvalidInput
	}
	report, err := s.repository.GetReport(ctx, id)
	if err != nil {
		return Report{}, err
	}
	if request.Name != nil {
		report.Name = strings.TrimSpace(*request.Name)
	}
	if request.Status != nil {
		report.Status = strings.TrimSpace(*request.Status)
	}
	if request.Type != nil {
		report.Type = strings.TrimSpace(*request.Type)
	}
	if request.DueAt != nil {
		report.DueAt, err = parseTime(*request.DueAt)
		if err != nil {
			return Report{}, ErrInvalidInput
		}
	}
	if report.Name == "" || utf8.RuneCountInString(report.Name) > 250 || !validStatus(report.Status) || !validType(report.Type) || !validGovernorDate(report.Type, report.DueAt) {
		return Report{}, ErrInvalidInput
	}
	return s.repository.UpdateReport(ctx, report)
}

func parseTime(value string) (time.Time, error) {
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}
	return time.Parse("2006-01-02", value)
}

func validStatus(value string) bool {
	return value == "planned" || value == "completed" || value == "cancelled"
}

func validType(value string) bool { return value == "report" || value == "governor_award" }
func validGovernorDate(kind string, dueAt time.Time) bool {
	return kind != "governor_award" || (dueAt.Month() == time.February && dueAt.Day() <= 1)
}

func unique(items []Item) []Item {
	result := make([]Item, 0, len(items))
	seen := make(map[string]bool, len(items))
	for _, item := range items {
		key := item.Type + ":" + strconv.FormatInt(item.ID, 10)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, item)
	}
	return result
}
