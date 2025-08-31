// OpenFreeMap style URLs - these are complete style definitions, not tile URLs
export const OPENFREEMAP_STYLES = {
  liberty: 'https://tiles.openfreemap.org/styles/liberty',
  bright: 'https://tiles.openfreemap.org/styles/bright',
  positron: 'https://tiles.openfreemap.org/styles/positron'
} as const

// Use OpenFreeMap's liberty style directly
// This is a complete style.json that includes sources, layers, etc.
export const MAP_STYLE = OPENFREEMAP_STYLES.liberty

// Legacy tile providers (kept for reference)
export const TILE_SOURCES = {
  cartodb: {
    light: 'https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png',
    dark: 'https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png',
    voyager: 'https://{s}.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}{r}.png'
  },
  osm: {
    standard: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png'
  }
} as const