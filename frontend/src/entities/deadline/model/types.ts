export interface DeadlineItem {
  type: string;
  id: number;
  title: string;
  due_at: string;
  status: string;
  user_id?: number;
  user_name?: string;
  competition_id?: number;
}

export interface ReportDeadline {
  id: number;
  name: string;
  due_at: string;
  status: string;
  type: "report" | "governor_award";
  created_by_user_id: number;
  created_at: string;
  updated_at: string;
}

export interface CreateReportDeadlinePayload {
  name: string;
  due_at: string;
  status?: string;
  type?: "report" | "governor_award";
}

export type UpdateReportDeadlinePayload = Partial<CreateReportDeadlinePayload>;
