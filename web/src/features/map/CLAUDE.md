# Map Feature Documentation

## Architecture Overview

The map feature provides a MapLibre-based map renderer with a sophisticated overlay system that supports both WebGL (for performance) and DOM (for flexibility) rendering.

## Core Components

### Map Rendering
- **MapRenderer**: Main component that renders the MapLibre map
- **MapControls**: UI controls for map navigation
- **LocateButton**: User geolocation with one-shot positioning

### Overlay System
The overlay system solves the fundamental problem of DOM/WebGL synchronization:

#### Problem
- DOM elements overlaid on WebGL canvas cause jittering during map interactions
- Different frame rates between DOM updates and WebGL rendering
- Sub-pixel positioning errors during zoom/pan operations

#### Solution
Three-tier rendering architecture:

1. **WebGL Layer** (Performance Critical)
   - User location dots
   - Large numbers of markers (>100)
   - Heatmaps and clusters
   - Uses `LocationLayer` class for perfect sync

2. **DOM Layer** (Interactive UI)
   - Control panels
   - Popups and tooltips
   - External React components
   - Uses `OverlayProvider` and `OverlayContainer`

3. **Canvas Layer** (Future)
   - Custom graphics
   - Smooth animations
   - HUD elements

## API Usage

### Internal WebGL Overlays
```typescript
// For performance-critical overlays like user location
const locationLayer = new LocationLayer(map);
locationLayer.updateLocation(lat, lng, accuracy);
```

### External DOM Overlays
```typescript
// External features can add React components
const { addOverlay } = useMapOverlays();
const overlayId = addOverlay(
  <CustomComponent />,
  { position: { lat, lng }, zIndex: 100 }
);
```

### Block Composition
```typescript
// Blocks compose map with overlays
<MapRenderer>
  <ExternalFeatureOverlays />
</MapRenderer>
```

## Key Design Decisions

### 1. User Location in Map Feature
User location stays within the map feature because:
- It's a core map functionality
- Needs tight integration with map controls
- Requires WebGL for smooth rendering

### 2. Overlay System Architecture
- Map feature owns rendering concerns
- External features provide pure React components
- Map decides optimal renderer (WebGL vs DOM)
- No cross-feature imports required

### 3. WebGL for User Location
Switched from DOM markers to WebGL layers because:
- Eliminates jittering during map interactions
- 50-100x better performance
- Perfect synchronization with map transforms
- Smooth 60fps even on mobile

## File Structure
```
features/map/
├── components/           # React components
│   ├── map-renderer.tsx # Main map component
│   ├── map-controls.tsx # UI controls
│   └── locate-button.tsx # Geolocation button
├── hooks/               # Custom hooks
│   ├── use-map-instance.ts
│   ├── use-map-controls.ts
│   └── use-geolocation.ts
├── lib/                 # Utilities
│   ├── location-layer.ts # WebGL location rendering
│   ├── location-utils.ts # Helper functions
│   ├── config.ts        # Map configuration
│   └── styles.ts        # Styling constants
├── overlays/            # Overlay system
│   ├── overlay-container.tsx
│   ├── overlay-context.tsx
│   └── index.ts
├── constants/           # Configuration
│   └── map-config.ts
└── index.ts            # Public API

blocks/
└── interactive-map/     # Composition example
    └── interactive-map.tsx
```

## Performance Considerations

### WebGL Layers
- Use for >50 markers
- Automatic clustering at scale
- GPU-accelerated rendering
- No DOM manipulation

### DOM Overlays
- Use for <50 interactive elements
- Full React component support
- Event handling capabilities
- CSS styling flexibility

## Future Enhancements

### Planned
- Canvas renderer for custom graphics
- Smart renderer selection based on device capabilities
- Automatic promotion/demotion between renderers
- AR/VR overlay support

### Considerations
- Performance monitoring for renderer switching
- Lazy loading for overlay components
- WebWorker for heavy calculations
- Offline map tile caching

## Testing Notes

