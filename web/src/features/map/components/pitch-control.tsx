import type { Map } from "maplibre-gl";
import { useEffect, useRef, useState } from "react";
import { useMapOrientation } from "../hooks/use-map-orientation";

interface PitchControlProps {
  map: Map;
  showAlways?: boolean;
}

const TiltIcon = ({ pitch }: { pitch: number }) => {
  const tiltAmount = Math.min(pitch / 60, 1); // Normalize to 0-1

  return (
    <svg
      width="20"
      height="20"
      viewBox="0 0 20 20"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
    >
      <title>3D Tilt Control</title>
      {/* Base plane */}
      <ellipse
        cx="10"
        cy="14"
        rx={8 - tiltAmount * 2}
        ry={3 - tiltAmount * 1.5}
        fill="none"
        opacity={0.5}
      />
      {/* Tilted plane */}
      <ellipse
        cx="10"
        cy={14 - tiltAmount * 8}
        rx="8"
        ry={3 + tiltAmount * 2}
        fill="currentColor"
        fillOpacity={0.2}
      />
      {/* Building representation */}
      <rect
        x="8"
        y={10 - tiltAmount * 4}
        width="4"
        height={4 + tiltAmount * 4}
        fill="currentColor"
        fillOpacity={0.4}
      />
    </svg>
  );
};

