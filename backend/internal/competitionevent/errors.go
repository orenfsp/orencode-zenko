package competitionevent

import "errors"

var (
	ErrInvalidID          = errors.New("invalid competition or event id")
	ErrInvalidCreate      = errors.New("invalid competition event data")
	ErrInvalidUpdate      = errors.New("invalid competition event update")
	ErrCompetitionMissing = errors.New("competition not found")
	ErrEventMissing       = errors.New("competition event not found")
	ErrAgeGroupMissing    = errors.New("age group not found")
	ErrNotDraft           = errors.New("competition is not a draft")
	ErrHasApplications    = errors.New("competition event has applications")
)
