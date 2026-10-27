package competition

import "errors"

var (
	ErrInvalidJudgeRequest     = errors.New("invalid competition judge request")
	ErrJudgeCompetitionMissing = errors.New("competition for judge assignment not found")
	ErrJudgeCompetitionPrivate = errors.New("competition judges are not public")
	ErrJudgeCompetitionLocked  = errors.New("competition judges can only be changed for a draft")
	ErrJudgeUserMissing        = errors.New("judge user not found")
	ErrUserIsNotJudge          = errors.New("user does not have judge role")
	ErrJudgePositionMissing    = errors.New("judge position not found")
	ErrJudgeAssignmentExists   = errors.New("competition judge assignment already exists")
	ErrJudgeAssignmentMissing  = errors.New("competition judge assignment not found")
	ErrInvalidJudgeReference   = errors.New("invalid competition judge reference")
)
