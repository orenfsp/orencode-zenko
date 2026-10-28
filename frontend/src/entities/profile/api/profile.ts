import { api, type ApiEnvelope } from "@/shared/api/client";
import type { OwnProfile, PublicProfile, UpdateProfilePayload, UpdateProfileResult } from "../model/types";
export const profileApi = {
  me: async () => (await api.get<ApiEnvelope<OwnProfile>>("/api/v1/profile")).data,
  public: async (userId: number) => (await api.get<ApiEnvelope<PublicProfile>>(`/api/v1/profiles/${userId}`, { auth: false })).data,
  update: async (body: UpdateProfilePayload) => (await api.patch<ApiEnvelope<UpdateProfileResult>>("/api/v1/profile", body)).data,
};
