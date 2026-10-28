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
