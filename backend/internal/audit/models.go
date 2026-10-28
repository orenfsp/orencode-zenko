package audit

import "time"

const (
	EventLogin                         = "auth.login"
	EventLogout                        = "auth.logout"
	EventRoleCreated                   = "role.created"
	EventUserRoleAssigned              = "user.role.assigned"
	EventProfileUpdated                = "user.profile.updated"
	EventCompetitionCreated            = "competition.created"
	EventCompetitionUpdated            = "competition.updated"
	EventCompetitionSubmitted          = "competition.submitted"
	EventCompetitionPublished          = "competition.published"
	EventCompetitionCancelled          = "competition.cancelled"
	EventCompetitionReturned           = "competition.returned_to_draft"
	EventCompetitionEventCreated       = "competition_event.created"
	EventCompetitionEventUpdated       = "competition_event.updated"
	EventCompetitionEventDeleted       = "competition_event.deleted"
	EventCompetitionJudgeAssigned      = "competition_judge.assigned"
	EventCompetitionJudgeRemoved       = "competition_judge.removed"
	EventCompetitionApplicationApprove = "competition_application.approved"
	EventCompetitionApplicationReject  = "competition_application.rejected"
	EventJudgeCategoryReviewed         = "judge_category_application.reviewed"
	EventRankApplicationReviewed       = "rank_application.reviewed"
	EventAntiDopingReviewed            = "anti_doping.reviewed"
	EventParentalConsentReviewed       = "parental_consent.reviewed"
	EventResultConfigured              = "competition_result.configured"
	EventResultUpdated                 = "competition_result.updated"
	EventResultCalculated              = "competition_result.calculated"
	EventResultApproved                = "competition_result.approved"
	EventProtocolUploaded              = "competition_protocol.uploaded"
	EventProtocolDeleted               = "competition_protocol.deleted"
	EventRatingSelectionSaved          = "rating_selection.saved"
	EventRatingSelectionLocked         = "rating_selection.locked"
)

type EventCreate struct {
	EventType   string
	ActorUserID *int64
	EntityType  string
	EntityID    *int64
	Method      string
	Route       string
	Status      int16
}

type Event struct {
	ID          int64
	EventType   string
	ActorUserID *int64
	EntityType  string
	EntityID    *int64
	Method      string
	Route       string
	Status      int16
	OccurredAt  time.Time
}
