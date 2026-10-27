package dashboard

import (
	"context"
	"sort"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Get(ctx context.Context) (Data, error) {
	data, err := s.repository.Load(ctx, defaultLimit, expiringDays)
	if err != nil {
		return Data{}, err
	}
	return normalize(data), nil
}

func normalize(data Data) Data {
	if data.UpcomingCompetitions == nil {
		data.UpcomingCompetitions = make([]UpcomingCompetition, 0)
	}
	if data.PendingCompetitionApplications == nil {
		data.PendingCompetitionApplications = make([]PendingApplication, 0)
	}
	if data.PendingRankApplications == nil {
		data.PendingRankApplications = make([]PendingApplication, 0)
	}
	if data.PendingJudgeApplications == nil {
		data.PendingJudgeApplications = make([]PendingApplication, 0)
	}
	if data.ExpiringRanks == nil {
		data.ExpiringRanks = make([]ExpiringCredential, 0)
	}
	if data.ExpiringJudgeCategories == nil {
		data.ExpiringJudgeCategories = make([]ExpiringCredential, 0)
	}
	if data.Deadlines == nil {
		data.Deadlines = make([]Deadline, 0)
	}
	sort.SliceStable(data.Deadlines, func(i, j int) bool {
		if data.Deadlines[i].DueAt.Equal(data.Deadlines[j].DueAt) {
			return data.Deadlines[i].Title < data.Deadlines[j].Title
		}
		return data.Deadlines[i].DueAt.Before(data.Deadlines[j].DueAt)
	})
	return data
}
