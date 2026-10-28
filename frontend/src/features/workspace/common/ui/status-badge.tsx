import { Badge } from "@/shared/ui/badge";

const statusLabels: Record<string, string> = {
  pending: "На рассмотрении",
  approved: "Одобрено",
  rejected: "Отклонено",
  active: "Активно",
  draft: "Черновик",
  under_review: "На проверке",
  published: "Опубликовано",
  registration_open: "Регистрация открыта",
  registration_closed: "Регистрация закрыта",
  ongoing: "Идёт сейчас",
  results_pending: "Ожидаются результаты",
  completed: "Завершено",
  cancelled: "Отменено",
  confirmed: "Подтверждено",
  ended: "Завершено",
  manual_review: "На проверке",
  valid: "Действует",
  expired: "Истёк",
  uploaded: "Загружен",
  attached: "Прикреплён",
  delete_pending: "Удаляется",
  withdrawn: "Отозвано",
  planned: "Запланировано",
};

export function StatusBadge({ status }: { status?: string | null }) {
  const normalized = status ?? "unknown";
  const variant =
    normalized === "rejected" || normalized === "cancelled"
      ? "destructive"
      : normalized === "approved" || normalized === "valid" || normalized === "active"
        ? "default"
        : "secondary";

  return <Badge variant={variant}>{statusLabels[normalized] ?? normalized}</Badge>;
}
