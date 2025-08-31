// Re-export from new config structure for backward compatibility
import { getStyleUrl, MAP_STYLES } from "@/features/map/config/styles";

// OpenFreeMap style URLs - now imported from config/styles.ts
export const OPENFREEMAP_STYLES = {
  liberty: MAP_STYLES.liberty.url,
  bright: MAP_STYLES.bright.url,
  positron: MAP_STYLES.positron.url,
} as const;

// Default style - now uses configuration
export const MAP_STYLE = getStyleUrl("liberty");

// Legacy tile providers (kept for reference)
export const TILE_SOURCES = {
  cartodb: {
    light: "https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png",
    dark: "https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png",
    voyager:
      "https://{s}.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}{r}.png",
  },
  osm: {
    standard: "https://tile.openstreetmap.org/{z}/{x}/{y}.png",
  },
} as const;
