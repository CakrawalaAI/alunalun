import { Suspense, lazy } from "react";
import { MapLoading } from "./map-loading";

// Lazy load the map component
const MapCore = lazy(() => 
  import("./map-core").then(module => ({ default: module.MapCore }))
);

interface MapContainerProps {
  className?: string;
  center?: [number, number];
  zoom?: number;
}

export function MapContainer({ 
  className = "", 
  center,
  zoom 
}: MapContainerProps) {
  return (
    <Suspense fallback={<MapLoading />}>
      <MapCore 
        className={className}
        center={center}
        zoom={zoom}
      />
    </Suspense>
  );
}