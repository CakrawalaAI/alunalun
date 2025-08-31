/**
 * OpenFreeMap style configurations
 *
 * OpenFreeMap provides free, open-source map tiles and styles
 * @see https://openfreemap.org/
 * @see https://github.com/openfreemap/styles
 */

export interface MapStyle {
  name: string;
  description: string;
  url: string;
  preview?: string;
  sprite?: string;
  glyphs?: string;
  features: {
    labels: boolean;
    pois: boolean;
    terrain3d: boolean;
    buildings3d: boolean;
    traffic?: boolean;
    transit?: boolean;
  };
}

/**
 * Available OpenFreeMap styles
 * Each style is optimized for different use cases
 *
 * Note: Sprites and glyphs are loaded from assets.openfreemap.com
 * These URLs are automatically included when the style is fetched
 */
export const MAP_STYLES = {
  liberty: {
    name: "Liberty",
    description:
      "Modern, colorful style with good contrast. Best for general use.",
    url: "https://tiles.openfreemap.org/styles/liberty",
    // Sprites and glyphs are included in the style JSON from OpenFreeMap
    sprite: "https://assets.openfreemap.com/sprites/liberty",
    glyphs: "https://assets.openfreemap.com/fonts/{fontstack}/{range}.pbf",
    features: {
      labels: true,
      pois: true,
      terrain3d: false,
      buildings3d: false,
    },
  },

  bright: {
    name: "Bright",
    description:
      "Clean, light theme with minimal colors. Good for data overlays.",
    url: "https://tiles.openfreemap.org/styles/bright",
    sprite: "https://assets.openfreemap.com/sprites/bright",
    glyphs: "https://assets.openfreemap.com/fonts/{fontstack}/{range}.pbf",
    features: {
      labels: true,
      pois: true,
      terrain3d: false,
      buildings3d: false,
    },
  },

  positron: {
    name: "Positron",
    description: "Minimal grayscale style. Perfect for data visualization.",
    url: "https://tiles.openfreemap.org/styles/positron",
    sprite: "https://assets.openfreemap.com/sprites/positron",
    glyphs: "https://assets.openfreemap.com/fonts/{fontstack}/{range}.pbf",
    features: {
      labels: true,
      pois: false,
      terrain3d: false,
      buildings3d: false,
    },
  },

  // Future styles (when available from OpenFreeMap)
  // dark: {
  //   name: "Dark",
  //   description: "Dark theme for reduced eye strain",
  //   url: "https://tiles.openfreemap.org/styles/dark",
  //   features: {
  //     labels: true,
  //     pois: true,
  //     terrain3d: false,
  //     buildings3d: false,
  //   },
  // },

  // "3d": {
  //   name: "3D",
  //   description: "Style with 3D building extrusions",
  //   url: "https://tiles.openfreemap.org/styles/3d",
  //   features: {
  //     labels: true,
  //     pois: true,
  //     terrain3d: false,
  //     buildings3d: true,
  //   },
  // },

  // satellite: {
  //   name: "Satellite",
  //   description: "Satellite imagery with labels",
  //   url: "custom-satellite-source",
  //   features: {
  //     labels: true,
  //     pois: false,
  //     terrain3d: false,
  //     buildings3d: false,
  //   },
  // },
} as const;

/**
 * Style modification helpers
 * These will be useful when we implement runtime style changes
 */
export const STYLE_MODIFICATIONS = {
  // Remove all text labels
  removeLabels: {
    filter: ["!", ["has", "text-field"]],
  },

  // Remove POI (Points of Interest) markers
  removePOIs: {
    filter: ["!", ["in", "$type", "Point"]],
  },

  // Adjust label density
  labelDensity: {
    sparse: { "text-padding": 20, "text-size": 10 },
    normal: { "text-padding": 2, "text-size": 12 },
    dense: { "text-padding": 0, "text-size": 14 },
  },

  // Color themes (future use with style manipulation)
  colorThemes: {
    grayscale: { saturation: -100 },
    muted: { saturation: -50 },
    vibrant: { saturation: 20 },
  },
};

/**
 * Custom style URLs (for user-provided or self-hosted styles)
 */
export const CUSTOM_STYLE_SOURCES = {
  // Maputnik: Visual style editor
  // Users can create custom styles at https://maputnik.github.io/
  maputnik: "https://maputnik.github.io/",

  // Alternative tile providers (future consideration)
  // mapTiler: "https://api.maptiler.com/maps/{style}/style.json?key={key}",
  // stadia: "https://tiles.stadiamaps.com/styles/{style}.json",
  // jawg: "https://api.jawg.io/styles/{style}.json?access-token={token}",

  // Self-hosted style example
  // selfHosted: "/api/map-styles/{style}.json",
};

/**
 * Get style URL by preset name
 */
export function getStyleUrl(
  preset: keyof typeof MAP_STYLES | "custom",
  customUrl?: string,
): string {
  if (preset === "custom" && customUrl) {
    return customUrl;
  }

  const style = MAP_STYLES[preset as keyof typeof MAP_STYLES];
  if (!style) {
    console.warn(`Unknown style preset: ${preset}, falling back to liberty`);
    return MAP_STYLES.liberty.url;
  }

  return style.url;
}

/**
 * Style capabilities detection
 * Check what features a style supports
 */
export function getStyleCapabilities(styleUrl: string): MapStyle["features"] {
  // Match against known styles
  const knownStyle = Object.values(MAP_STYLES).find((s) => s.url === styleUrl);
  if (knownStyle) {
    return knownStyle.features;
  }

  // Default capabilities for unknown styles
  return {
    labels: true,
    pois: true,
    terrain3d: false,
    buildings3d: false,
  };
}

/**
 * OpenFreeMap Asset URLs
 * These are the public buckets for accessing map assets
 */
export const OPENFREEMAP_ASSETS = {
  // Main assets bucket - contains fonts, sprites, styles, versions
  assets: "https://assets.openfreemap.com",

  // Sprite URLs for each style (automatically included in style JSON)
  sprites: {
    liberty: "https://assets.openfreemap.com/sprites/liberty",
    bright: "https://assets.openfreemap.com/sprites/bright",
    positron: "https://assets.openfreemap.com/sprites/positron",
  },

  // Font glyphs URL pattern (shared across all styles)
  glyphs: "https://assets.openfreemap.com/fonts/{fontstack}/{range}.pbf",

  // Full planet downloads (weekly updates)
  planet: "https://btrfs.openfreemap.com",
};

/**
 * Custom Sprite Configuration
 * For adding custom icons on top of OpenFreeMap defaults
 * See CUSTOM_SPRITES.md for implementation details
 */
export const CUSTOM_SPRITE_CONFIG = {
  // Example: Add custom Indonesian landmark sprites
  // customSprite: "/assets/sprites/indonesian-landmarks",

  // Merge strategy: How to combine custom sprites with OpenFreeMap sprites
  // Options: "overlay" (custom on top), "replace" (custom only), "merge" (combine both)
  mergeStrategy: "overlay" as const,

  // Priority: When there's a naming conflict, which sprite wins
  // Higher number = higher priority
  priorities: {
    custom: 100,
    openfreemap: 50,
  },
};
