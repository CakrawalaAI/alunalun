import type { Map as MapLibreMap } from "maplibre-gl";
import { useState } from "react";
import { useMapControls } from "./use-map-controls";
import { useLocationTracking } from "./use-location-tracking";

/**
 * Orchestrator hook that manages all map state and logic.
 * This is the single source of truth for the map feature.
 * Returns props that can be injected into pure UI components.
 */
export function useMapOrchestrator() {
  // All state management happens here
  const [mapInstance, setMapInstance] = useState<MapLibreMap | null>(null);

  // Compose business logic hooks
  const controls = useMapControls(mapInstance);
  const location = useLocationTracking(mapInstance);

  // Return props for pure UI components
  return {
    // Props for MapCanvas component
    canvas: {
      onMapReady: setMapInstance,
      className: "h-full w-full",
    },

    // Props for ControlPanel component
    controls: {
      zoom: {
        current: controls.zoom.current,
        canZoomIn: controls.zoom.canZoomIn,
        canZoomOut: controls.zoom.canZoomOut,
        onZoomIn: controls.zoom.onZoomIn,
        onZoomOut: controls.zoom.onZoomOut,
      },
      orientation: {
        bearing: controls.orientation.bearing,
        pitch: controls.orientation.pitch,
        isRotated: controls.orientation.isRotated,
        onReset: controls.orientation.onReset,
      },
      location: {
        isLocating: location.isLocating,
        hasError: location.hasError,
        errorMessage: location.errorMessage,
        onLocate: location.requestLocation,
      },
      isReady: !!mapInstance,
    },

    // Expose map instance if needed for extensions
    mapInstance,
  };
}