export interface AuditLogEntry {
  id: number;
  event_type: string;
  actor_user_id: number;
  entity_type: string;
  entity_id?: number;
  method: string;
  route: string;
  status: number;
  occurred_at: string;
}

export interface AuditLogQuery {
  limit?: number;
  offset?: number;
}
