import { cn } from "@/common/lib/utils";
import { Button } from "@/common/components/ui/button";

/**
 * Pure UI component for map controls.
 * Zero business logic - just renders based on props.
 * All handlers and state are injected from the parent.
 */
export interface ControlPanelProps {
  zoom: {
    current: number;
    canZoomIn: boolean;
    canZoomOut: boolean;
    onZoomIn: () => void;
    onZoomOut: () => void;
  };
  orientation: {
    bearing: number;
    pitch: number;
    isRotated: boolean;
    onReset: () => void;
  };
  location: {
    isLocating: boolean;
    hasError: boolean;
    errorMessage?: string;
    onLocate: () => void;
  };
  isReady: boolean;
}

export function ControlPanel(props: ControlPanelProps) {
  // Don't render if map isn't ready
  if (!props.isReady) return null;

  return (
    <div
      className="absolute top-4 right-4 z-[999999]"
      style={{
        position: 'absolute',
        top: '1rem',
        right: '1rem',
        zIndex: 999999,
        pointerEvents: 'auto'
      }}
    >
      <div className={cn(
        "flex flex-col",
        "bg-white/95 backdrop-blur-md",
        "rounded-xl shadow-2xl",
        "border border-gray-200/60",
        "overflow-hidden",
        "isolate"
      )}>
        {/* Zoom In Button */}
        <Button
          type="button"
          onClick={props.zoom.onZoomIn}
          disabled={!props.zoom.canZoomIn}
          className={cn(
            "relative w-11 h-11",
            "flex items-center justify-center",
            "bg-white hover:bg-gray-50 active:bg-gray-100",
            "transition-colors duration-150",
            "focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-inset",
            "disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:bg-white",
            "select-none cursor-pointer"
          )}
          style={{ pointerEvents: 'auto', position: 'relative', zIndex: 1 }}
          title="Zoom in"
          aria-label="Zoom in"
        >
          <svg
            width="20"
            height="20"
            viewBox="0 0 20 20"
            className="text-gray-700 pointer-events-none"
          >
            <path
              d="M10 5v10M5 10h10"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
            />
          </svg>
        </Button>

        {/* Zoom Level Display */}
        <div className={cn(
          "h-8 px-3",
          "flex items-center justify-center",
          "text-xs font-semibold text-gray-600",
          "bg-gray-50/50 border-y border-gray-200/50",
          "select-none"
        )}>
          {props.zoom.current.toFixed(1)}
        </div>

        {/* Zoom Out Button */}
        <button
          type="button"
          onClick={props.zoom.onZoomOut}
          disabled={!props.zoom.canZoomOut}
          className={cn(
            "relative w-11 h-11",
            "flex items-center justify-center",
            "bg-white hover:bg-gray-50 active:bg-gray-100",
            "transition-colors duration-150",
            "focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-inset",
            "disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:bg-white",
            "select-none cursor-pointer"
          )}
          style={{ pointerEvents: 'auto', position: 'relative', zIndex: 1 }}
          title="Zoom out"
          aria-label="Zoom out"
        >
          <svg
            width="20"
            height="20"
            viewBox="0 0 20 20"
            className="text-gray-700 pointer-events-none"
          >
            <path
              d="M5 10h10"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
            />
          </svg>
        </button>

        {/* Divider */}
        <div className="h-px bg-gray-200/50" />

        {/* Locate Button */}
        <button
          type="button"
          onClick={props.location.onLocate}
          disabled={props.location.isLocating}
          className={cn(
            "relative w-11 h-11",
            "flex items-center justify-center",
            "bg-white hover:bg-gray-50 active:bg-gray-100",
            "transition-all duration-150",
            "focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-inset",
            props.location.isLocating && "animate-pulse cursor-wait",
            props.location.hasError ? "text-red-500 hover:text-red-600" : "text-gray-700",
            "select-none cursor-pointer"
          )}
          style={{ pointerEvents: 'auto', position: 'relative', zIndex: 1 }}
          title={
            props.location.hasError
              ? `Error: ${props.location.errorMessage}. Click to retry`
              : props.location.isLocating
                ? "Getting your location..."
                : "Locate me"
          }
          aria-label="Get current location"
        >
          {props.location.isLocating ? (
            <svg
              width="20"
              height="20"
              viewBox="0 0 20 20"
              className="animate-spin pointer-events-none"
            >
              <circle
                cx="10"
                cy="10"
                r="8"
                stroke="currentColor"
                strokeWidth="2"
                fill="none"
                strokeDasharray="40"
                strokeDashoffset="10"
              />
            </svg>
          ) : (
            <svg
              width="20"
              height="20"
              viewBox="0 0 20 20"
              className="pointer-events-none"
            >
              <circle cx="10" cy="10" r="8" stroke="currentColor" strokeWidth="2" fill="none" />
              <line x1="10" y1="2" x2="10" y2="6" stroke="currentColor" strokeWidth="2" />
              <line x1="10" y1="14" x2="10" y2="18" stroke="currentColor" strokeWidth="2" />
              <line x1="2" y1="10" x2="6" y2="10" stroke="currentColor" strokeWidth="2" />
              <line x1="14" y1="10" x2="18" y2="10" stroke="currentColor" strokeWidth="2" />
            </svg>
          )}
        </button>

        {/* Divider */}
        <div className="h-px bg-gray-200/50" />

        {/* Reset Orientation Button */}
        <button
          type="button"
          onClick={props.orientation.onReset}
          className={cn(
            "relative w-11 h-11",
            "flex items-center justify-center",
            "bg-white hover:bg-gray-50 active:bg-gray-100",
            "transition-all duration-150",
            "focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-inset",
            !props.orientation.isRotated && "opacity-60",
            "select-none cursor-pointer",
            "rounded-b-xl"
          )}
          style={{ pointerEvents: 'auto', position: 'relative', zIndex: 1 }}
          title={`Reset orientation (bearing: ${props.orientation.bearing.toFixed(0)}°, pitch: ${props.orientation.pitch.toFixed(0)}°)`}
          aria-label="Reset map orientation"
        >
          <svg
            width="20"
            height="20"
            viewBox="0 0 20 20"
            className="text-gray-700 transition-transform duration-300 pointer-events-none"
            style={{ transform: `rotate(${-props.orientation.bearing}deg)` }}
          >
            <circle cx="10" cy="10" r="8" stroke="currentColor" strokeWidth="2" fill="none" />
            <path d="M10 6 L7 13 L10 11 L13 13 Z" fill="currentColor" stroke="none" />
          </svg>
        </button>
      </div>
    </div>
  );
}