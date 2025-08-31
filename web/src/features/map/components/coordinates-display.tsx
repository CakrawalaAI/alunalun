import { MapPin } from "lucide-react";
import type { Map } from "maplibre-gl";
import { useEffect, useState } from "react";

interface CoordinatesDisplayProps {
  map: Map;
  className?: string;
}

/**
 * Displays the current center coordinates of the map viewport
 * Updates in real-time as the map moves
 */
export function CoordinatesDisplay({
  map,
  className,
}: CoordinatesDisplayProps) {
  const [center, setCenter] = useState(() => {
    const c = map.getCenter();
    return { lng: c.lng, lat: c.lat };
  });
  const [showDetails, setShowDetails] = useState(true); // Start visible for now

  useEffect(() => {
    const updateCoordinates = () => {
      const c = map.getCenter();
      setCenter({ lng: c.lng, lat: c.lat });
    };

    // Update on map movement
    map.on("move", updateCoordinates);

    // Initial update
    updateCoordinates();

    return () => {
      map.off("move", updateCoordinates);
    };
  }, [map]);

  // Format coordinates to reasonable precision
  const formatCoord = (num: number, isLng: boolean) => {
    // Show 6 decimal places (roughly 0.1 meter precision)
    const formatted = num.toFixed(6);
    const label = isLng ? (num > 0 ? "E" : "W") : num > 0 ? "N" : "S";
    return `${Math.abs(Number(formatted))}°${label}`;
  };

  return (
    <div
      className={`absolute right-4 bottom-4 overflow-hidden rounded-lg bg-white/95 shadow-lg backdrop-blur-sm ${className}`}
      style={{
        transition: "all 0.3s ease",
        minWidth: showDetails ? "280px" : "48px",
        pointerEvents: "auto",
      }}
    >
      <button
        type="button"
        onClick={() => setShowDetails(!showDetails)}
        className="flex w-full items-center gap-2 px-3 py-2 text-left hover:bg-gray-50"
        title="Toggle coordinates display"
      >
        <MapPin className="h-4 w-4 text-gray-600" />
        {showDetails && (
          <span className="font-medium text-gray-700 text-xs">Map Center</span>
        )}
      </button>

      {showDetails && (
        <div className="space-y-1 border-gray-100 border-t px-3 pb-3">
          <div className="mt-2 space-y-1">
            <div className="flex items-center justify-between">
              <span className="text-gray-500 text-xs">Longitude:</span>
              <span className="font-mono text-gray-900 text-xs">
                {center.lng.toFixed(7)}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-gray-500 text-xs">Latitude:</span>
              <span className="font-mono text-gray-900 text-xs">
                {center.lat.toFixed(7)}
              </span>
            </div>
          </div>

          <div className="border-gray-100 border-t pt-2">
            <div className="text-center font-mono text-gray-600 text-xs">
              {formatCoord(center.lat, false)} {formatCoord(center.lng, true)}
            </div>
          </div>

          <div className="border-gray-100 border-t pt-2">
            <button
              type="button"
              onClick={() => {
                // Copy coordinates to clipboard
                const coords = `${center.lat.toFixed(7)}, ${center.lng.toFixed(7)}`;
                navigator.clipboard.writeText(coords);

                // Visual feedback (you could add a toast later)
                const button = event?.currentTarget as HTMLButtonElement;
                if (button) {
                  const originalText = button.textContent;
                  button.textContent = "Copied!";
                  setTimeout(() => {
                    button.textContent = originalText;
                  }, 1500);
                }
              }}
              className="w-full rounded px-2 py-1 text-blue-600 text-xs transition-colors hover:bg-blue-50 hover:text-blue-700"
            >
              Copy Coordinates
            </button>
          </div>

          <div className="border-gray-100 border-t pt-2">
            <button
              type="button"
              onClick={() => {
                // Recenter to DPR Jakarta
                map.flyTo({
                  center: [106.8001307, -6.2099592],
                  zoom: 15,
                  duration: 2000,
                  essential: true,
                });
              }}
              className="w-full rounded px-2 py-1 text-green-600 text-xs transition-colors hover:bg-green-50 hover:text-green-700"
            >
              Recenter to DPR Jakarta
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
