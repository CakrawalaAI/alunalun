import type { Map } from "maplibre-gl";
import { useEffect } from "react";
import { MAP_CONFIG } from "../constants/map-config";

export function useMapControls(map: Map | null) {
  useEffect(() => {
    if (!map) {
      return;
    }

    // Fly to Jakarta view
    const flyToJakarta = () => {
      map.flyTo({
        center: MAP_CONFIG.initialView.center,
        zoom: MAP_CONFIG.initialView.zoom,
        duration: 2000,
        essential: true,
      });
    };

    // Add keyboard shortcuts
    const handleKeyPress = (e: KeyboardEvent) => {
      if (e.key === "h" || e.key === "H") {
        // H for Jakarta (home city)
        flyToJakarta();
      }
    };

    window.addEventListener("keypress", handleKeyPress);

    return () => {
      window.removeEventListener("keypress", handleKeyPress);
    };
  }, [map]);
}
