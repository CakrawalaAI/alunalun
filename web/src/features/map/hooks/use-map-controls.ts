import type { Map } from "maplibre-gl";
import { useEffect } from "react";
import { MAP_CONFIG } from "../constants/map-config";

export function useMapControls(map: Map | null) {
  useEffect(() => {
    if (!map) {
      return;
    }

    // Fly to DPR Jakarta view
    const flyToJakarta = () => {
      map.flyTo({
        center: MAP_CONFIG.initialView.center,
        zoom: MAP_CONFIG.initialView.zoom,
        duration: 2000,
        essential: true,
      });
    };

    // Reset orientation (bearing and pitch)
    const resetOrientation = () => {
      map.easeTo({
        bearing: 0,
        pitch: 0,
        duration: 1000,
        essential: true,
      });
    };

    // Zoom controls
    const zoomIn = () => {
      const currentZoom = map.getZoom();
      if (currentZoom < MAP_CONFIG.limits.maxZoom) {
        map.zoomTo(Math.min(currentZoom + 1, MAP_CONFIG.limits.maxZoom), {
          duration: 300,
          essential: true,
        });
      }
    };

    const zoomOut = () => {
      const currentZoom = map.getZoom();
      if (currentZoom > MAP_CONFIG.limits.minZoom) {
        map.zoomTo(Math.max(currentZoom - 1, MAP_CONFIG.limits.minZoom), {
          duration: 300,
          essential: true,
        });
      }
    };

    // Pitch controls (3D tilt)
    const increasePitch = () => {
      const currentPitch = map.getPitch();
      if (currentPitch < 85) {
        map.easeTo({
          pitch: Math.min(currentPitch + 10, 85),
          duration: 300,
          essential: true,
        });
      }
    };

    const decreasePitch = () => {
      const currentPitch = map.getPitch();
      if (currentPitch > 0) {
        map.easeTo({
          pitch: Math.max(currentPitch - 10, 0),
          duration: 300,
          essential: true,
        });
      }
    };

    // Add keyboard shortcuts
    const handleKeyPress = (e: KeyboardEvent) => {
      // Ignore if user is typing in an input
      if (
        e.target instanceof HTMLInputElement ||
        e.target instanceof HTMLTextAreaElement
      ) {
        return;
      }

      switch (e.key) {
        case "h":
        case "H":
          // H for Jakarta (home city)
          flyToJakarta();
          break;
        case "r":
        case "R":
          // R for reset orientation
          resetOrientation();
          break;
        case "=":
        case "+":
          // + for zoom in
          zoomIn();
          break;
        case "-":
        case "_":
          // - for zoom out
          zoomOut();
          break;
        case "p":
        case "P":
          // P for increase pitch (tilt up)
          increasePitch();
          break;
        case "l":
        case "L":
          // L for decrease pitch (level/flat)
          decreasePitch();
          break;
        case "3":
          // 3 for 3D view (set pitch to 45°)
          map.easeTo({
            pitch: 45,
            duration: 500,
            essential: true,
          });
          break;
      }
    };

    window.addEventListener("keypress", handleKeyPress);

    return () => {
      window.removeEventListener("keypress", handleKeyPress);
    };
  }, [map]);
}
