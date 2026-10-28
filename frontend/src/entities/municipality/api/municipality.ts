import { api, type ApiEnvelope } from "@/shared/api/client";
import type {
  CreateMunicipalityPayload,
  Municipality,
  MunicipalityListQuery,
} from "../model/types";

export const municipalityApi = {
  list: async (query: MunicipalityListQuery = {}) =>
    (
      await api.get<ApiEnvelope<Municipality[]>>("/api/v1/municipalities", {
        auth: false,
        params: { ...query },
      })
    ).data,
  get: async (municipalityId: number) =>
    (
      await api.get<ApiEnvelope<Municipality>>(
        `/api/v1/municipalities/${municipalityId}`,
        { auth: false },
      )
    ).data,
  create: async (body: CreateMunicipalityPayload) =>
    (
      await api.post<ApiEnvelope<Municipality>>(
        "/api/v1/admin/municipalities",
        body,
        { permission: "municipalities.create" },
      )
    ).data,
};
