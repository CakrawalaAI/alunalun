import { MapContainer } from "./map-container";
import { MapErrorBoundary } from "./map-error-boundary";

interface MapProps {
  className?: string;
  center?: [number, number];
  zoom?: number;
}

export function Map({ 
  className = "", 
  center = [106.827, -6.175], // Jakarta coordinates
  zoom = 12 
}: MapProps) {
  return (
    <MapErrorBoundary>
      <MapContainer 
        className={className}
        center={center}
        zoom={zoom}
      />
    </MapErrorBoundary>
  );
}