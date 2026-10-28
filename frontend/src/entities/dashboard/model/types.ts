export interface DashboardMetrics {
  athletes: number;
  competitions: number;
  participants: number;
  active_ranks: number;
}

export interface DashboardUpcomingCompetition {
  id: number;
  name: string;
  start_at: string;
  end_at: string;
  status: string;
  discipline_name: string;
}

export interface DashboardPendingApplication {
  id: number;
  submitted_at: string;
  athlete_user_id: number;
  athlete_name: string;
  competition_id?: number;
  competition_name?: string;
  rank_name?: string;
  discipline_name?: string;
  category_name?: string;
  application_type?: string;
}

export interface DashboardExpiringCredential {
  id: number;
  user_id: number;
  user_name: string;
  credential_name: string;
  discipline_name?: string;
  expires_at: string;
}

export interface DashboardDeadline {
  type: string;
  title: string;
  due_at: string;
  competition_id?: number;
}

export interface DashboardData {
  metrics: DashboardMetrics;
  upcoming_competitions: DashboardUpcomingCompetition[];
  pending_competition_applications: DashboardPendingApplication[];
  pending_rank_applications: DashboardPendingApplication[];
  pending_judge_category_applications: DashboardPendingApplication[];
  expiring_ranks: DashboardExpiringCredential[];
  expiring_judge_categories: DashboardExpiringCredential[];
  deadlines: DashboardDeadline[];
}
