export interface Notification {
  id: number;
  title: string;
  message: string;
  is_read: boolean;
  created_at: string;
}

export interface NotificationListQuery {
  limit?: number;
  offset?: number;
}
