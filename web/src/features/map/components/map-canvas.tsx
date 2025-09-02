import type { Map as MapLibreMap } from "maplibre-gl";
import { MapBase } from "./map-base";

/**
 * Pure UI component for the map canvas.
 * Just a thin wrapper around MapBase with injected props.
 * Zero business logic.
 */
export interface MapCanvasProps {
  onMapReady: (map: MapLibreMap) => void;
  className?: string;
}

export function MapCanvas({ onMapReady, className }: MapCanvasProps) {
  return <MapBase onMapReady={onMapReady} className={className} />;
}