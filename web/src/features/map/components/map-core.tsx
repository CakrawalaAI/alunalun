import { useMapContainer } from "../hooks/useMapContainer";
import { useMapLibre } from "../hooks/useMapLibre";
import { MapLoading } from "./map-loading";
import "../styles/map.css";

interface MapCoreProps {
  className?: string;
  center?: [number, number];
  zoom?: number;
}

export function MapCore({ 
  className = "", 
  center,
  zoom 
}: MapCoreProps) {
  // First ensure container is ready with dimensions
  const { containerRef, isReady, dimensions } = useMapContainer();
  
  // Then initialize map once container is ready
  const { isLoading, isLoaded, error } = useMapLibre({
    container: isReady ? containerRef.current : null,
    center,
    zoom,
  });

  // Show loading while container is being measured or map is loading
  if (!isReady || isLoading) {
    return <MapLoading />;
  }

  // Show error if map failed to load
  if (error) {
    return (
      <div className="flex h-full w-full items-center justify-center bg-red-50">
        <div className="text-center">
          <p className="mb-2 text-red-600 font-semibold">Map Error</p>
          <p className="text-sm text-red-500">{error}</p>
          <p className="mt-2 text-xs text-gray-500">
            Container: {dimensions?.width}x{dimensions?.height}px
          </p>
        </div>
      </div>
    );
  }

  return (
    <div 
      ref={containerRef}
      className={`map-container ${className}`}
      style={{
        width: "100%",
        height: "100%",
        visibility: isLoaded ? "visible" : "hidden",
      }}
    />
  );
}