export interface Municipality {
  id: number;
  name: string;
}

export interface CreateMunicipalityPayload {
  name: string;
}

export interface MunicipalityListQuery {
  limit?: number;
  offset?: number;
}
