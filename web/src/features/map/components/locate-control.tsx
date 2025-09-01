import type { Map } from "maplibre-gl";
import { useEffect, useRef } from "react";
import { useGeolocation } from "../hooks/use-geolocation";
import { LocationLayer } from "../lib/location-layer";
import { getZoomFromAccuracy } from "../lib/location-utils";
import { controlButtonStyles } from "../lib/styles";

interface LocateControlProps {
  map: Map;
  onLocationFound?: (position: {
    latitude: number;
    longitude: number;
    accuracy: number;
  }) => void;
}

const CrosshairIcon = () => (
  <svg
    width="20"
    height="20"
    viewBox="0 0 20 20"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    role="img"
    aria-label="Crosshair icon"
  >
    <title>Location crosshair</title>
    <circle cx="10" cy="10" r="8" />
    <line x1="10" y1="2" x2="10" y2="6" />
    <line x1="10" y1="14" x2="10" y2="18" />
    <line x1="2" y1="10" x2="6" y2="10" />
    <line x1="14" y1="10" x2="18" y2="10" />
  </svg>
);

const SpinnerIcon = () => (
  <svg
    width="20"
    height="20"
    viewBox="0 0 20 20"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    style={{ animation: "spin 1s linear infinite" }}
    role="img"
    aria-label="Loading spinner"
  >
    <title>Loading location</title>
    <circle cx="10" cy="10" r="8" strokeDasharray="40" strokeDashoffset="10" />
  </svg>
);

export function LocateControl({ map, onLocationFound }: LocateControlProps) {
  const { position, isLocating, error, requestLocation, clearError } =
    useGeolocation();
  const locationLayerRef = useRef<LocationLayer | null>(null);

  useEffect(() => {
    if (position && map) {
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

      if (onLocationFound) {
        onLocationFound(position);
      }
    }

    return () => {
      if (locationLayerRef.current) {
        locationLayerRef.current.removeLocation();
        locationLayerRef.current = null;
      }
    };
  }, [position, map, onLocationFound]);

  const handleClick = () => {
    if (error) {
      clearError();
    }
    requestLocation();
  };

  return (
    <>
      <style>
        {`
          @keyframes spin {
            from { transform: rotate(0deg); }
            to { transform: rotate(360deg); }
          }
        `}
      </style>
      <button
        type="button"
        onClick={handleClick}
        disabled={isLocating}
        style={{
          ...controlButtonStyles,
          backgroundColor: "white",
          borderRadius: "4px",
          padding: "8px",
          boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
          cursor: isLocating ? "wait" : "pointer",
          color: error ? "#dc2626" : "#374151",
          position: "relative",
          zIndex: 10,
        }}
        title={
          error
            ? `Error: ${error.message}. Click to retry`
            : isLocating
              ? "Getting your location..."
              : "Locate me"
        }
        aria-label="Get current location"
      >
        {isLocating ? <SpinnerIcon /> : <CrosshairIcon />}
      </button>
    </>
  );
}
