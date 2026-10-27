import { api, type ApiEnvelope } from "@/shared/api/client";
import type {
  AgeGroup,
  CompetitionLevel,
  CompetitionType,
  Discipline,
} from "../model/types";

export const referenceDataApi = {
  disciplines: async () =>
    (
      await api.get<ApiEnvelope<Discipline[]>>("/api/v1/disciplines", {
        auth: false,
      })
    ).data,
  competitionTypes: async () =>
    (
      await api.get<ApiEnvelope<CompetitionType[]>>(
        "/api/v1/competition-types",
        { auth: false },
      )
    ).data,
  competitionLevels: async () =>
    (
      await api.get<ApiEnvelope<CompetitionLevel[]>>(
        "/api/v1/competition-levels",
        { auth: false },
      )
    ).data,
  ageGroups: async () =>
    (
      await api.get<ApiEnvelope<AgeGroup[]>>("/api/v1/age-groups", {
        auth: false,
      })
    ).data,
};
