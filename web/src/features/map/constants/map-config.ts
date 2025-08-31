import type { LngLatBoundsLike } from 'maplibre-gl'

// Map configuration for worldwide view
export const MAP_CONFIG = {
  initialView: {
    center: [0, 20] as [number, number], // Center of world, slightly north for better land visibility
    zoom: 2, // Show entire world
    pitch: 0,
    bearing: 0
  },
  limits: {
    minZoom: 1, // Allow zooming out to see the whole world
    maxZoom: 18, // Allow detailed street-level zoom
    // No maxBounds - allow navigation anywhere in the world
  }
} as const

// Optional: Common world cities for quick navigation
// Can be customized or passed as props
export const WORLD_CITIES = {
  london: { lng: -0.1276, lat: 51.5074, name: 'London' },
  newYork: { lng: -74.0060, lat: 40.7128, name: 'New York' },
  tokyo: { lng: 139.6503, lat: 35.6762, name: 'Tokyo' },
  paris: { lng: 2.3522, lat: 48.8566, name: 'Paris' },
  sydney: { lng: 151.2093, lat: -33.8688, name: 'Sydney' },
  dubai: { lng: 55.2708, lat: 25.2048, name: 'Dubai' },
  singapore: { lng: 103.8198, lat: 1.3521, name: 'Singapore' },
  moscow: { lng: 37.6173, lat: 55.7558, name: 'Moscow' },
  cairo: { lng: 31.2357, lat: 30.0444, name: 'Cairo' },
  rio: { lng: -43.1729, lat: -22.9068, name: 'Rio de Janeiro' }
} as const