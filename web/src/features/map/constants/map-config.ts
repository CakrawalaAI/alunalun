// Map configuration for Jakarta, Indonesia focus
export const MAP_CONFIG = {
  initialView: {
    center: [106.8001307, -6.2099592] as [number, number], // DPR RI, Jakarta
    zoom: 15, // Neighborhood-level view (good detail of DPR complex)
    bearing: 0, // North-facing
    pitch: 0, // Top-down view
  },
  limits: {
    minZoom: 1,
    maxZoom: 18,
    maxPitch: 85, // Allow dramatic 3D viewing angles
    // No maxBounds - allow navigation anywhere in the world
  },
  interactions: {
    scrollZoom: true,
    dragPan: true,
    dragRotate: true,
    doubleClickZoom: true,
    touchZoomRotate: true,
    keyboard: true, // Built-in keyboard navigation
  },
  controls: {
    navigation: true,
    scale: true,
    attribution: { compact: true },
  },
  style: {
    preset: "liberty" as const, // Modern, colorful OpenFreeMap style
  },
} as const;

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
