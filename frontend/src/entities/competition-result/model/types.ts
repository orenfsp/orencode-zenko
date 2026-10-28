export interface ResultTeamSetup {
  team_id: number;
  application_ids: number[];
}

export interface ResultSettingsPayload {
  mode: string;
  scoring_mode?: "algorithmic" | "manual";
  task_count: number;
  teams?: ResultTeamSetup[];
  template_id?: number;
  confirm_reset?: boolean;
}

export interface ResultTemplateField {
  key: string;
  label: string;
  type: "number" | "text";
  required: boolean;
}

export interface ResultTemplate {
  id: number;
  discipline_id: number;
  name: string;
  fields: ResultTemplateField[];
  is_active: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface CreateResultTemplatePayload {
  discipline_id: number;
  name: string;
  fields: ResultTemplateField[];
}

export type UpdateResultTemplatePayload = Partial<Omit<CreateResultTemplatePayload, "discipline_id">> & { is_active?: boolean };

export interface ResultTaskPayload {
  task_number: number;
  status: string;
  solve_time_deciseconds: number | null;
  negative_attempts: number;
}

export interface ResultEntryPayload {
  entry_id: number;
  result_status: string;
  tasks: ResultTaskPayload[];
}

export interface SaveResultsPayload {
  entries: ResultEntryPayload[];
}

export interface ManualResultEntryPayload {
  entry_id: number;
  result_status: string;
  place: number | null;
  custom_values: Record<string, string | number | null>;
}

export interface SaveManualResultsPayload {
  revision: number;
  entries: ManualResultEntryPayload[];
}

export interface ResultCompetition {
  id: number;
  name: string;
}

export interface ResultEvent {
  id: number;
  discipline_id: number;
  discipline_name: string;
  age_group_name: string | null;
}

export interface ResultTeam {
  id: number;
  name: string;
}

export interface ResultParticipant {
  application_id: number;
  user_id: number;
  last_name: string;
  first_name: string;
  middle_name: string | null;
}

export interface ResultTask {
  task_number: number;
  status: string;
  solve_time_deciseconds: number | null;
  negative_attempts: number;
}

export interface ResultEntry {
  id: number;
  team: ResultTeam | null;
  participants: ResultParticipant[];
  result_status: string;
  tasks: ResultTask[];
  solved_count: number;
  total_time_deciseconds: number;
  last_solve_time_deciseconds: number;
  negative_attempts: number;
  place: number | null;
  custom_values?: Record<string, string | number | null>;
}

export interface ResultTemplateSnapshot {
  id: number;
  name: string;
  fields: ResultTemplateField[];
}

export interface CompetitionResultTable {
  competition: ResultCompetition;
  event: ResultEvent;
  mode: string;
  task_count: number;
  revision: number;
  status: string;
  scoring_mode: "algorithmic" | "manual";
  template_id: number | null;
  template_snapshot: ResultTemplateSnapshot | null;
  calculated_at: string | null;
  approved_at: string | null;
  protocol_url: string | null;
  entries: ResultEntry[];
}

export interface CompetitionResultApproval {
  status: string;
  approved_at: string;
  protocol_url: string;
}

export interface CompetitionResultHistoryItem {
  competition_id: number;
  competition_name: string;
  event_id: number;
  discipline_id: number;
  discipline_name: string;
  start_at: string;
  place: number | null;
  score: number | null;
  result_status: string | null;
  protocol_url: string | null;
}

export interface CompetitionProtocol {
  blob: Blob;
  filename: string;
}
