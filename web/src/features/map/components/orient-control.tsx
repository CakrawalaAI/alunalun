import type { Map } from "maplibre-gl";
import { cn } from "@/common/lib/utils";
import { useMapOrientation } from "../hooks/use-map-orientation";

interface OrientControlProps {
  map: Map;
}

const CompassIcon = ({ rotation }: { rotation: number }) => (
  <svg
    width="20"
    height="20"
    viewBox="0 0 20 20"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    style={{
      transform: `rotate(${rotation}deg)`,
      transition: "transform 0.3s ease",
    }}
    role="img"
    aria-label="Compass icon"
  >
    <title>Compass - click to reset orientation</title>
    <circle cx="10" cy="10" r="8" />
    <path d="M10 6 L7 13 L10 11 L13 13 Z" fill="currentColor" stroke="none" />
  </svg>
);

export function OrientControl({ map }: OrientControlProps) {
  const { bearing, pitch, resetOrientation } = useMapOrientation(map);
  const isRotated = bearing !== 0 || pitch !== 0;

  return (
    <button
      type="button"
      onClick={resetOrientation}
      className={cn(
        "flex items-center justify-center w-10 h-10 hover:bg-gray-50",
        !isRotated && "opacity-70"
      )}
      title={`Reset orientation (bearing: ${bearing.toFixed(0)}°, pitch: ${pitch.toFixed(0)}°)`}
      aria-label="Reset map orientation"
    >
      <CompassIcon rotation={-bearing} />
    </button>
  );
}
