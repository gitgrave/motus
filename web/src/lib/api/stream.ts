import { getStoredAuthToken } from "$lib/auth-token-store";
import type { Position } from "$lib/types/api";

const BYTES_PER_POSITION_EST = 500;

export async function streamPositions(
  params: { deviceId?: number; from?: string; to?: string; limit?: number },
  onProgress: (estimated: number) => void,
): Promise<Position[]> {
  const query = new URLSearchParams();
  if (params.deviceId) query.set("deviceId", String(params.deviceId));
  if (params.from) query.set("from", params.from);
  if (params.to) query.set("to", params.to);
  if (params.limit) query.set("limit", String(params.limit));

  const headers: Record<string, string> = { "Content-Type": "application/json" };
  const authToken = await getStoredAuthToken();
  if (authToken) headers["X-Auth-Token"] = authToken;

  const response = await fetch(`/api/positions?${query}`, {
    credentials: "include",
    headers,
  });

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${await response.text()}`);
  }

  const reader = response.body!.getReader();
  const chunks: Uint8Array[] = [];
  let bytesReceived = 0;

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;
    chunks.push(value);
    bytesReceived += value.byteLength;
    onProgress(Math.floor(bytesReceived / BYTES_PER_POSITION_EST));
  }

  const buffer = new Uint8Array(bytesReceived);
  let offset = 0;
  for (const chunk of chunks) {
    buffer.set(chunk, offset);
    offset += chunk.byteLength;
  }

  const raw = JSON.parse(new TextDecoder().decode(buffer)) as Position[];
  return raw.map((pos) =>
    pos.speed != null ? { ...pos, speed: pos.speed * 1.852 } : pos,
  );
}
