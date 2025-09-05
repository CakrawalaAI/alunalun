import { useEffect, useRef, useState } from "react";
import maplibregl from "maplibre-gl";

interface UseMapOptions {
  containerId: string;
  center?: [number, number];
  zoom?: number;
  style?: string;
}

export function useMap({
  containerId,
  center = [106.827, -6.175], // Default to Jakarta
  zoom = 12,
  style = "https://tiles.openfreemap.org/styles/liberty",
}: UseMapOptions) {
  const mapRef = useRef<maplibregl.Map | null>(null);
  const [isMapLoaded, setIsMapLoaded] = useState(false);

  useEffect(() => {
    // Check if container exists
    const container = document.getElementById(containerId);
    if (!container) return;

    // Initialize map
    const map = new maplibregl.Map({
      container: containerId,
      style,
      center,
      zoom,
      attributionControl: true,
    });

    // Set map reference
    mapRef.current = map;

    // Handle map load event
    map.on("load", () => {
      setIsMapLoaded(true);
    });

    // Cleanup
    return () => {
      map.remove();
      mapRef.current = null;
      setIsMapLoaded(false);
    };
  }, [containerId, style, center, zoom]);

  return {
    map: mapRef.current,
    isMapLoaded,
  };
}