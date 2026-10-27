"use client";

import { createContext, useContext } from "react";

import { toast, Toaster } from "@/shared/ui/toast";

interface NotificationOptions {
  description?: string;
}

interface NotificationContextValue {
  success: (message: string, options?: NotificationOptions) => void;
  warning: (message: string, options?: NotificationOptions) => void;
  error: (message: string, options?: NotificationOptions) => void;
}

const notifications: NotificationContextValue = {
  success(message, options) {
    toast.add({
      type: "success",
      title: message,
      description: options?.description,
    });
  },
  warning(message, options) {
    toast.add({
      type: "warning",
      title: message,
      description: options?.description,
    });
  },
  error(message, options) {
    toast.add({
      type: "error",
      title: message,
      description: options?.description,
      priority: "high",
    });
  },
};

const NotificationContext = createContext<NotificationContextValue | null>(null);

export function NotificationProvider({ children }: { children: React.ReactNode }) {
  return (
    <NotificationContext.Provider value={notifications}>
      {children}
      <Toaster limit={3} timeout={5000} />
    </NotificationContext.Provider>
  );
}

export function useNotifications() {
  const context = useContext(NotificationContext);

  if (!context) {
    throw new Error("useNotifications must be used inside NotificationProvider");
  }

  return context;
}
