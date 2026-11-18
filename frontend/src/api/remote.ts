import { createURL, fetchJSON, fetchURL } from "./utils";

export interface RemoteServer {
  id: string;
  name: string;
  host: string;
  port: number;
  username: string;
  root: string;
  fingerprint: string;
  keyAuth: boolean;
}

export interface RemoteEntry {
  name: string;
  isDir: boolean;
  size: number;
  modified: string;
}

export const list = () => fetchJSON<RemoteServer[]>("/api/remote-servers");
export const add = (
  server: Omit<RemoteServer, "id"> & { credential: string }
) =>
  fetchJSON<RemoteServer>("/api/remote-servers", {
    method: "POST",
    body: JSON.stringify(server),
  });
export const remove = (id: string) =>
  fetchURL(`/api/remote-servers/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
export const files = (id: string, path: string, signal?: AbortSignal) =>
  fetchJSON<RemoteEntry[]>(
    `/api/remote-servers/${encodeURIComponent(id)}/files?${new URLSearchParams({ path })}`,
    { signal }
  );
export const upload = (id: string, path: string, file: File) =>
  fetchURL(
    `/api/remote-servers/${encodeURIComponent(id)}/files?${new URLSearchParams({ path, name: file.name })}`,
    { method: "POST", body: file }
  );
export const downloadURL = (id: string, path: string) =>
  createURL(`api/remote-servers/${id}/files`, { path, download: "true" });
