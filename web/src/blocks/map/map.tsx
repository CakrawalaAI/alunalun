import type { Map as MapLibreMap } from "maplibre-gl";
import { useEffect, useRef, useState } from "react";
import { cn } from "@/common/lib/utils";
import { logger } from "@/common/logger/logger";
import { 
  MapBase, 
  useGeolocation, 
  useMapOrientation,
  useMapZoom 
} from "@/features/map";
import { LocationLayer } from "@/features/map/lib/location-layer";
import { getZoomFromAccuracy } from "@/features/map/lib/location-utils";

export function Map() {
  const [map, setMap] = useState<MapLibreMap | null>(null);
  const locationLayerRef = useRef<LocationLayer | null>(null);
  
  // Hooks for map controls
  const { currentZoom, canZoomIn, canZoomOut, zoomIn, zoomOut } = useMapZoom(map!);
  const { bearing, pitch, resetOrientation } = useMapOrientation(map!);
  const { position, isLocating, error, requestLocation, clearError } = useGeolocation();
  
  const isMapRotated = bearing !== 0 || pitch !== 0;

  // Debug logging
  useEffect(() => {
    if (map) {
      logger.debug("Map initialized, setting up controls", {
        zoom: currentZoom,
        bearing,
        pitch,
        canZoomIn,
        canZoomOut
      });
    }
  }, [map, currentZoom, bearing, pitch, canZoomIn, canZoomOut]);

  // Handle location updates
  useEffect(() => {
    if (position && map) {
      logger.debug("Location update received", position);
      map.flyTo({
        center: [position.longitude, position.latitude],
        zoom: getZoomFromAccuracy(position.accuracy),
        duration: 1500,
        essential: true,
      });

      if (!locationLayerRef.current) {
        locationLayerRef.current = new LocationLayer(map);
      }
      locationLayerRef.current.updateLocation(
        position.latitude,
        position.longitude,
        position.accuracy,
      );
    }

    return () => {
      if (locationLayerRef.current) {
        locationLayerRef.current.removeLocation();
        locationLayerRef.current = null;
      }
    };
  }, [position, map]);

  const handleLocate = () => {
    logger.debug("Location button clicked", { error, isLocating });
    if (error) {
      clearError();
    }
    requestLocation();
  };

  const handleZoomIn = () => {
    logger.debug("Zoom in clicked", { currentZoom, canZoomIn });
    zoomIn();
  };

  const handleZoomOut = () => {
    logger.debug("Zoom out clicked", { currentZoom, canZoomOut });
    zoomOut();
  };

  const handleResetOrientation = () => {
    logger.debug("Reset orientation clicked", { bearing, pitch });
    resetOrientation();
  };

  return (
    <div className="relative h-full w-full overflow-hidden">
      {/* Map layer */}
      <MapBase onMapReady={setMap} className="h-full w-full" />

      {/* Controls overlay - using very high z-index and explicit pointer-events */}
      {map && (
        <div 
          className="absolute top-4 right-4 z-[999999]"
          style={{ 
            position: 'absolute',
            top: '1rem',
            right: '1rem',
            zIndex: 999999,
            pointerEvents: 'auto'
          }}
          onMouseEnter={() => logger.debug("Mouse entered control panel")}
          onMouseLeave={() => logger.debug("Mouse left control panel")}
        >
          <div className={cn(
            "flex flex-col",
            "bg-white/95 backdrop-blur-md",
            "rounded-xl shadow-2xl",
            "border border-gray-200/60",
            "overflow-hidden",
            "isolate" // Create new stacking context
          )}>
            {/* Zoom In Button */}
            <button
              type="button"
              onClick={handleZoomIn}
              onMouseDown={() => logger.debug("Zoom in mouse down")}
              disabled={!canZoomIn}
              className={cn(
                "relative w-11 h-11",
                "flex items-center justify-center",
                "bg-white hover:bg-gray-50 active:bg-gray-100",
                "transition-colors duration-150",
                "focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-inset",
                "disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:bg-white",
                "select-none cursor-pointer"
              )}
              style={{ pointerEvents: 'auto', position: 'relative', zIndex: 1 }}
              title="Zoom in"
              aria-label="Zoom in"
            >
              <svg 
                width="20" 
                height="20" 
                viewBox="0 0 20 20" 
                className="text-gray-700 pointer-events-none"
              >
                <path
                  d="M10 5v10M5 10h10"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                />
              </svg>
            </button>

            {/* Zoom Level Display */}
            <div className={cn(
              "h-8 px-3",
              "flex items-center justify-center",
              "text-xs font-semibold text-gray-600",
              "bg-gray-50/50 border-y border-gray-200/50",
              "select-none"
            )}>
              {currentZoom.toFixed(1)}
            </div>

            {/* Zoom Out Button */}
            <button
              type="button"
              onClick={handleZoomOut}
              onMouseDown={() => logger.debug("Zoom out mouse down")}
              disabled={!canZoomOut}
              className={cn(
                "relative w-11 h-11",
                "flex items-center justify-center",
                "bg-white hover:bg-gray-50 active:bg-gray-100",
                "transition-colors duration-150",
                "focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-inset",
                "disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:bg-white",
                "select-none cursor-pointer"
              )}
              style={{ pointerEvents: 'auto', position: 'relative', zIndex: 1 }}
              title="Zoom out"
              aria-label="Zoom out"
            >
              <svg 
                width="20" 
                height="20" 
                viewBox="0 0 20 20" 
                className="text-gray-700 pointer-events-none"
              >
                <path
                  d="M5 10h10"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                />
              </svg>
            </button>

            {/* Divider */}
            <div className="h-px bg-gray-200/50" />

            {/* Locate Button */}
            <button
              type="button"
              onClick={handleLocate}
              onMouseDown={() => logger.debug("Locate mouse down")}
              disabled={isLocating}
              className={cn(
                "relative w-11 h-11",
                "flex items-center justify-center",
                "bg-white hover:bg-gray-50 active:bg-gray-100",
                "transition-all duration-150",
                "focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-inset",
                isLocating && "animate-pulse cursor-wait",
                error ? "text-red-500 hover:text-red-600" : "text-gray-700",
                "select-none cursor-pointer"
              )}
              style={{ pointerEvents: 'auto', position: 'relative', zIndex: 1 }}
              title={
                error
                  ? `Error: ${error.message}. Click to retry`
                  : isLocating
                    ? "Getting your location..."
                    : "Locate me"
              }
              aria-label="Get current location"
            >
              {isLocating ? (
                <svg
                  width="20"
                  height="20"
                  viewBox="0 0 20 20"
                  className="animate-spin pointer-events-none"
                >
                  <circle
                    cx="10"
                    cy="10"
                    r="8"
                    stroke="currentColor"
                    strokeWidth="2"
                    fill="none"
                    strokeDasharray="40"
                    strokeDashoffset="10"
                  />
                </svg>
              ) : (
                <svg 
                  width="20" 
                  height="20" 
                  viewBox="0 0 20 20"
                  className="pointer-events-none"
                >
                  <circle cx="10" cy="10" r="8" stroke="currentColor" strokeWidth="2" fill="none" />
                  <line x1="10" y1="2" x2="10" y2="6" stroke="currentColor" strokeWidth="2" />
                  <line x1="10" y1="14" x2="10" y2="18" stroke="currentColor" strokeWidth="2" />
                  <line x1="2" y1="10" x2="6" y2="10" stroke="currentColor" strokeWidth="2" />
                  <line x1="14" y1="10" x2="18" y2="10" stroke="currentColor" strokeWidth="2" />
                </svg>
              )}
            </button>

            {/* Divider */}
            <div className="h-px bg-gray-200/50" />

            {/* Reset Orientation Button */}
            <button
              type="button"
              onClick={handleResetOrientation}
              onMouseDown={() => logger.debug("Reset orientation mouse down")}
              className={cn(
                "relative w-11 h-11",
                "flex items-center justify-center",
                "bg-white hover:bg-gray-50 active:bg-gray-100",
                "transition-all duration-150",
                "focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-inset",
                !isMapRotated && "opacity-60",
                "select-none cursor-pointer",
                "rounded-b-xl" // Round bottom corners
              )}
              style={{ pointerEvents: 'auto', position: 'relative', zIndex: 1 }}
              title={`Reset orientation (bearing: ${bearing.toFixed(0)}°, pitch: ${pitch.toFixed(0)}°)`}
              aria-label="Reset map orientation"
            >
              <svg
                width="20"
                height="20"
                viewBox="0 0 20 20"
                className="text-gray-700 transition-transform duration-300 pointer-events-none"
                style={{ transform: `rotate(${-bearing}deg)` }}
              >
                <circle cx="10" cy="10" r="8" stroke="currentColor" strokeWidth="2" fill="none" />
                <path d="M10 6 L7 13 L10 11 L13 13 Z" fill="currentColor" stroke="none" />
              </svg>
            </button>
          </div>
        </div>
      )}
    </div>
  );
}