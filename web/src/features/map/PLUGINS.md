# MapLibre GL JS Plugin Ecosystem

This document catalogs available plugins and extensions for MapLibre GL JS that can enhance our map feature in the future.

## 🔍 Search & Geocoding

### maplibre-gl-geocoder
- **Purpose**: Add a search box for finding places
- **Install**: `bun add @maplibre/maplibre-gl-geocoder`
- **Use Case**: Let users search for addresses, cities, or POIs
- **Priority**: High - Very common user need
```typescript
import MaplibreGeocoder from '@maplibre/maplibre-gl-geocoder';
map.addControl(new MaplibreGeocoder({ api: 'mapbox' }));
```

### MapTiler Geocoding Control
- **Purpose**: Advanced place search with autocomplete
- **Install**: `bun add @maptiler/geocoding-control`
- **Use Case**: Professional geocoding with better international support
- **Priority**: Medium - When basic geocoder isn't enough

## ✏️ Drawing & Editing

### mapbox-gl-draw
- **Purpose**: Draw and edit shapes on the map
- **Install**: `bun add @mapbox/mapbox-gl-draw`
- **Use Case**: User-generated content, area selection, route planning
- **Priority**: High - Essential for interactive features
```typescript
import MapboxDraw from '@mapbox/mapbox-gl-draw';
const draw = new MapboxDraw();
map.addControl(draw);
```

### Terra Draw
- **Purpose**: Advanced drawing with snapping and constraints
- **Install**: `bun add terra-draw`
- **Use Case**: Professional GIS editing, precise geometry creation
- **Priority**: Low - Only for specialized apps

### maplibre-gl-measures
- **Purpose**: Measure distances and areas
- **Install**: `bun add maplibre-gl-measures`
- **Use Case**: Real estate, logistics, planning applications
- **Priority**: Medium - Useful utility feature

## 🎨 Enhanced UI

### maplibre-gl-compare
- **Purpose**: Side-by-side map comparison with swipe control
- **Install**: `bun add maplibre-gl-compare`
- **Use Case**: Before/after views, style comparisons
- **Priority**: Low - Specific use case
```typescript
import Compare from 'maplibre-gl-compare';
new Compare(map1, map2, '#comparison-container');
```

### maplibregl-minimap
- **Purpose**: Overview minimap in corner
- **Install**: `bun add maplibregl-minimap`
- **Use Case**: Context awareness in detailed views
- **Priority**: Low - Nice to have

### maplibre-gl-export
- **Purpose**: Export map to PDF, PNG, or SVG
- **Install**: `bun add maplibre-gl-export`
- **Use Case**: Reports, printing, sharing static maps
- **Priority**: Medium - Common business requirement

### maplibre-gl-basemaps
- **Purpose**: Switch between different basemap styles
- **Install**: Custom implementation needed
- **Use Case**: Let users choose map appearance
- **Priority**: Medium - Good for user preference

### maplibre-gl-opacity
- **Purpose**: Control layer opacity with slider
- **Install**: Custom implementation needed
- **Use Case**: Overlay comparison, data visualization
- **Priority**: Low - Specialized feature

## 📊 Data Visualization

### deck.gl
- **Purpose**: Advanced WebGL-powered data visualization
- **Install**: `bun add deck.gl`
- **Use Case**: Heatmaps, 3D data, large datasets, animations
- **Priority**: High - Powerful for data-heavy apps
```typescript
import { MapboxOverlay } from '@deck.gl/mapbox';
import { HeatmapLayer } from '@deck.gl/aggregation-layers';

const overlay = new MapboxOverlay({
  layers: [new HeatmapLayer({...})]
});
map.addControl(overlay);
```

### Turf.js
- **Purpose**: Geospatial analysis and processing
- **Install**: `bun add @turf/turf`
- **Use Case**: Buffer zones, intersections, spatial calculations
- **Priority**: Medium - Essential for GIS operations
```typescript
import * as turf from '@turf/turf';
const buffered = turf.buffer(point, 500, { units: 'meters' });
```

### maplibre-gl-inspect
- **Purpose**: Debug vector tiles and data sources
- **Install**: `bun add maplibre-gl-inspect`
- **Use Case**: Development, debugging data issues
- **Priority**: Low - Developer tool only

## 🎯 Specialized Features

### maplibre-gl-directions
- **Purpose**: Turn-by-turn routing and navigation
- **Install**: `bun add maplibre-gl-directions`
- **Use Case**: Navigation apps, delivery routes
- **Priority**: Medium - Depends on app purpose

