export interface CompetitionEventReference {
  id: number;
  name: string;
}

export interface CompetitionEvent {
  id: number;
  competition_id: number;
  name: string;
  discipline: CompetitionEventReference;
  age_group: CompetitionEventReference | null;
  protocol_url: string | null;
}

export interface CreateCompetitionEventPayload {
  name: string;
  age_group_id?: number | null;
}

export interface UpdateCompetitionEventPayload {
  name?: string;
  age_group_id?: number | null;
}
