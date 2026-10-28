import { api, type ApiEnvelope } from "@/shared/api/client";
import type { RatingPage, RatingQuery } from "../model/types";

export const ratingApi = {
  list: async (query: RatingQuery = {}) =>
    (await api.get<ApiEnvelope<RatingPage>>("/api/v1/admin/rating", {
      permission: "rating.read",
      params: { ...query },
    })).data,
};
