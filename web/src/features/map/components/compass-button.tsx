import type { Map } from "maplibre-gl";
import { useEffect, useRef, useState } from "react";
import { useMapOrientation } from "../hooks/use-map-orientation";

interface CompassButtonProps {
  map: Map;
  showAlways?: boolean;
}

const CompassIcon = ({
  bearing,
  className,
}: {
  bearing: number;
  className?: string;
}) => (
  <svg
    width="24"
    height="24"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    className={className}
    style={{
      transform: `rotate(${-bearing}deg)`,
      transition: "transform 0.3s ease-out",
    }}
    role="img"
    aria-label="Compass icon"
  >
    <title>Compass - Drag to rotate, click to reset</title>
    <circle cx="12" cy="12" r="10" />
    <path
      d="M12 2 L15 9 L12 7 L9 9 Z"
      fill="#ef4444"
      stroke="#ef4444"
      strokeWidth="1"
    />
    <text
      x="12"
      y="6"
      textAnchor="middle"
      fontSize="6"
      fill="currentColor"
      stroke="none"
    >
      N
    </text>
  </svg>
);

export function CompassButton({ map, showAlways = false }: CompassButtonProps) {
  const { bearing, pitch, isOrientationModified, resetOrientation } =
    useMapOrientation(map);
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState<{
    x: number;
    y: number;
    bearing: number;
  } | null>(null);
  const buttonRef = useRef<HTMLButtonElement>(null);

  // Handle drag rotation
  useEffect(() => {
    if (!(isDragging && dragStart && buttonRef.current)) {
      return;
    }

    const handleMouseMove = (e: MouseEvent) => {
      const rect = buttonRef.current!.getBoundingClientRect();
      const centerX = rect.left + rect.width / 2;
      const centerY = rect.top + rect.height / 2;

      // Calculate angle from center
      const currentAngle = Math.atan2(e.clientY - centerY, e.clientX - centerX);
      const startAngle = Math.atan2(
        dragStart.y - centerY,
        dragStart.x - centerX,
      );
      const deltaAngle = (currentAngle - startAngle) * (180 / Math.PI);

      // Update map bearing
      const newBearing = (dragStart.bearing - deltaAngle + 360) % 360;
      map.setBearing(newBearing > 180 ? newBearing - 360 : newBearing);
    };

    const handleMouseUp = () => {
      setIsDragging(false);
      setDragStart(null);
      document.body.style.cursor = "";
    };

    document.addEventListener("mousemove", handleMouseMove);
    document.addEventListener("mouseup", handleMouseUp);
    document.body.style.cursor = "grabbing";

    return () => {
      document.removeEventListener("mousemove", handleMouseMove);
      document.removeEventListener("mouseup", handleMouseUp);
      document.body.style.cursor = "";
    };
  }, [isDragging, dragStart, map]);

  // Handle touch drag
  useEffect(() => {
    if (!buttonRef.current) {
      return;
    }

    let touchStart: {
      x: number;
      y: number;
      bearing: number;
      identifier: number;
    } | null = null;

    const handleTouchStart = (e: TouchEvent) => {
      e.preventDefault();
      const touch = e.touches[0];
      touchStart = {
        x: touch.clientX,
        y: touch.clientY,
        bearing: map.getBearing(),
        identifier: touch.identifier,
      };
    };

    const handleTouchMove = (e: TouchEvent) => {
      if (!(touchStart && buttonRef.current)) {
        return;
      }
      e.preventDefault();

      const touch = Array.from(e.touches).find(
        (t) => t.identifier === touchStart?.identifier,
      );
      if (!touch) {
        return;
      }

      const rect = buttonRef.current.getBoundingClientRect();
      const centerX = rect.left + rect.width / 2;
      const centerY = rect.top + rect.height / 2;

      const currentAngle = Math.atan2(
        touch.clientY - centerY,
        touch.clientX - centerX,
      );
      const startAngle = Math.atan2(
        touchStart.y - centerY,
        touchStart.x - centerX,
      );
      const deltaAngle = (currentAngle - startAngle) * (180 / Math.PI);

      const newBearing = (touchStart.bearing - deltaAngle + 360) % 360;
      map.setBearing(newBearing > 180 ? newBearing - 360 : newBearing);
    };

    const handleTouchEnd = (e: TouchEvent) => {
      const touch = e.changedTouches[0];
      if (touchStart && touch.identifier === touchStart.identifier) {
        // Check if it was a tap (no significant movement)
        const dx = touch.clientX - touchStart.x;
        const dy = touch.clientY - touchStart.y;
        if (Math.abs(dx) < 5 && Math.abs(dy) < 5) {
          resetOrientation();
        }
        touchStart = null;
      }
    };

    const button = buttonRef.current;
    button.addEventListener("touchstart", handleTouchStart, { passive: false });
    button.addEventListener("touchmove", handleTouchMove, { passive: false });
    button.addEventListener("touchend", handleTouchEnd);

    return () => {
      button.removeEventListener("touchstart", handleTouchStart);
      button.removeEventListener("touchmove", handleTouchMove);
      button.removeEventListener("touchend", handleTouchEnd);
    };
  }, [map, resetOrientation]);

  if (!(showAlways || isOrientationModified)) {
    return null;
  }

  const handleMouseDown = (e: React.MouseEvent) => {
    // Prevent text selection during drag
    e.preventDefault();
    setIsDragging(true);
    setDragStart({
      x: e.clientX,
      y: e.clientY,
      bearing: map.getBearing(),
    });
  };

  const handleClick = () => {
    // Only reset if not dragging
    if (!isDragging) {
      resetOrientation();
    }
  };

  const getButtonStyle = () => ({
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    width: "40px",
    height: "40px",
    padding: "0",
    border: isDragging ? "2px solid #3b82f6" : "1px solid #ddd",
    borderRadius: "4px",
    backgroundColor: isDragging ? "#eff6ff" : "white",
    cursor: isDragging ? "grabbing" : "grab",
    fontSize: "14px",
    transition: "all 0.2s",
    color: "#374151",
    opacity: showAlways || isOrientationModified ? 1 : 0.5,
    userSelect: "none" as const,
  });

  return (
    <button
      ref={buttonRef}
      type="button"
      onClick={handleClick}
      onMouseDown={handleMouseDown}
      style={getButtonStyle()}
      title={`Compass\nBearing: ${Math.round(bearing)}°\nPitch: ${Math.round(pitch)}°\n\nDrag to rotate • Click to reset north\nKeyboard: R to reset`}
      onMouseEnter={(e) => {
        if (!isDragging) {
          e.currentTarget.style.backgroundColor = "#f5f5f5";
        }
      }}
      onMouseLeave={(e) => {
        if (!isDragging) {
          e.currentTarget.style.backgroundColor = "white";
        }
      }}
    >
      <CompassIcon bearing={bearing} />
    </button>
  );
}
