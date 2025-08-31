import type { Map } from "maplibre-gl";
import { useEffect, useState } from "react";
import { MAP_CONFIG } from "../constants/map-config";

interface ZoomState {
  currentZoom: number;
  canZoomIn: boolean;
  canZoomOut: boolean;
}

/**
 * Hook to track and control map zoom level
 */
export function useMapZoom(map: Map | null) {
  const [zoomState, setZoomState] = useState<ZoomState>({
    currentZoom: MAP_CONFIG.initialView.zoom,
    canZoomIn: true,
    canZoomOut: true,
  });

  useEffect(() => {
    if (!map) {
      return;
    }

    const updateZoomState = () => {
      const zoom = map.getZoom();
      setZoomState({
        currentZoom: zoom,
        canZoomIn: zoom < MAP_CONFIG.limits.maxZoom,
        canZoomOut: zoom > MAP_CONFIG.limits.minZoom,
      });
    };

    // Initial state
    updateZoomState();

    // Listen to zoom changes
    map.on("zoom", updateZoomState);
    map.on("zoomend", updateZoomState);

    return () => {
      map.off("zoom", updateZoomState);
      map.off("zoomend", updateZoomState);
    };
  }, [map]);

  const zoomIn = () => {
    if (!(map && zoomState.canZoomIn)) {
      return;
    }

    map.zoomTo(Math.min(zoomState.currentZoom + 1, MAP_CONFIG.limits.maxZoom), {
      duration: 300,
      essential: true,
    });
  };

  const zoomOut = () => {
    if (!(map && zoomState.canZoomOut)) {
      return;
    }

    map.zoomTo(Math.max(zoomState.currentZoom - 1, MAP_CONFIG.limits.minZoom), {
      duration: 300,
      essential: true,
    });
  };

  return {
    currentZoom: zoomState.currentZoom,
    canZoomIn: zoomState.canZoomIn,
    canZoomOut: zoomState.canZoomOut,
    zoomIn,
    zoomOut,
  };
}
