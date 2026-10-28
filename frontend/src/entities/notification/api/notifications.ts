import { api } from "@/shared/api/client";
import type { Notification, NotificationListQuery } from "../model/types";

export const notificationApi = {
  list: (query: NotificationListQuery = {}) =>
    api.get<Notification[]>("/api/v1/notifications", { params: { ...query } }),
  markRead: (notificationId: number) =>
    api.patch<void>(`/api/v1/notifications/${notificationId}/read`),
};
