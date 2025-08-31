/**
 * Comprehensive MapLibre GL JS configuration options
 * 80/20 Rule: Common options are active, advanced options are commented
 *
 * @see https://maplibre.org/maplibre-gl-js/docs/API/classes/Map/
 */

/**
 * Extended map configuration with all MapLibre options documented
 */
export interface MapOptions {
  // Style configuration (wrapped for our use)
  style: {
    preset: "liberty" | "bright" | "positron" | "custom";
    customUrl?: string;
    // labels?: boolean;  // Future: Toggle place labels
    // pois?: boolean;     // Future: Toggle POI markers
  };

  // Viewport settings (ACTIVE - 80% use case)
  viewport: {
    center: [number, number]; // [lng, lat]
    zoom: number; // 0-24 scale
    bearing?: number; // 0-360 degrees rotation
    pitch?: number; // 0-85 degrees tilt
    // bounds?: LngLatBounds;  // Future: Fit to specific area
  };

  // Interaction controls (ACTIVE - 80% use case)
  interactions: {
    scrollZoom: boolean; // Mouse wheel zoom
    dragPan: boolean; // Click and drag to pan
    dragRotate: boolean; // Right-click drag to rotate
    doubleClickZoom: boolean; // Double-click to zoom in
    touchZoomRotate: boolean; // Pinch and rotate on mobile
    keyboard: boolean; // Arrow keys and +/- for navigation

    // cooperativeGestures?: boolean;  // Future: Require Ctrl/Cmd for scroll zoom
    // boxZoom?: boolean;              // Future: Shift+drag to zoom to box
    // touchPitch?: boolean;           // Future: Two-finger drag for pitch
  };

  // Built-in controls (ACTIVE - 80% use case)
  controls: {
    navigation: boolean; // Zoom +/- and compass
    scale: boolean; // Distance scale bar
    attribution: boolean | { compact: boolean }; // Map data attribution

    // fullscreen?: boolean;    // Future: Fullscreen toggle button
    // geolocate?: boolean;     // Already custom implemented
    // terrain?: boolean;       // Future: 3D terrain toggle
  };

  // Performance settings (DOCUMENTED - 20% use case)
  performance: {
    maxZoom: number; // Maximum allowed zoom (default: 22)
    minZoom: number; // Minimum allowed zoom (default: 0)
    maxPitch: number; // Maximum tilt angle (default: 85)

    // renderWorldCopies?: boolean;     // Show multiple world copies when zoomed out (default: true)
    // crossSourceCollisions?: boolean;  // Prevent symbol collisions across sources (default: true)
    // fadeDuration?: number;           // Label fade animation duration in ms (default: 300)
    // antialias?: boolean;            // Smooth edges (default: false, performance cost)
    // refreshExpiredTiles?: boolean;   // Reload expired tiles (default: true)
    // maxTileCacheSize?: number;      // Max tiles in cache (default: null)
    // localIdeographFontFamily?: string; // Font for CJK characters
    // transformRequest?: Function;     // Modify resource requests
    // collectResourceTiming?: boolean; // Performance metrics (default: false)
    // optimizeForTerrain?: boolean;   // Optimize for 3D terrain (default: true)
    // maxParallelImageRequests?: number; // Parallel image loads (default: 16)
  };

  // Advanced map options (DOCUMENTED - rarely used)
  advanced?: {
    // hash?: boolean;          // Sync map state with URL hash
    // bearingSnap?: number;    // Snap bearing to north (default: 7 degrees)
    // clickTolerance?: number; // Pixels for click detection (default: 3)
    // pitchWithRotate?: boolean; // Allow pitch during rotation (default: true)
    // preserveDrawingBuffer?: boolean; // Allow canvas export (default: false)
    // failIfMajorPerformanceCaveat?: boolean; // Fail on slow GPU (default: false)
    // trackResize?: boolean;   // Auto-resize with container (default: true)
    // center?: [number, number]; // Alternative to viewport.center
    // zoom?: number;           // Alternative to viewport.zoom
    // bounds?: LngLatBounds;   // Initial bounds to fit
    // fitBoundsOptions?: object; // Options for bounds fitting
    // locale?: object;         // Localization settings
    // testMode?: boolean;      // Suppress GL errors in tests
  };
}

/**
 * Default configuration (80% use case - sensible defaults)
 */
export const DEFAULT_MAP_OPTIONS: MapOptions = {
  // Style (using our wrapper)
  style: {
    preset: "liberty", // Modern, colorful OpenFreeMap style
  },

  // Viewport
  viewport: {
    center: [106.8001307, -6.2099592], // DPR RI, Jakarta
    zoom: 15, // Neighborhood-level view (good detail of DPR complex)
    bearing: 0, // North-facing
    pitch: 0, // Top-down view
  },

  // Interactions (all enabled for rich UX)
  interactions: {
    scrollZoom: true,
    dragPan: true,
    dragRotate: true,
    doubleClickZoom: true,
    touchZoomRotate: true,
    keyboard: true, // Built-in keyboard navigation
  },

  // Controls
  controls: {
    navigation: true,
    scale: true,
    attribution: { compact: true },
  },

  // Performance
  performance: {
    minZoom: 1,
    maxZoom: 18,
    maxPitch: 85, // Allow dramatic 3D viewing angles
  },
};

/**
 * Configuration presets for different use cases
 */
export const MAP_PRESETS = {
  // Minimal UI for embedded maps
  minimal: {
    ...DEFAULT_MAP_OPTIONS,
    style: { preset: "positron" as const },
    controls: {
      navigation: false,
      scale: false,
      attribution: { compact: true },
    },
    interactions: {
      ...DEFAULT_MAP_OPTIONS.interactions,
      dragRotate: false,
      keyboard: false,
    },
  },

  // Standard interactive map
  standard: DEFAULT_MAP_OPTIONS,

  // Optimized for data visualization
  dataViz: {
    ...DEFAULT_MAP_OPTIONS,
    style: { preset: "bright" as const },
    interactions: {
      ...DEFAULT_MAP_OPTIONS.interactions,
      dragRotate: false, // Keep north-up for clarity
    },
    performance: {
      ...DEFAULT_MAP_OPTIONS.performance,
      maxPitch: 0, // Force 2D view
    },
  },

  // Full-featured interactive map
  interactive: {
    ...DEFAULT_MAP_OPTIONS,
    style: { preset: "liberty" as const },
    controls: {
      navigation: true,
      scale: true,
      attribution: { compact: false },
      // fullscreen: true,  // Future: Add fullscreen button
      // terrain: true,     // Future: 3D terrain control
    },
    performance: {
      ...DEFAULT_MAP_OPTIONS.performance,
      // renderWorldCopies: true,  // Future: Seamless world wrapping
      // crossSourceCollisions: true,  // Future: Better label placement
    },
  },
} as const;

/**
 * Merge user options with defaults
 */
export function mergeMapOptions(
  userOptions?: Partial<MapOptions>,
  preset: keyof typeof MAP_PRESETS = "standard",
): MapOptions {
  const baseOptions = MAP_PRESETS[preset];

  if (!userOptions) {
    return baseOptions;
  }

  return {
    ...baseOptions,
    ...userOptions,
    style: {
      ...baseOptions.style,
      ...userOptions.style,
    },
    viewport: {
      ...baseOptions.viewport,
      ...userOptions.viewport,
    },
    interactions: {
      ...baseOptions.interactions,
      ...userOptions.interactions,
    },
    controls: {
      ...baseOptions.controls,
      ...userOptions.controls,
    },
    performance: {
      ...baseOptions.performance,
      ...userOptions.performance,
    },
  };
}
