import { api, type ApiEnvelope } from "@/shared/api/client";
import type { UpdateProfilePayload, UpdateProfileResult } from "../model/types";
export const profileApi = { update: async (body: UpdateProfilePayload) => (await api.patch<ApiEnvelope<UpdateProfileResult>>("/api/v1/profile", body)).data };
