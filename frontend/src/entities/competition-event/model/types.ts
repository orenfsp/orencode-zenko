export interface CompetitionEventReference {
  id: number;
  name: string;
}

export interface CompetitionEvent {
  id: number;
  competition_id: number;
  discipline: CompetitionEventReference;
  age_group: CompetitionEventReference | null;
  protocol_url: string | null;
}

export interface CreateCompetitionEventPayload {
  age_group_id?: number | null;
}

export interface UpdateCompetitionEventPayload {
  age_group_id: number | null;
}
