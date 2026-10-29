import { api, type ApiEnvelope } from "@/shared/api/client";
import type {
  OwnProfile,
  PublicProfile,
  UpdateProfilePayload,
  UpdateProfileResult,
} from "../model/types";

type ProfilePayload = Partial<OwnProfile> & {
  birth_date?: string;
};

function calculateAge(birthDate?: string) {
  if (!birthDate) return 0;
  const birth = new Date(`${birthDate.slice(0, 10)}T00:00:00Z`);
  if (Number.isNaN(birth.getTime())) return 0;

  const today = new Date();
  let age = today.getUTCFullYear() - birth.getUTCFullYear();
  const birthdayNotReached =
    today.getUTCMonth() < birth.getUTCMonth() ||
    (today.getUTCMonth() === birth.getUTCMonth() &&
      today.getUTCDate() < birth.getUTCDate());
  if (birthdayNotReached) age -= 1;
  return Math.max(0, age);
}

function normalizeProfile<T extends ProfilePayload>(profile: T) {
  return {
    ...profile,
    age: typeof profile.age === "number" ? profile.age : calculateAge(profile.birth_date),
    organization: profile.organization ?? null,
    municipality: profile.municipality ?? null,
    roles: Array.isArray(profile.roles) ? profile.roles : [],
    disciplines: Array.isArray(profile.disciplines) ? profile.disciplines : [],
    ranks: Array.isArray(profile.ranks) ? profile.ranks : [],
    judge_category: profile.judge_category ?? null,
  };
}

export const profileApi = {
  report: async (athleteUserId?: number) => {
    const response = await api.get<Response>("/api/v1/profile/report.pdf", {
      params: athleteUserId ? { athlete_user_id: athleteUserId } : undefined,
      responseType: "response",
    });
    return {
      blob: await response.blob(),
      filename: response.headers.get("Content-Disposition")?.match(/filename="([^"]+)"/)?.[1] ?? `athlete-report${athleteUserId ? `-${athleteUserId}` : ""}.pdf`,
    };
  },
  me: async () =>
    normalizeProfile(
      (await api.get<ApiEnvelope<ProfilePayload>>("/api/v1/profile")).data,
    ) as OwnProfile,
  public: async (userId: number) =>
    normalizeProfile(
      (
        await api.get<ApiEnvelope<ProfilePayload>>(`/api/v1/profiles/${userId}`, {
          auth: false,
        })
      ).data,
    ) as PublicProfile,
  update: async (body: UpdateProfilePayload) =>
    (await api.patch<ApiEnvelope<UpdateProfileResult>>("/api/v1/profile", body))
      .data,
};