export function PitchControl({ map, showAlways = false }: PitchControlProps) {
  const { pitch } = useMapOrientation(map);
  const [isDragging, setIsDragging] = useState(false);
  const [isExpanded, setIsExpanded] = useState(false);
  const sliderRef = useRef<HTMLDivElement>(null);
  const MAX_PITCH = 85; // MapLibre max pitch

  useEffect(() => {
    if (!(isDragging && sliderRef.current)) {
      return;
    }

    const handleMouseMove = (e: MouseEvent) => {
      if (!sliderRef.current) {
        return;
      }

      const rect = sliderRef.current.getBoundingClientRect();
      const y = e.clientY - rect.top;
      const height = rect.height;

      // Invert: top = max pitch, bottom = 0 pitch
      const normalizedPitch = Math.max(0, Math.min(1, 1 - y / height));
      const newPitch = normalizedPitch * MAX_PITCH;

      map.setPitch(newPitch);
    };

    const handleMouseUp = () => {
      setIsDragging(false);
      document.body.style.cursor = "";
    };

    document.addEventListener("mousemove", handleMouseMove);
    document.addEventListener("mouseup", handleMouseUp);
    document.body.style.cursor = "ns-resize";

    return () => {
      document.removeEventListener("mousemove", handleMouseMove);
      document.removeEventListener("mouseup", handleMouseUp);
      document.body.style.cursor = "";
    };
  }, [isDragging, map]);

  // Touch support for slider
  useEffect(() => {
    if (!sliderRef.current) {
      return;
    }

    let touchActive = false;

    const handleTouchStart = (e: TouchEvent) => {
      e.preventDefault();
      touchActive = true;
    };

    const handleTouchMove = (e: TouchEvent) => {
      if (!(touchActive && sliderRef.current)) {
        return;
      }
      e.preventDefault();

      const touch = e.touches[0];
      const rect = sliderRef.current.getBoundingClientRect();
      const y = touch.clientY - rect.top;
      const height = rect.height;

      const normalizedPitch = Math.max(0, Math.min(1, 1 - y / height));
      const newPitch = normalizedPitch * MAX_PITCH;

      map.setPitch(newPitch);
    };

    const handleTouchEnd = () => {
      touchActive = false;
    };

    const slider = sliderRef.current;
    slider.addEventListener("touchstart", handleTouchStart, { passive: false });
    slider.addEventListener("touchmove", handleTouchMove, { passive: false });
    slider.addEventListener("touchend", handleTouchEnd);

    return () => {
      slider.removeEventListener("touchstart", handleTouchStart);
      slider.removeEventListener("touchmove", handleTouchMove);
      slider.removeEventListener("touchend", handleTouchEnd);
    };
  }, [map, isExpanded]);

  const handleButtonClick = () => {
    if (pitch > 0) {
      // Reset pitch
      map.easeTo({
        pitch: 0,
        duration: 500,
        essential: true,
      });
    } else {
      // Set to default 3D view
      map.easeTo({
        pitch: 45,
        duration: 500,
        essential: true,
      });
    }
  };

  const handleSliderMouseDown = (e: React.MouseEvent) => {
    e.preventDefault();
    setIsDragging(true);

    // Set pitch based on click position
    const rect = sliderRef.current!.getBoundingClientRect();
    const y = e.clientY - rect.top;
    const height = rect.height;
    const normalizedPitch = Math.max(0, Math.min(1, 1 - y / height));
    const newPitch = normalizedPitch * MAX_PITCH;
    map.setPitch(newPitch);
  };

  if (!showAlways && pitch === 0 && !isExpanded) {
    return null;
  }

  return (
    <div
      style={{
        position: "relative",
        display: "flex",
        flexDirection: "column",
        gap: "4px",
      }}
      onMouseEnter={() => setIsExpanded(true)}
      onMouseLeave={() => !isDragging && setIsExpanded(false)}
    >
      {/* 3D Tilt Button */}
      <button
        type="button"
        onClick={handleButtonClick}
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          width: "40px",
          height: "40px",
          padding: "0",
          border: pitch > 0 ? "1px solid #3b82f6" : "1px solid #ddd",
          borderRadius: "4px",
          backgroundColor: pitch > 0 ? "#eff6ff" : "white",
          cursor: "pointer",
          fontSize: "14px",
          transition: "all 0.2s",
          color: pitch > 0 ? "#3b82f6" : "#374151",
        }}
        title={`3D Tilt: ${Math.round(pitch)}°\nClick to ${pitch > 0 ? "reset" : "enable 3D view"}`}
        onMouseEnter={(e) => {
          e.currentTarget.style.backgroundColor =
            pitch > 0 ? "#dbeafe" : "#f5f5f5";
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.backgroundColor =
            pitch > 0 ? "#eff6ff" : "white";
        }}
      >
        <TiltIcon pitch={pitch} />
      </button>

      {/* Pitch Slider */}
      {(isExpanded || pitch > 0) && (
        <div
          ref={sliderRef}
          onMouseDown={handleSliderMouseDown}
          style={{
            position: "absolute",
            top: "44px",
            left: "50%",
            transform: "translateX(-50%)",
            width: "32px",
            height: "100px",
            backgroundColor: "white",
            border: "1px solid #ddd",
            borderRadius: "4px",
            cursor: isDragging ? "ns-resize" : "pointer",
            padding: "4px",
            boxShadow: "0 2px 8px rgba(0,0,0,0.1)",
            zIndex: 10,
            opacity: isExpanded || pitch > 0 ? 1 : 0,
            transition: "opacity 0.2s",
            pointerEvents: isExpanded || pitch > 0 ? "auto" : "none",
          }}
        >
          {/* Slider Track */}
          <div
            style={{
              position: "relative",
              width: "2px",
              height: "100%",
              backgroundColor: "#e5e7eb",
              margin: "0 auto",
              borderRadius: "1px",
            }}
          >
            {/* Slider Fill */}
            <div
              style={{
                position: "absolute",
                bottom: 0,
                width: "100%",
                height: `${(pitch / MAX_PITCH) * 100}%`,
                backgroundColor: "#3b82f6",
                borderRadius: "1px",
                transition: isDragging ? "none" : "height 0.2s",
              }}
            />

            {/* Slider Thumb */}
            <div
              style={{
                position: "absolute",
                bottom: `${(pitch / MAX_PITCH) * 100}%`,
                left: "50%",
                transform: "translate(-50%, 50%)",
                width: "12px",
                height: "12px",
                backgroundColor: "#3b82f6",
                border: "2px solid white",
                borderRadius: "50%",
                boxShadow: "0 1px 3px rgba(0,0,0,0.3)",
                transition: isDragging ? "none" : "bottom 0.2s",
              }}
            />
          </div>

          {/* Labels */}
          <div
            style={{
              position: "absolute",
              top: "4px",
              right: "100%",
              marginRight: "4px",
              fontSize: "9px",
              color: "#9ca3af",
              whiteSpace: "nowrap",
            }}
          >
            {MAX_PITCH}°
          </div>
          <div
            style={{
              position: "absolute",
              bottom: "4px",
              right: "100%",
              marginRight: "4px",
              fontSize: "9px",
              color: "#9ca3af",
              whiteSpace: "nowrap",
            }}
          >
            0°
          </div>
        </div>
      )}
    </div>
  );
}
