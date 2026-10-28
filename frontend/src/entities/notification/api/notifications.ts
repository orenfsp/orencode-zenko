import { api, type ApiEnvelope } from "@/shared/api/client";
import type { Notification, NotificationListQuery } from "../model/types";

export const notificationApi = {
  list: async (query: NotificationListQuery = {}) =>
    (
      await api.get<ApiEnvelope<Notification[]>>("/api/v1/notifications", {
        params: { ...query },
      })
    ).data,
  markRead: (notificationId: number) =>
    api.patch<void>(`/api/v1/notifications/${notificationId}/read`),
};
