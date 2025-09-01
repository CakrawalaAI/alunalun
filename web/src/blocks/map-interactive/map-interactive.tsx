import type { Map } from "maplibre-gl";
import { useState } from "react";
import {
  LocateControl,
  MapBase,
  OrientControl,
  ZoomControl,
} from "@/features/map";

export function MapInteractive() {
  const [map, setMap] = useState<Map | null>(null);

  return (
    <div className="relative h-full w-full">
      <MapBase onMapReady={setMap} className="h-full w-full" />

      {map && (
        <div className="pointer-events-none absolute top-4 right-4 space-y-2 z-10">
          <div className="pointer-events-auto">
            <ZoomControl map={map} showZoomLevel={true} />
          </div>
          <div className="pointer-events-auto">
            <LocateControl map={map} />
          </div>
          <div className="pointer-events-auto">
            <OrientControl map={map} />
          </div>
        </div>
      )}
    </div>
  );
}
