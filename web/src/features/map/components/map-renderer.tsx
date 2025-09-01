import { type ReactNode, useEffect, useRef, useState } from "react";
import "maplibre-gl/dist/maplibre-gl.css";
import { logger } from "@/common/logger/logger";
import { useMapControls } from "../hooks/use-map-controls";
import { useMapInstance } from "../hooks/use-map-instance";
import { debounce, useMapState } from "../hooks/use-map-state";
import { mapContainerStyles } from "../lib/styles";
import { OverlayProvider } from "../overlays";
import { MapUIOverlays } from "./map-ui-overlays";

export interface MapRendererProps {
  showControls?: boolean;
  className?: string;
  children?: ReactNode;
}

export function MapRenderer({
  showControls = true,
  className,
  children,
}: MapRendererProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [containerElement, setContainerElement] =
    useState<HTMLDivElement | null>(null);
  const { map, isLoaded, error } = useMapInstance(containerElement);
  const { setViewState } = useMapState();

  // Apply map controls and keyboard shortcuts
  useMapControls(map);

  useEffect(() => {
    // Set container element after mount
    if (containerRef.current) {
      setContainerElement(containerRef.current);
    }

    // Check for WebGL support
    const canvas = document.createElement("canvas");
    const gl =
      canvas.getContext("webgl") || canvas.getContext("experimental-webgl");
    if (!gl) {
      logger.error("WebGL is not supported in your browser");
    }
  }, []);

  // Save map state when user interacts with the map
  useEffect(() => {
    if (!(map && isLoaded)) {
      return;
    }

    // Create debounced save function (500ms delay)
    const saveMapState = debounce(() => {
      const center = map.getCenter();
      const zoom = map.getZoom();
      const bearing = map.getBearing();
      const pitch = map.getPitch();

      setViewState({
        center: [center.lng, center.lat],
        zoom,
        bearing,
        pitch,
      });

      logger.debug("Map state saved", { center, zoom, bearing, pitch });
    }, 500);

    // Listen to map events
    map.on("moveend", saveMapState);
    map.on("zoomend", saveMapState);
    map.on("rotateend", saveMapState);
    map.on("pitchend", saveMapState);

    // Cleanup
    return () => {
      map.off("moveend", saveMapState);
      map.off("zoomend", saveMapState);
      map.off("rotateend", saveMapState);
      map.off("pitchend", saveMapState);
    };
  }, [map, isLoaded, setViewState]);

  if (error) {
    return (
      <div
        style={{
          ...mapContainerStyles,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
        }}
      >
        <div style={{ textAlign: "center", padding: "2rem" }}>
          <h2>Failed to load map</h2>
          <p>{error}</p>
          <button type="button" onClick={() => window.location.reload()}>
            Reload
          </button>
        </div>
      </div>
    );
  }

  return (
    <OverlayProvider>
      <div style={{ position: "relative", width: "100%", height: "100%" }}>
        <div
          ref={containerRef}
          style={mapContainerStyles}
          className={className}
        />
        {isLoaded && map && (
          <MapUIOverlays map={map} showControls={showControls} />
        )}
        {!isLoaded && (
          <div
            style={{
              position: "absolute",
              top: "50%",
              left: "50%",
              transform: "translate(-50%, -50%)",
              zIndex: 1000,
              background: "white",
              padding: "1rem 2rem",
              borderRadius: "8px",
              boxShadow: "0 2px 8px rgba(0,0,0,0.1)",
            }}
          >
            Loading map...
          </div>
        )}
        {/* External features can add React components as overlays */}
        {children}
      </div>
    </OverlayProvider>
  );
}
