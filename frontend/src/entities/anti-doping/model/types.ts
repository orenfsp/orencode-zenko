export type AntiDopingCertificateStatus =
  | "manual_review"
  | "valid"
  | "expired"
  | "rejected";

export interface AntiDopingCertificate {
  id: number;
  user_id: number;
  file_id: number;
  status: AntiDopingCertificateStatus;
  completed_at?: string;
  valid_until?: string;
  reviewed_at?: string;
  rejection_reason?: string;
  created_at: string;
}

export interface ReviewAntiDopingCertificatePayload {
  status: Exclude<AntiDopingCertificateStatus, "manual_review">;
  completed_at?: string;
  rejection_reason?: string;
}
