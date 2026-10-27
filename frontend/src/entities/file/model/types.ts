export interface DownloadedFile {
  blob: Blob;
  filename: string;
  contentType: string;
  contentLength: number | null;
  internalPath: string | null;
}
