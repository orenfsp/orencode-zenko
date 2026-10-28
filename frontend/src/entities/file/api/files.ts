import { api, type ApiEnvelope } from "@/shared/api/client";
import type { DownloadedFile, FileListQuery, UploadedFile } from "../model/types";

function getFilename(disposition: string | null, fallback: string) {
  if (!disposition) return fallback;
  const encoded = disposition.match(/filename\*=UTF-8''([^;]+)/i)?.[1];
  if (encoded) {
    try { return decodeURIComponent(encoded); } catch { return encoded; }
  }
  return disposition.match(/filename="([^"]+)"/i)?.[1] ?? fallback;
}

export const fileApi = {
  list: async (query: FileListQuery = {}) =>
    (
      await api.get<ApiEnvelope<UploadedFile[]>>("/api/v1/admin/files", {
        permission: "files.read",
        params: { ...query },
      })
    ).data,
  async download(fileId: number): Promise<DownloadedFile> {
    const response = await api.get<Response>(`/api/v1/files/${fileId}/download`, { responseType: "response" });
    const blob = await response.blob();
    const contentLength = Number(response.headers.get("Content-Length"));
    return {
      blob,
      filename: getFilename(response.headers.get("Content-Disposition"), `document-${fileId}`),
      contentType: response.headers.get("Content-Type") || blob.type,
      contentLength: Number.isFinite(contentLength) && contentLength > 0 ? contentLength : null,
      internalPath: response.headers.get("X-File-Path"),
    };
  },
};
