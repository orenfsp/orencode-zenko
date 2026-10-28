import { api, type ApiEnvelope } from "@/shared/api/client";
import type { Rank } from "../model/types";

export const rankApi = {
  list: async () =>
    (
      await api.get<ApiEnvelope<Rank[]>>("/api/v1/rank/", { auth: false })
    ).data,
  get: async (rankId: number) =>
    (
      await api.get<ApiEnvelope<Rank>>(`/api/v1/rank/${rankId}`, {
        auth: false,
      })
    ).data,
};
