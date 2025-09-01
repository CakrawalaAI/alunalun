import type { Map } from "maplibre-gl";
import { cn } from "@/common/lib/utils";
import { useMapZoom } from "../hooks/use-map-zoom";

interface ZoomControlProps {
  map: Map;
  showZoomLevel?: boolean;
}

export function ZoomControl({ map, showZoomLevel = false }: ZoomControlProps) {
  const { currentZoom, canZoomIn, canZoomOut, zoomIn, zoomOut } =
    useMapZoom(map);

  return (
    <>
      <button
        type="button"
        onClick={zoomIn}
        disabled={!canZoomIn}
        className={cn(
          "flex items-center justify-center w-10 h-10 hover:bg-gray-50 border-b border-gray-200",
          !canZoomIn && "opacity-50 cursor-not-allowed"
        )}
        title="Zoom in"
        aria-label="Zoom in"
      >
        <svg
          width="20"
          height="20"
          viewBox="0 0 20 20"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
        >
          <line x1="10" y1="5" x2="10" y2="15" />
          <line x1="5" y1="10" x2="15" y2="10" />
        </svg>
      </button>

      {showZoomLevel && (
        <div className="flex items-center justify-center h-8 text-xs text-gray-600 font-medium border-b border-gray-200 select-none">
          {currentZoom.toFixed(1)}
        </div>
      )}

      <button
        type="button"
        onClick={zoomOut}
        disabled={!canZoomOut}
        className={cn(
          "flex items-center justify-center w-10 h-10 hover:bg-gray-50 border-b border-gray-200",
          !canZoomOut && "opacity-50 cursor-not-allowed"
        )}
        title="Zoom out"
        aria-label="Zoom out"
      >
        <svg
          width="20"
          height="20"
          viewBox="0 0 20 20"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
        >
          <line x1="5" y1="10" x2="15" y2="10" />
        </svg>
      </button>
    </>
  );
}
