// Core map components
export { MapCanvas } from "./components/map-canvas";
export type { MapCanvasProps } from "./components/map-canvas";
export { ControlPanel } from "./components/control-panel";
export type { ControlPanelProps } from "./components/control-panel";
export { MapBase } from "./components/map-base";
export type { MapBaseProps } from "./components/map-base";

// Configuration and constants
export { MAP_CONFIG, WORLD_CITIES } from "./constants/map-config";

// Core hooks
export { useMapOrchestrator } from "./hooks/use-map-orchestrator";

// Supporting hooks (exported for extensibility)
export { useMapControls } from "./hooks/use-map-controls";
export { useLocationTracking } from "./hooks/use-location-tracking";
export { useMapInstance } from "./hooks/use-map-instance";
export { useMapOrientation } from "./hooks/use-map-orientation";
export { useMapState } from "./hooks/use-map-state";
export type { MapViewState } from "./hooks/use-map-state";
export { useMapZoom } from "./hooks/use-map-zoom";
export { useGeolocation } from "./hooks/use-geolocation";
export type {
  GeolocationError,
  GeolocationPosition,
} from "./hooks/use-geolocation";

// Utilities (keeping only what's actually used)
export { getZoomFromAccuracy } from "./lib/location-utils";
