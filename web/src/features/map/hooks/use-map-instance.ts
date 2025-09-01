import maplibregl, { type Map } from "maplibre-gl";
import { useEffect, useRef, useState } from "react";
import { logger } from "@/common/logger/logger";
import { getStyleUrl } from "../config/styles";
import { MAP_CONFIG } from "../constants/map-config";
import { useMapState } from "./use-map-state";

export function useMapInstance(container: string | HTMLElement | null) {
  const mapRef = useRef<Map | null>(null);
  const [isLoaded, setIsLoaded] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { loadSavedState } = useMapState();

  useEffect(() => {
    if (!container) {
      return;
    }

    try {
      // Load saved state or use default config
      const savedState = loadSavedState();
      const initialView = savedState || MAP_CONFIG.initialView;

      // Get style URL from configuration
      const styleUrl = getStyleUrl(
        MAP_CONFIG.style.preset,
        MAP_CONFIG.style.customUrl,
      );

      // Initialize map with comprehensive configuration
      // NOTE: You may see console warnings "Expected value to be of type number, but found null"
      // These are harmless warnings from MapLibre's Web Workers processing OpenFreeMap vector tiles
      // The warnings occur when tile data contains null values where the style expects numbers
      // They don't affect map functionality and can be safely ignored
      const map = new maplibregl.Map({
        container,
        style: styleUrl,
        center: initialView.center,
        zoom: initialView.zoom,
        pitch: initialView.pitch,
        bearing: initialView.bearing,
        minZoom: MAP_CONFIG.limits.minZoom,
        maxZoom: MAP_CONFIG.limits.maxZoom,
        maxPitch: MAP_CONFIG.limits.maxPitch,
        // Interaction settings from config
        scrollZoom: MAP_CONFIG.interactions.scrollZoom,
        dragPan: MAP_CONFIG.interactions.dragPan,
        dragRotate: MAP_CONFIG.interactions.dragRotate,
        doubleClickZoom: MAP_CONFIG.interactions.doubleClickZoom,
        touchZoomRotate: MAP_CONFIG.interactions.touchZoomRotate,
        keyboard: MAP_CONFIG.interactions.keyboard,
        // Attribution control from config
        attributionControl:
          MAP_CONFIG.controls.attribution === true
            ? undefined
            : (MAP_CONFIG.controls.attribution as
                | false
                | maplibregl.AttributionControlOptions),
        // No maxBounds - allow worldwide navigation
      });

      mapRef.current = map;

      // Add controls based on configuration
      if (MAP_CONFIG.controls.navigation) {
        // Note: We have custom navigation controls, but can enable built-in too
        // For now, keeping custom implementation for better UI
      }

      if (MAP_CONFIG.controls.scale) {
        map.addControl(
          new maplibregl.ScaleControl({
            maxWidth: 200,
            unit: "metric",
          }),
          "bottom-left",
        );
      }

      // Handle missing sprites/icons gracefully
      // OpenFreeMap Liberty style references some icons that aren't in the Maki sprite set
      // We map these to available alternatives or create placeholders
      map.on("styleimagemissing", (e) => {
        const id = e.id;

        // Map missing icons to available alternatives in OSM Liberty sprites
        const iconMappings: Record<string, string> = {
          // Map missing icons to closest available alternatives
          office: "commercial", // Office buildings → commercial
          atm: "bank", // ATM → bank icon
          sports_centre: "pitch", // Sports centre → sports pitch
          swimming_pool: "swimming", // Pool → swimming icon
          gate: "entrance", // Gate → entrance
          lift_gate: "entrance", // Lift gate → entrance
          bollard: "barrier", // Bollard → barrier
          brownfield: "industrial", // Brownfield → industrial
        };

        // Check if we have a mapping for this icon
        const mappedIcon = iconMappings[id];
        if (mappedIcon && map.hasImage(mappedIcon)) {
          // Copy the mapped icon to the requested id
          const imageData = map.getImage(mappedIcon);
          if (imageData) {
            map.addImage(id, imageData.data, {
              pixelRatio: imageData.pixelRatio || 1,
              sdf: imageData.sdf,
            });
            logger.debug(`Mapped missing icon '${id}' to '${mappedIcon}'`);
            return;
          }
        }

        // Only create placeholder if no mapping exists
        if (!map.hasImage(id)) {
          // Check if this is a known missing icon we can safely ignore
          const ignorableIcons: string[] = [
            // Add any icons here that are safe to ignore
            // e.g., custom icons that aren't critical for map display
          ];

          if (ignorableIcons.some((pattern) => id.includes(pattern))) {
            return; // Skip creating placeholder for ignorable icons
          }

          // Create a subtle placeholder for unmapped icons
          // This prevents console errors while maintaining visual consistency
          const size = 16; // Standard icon size
          const placeholder = new ImageData(size, size);

          // Optional: Create a visible placeholder (gray circle)
          // Uncomment below to make missing icons visible for debugging
          /*
          const data = placeholder.data;
          const centerX = size / 2;
          const centerY = size / 2;
          const radius = size / 3;
          
          for (let y = 0; y < size; y++) {
            for (let x = 0; x < size; x++) {
              const distance = Math.sqrt(Math.pow(x - centerX, 2) + Math.pow(y - centerY, 2));
              if (distance < radius) {
                const index = (y * size + x) * 4;
                data[index] = 128;     // R
                data[index + 1] = 128; // G
                data[index + 2] = 128; // B
                data[index + 3] = 64;  // A (semi-transparent)
              }
            }
          }
          */

          map.addImage(id, placeholder, { pixelRatio: 1 });

          // Log for monitoring which icons might need custom sprites
          logger.info(
            `Created placeholder for missing icon: ${id}. Consider adding custom sprite if needed.`,
          );
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
