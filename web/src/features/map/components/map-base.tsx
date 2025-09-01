import type { Map } from "maplibre-gl";
import { type ReactNode, useEffect, useRef, useState } from "react";
import "maplibre-gl/dist/maplibre-gl.css";
import { useMapInstance } from "../hooks/use-map-instance";
import { mapContainerStyles } from "../lib/styles";

export interface MapBaseProps {
  onMapReady?: (map: Map) => void;
  className?: string;
  children?: ReactNode;
}

export function MapBase({ onMapReady, className, children }: MapBaseProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [containerElement, setContainerElement] =
    useState<HTMLDivElement | null>(null);
  const { map, isLoaded, error } = useMapInstance(containerElement);

  useEffect(() => {
    if (containerRef.current) {
      setContainerElement(containerRef.current);
    }
  }, []);

  useEffect(() => {
    if (map && isLoaded && onMapReady) {
      onMapReady(map);
    }
  }, [map, isLoaded, onMapReady]);

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
    <>
      <div
        ref={containerRef}
        style={{
          ...mapContainerStyles,
          position: "absolute" as const, // Use absolute instead of fixed when className is provided
          ...(className ? { width: "100%", height: "100%" } : {}),
        }}
        className={className}
      />
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
      {children}
    </>
  );
}
