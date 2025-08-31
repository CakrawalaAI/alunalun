import type { Map } from "maplibre-gl";
import { useEffect, useState } from "react";

interface MapOrientation {
  bearing: number;
  pitch: number;
  isModified: boolean;
}

/**
 * Hook to track and control map orientation (bearing and pitch)
 */
export function useMapOrientation(map: Map | null) {
  const [orientation, setOrientation] = useState<MapOrientation>({
    bearing: 0,
    pitch: 0,
    isModified: false,
  });

  useEffect(() => {
    if (!map) {
      return;
    }

    const updateOrientation = () => {
      const bearing = map.getBearing();
      const pitch = map.getPitch();
      const isModified = Math.abs(bearing) > 5 || pitch > 5;

      setOrientation({
        bearing,
        pitch,
        isModified,
      });
    };

    // Initial state
    updateOrientation();

    // Listen to map movements
    map.on("rotate", updateOrientation);
    map.on("pitch", updateOrientation);
    map.on("moveend", updateOrientation);

    return () => {
      map.off("rotate", updateOrientation);
      map.off("pitch", updateOrientation);
      map.off("moveend", updateOrientation);
    };
  }, [map]);

  const resetOrientation = () => {
    if (!map) {
      return;
    }

    map.easeTo({
      bearing: 0,
      pitch: 0,
      duration: 1000,
      essential: true,
    });
  };

  return {
    bearing: orientation.bearing,
    pitch: orientation.pitch,
    isOrientationModified: orientation.isModified,
    resetOrientation,
  };
}
