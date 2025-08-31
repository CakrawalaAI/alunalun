import { type ReactNode, useEffect, useRef, useState } from "react";
import "maplibre-gl/dist/maplibre-gl.css";
import { logger } from "@/common/logger/logger";
import { useMapControls } from "../hooks/use-map-controls";
import { useMapInstance } from "../hooks/use-map-instance";
import { mapContainerStyles } from "../lib/styles";
import { OverlayProvider } from "../overlays";
import { MapControls } from "./map-controls";

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
        {isLoaded && map && showControls && <MapControls map={map} />}
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
