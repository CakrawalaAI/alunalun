import type { Map } from "maplibre-gl";
import maplibregl from "maplibre-gl";

/**
 * Calculate appropriate zoom level based on location accuracy
 * Higher accuracy (smaller number) = higher zoom
 */
export function getZoomFromAccuracy(accuracyInMeters: number): number {
  if (accuracyInMeters <= 5) {
    return 18; // Building level
  }
  if (accuracyInMeters <= 10) {
    return 17; // Street level
  }
  if (accuracyInMeters <= 25) {
    return 16; // Block level
  }
  if (accuracyInMeters <= 50) {
    return 15; // Neighborhood level
  }
  if (accuracyInMeters <= 100) {
    return 14; // District level
  }
  if (accuracyInMeters <= 500) {
    return 13; // City sector
  }
  if (accuracyInMeters <= 1000) {
    return 12; // City level
  }
  if (accuracyInMeters <= 5000) {
    return 11; // Metro area
  }
  return 10; // Regional level
}

/**
 * Format accuracy for user-friendly display
 */
export function formatAccuracy(meters: number): string {
  if (meters < 1) {
    return "±<1m";
  } else if (meters < 1000) {
    return `±${Math.round(meters)}m`;
  } else {
    const km = (meters / 1000).toFixed(1);
    return `±${km}km`;
  }
}

/**
 * Add a location marker with accuracy circle to the map
 */
export function addLocationMarker(
  map: Map,
  latitude: number,
  longitude: number,
  accuracyInMeters: number,
): maplibregl.Marker {
  // Create marker element
  const el = document.createElement("div");
  el.style.width = "20px";
  el.style.height = "20px";
  el.style.borderRadius = "50%";
  el.style.backgroundColor = "#4285F4";
  el.style.border = "3px solid white";
  el.style.boxShadow = "0 2px 6px rgba(0,0,0,0.3)";
  el.style.zIndex = "1000";

  // Create accuracy circle
  const accuracyEl = document.createElement("div");
  accuracyEl.style.position = "absolute";
  accuracyEl.style.borderRadius = "50%";
  accuracyEl.style.backgroundColor = "rgba(66, 133, 244, 0.15)";
  accuracyEl.style.border = "1px solid rgba(66, 133, 244, 0.3)";
  accuracyEl.style.pointerEvents = "none";

  // Calculate accuracy circle size based on zoom and accuracy
  const updateAccuracyCircle = () => {
    const zoom = map.getZoom();
    const metersPerPixel =
      (156543.03392 * Math.cos((latitude * Math.PI) / 180)) / 2 ** zoom;
    const diameterInPixels = (accuracyInMeters * 2) / metersPerPixel;

    accuracyEl.style.width = `${diameterInPixels}px`;
    accuracyEl.style.height = `${diameterInPixels}px`;
    accuracyEl.style.left = `${10 - diameterInPixels / 2}px`;
    accuracyEl.style.top = `${10 - diameterInPixels / 2}px`;
  };

  // Create container for both elements
  const container = document.createElement("div");
  container.style.position = "relative";
  container.appendChild(accuracyEl);
  container.appendChild(el);

  // Update accuracy circle on zoom
  map.on("zoom", updateAccuracyCircle);
  updateAccuracyCircle();

  // Create and add marker
  const marker = new maplibregl.Marker({
    element: container,
    anchor: "center",
  })
    .setLngLat([longitude, latitude])
    .addTo(map);

  // Add popup with accuracy info
  const popup = new maplibregl.Popup({ offset: 25 }).setText(
    `Your location (${formatAccuracy(accuracyInMeters)})`,
  );

  marker.setPopup(popup);

  return marker;
}

/**
 * Check if a location timestamp is stale
 */
export function isLocationStale(
  timestamp: number,
  maxAgeInSeconds = 300,
): boolean {
  const ageInSeconds = (Date.now() - timestamp) / 1000;
  return ageInSeconds > maxAgeInSeconds;
}
