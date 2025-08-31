import type { Map } from "maplibre-gl";
import { MAP_CONFIG, WORLD_CITIES } from "../constants/map-config";
import { LocateButton } from "./locate-button";

interface MapControlsProps {
  map: Map;
  showCityJumps?: boolean;
  cities?: typeof WORLD_CITIES;
  onLocationFound?: (position: {
    latitude: number;
    longitude: number;
    accuracy: number;
  }) => void;
}

export function MapControls({
  map,
  showCityJumps = true,
  cities = WORLD_CITIES,
  onLocationFound,
}: MapControlsProps) {
  const handleHomeClick = () => {
    map.flyTo({
      center: MAP_CONFIG.initialView.center,
      zoom: MAP_CONFIG.initialView.zoom,
      duration: 1500,
      essential: true,
    });
  };

  const handleCityClick = (
    city: (typeof WORLD_CITIES)[keyof typeof WORLD_CITIES],
  ) => {
    map.flyTo({
      center: [city.lng, city.lat],
      zoom: 11,
      duration: 1500,
      essential: true,
    });
  };

  return (
    <div
      style={{
        position: "absolute",
        top: "60px",
        right: "10px",
        zIndex: 10,
        backgroundColor: "white",
        borderRadius: "8px",
        padding: "0.5rem",
        boxShadow: "0 2px 8px rgba(0,0,0,0.15)",
      }}
    >
      <button
        type="button"
        onClick={handleHomeClick}
        style={{
          display: "block",
          width: "100%",
          padding: "0.5rem 1rem",
          marginBottom: "0.5rem",
          border: "1px solid #ddd",
          borderRadius: "4px",
          backgroundColor: "#f0f0f0",
          cursor: "pointer",
          fontSize: "14px",
          fontWeight: "bold",
        }}
        title="Press H key"
      >
        📍 Jakarta
      </button>

      <div style={{ marginBottom: "0.5rem" }}>
        <LocateButton map={map} onLocationFound={onLocationFound} />
      </div>

      {showCityJumps && (
        <div style={{ borderTop: "1px solid #eee", paddingTop: "0.5rem" }}>
          <div
            style={{ fontSize: "12px", color: "#666", marginBottom: "0.25rem" }}
          >
            Quick Jump:
          </div>
          {Object.entries(cities).map(([key, city]) => (
            <button
              type="button"
              key={key}
              onClick={() => handleCityClick(city)}
              style={{
                display: "block",
                width: "100%",
                padding: "0.25rem 0.5rem",
                marginBottom: "0.25rem",
                border: "1px solid #eee",
                borderRadius: "3px",
                backgroundColor: "white",
                cursor: "pointer",
                fontSize: "12px",
                textAlign: "left",
                transition: "background-color 0.2s",
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.backgroundColor = "#f5f5f5";
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.backgroundColor = "white";
              }}
            >
              📍 {city.name}
            </button>
          ))}
        </div>
      )}

      <div
        style={{
          borderTop: "1px solid #eee",
          paddingTop: "0.5rem",
          marginTop: "0.5rem",
          fontSize: "11px",
          color: "#999",
        }}
      >
        Tip: Press H for Jakarta
      </div>
    </div>
  );
}
