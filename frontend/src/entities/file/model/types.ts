export interface DownloadedFile {
  blob: Blob;
  filename: string;
  contentType: string;
  contentLength: number | null;
  internalPath: string | null;
}

export type UploadedFileStatus = "uploaded" | "attached" | "delete_pending";

export interface UploadedFile {
  id: number;
  purpose: string;
  storage_key: string;
  original_filename: string;
  content_type: string;
  size_bytes: number;
  sha256: string;
  status: UploadedFileStatus;
  created_at: string;
  updated_at: string;
}

export interface FileListQuery {
  limit?: number;
  offset?: number;
}