### maplibre-gl-traffic
- **Purpose**: Real-time traffic overlay
- **Install**: Requires traffic data source
- **Use Case**: Navigation, logistics
- **Priority**: Low - Needs data subscription

### maplibre-gl-indoor
- **Purpose**: Indoor mapping and navigation
- **Install**: Custom implementation
- **Use Case**: Malls, airports, large buildings
- **Priority**: Low - Very specialized

## 📱 Mobile & Touch

### maplibre-gl-gesture-handling
- **Purpose**: Better mobile gesture controls
- **Install**: `bun add maplibre-gl-gesture-handling`
- **Use Case**: Improved mobile UX
- **Priority**: High - Important for mobile users
```typescript
import GestureHandling from 'maplibre-gl-gesture-handling';
map.addControl(new GestureHandling());
```

## 🎨 Sprites and Icons

### Custom Sprite Management
- **Purpose**: Add custom icons on top of OpenFreeMap defaults
- **Install**: See `CUSTOM_SPRITES.md` for detailed implementation
- **Use Case**: Indonesian landmarks, organization-specific POIs
- **Priority**: High - Essential for localized content
- **Current Status**: OpenFreeMap sprites integrated, custom sprites documented

### Spreet
- **Purpose**: Generate sprite sheets from SVG icons
- **Install**: `bun add -D @flother/spreet`
- **Use Case**: Convert SVG icons to MapLibre-compatible sprites
- **Priority**: High - Required for custom icons
```bash
spreet icons/ output/sprites
spreet --retina icons/ output/sprites  # For @2x version
```

### Maputnik
- **Purpose**: Visual style editor for MapLibre/Mapbox styles
- **URL**: https://maputnik.github.io/
- **Use Case**: Customize OpenFreeMap styles, adjust sprites/icons
- **Priority**: Medium - Useful for style customization
- **Features**:
  - Visual editing of all style properties
  - Sprite management and preview
  - Export modified styles as JSON
  - Works with OpenFreeMap styles

## 🔧 Developer Tools

### maplibre-gl-style-switcher
- **Purpose**: Runtime style switching UI
- **Install**: Custom implementation
- **Use Case**: Development, demos
- **Priority**: Low - Mainly for development

### maplibre-gl-fps
- **Purpose**: FPS counter for performance monitoring
- **Install**: Custom implementation
- **Use Case**: Performance optimization
- **Priority**: Low - Developer tool

## 🚀 Implementation Roadmap

### Phase 1: Essential (Next Sprint)
1. **Geocoding**: Basic search functionality
2. **Gesture Handling**: Better mobile experience
3. **Drawing**: Basic shape creation

### Phase 2: Enhanced (Q2 2025)
4. **Data Visualization**: deck.gl integration
5. **Export**: PDF/PNG generation
6. **Measurements**: Distance and area tools

### Phase 3: Advanced (Q3 2025)
7. **Routing**: Navigation features
8. **Comparison**: Side-by-side views
9. **Advanced Analysis**: Turf.js integration

### Phase 4: Specialized (Future)
10. **Indoor Mapping**: If needed
11. **Traffic Data**: If data source available
12. **Custom Plugins**: Business-specific needs

## 📝 Notes on Implementation

### Plugin Compatibility
- Most Mapbox GL JS plugins work with MapLibre GL JS
- Check version compatibility before installing
- Some plugins may need minor modifications

### Bundle Size Considerations
- Each plugin adds to bundle size
- Use dynamic imports for optional features
- Consider code splitting by route

### Performance Impact
- Test performance with each new plugin
- Monitor FPS and memory usage
- Some plugins (like deck.gl) need GPU consideration

### Example Dynamic Import
```typescript
// Load drawing plugin only when needed
async function enableDrawing() {
  const { default: MapboxDraw } = await import('@mapbox/mapbox-gl-draw');
  const draw = new MapboxDraw();
  map.addControl(draw);
}
```

## 🔗 Resources

- [MapLibre GL JS Plugins](https://maplibre.org/maplibre-gl-js/docs/plugins/)
- [Awesome MapLibre](https://github.com/maplibre/awesome-maplibre)
- [OpenFreeMap Styles](https://github.com/openfreemap/styles)
- [Maputnik Style Editor](https://maputnik.github.io/)
- [deck.gl Gallery](https://deck.gl/gallery)
- [Turf.js Docs](https://turfjs.org/)