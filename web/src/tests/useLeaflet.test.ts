import { describe, it, expect } from "vitest";
import {
  LEAFLET_MARKER_ICON_URL,
  LEAFLET_MARKER_ICON_RETINA_URL,
  LEAFLET_MARKER_SHADOW_URL,
} from "$lib/composables/useLeaflet";

describe("useLeaflet marker icons", () => {
  it("resolves marker-icon.png to a bundled asset path", () => {
    expect(LEAFLET_MARKER_ICON_URL).toMatch(/\.png(\?.*)?$/);
    expect(LEAFLET_MARKER_ICON_URL).not.toContain("unpkg.com");
  });

  it("resolves marker-icon-2x.png to a bundled asset path", () => {
    expect(LEAFLET_MARKER_ICON_RETINA_URL).toMatch(/\.png(\?.*)?$/);
    expect(LEAFLET_MARKER_ICON_RETINA_URL).not.toContain("unpkg.com");
  });

  it("resolves marker-shadow.png to a bundled asset path", () => {
    expect(LEAFLET_MARKER_SHADOW_URL).toMatch(/\.png(\?.*)?$/);
    expect(LEAFLET_MARKER_SHADOW_URL).not.toContain("unpkg.com");
  });
});
