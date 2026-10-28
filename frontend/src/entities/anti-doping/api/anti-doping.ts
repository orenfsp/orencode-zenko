import { api, type ApiEnvelope } from "@/shared/api/client";
import type {
  AntiDopingCertificate,
  ReviewAntiDopingCertificatePayload,
} from "../model/types";

export const antiDopingApi = {
  upload: async (file: File) => {
    const body = new FormData();
    body.append("certificate_file", file);
    return (
      await api.post<ApiEnvelope<AntiDopingCertificate>>(
        "/api/v1/anti-doping/self",
        body,
      )
    ).data;
  },
  mine: async () =>
    (
      await api.get<ApiEnvelope<AntiDopingCertificate[]>>(
        "/api/v1/anti-doping/self",
      )
    ).data,
  review: async (
    certificateId: number,
    body: ReviewAntiDopingCertificatePayload,
  ) =>
    (
      await api.patch<ApiEnvelope<AntiDopingCertificate>>(
        `/api/v1/admin/anti-doping/${certificateId}`,
        body,
        { permission: "anti_doping.review" },
      )
    ).data,
};