### Critical Test Cases
1. Rapid zoom/pan with location marker visible
2. Multiple overlay types rendered simultaneously
3. External feature overlay integration
4. Mobile device performance
5. Memory leak prevention during overlay lifecycle

### Known Issues
- Console warnings "Expected value to be of type number, but found null instead" appear when loading OpenFreeMap tiles. These are harmless warnings from MapLibre's Web Workers processing vector tiles where style filter expressions expect numeric values but encounter nulls in the tile data. They don't affect functionality and can be safely ignored.

## Troubleshooting: Container Height Requirements

### Problem
MapLibre shows "Container has zero dimensions! Map cannot render" error.

### Root Cause
MapLibre GL requires a container with **explicit dimensions** (width/height > 0). When using percentage heights (`h-full`/`height: 100%`), ALL parent elements in the DOM chain must have explicit heights.

### Solution
Ensure complete height chain from `html` to map container:

1. **Global styles** (`globals.css`):
```css
html, body { @apply h-full; }
#app { @apply h-full; }
```

2. **Page wrapper**:
```jsx
<div className="h-screen w-screen">
  <Map />
</div>
```

3. **Map container**:
```jsx
// Use absolute positioning to fill parent
style={{ position: "absolute", top: 0, left: 0, right: 0, bottom: 0 }}
```

### Best Practices
- Use **viewport units** (`h-screen`/`100vh`) for full-screen maps
- Use **absolute positioning** when parent has defined dimensions
- Avoid percentage heights unless all parents have explicit heights
- Add debug logging to verify container dimensions before map init

## Troubleshooting: WebGL Canvas Pointer Events

### Problem
MapLibre GL's canvas element intercepts all pointer events, making overlaid DOM controls (buttons, panels) unclickable even when they appear visually on top. This is a fundamental issue with WebGL/Canvas layers that capture mouse events for map interactions (pan, zoom, rotate).

### Symptoms
- Controls appear visually on top but can't be clicked
- Playwright/browser automation shows "canvas intercepts pointer events" errors
- Z-index alone doesn't solve the problem
- Mouse hover effects may work but clicks don't register

### Root Cause
1. **MapLibre's event system** - The WebGL canvas has its own event handling for map interactions
2. **Event bubbling** - Canvas captures events before they reach DOM elements
3. **Stacking context conflicts** - Complex component hierarchies can create unexpected stacking contexts

### Solution

#### 1. Use Extreme Z-Index Values
```tsx
<div 
  className="absolute top-4 right-4 z-[999999]"
  style={{ 
    zIndex: 999999,
    pointerEvents: 'auto'
  }}
>
  {/* Controls here */}
</div>
```

#### 2. Simplify DOM Structure
- Remove unnecessary wrapper components
- Implement controls directly inline when possible
- Avoid deep component nesting for controls

#### 3. Explicit Pointer Events Management
```tsx
// Container: pointer-events-none to pass through
<div className="absolute inset-0 pointer-events-none z-10">
  // Control panel: pointer-events-auto to capture
  <div className="pointer-events-auto">
    {/* Buttons here */}
  </div>
</div>
```

#### 4. Create New Stacking Context
```tsx
className={cn(
  "isolate", // Creates new stacking context
  "relative z-[999999]"
)}
```

### Best Practices
1. **Test with browser DevTools** - Use element inspector to verify z-index stacking
2. **Add debug logging** - Track mouse events to identify where they're being captured
3. **Use Playwright for testing** - Automated tests reveal pointer-event issues clearly
4. **Keep controls outside map container** when possible
5. **Use inline styles for critical z-index** - Ensures specificity

### Example Working Structure
```tsx
<div className="relative h-full w-full">
  {/* Map at base level */}
  <MapBase className="h-full w-full" />
  
  {/* Controls at top level with extreme z-index */}
  <div style={{ position: 'absolute', zIndex: 999999 }}>
    <button onClick={handleClick}>Works!</button>
  </div>
</div>
```

## References
- MapLibre GL JS: https://maplibre.org/
- OpenFreeMap tiles: https://openfreemap.org/
- Geolocation API: https://developer.mozilla.org/en-US/docs/Web/API/Geolocation_API