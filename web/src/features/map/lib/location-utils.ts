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
