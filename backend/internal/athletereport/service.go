package athletereport

import (
	"context"
	"time"

	"github.com/orenfsp/orencode-zenko.git/internal/auth"
	rankuser "github.com/orenfsp/orencode-zenko.git/internal/rank_user"
	"github.com/orenfsp/orencode-zenko.git/internal/result"
)

type ProfileRepository interface {
	GetProfile(ctx context.Context, id int64) (auth.Profile, error)
}

type ResultRepository interface {
	ListHistory(ctx context.Context, userID int64, disciplineID *int64) ([]result.HistoryItem, error)
}

type RankRepository interface {
	ListHistory(ctx context.Context, athleteUserID int64) ([]rankuser.RankHistory, error)
}

type Generator interface {
	Generate(data Data) ([]byte, error)
}

type Data struct {
	Profile   auth.Profile
	Results   []result.HistoryItem
	Ranks     []rankuser.RankHistory
	CreatedAt time.Time
}

type Service struct {
	profiles  ProfileRepository
	results   ResultRepository
	ranks     RankRepository
	generator Generator
}

func NewService(profiles ProfileRepository, results ResultRepository, ranks RankRepository, generator Generator) *Service {
	return &Service{profiles: profiles, results: results, ranks: ranks, generator: generator}
}

func (s *Service) Generate(ctx context.Context, userID int64) ([]byte, error) {
	profile, err := s.profiles.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	results, err := s.results.ListHistory(ctx, userID, nil)
	if err != nil {
		return nil, err
	}
	ranks, err := s.ranks.ListHistory(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.generator.Generate(Data{Profile: profile, Results: results, Ranks: ranks, CreatedAt: time.Now()})
}
