// Public API for map feature

export { LocateButton } from "./components/locate-button";
export { MapControls } from "./components/map-controls";
export type { MapRendererProps } from "./components/map-renderer";
export { MapRenderer } from "./components/map-renderer";
export { MAP_CONFIG, WORLD_CITIES } from "./constants/map-config";
export type {
  GeolocationError,
  GeolocationPosition,
} from "./hooks/use-geolocation";
export { useGeolocation } from "./hooks/use-geolocation";
export { useMapControls } from "./hooks/use-map-controls";
export { useMapInstance } from "./hooks/use-map-instance";
export { MAP_STYLE, OPENFREEMAP_STYLES, TILE_SOURCES } from "./lib/config";
export {
  addLocationMarker,
  formatAccuracy,
  getZoomFromAccuracy,
  isLocationStale,
} from "./lib/location-utils";
export {
  controlButtonStyles,
  controlPanelStyles,
  mapContainerStyles,
} from "./lib/styles";

// Overlay system exports
export { OverlayProvider, useMapOverlays } from "./overlays";
