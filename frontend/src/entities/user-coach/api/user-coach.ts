import { api, type ApiEnvelope } from "@/shared/api/client";
import type {
  CreateUserCoachPayload,
  UserCoachListQuery,
  UserCoachRelation,
} from "../model/types";

export const userCoachApi = {
  create: async (body: CreateUserCoachPayload) =>
    (
      await api.post<ApiEnvelope<UserCoachRelation>>(
        "/api/v1/coach/user-coaches",
        body,
      )
    ).data,
  byCoach: async (query: UserCoachListQuery = {}) =>
    (
      await api.get<ApiEnvelope<UserCoachRelation[]>>(
        "/api/v1/coach/user-coaches",
        { params: { ...query } },
      )
    ).data,
  byParticipant: async (query: UserCoachListQuery = {}) =>
    (
      await api.get<ApiEnvelope<UserCoachRelation[]>>(
        "/api/v1/participant/coaches",
        { params: { ...query } },
      )
    ).data,
  confirm: async (relationId: number) =>
    (
      await api.patch<ApiEnvelope<UserCoachRelation>>(
        `/api/v1/user-coaches/${relationId}/confirm`,
      )
    ).data,
  end: async (relationId: number) =>
    (
      await api.patch<ApiEnvelope<UserCoachRelation>>(
        `/api/v1/user-coaches/${relationId}/end`,
      )
    ).data,
};
