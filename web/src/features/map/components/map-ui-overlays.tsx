import type { Map } from "maplibre-gl";
import { useEffect } from "react";
import { useMapOverlays } from "../overlays";
import { CoordinatesDisplay } from "./coordinates-display";
import { GestureOverlay } from "./gesture-overlay";
import { MapControls } from "./map-controls";

interface MapUIOverlaysProps {
  map: Map;
  showControls?: boolean;
}

/**
 * Manages all built-in UI overlays for the map
 * Uses the overlay system to ensure proper stacking and z-index management
 */
export function MapUIOverlays({
  map,
  showControls = true,
}: MapUIOverlaysProps) {
  const { addOverlay, removeOverlay } = useMapOverlays();

  useEffect(() => {
    const overlayIds: string[] = [];

    // Add map controls overlay (highest priority)
    if (showControls) {
      const controlsId = addOverlay(<MapControls map={map} />, { zIndex: 30 });
      overlayIds.push(controlsId);
    }

    // Add coordinates display overlay
    const coordsId = addOverlay(<CoordinatesDisplay map={map} />, {
      zIndex: 20,
    });
    overlayIds.push(coordsId);

    // Add gesture feedback overlay
    const gestureId = addOverlay(<GestureOverlay map={map} />, { zIndex: 10 });
    overlayIds.push(gestureId);

    // Cleanup on unmount
    return () => {
      overlayIds.forEach((id) => removeOverlay(id));
    };
  }, [map, showControls, addOverlay, removeOverlay]);

  return null;
}
