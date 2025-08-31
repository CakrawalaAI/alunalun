// Public API for map feature

export { CompassButton } from "./components/compass-button";
export { CoordinatesDisplay } from "./components/coordinates-display";
export { GestureOverlay } from "./components/gesture-overlay";
export { LocateButton } from "./components/locate-button";
export { MapControls } from "./components/map-controls";
export type { MapRendererProps } from "./components/map-renderer";
export { MapRenderer } from "./components/map-renderer";
export { PitchControl } from "./components/pitch-control";
export { ZoomControls } from "./components/zoom-controls";
export { MAP_CONFIG, WORLD_CITIES } from "./constants/map-config";
export type {
  GeolocationError,
  GeolocationPosition,
} from "./hooks/use-geolocation";
export { useGeolocation } from "./hooks/use-geolocation";
export {
  copyCoordinatesToClipboard,
  formatCoordinates,
  formatCoordinatesWithCardinal,
  recenterToDPR,
  useMapCenter,
} from "./hooks/use-map-center";
export { useMapControls } from "./hooks/use-map-controls";
export { useMapInstance } from "./hooks/use-map-instance";
export { useMapOrientation } from "./hooks/use-map-orientation";
export { useMapZoom } from "./hooks/use-map-zoom";
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
export type { MapViewState } from "./stores/use-map-store";
// Store exports for map state management
export { useMapStore } from "./stores/use-map-store";
