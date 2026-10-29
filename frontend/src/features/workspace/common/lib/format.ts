export function formatDate(value?: string | null, withTime = false) {
  if (!value) return "—";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  const includeYear = date.getFullYear() !== new Date().getFullYear();
  return new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "long",
    ...(includeYear ? { year: "numeric" as const } : {}),
    ...(withTime ? { hour: "2-digit", minute: "2-digit" } : {}),
  }).format(date);
}

export function fullName(...parts: Array<string | null | undefined>) {
  return parts.filter(Boolean).join(" ");
}

export function daysUntil(value?: string | null) {
  if (!value) return null;
  const expiresAt = new Date(value);
  if (Number.isNaN(expiresAt.getTime())) return null;
  return Math.ceil((expiresAt.getTime() - Date.now()) / 86_400_000);
}

export function isExpiringSoon(value?: string | null, withinDays = 90) {
  const days = daysUntil(value);
  return days !== null && days >= 0 && days <= withinDays;
}

export function expiryDescription(value?: string | null) {
  const days = daysUntil(value);
  if (days === null) return null;
  if (days < 0) return "срок действия истёк";
  if (days === 0) return "срок действия истекает сегодня";
  if (days === 1) return "срок действия истекает завтра";
  return `до истечения ${days} дн.`;
}
