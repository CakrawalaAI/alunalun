import type { Map } from "maplibre-gl";
import maplibregl from "maplibre-gl";

/**
 * Manages user location visualization using MapLibre native layers
 * This provides smooth, GPU-accelerated rendering without DOM jittering
 */
export class LocationLayer {
  private map: Map;
  private sourceId = "user-location-source";
  private circleLayerId = "user-location-accuracy-circle";
  private dotLayerId = "user-location-dot";
  private pulseLayerId = "user-location-pulse";
  private marker: maplibregl.Marker | null = null;

  constructor(map: Map) {
    this.map = map;
  }

  /**
   * Initialize the location layers
   */
  private initializeLayers(): void {
    // Source should already exist with data from updateLocation
    
    // Add accuracy circle layer (rendered first, below the dot)
    if (!this.map.getLayer(this.circleLayerId)) {
      this.map.addLayer({
        id: this.circleLayerId,
        type: "circle",
        source: this.sourceId,
        filter: ["==", "$type", "Point"],
        paint: {
          "circle-radius": [
            "case",
            ["has", "accuracy"],
            [
              "interpolate",
              ["exponential", 2],
              ["zoom"],
              0,
              0,
              20,
              [
                "*",
                ["coalesce", ["get", "accuracy"], 50],
                [
                  "/",
                  1,
                  [
                    "*",
                    [
                      "cos",
                      [
                        "*",
                        ["coalesce", ["get", "latitude"], 0],
                        ["/", Math.PI, 180],
                      ],
                    ],
                    ["/", 156543.03392, ["^", 2, ["zoom"]]],
                  ],
                ],
              ],
            ],
            0, // Default radius when no accuracy data
          ],
          "circle-color": "rgba(66, 133, 244, 0.15)",
          "circle-stroke-color": "rgba(66, 133, 244, 0.3)",
          "circle-stroke-width": 1,
          "circle-pitch-alignment": "map",
        },
      });
    }

    // Add pulsing animation layer
    if (!this.map.getLayer(this.pulseLayerId)) {
      this.map.addLayer({
        id: this.pulseLayerId,
        type: "circle",
        source: this.sourceId,
        filter: ["==", "$type", "Point"],
        paint: {
          "circle-radius": ["interpolate", ["linear"], ["zoom"], 10, 5, 20, 15],
          "circle-color": "rgba(66, 133, 244, 0.5)",
          "circle-opacity": 0,
          "circle-pitch-alignment": "map",
        },
      });

      // Animate the pulse
      this.animatePulse();
    }

    // Add location dot layer (rendered on top)
    if (!this.map.getLayer(this.dotLayerId)) {
      this.map.addLayer({
        id: this.dotLayerId,
        type: "circle",
        source: this.sourceId,
        filter: ["==", "$type", "Point"],
        paint: {
          "circle-radius": ["interpolate", ["linear"], ["zoom"], 10, 4, 20, 10],
          "circle-color": "#4285F4",
          "circle-stroke-color": "#ffffff",
          "circle-stroke-width": [
            "interpolate",
            ["linear"],
            ["zoom"],
            10,
            2,
            20,
            3,
          ],
          "circle-pitch-alignment": "map",
        },
      });
    }
  }

  /**
   * Animate the pulse effect
   */
  private animatePulse(): void {
    let radius = 5;
    let opacity = 1;
    let growing = true;

    const frame = () => {
      if (!(this.map && this.map.getLayer(this.pulseLayerId))) {
        return;
      }

      if (growing) {
        radius += 0.5;
        opacity -= 0.02;
        if (radius >= 25) {
          growing = false;
        }
      } else {
        radius = 5;
        opacity = 1;
        growing = true;
      }

      this.map.setPaintProperty(this.pulseLayerId, "circle-radius", radius);
      this.map.setPaintProperty(
        this.pulseLayerId,
        "circle-opacity",
        Math.max(0, opacity),
      );

      requestAnimationFrame(frame);
    };

    requestAnimationFrame(frame);
  }

  /**
   * Update the user's location on the map
   */
  updateLocation(latitude: number, longitude: number, accuracy: number): void {
    // Create source with initial data if not exists
    if (!this.map.getSource(this.sourceId)) {
      this.map.addSource(this.sourceId, {
        type: "geojson",
        data: {
          type: "FeatureCollection",
          features: [
            {
              type: "Feature",
              geometry: {
                type: "Point",
                coordinates: [longitude, latitude],
              },
              properties: {
                accuracy,
                latitude,
              },
            },
          ],
        },
      });
      // Now initialize layers with data already present
      this.initializeLayers();
    } else {
      // Update existing source
      const source = this.map.getSource(
        this.sourceId,
      ) as maplibregl.GeoJSONSource;
      if (source) {
        source.setData({
          type: "FeatureCollection",
          features: [
            {
              type: "Feature",
              geometry: {
                type: "Point",
                coordinates: [longitude, latitude],
              },
              properties: {
                accuracy,
                latitude,
              },
            },
          ],
        });
      }
    }

    // Add or update popup marker (optional, for showing accuracy text)
    if (this.marker) {
      this.marker.setLngLat([longitude, latitude]);
    } else {
      // Create invisible marker just for the popup
      const el = document.createElement("div");
      el.style.display = "none";

      this.marker = new maplibregl.Marker({ element: el })
        .setLngLat([longitude, latitude])
        .setPopup(
          new maplibregl.Popup({ offset: 25 }).setText(
            `Your location (±${Math.round(accuracy)}m)`,
          ),
        )
        .addTo(this.map);
    }
  }

  /**
   * Remove the location from the map
   */
  removeLocation(): void {
    // Remove layers
    if (this.map.getLayer(this.dotLayerId)) {
      this.map.removeLayer(this.dotLayerId);
    }
    if (this.map.getLayer(this.pulseLayerId)) {
      this.map.removeLayer(this.pulseLayerId);
    }
    if (this.map.getLayer(this.circleLayerId)) {
      this.map.removeLayer(this.circleLayerId);
    }

    // Remove source
    if (this.map.getSource(this.sourceId)) {
      this.map.removeSource(this.sourceId);
    }

    // Remove marker
    if (this.marker) {
      this.marker.remove();
      this.marker = null;
    }
  }

  /**
   * Get the marker instance (for popup control)
   */
  getMarker(): maplibregl.Marker | null {
    return this.marker;
  }
}

/**
 * Create a smooth location layer for the user's position
 */
export function createLocationLayer(
  map: Map,
  latitude: number,
  longitude: number,
  accuracy: number,
): LocationLayer {
  const layer = new LocationLayer(map);
  layer.updateLocation(latitude, longitude, accuracy);
  return layer;
}
