import type { Map } from "maplibre-gl";
import { useEffect, useState } from "react";

interface GestureOverlayProps {
  map: Map | null;
}

interface GestureState {
  isRotating: boolean;
  isPitching: boolean;
  isZooming: boolean;
  bearing: number;
  pitch: number;
  zoom: number;
}

export function GestureOverlay({ map }: GestureOverlayProps) {
  const [gestureState, setGestureState] = useState<GestureState>({
    isRotating: false,
    isPitching: false,
    isZooming: false,
    bearing: 0,
    pitch: 0,
    zoom: 12,
  });
  const [showHints, setShowHints] = useState(false);

  useEffect(() => {
    if (!map) {
      return;
    }

    let rotateTimeout: ReturnType<typeof setTimeout>;
    let pitchTimeout: ReturnType<typeof setTimeout>;
    let zoomTimeout: ReturnType<typeof setTimeout>;

    const handleRotateStart = () => {
      clearTimeout(rotateTimeout);
      setGestureState((prev) => ({ ...prev, isRotating: true }));
    };

    const handleRotate = () => {
      setGestureState((prev) => ({
        ...prev,
        bearing: map.getBearing(),
        isRotating: true,
      }));

      clearTimeout(rotateTimeout);
      rotateTimeout = setTimeout(() => {
        setGestureState((prev) => ({ ...prev, isRotating: false }));
      }, 500);
    };

    const handlePitchStart = () => {
      clearTimeout(pitchTimeout);
      setGestureState((prev) => ({ ...prev, isPitching: true }));
    };

    const handlePitch = () => {
      setGestureState((prev) => ({
        ...prev,
        pitch: map.getPitch(),
        isPitching: true,
      }));

      clearTimeout(pitchTimeout);
      pitchTimeout = setTimeout(() => {
        setGestureState((prev) => ({ ...prev, isPitching: false }));
      }, 500);
    };

    const handleZoomStart = () => {
      clearTimeout(zoomTimeout);
      setGestureState((prev) => ({ ...prev, isZooming: true }));
    };

    const handleZoom = () => {
      setGestureState((prev) => ({
        ...prev,
        zoom: map.getZoom(),
        isZooming: true,
      }));

      clearTimeout(zoomTimeout);
      zoomTimeout = setTimeout(() => {
        setGestureState((prev) => ({ ...prev, isZooming: false }));
      }, 500);
    };

    // Check if touch device
    const isTouchDevice = "ontouchstart" in window;
    if (isTouchDevice) {
      setShowHints(true);
      setTimeout(() => setShowHints(false), 5000);
    }

    map.on("rotatestart", handleRotateStart);
    map.on("rotate", handleRotate);
    map.on("pitchstart", handlePitchStart);
    map.on("pitch", handlePitch);
    map.on("zoomstart", handleZoomStart);
    map.on("zoom", handleZoom);

    return () => {
      map.off("rotatestart", handleRotateStart);
      map.off("rotate", handleRotate);
      map.off("pitchstart", handlePitchStart);
      map.off("pitch", handlePitch);
      map.off("zoomstart", handleZoomStart);
      map.off("zoom", handleZoom);
      clearTimeout(rotateTimeout);
      clearTimeout(pitchTimeout);
      clearTimeout(zoomTimeout);
    };
  }, [map]);

  const { isRotating, isPitching, isZooming, bearing, pitch, zoom } =
    gestureState;
  const isActive = isRotating || isPitching || isZooming;

  return (
    <>
      {/* Gesture Feedback Overlay */}
      {isActive && (
        <div
          style={{
            position: "absolute",
            top: "50%",
            left: "50%",
            transform: "translate(-50%, -50%)",
            backgroundColor: "rgba(0, 0, 0, 0.8)",
            color: "white",
            padding: "12px 20px",
            borderRadius: "8px",
            fontSize: "14px",
            fontWeight: "500",
            pointerEvents: "none",
            transition: "opacity 0.2s",
            minWidth: "120px",
            textAlign: "center",
          }}
        >
          {isRotating && <div>Bearing: {Math.round(bearing)}°</div>}
          {isPitching && <div>Pitch: {Math.round(pitch)}°</div>}
          {isZooming && <div>Zoom: {zoom.toFixed(1)}</div>}
        </div>
      )}

      {/* Touch Gesture Hints */}
      {showHints && (
        <div
          style={{
            position: "absolute",
            bottom: "80px",
            left: "50%",
            transform: "translateX(-50%)",
            backgroundColor: "rgba(0, 0, 0, 0.8)",
            color: "white",
            padding: "16px",
            borderRadius: "8px",
            fontSize: "12px",
            pointerEvents: "none",
            maxWidth: "280px",
            animation: "fadeIn 0.5s",
          }}
        >
          <style>
            {`
              @keyframes fadeIn {
                from { opacity: 0; transform: translateX(-50%) translateY(10px); }
                to { opacity: 1; transform: translateX(-50%) translateY(0); }
              }
            `}
          </style>
          <div style={{ marginBottom: "8px", fontWeight: "bold" }}>
            Touch Gestures:
          </div>
          <div style={{ display: "flex", flexDirection: "column", gap: "4px" }}>
            <div>
              👆 <strong>Drag:</strong> Pan the map
            </div>
            <div>
              🤏 <strong>Pinch:</strong> Zoom in/out
            </div>
            <div>
              ✌️ <strong>Two-finger rotate:</strong> Rotate map
            </div>
            <div>
              ☝️☝️ <strong>Two-finger drag:</strong> Tilt for 3D
            </div>
          </div>
        </div>
      )}

      {/* Desktop Hints (on first load) */}
      {!("ontouchstart" in window) && showHints && (
        <div
          style={{
            position: "absolute",
            top: "80px",
            left: "10px",
            backgroundColor: "rgba(0, 0, 0, 0.8)",
            color: "white",
            padding: "12px",
            borderRadius: "8px",
            fontSize: "11px",
            pointerEvents: "none",
            maxWidth: "200px",
          }}
        >
          <div style={{ marginBottom: "6px", fontWeight: "bold" }}>
            Mouse Controls:
          </div>
          <div style={{ display: "flex", flexDirection: "column", gap: "3px" }}>
            <div>
              <strong>Right-click drag:</strong> Rotate
            </div>
            <div>
              <strong>Ctrl + drag:</strong> Rotate & tilt
            </div>
            <div>
              <strong>Scroll:</strong> Zoom
            </div>
          </div>
        </div>
      )}
    </>
  );
}
