import type { Map } from "maplibre-gl";
import { type ReactNode, useEffect, useRef, useState } from "react";
import "maplibre-gl/dist/maplibre-gl.css";
import { cn } from "@/common/lib/utils";
import { useMapInstance } from "../hooks/use-map-instance";

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
      <div className="fixed inset-0 z-0 flex items-center justify-center">
        <div className="text-center p-8">
          <h2 className="text-xl font-semibold mb-2">Failed to load map</h2>
          <p className="text-gray-600 mb-4">{error}</p>
          <button
            type="button"
            onClick={() => window.location.reload()}
            className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors"
          >
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
        className={cn(
          // MapLibre needs a container with explicit dimensions
          // Using h-full/w-full with proper parent heights works best
          "h-full w-full overflow-hidden",
          // Allow custom className to override if needed
          className
        )}
        style={{
          // Ensure the container has dimensions for MapLibre
          position: "absolute",
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
        }}
      />
      {!isLoaded && (
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 z-[1000] bg-white px-8 py-4 rounded-lg shadow-lg">
          <div className="flex items-center space-x-2">
            <svg
              className="animate-spin h-5 w-5 text-blue-500"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                className="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                strokeWidth="4"
              />
              <path
                className="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              />
            </svg>
            <span className="text-gray-700">Loading map...</span>
          </div>
        </div>
      )}
      {children}
    </>
  );
}
