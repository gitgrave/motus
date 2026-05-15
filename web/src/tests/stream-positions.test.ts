import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("$app/environment", () => ({ browser: true }));
vi.mock("$lib/auth-token-store", () => ({ getStoredAuthToken: vi.fn().mockResolvedValue(null) }));

import { streamPositions } from "$lib/api/stream";

function makePositionJSON(id: number) {
  return {
    id,
    deviceId: 1,
    fixTime: "2024-01-01T00:00:00Z",
    valid: true,
    latitude: 48.0 + id * 0.001,
    longitude: 11.0 + id * 0.001,
    speed: 10,
  };
}

function encodeBody(positions: object[]): ReadableStream<Uint8Array> {
  const json = JSON.stringify(positions);
  const bytes = new TextEncoder().encode(json);
  return new ReadableStream({
    start(controller) {
      controller.enqueue(bytes);
      controller.close();
    },
  });
}

describe("streamPositions", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("returns all positions and normalizes speed from knots to km/h", async () => {
    const raw = [makePositionJSON(1), makePositionJSON(2)];
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        body: encodeBody(raw),
      }),
    );

    const progress: number[] = [];
    const result = await streamPositions({ deviceId: 1 }, (n) => progress.push(n));

    expect(result).toHaveLength(2);
    expect(result[0].speed).toBeCloseTo(10 * 1.852, 5);
    expect(result[1].speed).toBeCloseTo(10 * 1.852, 5);
    expect(progress.length).toBeGreaterThan(0);
  });

  it("sends deviceId, from, to as query params", async () => {
    const raw = [makePositionJSON(1)];
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      body: encodeBody(raw),
    });
    vi.stubGlobal("fetch", mockFetch);

    await streamPositions(
      { deviceId: 42, from: "2024-01-01T00:00:00Z", to: "2024-12-31T23:59:59Z" },
      () => {},
    );

    const url: string = mockFetch.mock.calls[0][0];
    expect(url).toContain("deviceId=42");
    expect(url).toContain("from=");
    expect(url).toContain("to=");
  });

  it("throws on non-ok response", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 502,
        text: vi.fn().mockResolvedValue("Bad Gateway"),
      }),
    );

    await expect(streamPositions({}, () => {})).rejects.toThrow("HTTP 502");
  });

  it("leaves null speed as null without crashing", async () => {
    const raw = [{ ...makePositionJSON(1), speed: null }];
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        body: encodeBody(raw),
      }),
    );

    const result = await streamPositions({}, () => {});
    expect(result[0].speed).toBeNull();
  });
});
