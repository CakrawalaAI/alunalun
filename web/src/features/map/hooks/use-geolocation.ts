import { useCallback, useState } from "react";

export interface GeolocationPosition {
  latitude: number;
  longitude: number;
  accuracy: number;
  timestamp: number;
}

export interface GeolocationError {
  code: "PERMISSION_DENIED" | "POSITION_UNAVAILABLE" | "TIMEOUT" | "UNKNOWN";
  message: string;
}

interface UseGeolocationReturn {
  position: GeolocationPosition | null;
  isLocating: boolean;
  error: GeolocationError | null;
  requestLocation: () => void;
  clearError: () => void;
}

export function useGeolocation(): UseGeolocationReturn {
  const [position, setPosition] = useState<GeolocationPosition | null>(null);
  const [isLocating, setIsLocating] = useState(false);
  const [error, setError] = useState<GeolocationError | null>(null);

  const requestLocation = useCallback(() => {
    // Check if geolocation is supported
    if (!navigator.geolocation) {
      setError({
        code: "POSITION_UNAVAILABLE",
        message: "Geolocation is not supported by your browser",
      });
      return;
    }

    // Clear any previous error
    setError(null);
    setIsLocating(true);

    // Request position with high accuracy
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        setPosition({
          latitude: pos.coords.latitude,
          longitude: pos.coords.longitude,
          accuracy: pos.coords.accuracy,
          timestamp: pos.timestamp,
        });
        setIsLocating(false);
        setError(null);
      },
      (err) => {
        let errorCode: GeolocationError["code"] = "UNKNOWN";
        let errorMessage = "Failed to get your location";

        switch (err.code) {
          case err.PERMISSION_DENIED:
            errorCode = "PERMISSION_DENIED";
            errorMessage = "Location permission denied";
            break;
          case err.POSITION_UNAVAILABLE:
            errorCode = "POSITION_UNAVAILABLE";
            errorMessage = "Location information unavailable";
            break;
          case err.TIMEOUT:
            errorCode = "TIMEOUT";
            errorMessage = "Location request timed out";
            break;
        }

        setError({
          code: errorCode,
          message: errorMessage,
        });
        setIsLocating(false);
      },
      {
        enableHighAccuracy: true,
        timeout: 10000, // 10 seconds
        maximumAge: 0, // Don't use cached position
      },
    );
  }, []);

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  return {
    position,
    isLocating,
    error,
    requestLocation,
    clearError,
  };
}
