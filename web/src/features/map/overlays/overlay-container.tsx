import type { ReactNode } from "react";

interface OverlayContainerProps {
  children: ReactNode;
}

/**
 * Container for DOM-based overlays on the map
 * Positioned absolutely to cover the map area
 * pointer-events-none allows map interaction to pass through
 * Child components can override pointer-events for interactivity
 */
export function OverlayContainer({ children }: OverlayContainerProps) {
  return (
    <div
      className="pointer-events-none absolute inset-0"
      style={{
        position: "absolute",
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        pointerEvents: "none",
        zIndex: 10,
      }}
    >
      {children}
    </div>
  );
}
