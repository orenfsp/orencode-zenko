package result

import "errors"

var (
	ErrInvalidRequest        = errors.New("invalid competition result request")
	ErrEventNotFound         = errors.New("competition event not found")
	ErrNotConfigured         = errors.New("event results are not configured")
	ErrAlreadyApproved       = errors.New("event results are already approved")
	ErrApplicationsChanged   = errors.New("approved applications changed after result setup")
	ErrStaleResults          = errors.New("results changed concurrently")
	ErrResetConfirmation     = errors.New("configuration reset requires confirmation")
	ErrUnsupportedDiscipline = errors.New("discipline result rules are not implemented")
	ErrInvalidTeamSetup      = errors.New("invalid team result setup")
	ErrEntryNotFound         = errors.New("result entry not found")
	ErrProtocolNotFound      = errors.New("event protocol not found")
	ErrProtocolNotApproved   = errors.New("event results are not approved")
	ErrInvalidPDF            = errors.New("invalid PDF file")
	ErrProtocolTooLarge      = errors.New("protocol file too large")
	ErrTemplateNotFound      = errors.New("result template not found")
	ErrTemplateInactive      = errors.New("result template is inactive")
	ErrTemplateDiscipline    = errors.New("result template belongs to another discipline")
	ErrWrongScoringMode      = errors.New("wrong result scoring mode")
)
