import { api } from "@/shared/api/client";
export const fileApi = { download: (fileId: number) => api.get<Blob>(`/api/v1/files/${fileId}/download`, { responseType: "blob" }) };
