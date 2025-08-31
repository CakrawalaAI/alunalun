import type { Map } from "maplibre-gl";
import { useEffect, useRef } from "react";
import { useGeolocation } from "../hooks/use-geolocation";
import { getZoomFromAccuracy } from "../lib/location-utils";
import { LocationLayer } from "../lib/location-layer";

interface LocateButtonProps {
  map: Map;
  onLocationFound?: (position: {
    latitude: number;
    longitude: number;
    accuracy: number;
  }) => void;
}

const CrosshairIcon = ({ className }: { className?: string }) => (
  <svg
    width="20"
    height="20"
    viewBox="0 0 20 20"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    className={className}
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

export function LocateButton({ map, onLocationFound }: LocateButtonProps) {
  const { position, isLocating, error, requestLocation, clearError } =
    useGeolocation();
  const locationLayerRef = useRef<LocationLayer | null>(null);

  useEffect(() => {
    if (position && map) {
      // Fly to user location
      map.flyTo({
        center: [position.longitude, position.latitude],
        zoom: getZoomFromAccuracy(position.accuracy),
        duration: 1500,
        essential: true,
      });

      // Create or update location layer (WebGL)
      if (!locationLayerRef.current) {
        locationLayerRef.current = new LocationLayer(map);
      }
      locationLayerRef.current.updateLocation(
        position.latitude,
        position.longitude,
        position.accuracy
      );

      // Call callback if provided
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

  const getButtonStyle = () => {
    const baseStyle = {
      display: "flex",
      alignItems: "center",
      justifyContent: "center",
      width: "40px",
      height: "40px",
      padding: "0",
      border: "1px solid #ddd",
      borderRadius: "4px",
      backgroundColor: "white",
      cursor: isLocating ? "wait" : "pointer",
      fontSize: "14px",
      transition: "all 0.2s",
      color: error ? "#dc2626" : "#374151",
    };

    return baseStyle;
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
        style={getButtonStyle()}
        title={
          error
            ? `Error: ${error.message}. Click to retry`
            : isLocating
              ? "Getting your location..."
              : "Locate me"
        }
        onMouseEnter={(e) => {
          if (!(isLocating || error)) {
            e.currentTarget.style.backgroundColor = "#f5f5f5";
          }
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.backgroundColor = "white";
        }}
      >
        {isLocating ? (
          <SpinnerIcon />
        ) : (
          <CrosshairIcon className={error ? "error" : ""} />
        )}
      </button>
    </>
  );
}
