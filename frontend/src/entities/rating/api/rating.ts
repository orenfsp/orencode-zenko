import { api, ApiError, type ApiEnvelope } from "@/shared/api/client";
import type {
  RatingPage,
  RatingQuery,
  RatingSelection,
  RatingSelectionQuery,
  SaveRatingSelectionPayload,
} from "../model/types";

export const ratingApi = {
  list: async (query: RatingQuery = {}) =>
    (await api.get<ApiEnvelope<RatingPage>>("/api/v1/admin/rating", {
      permission: "rating.read",
      params: { ...query },
    })).data,
  selection: async (query: RatingSelectionQuery) => {
    try {
      return (
        await api.get<ApiEnvelope<RatingSelection>>(
          "/api/v1/admin/rating/selections",
          { permission: "rating.read", params: { ...query } },
        )
      ).data;
    } catch (cause) {
      if (cause instanceof ApiError && cause.status === 404) return null;
      throw cause;
    }
  },
  saveSelection: async (payload: SaveRatingSelectionPayload) =>
    (
      await api.put<ApiEnvelope<RatingSelection>>(
        "/api/v1/admin/rating/selections",
        payload,
        { permission: "rating.manage" },
      )
    ).data,
  lockSelection: async (selectionId: number) =>
    (
      await api.post<ApiEnvelope<RatingSelection>>(
        `/api/v1/admin/rating/selections/${selectionId}/lock`,
        undefined,
        { permission: "rating.manage" },
      )
    ).data,
};
