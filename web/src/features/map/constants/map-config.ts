import {
  DEFAULT_MAP_OPTIONS,
  MAP_PRESETS,
} from "@/features/map/config/map-options";

// Map configuration for Jakarta, Indonesia focus
// Now using the comprehensive configuration structure
export const MAP_CONFIG = {
  initialView: DEFAULT_MAP_OPTIONS.viewport,
  limits: {
    minZoom: DEFAULT_MAP_OPTIONS.performance.minZoom,
    maxZoom: DEFAULT_MAP_OPTIONS.performance.maxZoom,
    maxPitch: DEFAULT_MAP_OPTIONS.performance.maxPitch,
    // No maxBounds - allow navigation anywhere in the world
  },
  interactions: DEFAULT_MAP_OPTIONS.interactions,
  controls: DEFAULT_MAP_OPTIONS.controls,
  style: DEFAULT_MAP_OPTIONS.style,
} as const;

// Export presets for easy switching
export { MAP_PRESETS };

// Indonesian cities and landmarks for quick navigation
export const INDONESIAN_CITIES = {
  dpr: { lng: 106.8001307, lat: -6.2099592, name: "DPR RI" },
  jakarta: { lng: 106.8033, lat: -6.2195, name: "Jakarta" },
  bandung: { lng: 107.6098, lat: -6.9147, name: "Bandung" },
  surabaya: { lng: 112.7688, lat: -7.2504, name: "Surabaya" },
  yogyakarta: { lng: 110.3695, lat: -7.7956, name: "Yogyakarta" },
  semarang: { lng: 110.4203, lat: -6.9932, name: "Semarang" },
  medan: { lng: 98.6722, lat: 3.5897, name: "Medan" },
  makassar: { lng: 119.4327, lat: -5.1477, name: "Makassar" },
} as const;

// Export alias for backward compatibility
export const WORLD_CITIES = INDONESIAN_CITIES;
