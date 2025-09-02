import type { Map as MapLibreMap } from "maplibre-gl";
import { useCallback } from "react";
import { logger } from "@/common/logger/logger";
import { useMapZoom } from "./use-map-zoom";
import { useMapOrientation } from "./use-map-orientation";

/**
 * Hook that encapsulates all map control logic.
 * Combines zoom and orientation controls with debug logging.
 * Returns a clean interface for UI components.
 */
export function useMapControls(map: MapLibreMap | null) {
  // Compose primitive hooks
  const zoom = useMapZoom(map!);
  const orientation = useMapOrientation(map!);

  // Encapsulate all handlers with logging
  const handleZoomIn = useCallback(() => {
    logger.debug("Zoom in clicked", { 
      currentZoom: zoom.currentZoom, 
      canZoomIn: zoom.canZoomIn 
    });
    zoom.zoomIn();
  }, [zoom]);

  const handleZoomOut = useCallback(() => {
    logger.debug("Zoom out clicked", { 
      currentZoom: zoom.currentZoom, 
      canZoomOut: zoom.canZoomOut 
    });
    zoom.zoomOut();
  }, [zoom]);

  const handleResetOrientation = useCallback(() => {
    logger.debug("Reset orientation clicked", { 
      bearing: orientation.bearing, 
      pitch: orientation.pitch 
    });
    orientation.resetOrientation();
  }, [orientation]);

  // Return clean interface for components
  return {
    zoom: {
      current: zoom.currentZoom,
      canZoomIn: zoom.canZoomIn,
      canZoomOut: zoom.canZoomOut,
      onZoomIn: handleZoomIn,
      onZoomOut: handleZoomOut,
    },
    orientation: {
      bearing: orientation.bearing,
      pitch: orientation.pitch,
      isRotated: orientation.bearing !== 0 || orientation.pitch !== 0,
      onReset: handleResetOrientation,
    },
  };
}
