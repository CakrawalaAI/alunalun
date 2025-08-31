import type { Map } from "maplibre-gl";
import { useEffect, useRef } from "react";
import { useMapZoom } from "../hooks/use-map-zoom";

interface ZoomControlsProps {
  map: Map;
  showZoomLevel?: boolean;
}

export function ZoomControls({
  map,
  showZoomLevel = false,
}: ZoomControlsProps) {
  const { currentZoom, canZoomIn, canZoomOut, zoomIn, zoomOut } =
    useMapZoom(map);
  const containerRef = useRef<HTMLDivElement>(null);

  // Handle scroll wheel zoom on the control
  useEffect(() => {
    if (!containerRef.current) {
      return;
    }

    const handleWheel = (e: WheelEvent) => {
      e.preventDefault();
      e.stopPropagation();

      // Determine zoom direction
      const delta = e.deltaY || e.deltaX;
      if (delta > 0) {
        zoomOut();
      } else if (delta < 0) {
        zoomIn();
      }
    };

    const container = containerRef.current;
    container.addEventListener("wheel", handleWheel, { passive: false });

    return () => {
      container.removeEventListener("wheel", handleWheel);
    };
  }, [zoomIn, zoomOut]);

  const getButtonStyle = (disabled: boolean) => ({
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    width: "40px",
    height: "40px",
    padding: "0",
    border: "1px solid #ddd",
    borderRadius: "4px",
    backgroundColor: "white",
    cursor: disabled ? "not-allowed" : "pointer",
    fontSize: "18px",
    fontWeight: "bold",
    transition: "all 0.2s",
    color: disabled ? "#9ca3af" : "#374151",
    opacity: disabled ? 0.5 : 1,
    userSelect: "none" as const,
  });

  const containerStyle = {
    display: "flex",
    flexDirection: "column" as const,
    gap: "2px",
    backgroundColor: "white",
    borderRadius: "4px",
    overflow: "hidden",
  };

  const zoomLevelStyle = {
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    height: "24px",
    padding: "0 8px",
    fontSize: "11px",
    color: "#6b7280",
    borderTop: "1px solid #e5e7eb",
    borderBottom: "1px solid #e5e7eb",
    cursor: "pointer",
    userSelect: "none" as const,
  };

  // Handle zoom level click - reset to default zoom
  const handleZoomLevelClick = () => {
    map.zoomTo(12, {
      duration: 500,
      essential: true,
    });
  };

  // Keyboard support for zoom buttons
  const handleKeyDown = (e: React.KeyboardEvent, action: () => void) => {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      action();
    }
  };

  return (
    <div
      ref={containerRef}
      style={containerStyle}
      title="Scroll to zoom • Click numbers to reset"
    >
      <button
        type="button"
        onClick={zoomIn}
        onKeyDown={(e) => handleKeyDown(e, zoomIn)}
        disabled={!canZoomIn}
        style={getButtonStyle(!canZoomIn)}
        title="Zoom in (+ key)"
        aria-label="Zoom in"
        onMouseEnter={(e) => {
          if (canZoomIn) {
            e.currentTarget.style.backgroundColor = "#f5f5f5";
          }
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.backgroundColor = "white";
        }}
      >
        +
      </button>

      {showZoomLevel && (
        <div
          style={zoomLevelStyle}
          onClick={handleZoomLevelClick}
          onKeyDown={(e) => handleKeyDown(e, handleZoomLevelClick)}
          role="button"
          tabIndex={0}
          title="Click to reset zoom"
          onMouseEnter={(e) => {
            e.currentTarget.style.backgroundColor = "#f9fafb";
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.backgroundColor = "transparent";
          }}
        >
          {Math.round(currentZoom)}
        </div>
      )}

      <button
        type="button"
        onClick={zoomOut}
        onKeyDown={(e) => handleKeyDown(e, zoomOut)}
        disabled={!canZoomOut}
        style={getButtonStyle(!canZoomOut)}
        title="Zoom out (- key)"
        aria-label="Zoom out"
        onMouseEnter={(e) => {
          if (canZoomOut) {
            e.currentTarget.style.backgroundColor = "#f5f5f5";
          }
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.backgroundColor = "white";
        }}
      >
        −
      </button>
    </div>
  );
}
