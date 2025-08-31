// Map configuration for Jakarta, Indonesia focus
export const MAP_CONFIG = {
  initialView: {
    center: [106.8033, -6.2195] as [number, number], // Senayan, Jakarta, Indonesia
    zoom: 12, // City-level zoom
    pitch: 0,
    bearing: 0,
  },
  limits: {
    minZoom: 1, // Allow zooming out to see the whole world
    maxZoom: 18, // Allow detailed street-level zoom
    // No maxBounds - allow navigation anywhere in the world
  },
} as const;

// Indonesian cities for quick navigation
export const INDONESIAN_CITIES = {
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
