// Core map components

export { CompassButton } from "./components/compass-button";
export { CoordinatesDisplay } from "./components/coordinates-display";
export { GestureOverlay } from "./components/gesture-overlay";
export { LocateButton } from "./components/locate-button";
export { LocateControl } from "./components/locate-control";
export type { MapBaseProps } from "./components/map-base";
export { MapBase } from "./components/map-base";
export { MapControls } from "./components/map-controls";
export type { MapRendererProps } from "./components/map-renderer";
// Legacy components (for backward compatibility)
export { MapRenderer } from "./components/map-renderer";
export { OrientControl } from "./components/orient-control";
export { PitchControl } from "./components/pitch-control";
// Control components
export { ZoomControl } from "./components/zoom-control";
export { ZoomControls } from "./components/zoom-controls";
// Configuration and constants
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
// Hooks
export { useMapInstance } from "./hooks/use-map-instance";
export { useMapOrientation } from "./hooks/use-map-orientation";
export type { MapViewState } from "./hooks/use-map-state";
export { useMapState } from "./hooks/use-map-state";
export { useMapZoom } from "./hooks/use-map-zoom";
export { MAP_STYLE, OPENFREEMAP_STYLES, TILE_SOURCES } from "./lib/config";

// Utilities
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

// Overlay system (legacy - will be removed later)
export { OverlayProvider, useMapOverlays } from "./overlays";

// Store exports (legacy - use useMapState instead)
export { useMapStore } from "./stores/use-map-store";
