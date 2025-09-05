import { useEffect, useRef, useState, useCallback } from "react";
import maplibregl from "maplibre-gl";

interface UseMapLibreOptions {
  container: HTMLDivElement | null;
  center?: [number, number];
  zoom?: number;
  style?: string;
}

interface MapState {
  map: maplibregl.Map | null;
  isLoading: boolean;
  isLoaded: boolean;
  error: string | null;
}

export function useMapLibre({
  container,
  center = [106.827, -6.175], // Jakarta
  zoom = 12,
  style = "https://tiles.openfreemap.org/styles/liberty",
}: UseMapLibreOptions): MapState {
  const mapRef = useRef<maplibregl.Map | null>(null);
  const [state, setState] = useState<MapState>({
    map: null,
    isLoading: false,
    isLoaded: false,
    error: null,
  });

  // Validate and sanitize numeric values
  const validateCoordinate = useCallback((coord: any): number => {
    const num = Number(coord);
    return !isNaN(num) && isFinite(num) ? num : 0;
  }, []);

  const validateZoom = useCallback((z: any): number => {
    const num = Number(z);
    if (isNaN(num) || !isFinite(num)) return 12;
    return Math.min(Math.max(num, 0), 24);
  }, []);

  const sanitizeOptions = useCallback(() => {
    const validatedCenter: [number, number] = [
      validateCoordinate(center[0]),
      validateCoordinate(center[1]),
    ];
    const validatedZoom = validateZoom(zoom);

    return {
      center: validatedCenter,
      zoom: validatedZoom,
      style,
    };
  }, [center, zoom, style, validateCoordinate, validateZoom]);

  useEffect(() => {
    // Don't initialize if no container or already initialized
    if (!container || mapRef.current) return;

    // Double-check container dimensions
    const rect = container.getBoundingClientRect();
    if (rect.width === 0 || rect.height === 0) {
      setState((prev) => ({
        ...prev,
        error: "Map container has no dimensions",
      }));
      return;
    }

    setState((prev) => ({ ...prev, isLoading: true, error: null }));

    // Use requestAnimationFrame to ensure DOM is ready
    const frameId = requestAnimationFrame(() => {
      try {
        const options = sanitizeOptions();
        
        const map = new maplibregl.Map({
          container,
          ...options,
          attributionControl: true,
        });

        mapRef.current = map;

        // Handle successful load
        map.on("load", () => {
          setState({
            map,
            isLoading: false,
            isLoaded: true,
            error: null,
          });

          // Add controls after load
          map.addControl(new maplibregl.NavigationControl(), "top-right");
          map.addControl(
            new maplibregl.GeolocateControl({
              positionOptions: {
                enableHighAccuracy: true,
              },
              trackUserLocation: true,
              showUserHeading: true,
            }),
            "top-right"
          );
          map.addControl(new maplibregl.FullscreenControl(), "top-right");
        });

        // Handle errors
        map.on("error", (e) => {
          console.error("MapLibre error:", e);
          setState((prev) => ({
            ...prev,
            isLoading: false,
            error: e.error?.message || "Failed to load map",
          }));
        });
      } catch (err) {
        console.error("Failed to initialize map:", err);
        setState({
          map: null,
          isLoading: false,
          isLoaded: false,
          error: err instanceof Error ? err.message : "Failed to initialize map",
        });
      }
    });

    // Cleanup
    return () => {
      cancelAnimationFrame(frameId);
      if (mapRef.current) {
        mapRef.current.remove();
        mapRef.current = null;
      }
      setState({
        map: null,
        isLoading: false,
        isLoaded: false,
        error: null,
      });
    };
  }, [container, sanitizeOptions]);

  return state;
}