package audit

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	ActorEmployee = "employee"
	ActorDevice   = "device"
	ActorSystem   = "system"

	EventHospitalEquipmentStatusChanged = "hospital_equipment.status_changed"
	EventHospitalOperatingStatusChanged = "hospital_operating.status_changed"
	EventHospitalOperationCreated       = "hospital_operating_operation.created"
	EventHospitalOperationUpdated       = "hospital_operating_operation.updated"
	EventHospitalOperationStatusChanged = "hospital_operating_operation.status_changed"
	EventDestinationResolved            = "emergency_call_destination.resolved"
	EventDestinationSelected            = "emergency_call_destination.selected"
	EventDestinationDispatched          = "emergency_call_destination.dispatcher_selected"
	EventChecklistVersionPublished      = "checklist_version.published"
	EventChecklistVersionRestored       = "checklist_version.restored"
	EventUserAccessStatusChanged        = "user.access_status_changed"
	EventUserRoleChanged                = "user.role_changed"
	EventUserProfileUpdated             = "user.profile_updated"
	EventUserPasswordReset              = "user.password_reset"
	EventDeviceDisabled                 = "device.disabled"
	EventDeviceDeleted                  = "device.deleted"
	EventUserDeleted                    = "user.deleted"
	EventGeoTrackingPolicyUpdated       = "geo_tracking_policy.updated"
)

type EventCreate struct {
	EventType       string
	ActorKind       string
	ActorUserID     *uuid.UUID
	ActorDeviceID   *uuid.UUID
	EntityType      string
	EntityID        uuid.UUID
	EmergencyCallID *uuid.UUID
	Payload         json.RawMessage
}

type Event struct {
	ID              uuid.UUID
	EventType       string
	ActorKind       string
	ActorUserID     *uuid.UUID
	ActorDeviceID   *uuid.UUID
	EntityType      string
	EntityID        uuid.UUID
	EmergencyCallID *uuid.UUID
	Payload         json.RawMessage
	OccurredAt      time.Time
}
