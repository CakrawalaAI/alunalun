import type { Map } from "maplibre-gl";
import { useEffect, useState } from "react";

/**
 * Hook to track the current center coordinates of the map
 * Updates in real-time as the map moves
 *
 * @param map - MapLibre GL map instance
 * @returns Current center coordinates
 */
export function useMapCenter(map: Map | null) {
  const [center, setCenter] = useState<{ lng: number; lat: number } | null>(
    null,
  );

  useEffect(() => {
    if (!map) {
      return;
    }

    const updateCenter = () => {
      const mapCenter = map.getCenter();
      setCenter({
        lng: mapCenter.lng,
        lat: mapCenter.lat,
      });
    };

    // Initial center
    updateCenter();

    // Update on move
    map.on("move", updateCenter);

    return () => {
      map.off("move", updateCenter);
    };
  }, [map]);

  return center;
}

/**
 * Utility function to recenter the map to DPR Jakarta
 * Can be called from anywhere with a map instance
 */
export function recenterToDPR(map: Map) {
  map.flyTo({
    center: [106.8001307, -6.2099592], // DPR RI coordinates
    zoom: 15, // Neighborhood-level zoom
    duration: 2000,
    essential: true,
  });
}

/**
 * Format coordinates for display
 */
export function formatCoordinates(
  lat: number,
  lng: number,
  precision = 6,
): string {
  return `${lat.toFixed(precision)}, ${lng.toFixed(precision)}`;
}

/**
 * Format coordinates with cardinal directions
 */
export function formatCoordinatesWithCardinal(
  lat: number,
  lng: number,
  precision = 6,
): string {
  const latLabel = lat >= 0 ? "N" : "S";
  const lngLabel = lng >= 0 ? "E" : "W";

  return `${Math.abs(lat).toFixed(precision)}°${latLabel} ${Math.abs(lng).toFixed(precision)}°${lngLabel}`;
}

/**
 * Copy coordinates to clipboard
 */
export async function copyCoordinatesToClipboard(
  lat: number,
  lng: number,
): Promise<boolean> {
  try {
    const coords = formatCoordinates(lat, lng, 7);
    await navigator.clipboard.writeText(coords);
    return true;
  } catch (error) {
    console.error("Failed to copy coordinates:", error);
    return false;
  }
}
