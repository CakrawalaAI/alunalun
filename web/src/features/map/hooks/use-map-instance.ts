import maplibregl, { type Map } from "maplibre-gl";
import { useEffect, useRef, useState } from "react";
import { logger } from "@/common/logger/logger";
import { MAP_CONFIG } from "../constants/map-config";
import { MAP_STYLE } from "../lib/config";

export function useMapInstance(container: string | HTMLElement | null) {
  const mapRef = useRef<Map | null>(null);
  const [isLoaded, setIsLoaded] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!container) {
      return;
    }

    try {
      // Initialize map with worldwide view
      // NOTE: You may see console warnings "Expected value to be of type number, but found null"
      // These are harmless warnings from MapLibre's Web Workers processing OpenFreeMap vector tiles
      // The warnings occur when tile data contains null values where the style expects numbers
      // They don't affect map functionality and can be safely ignored
      const map = new maplibregl.Map({
        container,
        style: MAP_STYLE,
        center: MAP_CONFIG.initialView.center,
        zoom: MAP_CONFIG.initialView.zoom,
        pitch: MAP_CONFIG.initialView.pitch,
        bearing: MAP_CONFIG.initialView.bearing,
        minZoom: MAP_CONFIG.limits.minZoom,
        maxZoom: MAP_CONFIG.limits.maxZoom,
        // No maxBounds - allow worldwide navigation
        attributionControl: { compact: true },
      });

      mapRef.current = map;

      // Add navigation controls
      map.addControl(new maplibregl.NavigationControl(), "top-left");

      // Add scale control
      map.addControl(
        new maplibregl.ScaleControl({
          maxWidth: 200,
          unit: "metric",
        }),
        "bottom-left",
      );

      // Handle missing sprites/icons gracefully
      map.on("styleimagemissing", (e) => {
        const id = e.id;

        // Create a simple placeholder for missing icons
        // This prevents console warnings for POI icons we don't use
        if (!map.hasImage(id)) {
          // Create a 1x1 transparent image as placeholder
          const placeholder = new ImageData(1, 1);
          map.addImage(id, placeholder, { pixelRatio: 1 });

          logger.debug(`Added placeholder for missing icon: ${id}`);
        }
      });

      // Handle map load
      map.on("load", () => {
        setIsLoaded(true);
        logger.debug("Map loaded successfully");
      });

      // Handle errors
      map.on("error", (e) => {
        logger.error("Map error:", e);
        setError(e.error?.message || "Map loading error");
      });

      // Cleanup
      return () => {
        map.remove();
        mapRef.current = null;
      };
    } catch (err) {
      logger.error("Failed to initialize map:", err);
      setError(err instanceof Error ? err.message : "Failed to initialize map");
    }
  }, [container]);

  return {
    map: mapRef.current,
    isLoaded,
    error,
  };
}
