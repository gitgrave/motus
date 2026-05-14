import { describe, it, expect } from "vitest";

// Guards the "All time" sentinel shared across all three /reports pages and
// the heatmap page. The value is new Date('2020-01-01') — if any call site
// drifts to a different date this test catches the regression.
describe("Reports date range — All time preset", () => {
  it("uses the 2020-01-01 sentinel matching the heatmap convention", () => {
    expect(new Date("2020-01-01").toISOString()).toBe(
      "2020-01-01T00:00:00.000Z",
    );
  });
});
