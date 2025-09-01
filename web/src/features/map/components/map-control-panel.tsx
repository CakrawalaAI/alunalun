import type { ReactNode } from "react";

interface MapControlPanelProps {
  children: ReactNode;
  position?: "top-left" | "top-right" | "bottom-left" | "bottom-right";
  className?: string;
}

export function MapControlPanel({
  children,
  position = "top-right",
  className = "",
}: MapControlPanelProps) {
  const positionClasses = {
    "top-left": "top-4 left-4",
    "top-right": "top-4 right-4",
    "bottom-left": "bottom-4 left-4",
    "bottom-right": "bottom-4 right-4",
  };

  return (
    <div
      className={`absolute ${positionClasses[position]} ${className}`}
      style={{ zIndex: 1000 }}
    >
      <div className="flex flex-col bg-white rounded-lg shadow-md pointer-events-auto">
        {children}
      </div>
    </div>
  );
}