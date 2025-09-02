import type { Map as MapLibreMap } from "maplibre-gl";
import { useCallback, useEffect, useRef } from "react";
import { logger } from "@/common/logger/logger";
import { useGeolocation } from "./use-geolocation";
import { LocationLayer } from "../lib/location-layer";
import { getZoomFromAccuracy } from "../lib/location-utils";

/**
 * Hook that manages location tracking and visualization.
 * Encapsulates geolocation API, location layer management, and map positioning.
 * Returns a clean interface for UI components.
 */
export function useLocationTracking(map: MapLibreMap | null) {
  const locationLayerRef = useRef<LocationLayer | null>(null);
  const { position, isLocating, error, requestLocation, clearError } = useGeolocation();

  // Handle location updates
  useEffect(() => {
    if (position && map) {
      logger.debug("Location update received", position);
      
      // Fly to location
      map.flyTo({
        center: [position.longitude, position.latitude],
        zoom: getZoomFromAccuracy(position.accuracy),
        duration: 1500,
        essential: true,
      });

      // Update or create location layer
      if (!locationLayerRef.current) {
        locationLayerRef.current = new LocationLayer(map);
      }
      locationLayerRef.current.updateLocation(
        position.latitude,
        position.longitude,
        position.accuracy,
      );
    }

    // Cleanup on unmount
    return () => {
      if (locationLayerRef.current) {
        locationLayerRef.current.removeLocation();
        locationLayerRef.current = null;
      }
    };
  }, [position, map]);

  // Handle locate request with error clearing
  const handleLocate = useCallback(() => {
    logger.debug("Location request initiated", { error, isLocating });
    if (error) {
      clearError();
    }
    requestLocation();
  }, [error, isLocating, clearError, requestLocation]);

  // Return clean interface for components
  return {
    isLocating,
    hasError: !!error,
    errorMessage: error?.message,
    requestLocation: handleLocate,
    position,
  };
}