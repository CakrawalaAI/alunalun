import type { Map } from "maplibre-gl";
import { useMapZoom } from "../hooks/use-map-zoom";
import { controlButtonStyles } from "../lib/styles";

interface ZoomControlProps {
  map: Map;
  showZoomLevel?: boolean;
}

export function ZoomControl({ map, showZoomLevel = false }: ZoomControlProps) {
  const { currentZoom, canZoomIn, canZoomOut, zoomIn, zoomOut } =
    useMapZoom(map);

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        gap: "4px",
        backgroundColor: "white",
        borderRadius: "4px",
        padding: "4px",
        boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
        position: "relative",
        zIndex: 10,
      }}
    >
      <button
        type="button"
        onClick={zoomIn}
        disabled={!canZoomIn}
        style={{
          ...controlButtonStyles,
          opacity: canZoomIn ? 1 : 0.5,
          cursor: canZoomIn ? "pointer" : "not-allowed",
        }}
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
        <div
          style={{
            padding: "4px",
            fontSize: "11px",
            textAlign: "center",
            color: "#666",
            userSelect: "none",
          }}
        >
          {currentZoom.toFixed(1)}
        </div>
      )}

      <button
        type="button"
        onClick={zoomOut}
        disabled={!canZoomOut}
        style={{
          ...controlButtonStyles,
          opacity: canZoomOut ? 1 : 0.5,
          cursor: canZoomOut ? "pointer" : "not-allowed",
        }}
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
    </div>
  );
}
