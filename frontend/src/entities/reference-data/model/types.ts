export interface NamedReference {
  id: number;
  name: string;
}

export type Discipline = NamedReference;
export type CompetitionType = NamedReference;

export interface CompetitionLevel {
  id: number;
  code: string | null;
  name: string | null;
}

export interface AgeGroup {
  id: number;
  name: string;
  min_age: number | null;
  max_age: number | null;
}
