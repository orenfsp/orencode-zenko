export interface ResultTeamSetup {
  team_id: number;
  application_ids: number[];
}

export interface ResultSettingsPayload {
  mode: string;
  task_count: number;
  teams?: ResultTeamSetup[];
  confirm_reset?: boolean;
}

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
}

export interface CompetitionResultTable {
  competition: ResultCompetition;
  event: ResultEvent;
  mode: string;
  task_count: number;
  status: string;
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
