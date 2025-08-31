# Custom Sprites and Icons for OpenFreeMap

This document explains how to create and integrate custom sprites (icons) on top of OpenFreeMap's default sprites, perfect for adding Indonesian landmarks or organization-specific icons.

## Overview

OpenFreeMap provides default sprites from `assets.openfreemap.com`, but you can layer custom sprites on top for:
- Indonesian landmarks (DPR, Monas, Borobudur, etc.)
- Organization-specific locations
- Custom POI categories
- Brand-specific icons

## Current Setup

Our map styles automatically load sprites from OpenFreeMap:
- **Liberty**: `https://assets.openfreemap.com/sprites/liberty`
- **Bright**: `https://assets.openfreemap.com/sprites/bright`
- **Positron**: `https://assets.openfreemap.com/sprites/positron`

These provide standard icons for common POIs, but may not include Indonesia-specific landmarks.

## Creating Custom Sprites

### Step 1: Prepare SVG Icons

Create SVG files for each icon you need:

```
custom-sprites/
├── dpr-ri.svg           # DPR RI building
├── monas.svg            # National Monument
├── borobudur.svg        # Borobudur Temple
├── mosque.svg           # Mosque icon
├── traditional-house.svg # Traditional house
└── volcano.svg          # Volcano icon
```

**SVG Guidelines:**
- Size: 24x24px or 48x48px for clarity
- Colors: Use solid colors (they can be modified at runtime)
- Format: Clean paths, no raster images embedded
- Naming: Use kebab-case, descriptive names

### Step 2: Generate Sprite Sheet

Use [Spreet](https://github.com/flother/spreet) to generate sprite sheets from SVGs:

```bash
# Install Spreet
npm install -g @flother/spreet

# Generate standard resolution sprite
spreet custom-sprites/ sprites/indonesian-landmarks

# Generate retina resolution sprite
spreet --retina custom-sprites/ sprites/indonesian-landmarks
```

This creates:
- `indonesian-landmarks.png` - Sprite image
- `indonesian-landmarks.json` - Sprite index
- `indonesian-landmarks@2x.png` - Retina sprite image  
- `indonesian-landmarks@2x.json` - Retina sprite index

### Step 3: Host Sprites

Option A: **In Your Repository** (Recommended for small sets)
```
public/
└── assets/
    └── sprites/
        ├── indonesian-landmarks.png
        ├── indonesian-landmarks.json
        ├── indonesian-landmarks@2x.png
        └── indonesian-landmarks@2x.json
```

Option B: **CDN/Object Storage** (For larger sets or multiple projects)
- Upload to S3, Cloudflare R2, or similar
- Ensure CORS headers are properly configured

### Step 4: Integrate with Map Styles

#### Method 1: Merge at Build Time

Combine custom sprites with OpenFreeMap sprites during build:

```typescript
// scripts/merge-sprites.ts
import { mergeSprites } from 'sprite-merger'; // hypothetical library

const merged = await mergeSprites([
  'https://assets.openfreemap.com/sprites/liberty',
  '/assets/sprites/indonesian-landmarks'
]);
```

#### Method 2: Runtime Layer Addition

Add custom icon layer on top of base map:

```typescript
// After map loads
map.on('load', () => {
  // Load custom sprite
  const spriteUrl = '/assets/sprites/indonesian-landmarks';
  
  // Add each icon from the sprite
  fetch(`${spriteUrl}.json`)
    .then(res => res.json())
    .then(icons => {
      Object.entries(icons).forEach(([name, data]) => {
        if (!map.hasImage(name)) {
          // Load and add the icon
          map.loadImage(`${spriteUrl}.png`, (error, image) => {
            if (!error) {
              map.addImage(name, image, {
                pixelRatio: 1,
                sdf: false // or true for recolorable icons
              });
            }
          });
        }
      });
    });
});
```

#### Method 3: Custom Style with Multiple Sprites

MapLibre GL JS v3+ supports multiple sprite sources:

```typescript
const style = {
  version: 8,
  sprite: [
    { id: 'default', url: 'https://assets.openfreemap.com/sprites/liberty' },
    { id: 'custom', url: '/assets/sprites/indonesian-landmarks' }
  ],
  // ... rest of style
};
```

## Using Custom Icons

### In Map Layers

Reference custom icons in style layers:

```typescript
map.addLayer({
  id: 'indonesian-landmarks',
  type: 'symbol',
  source: 'landmarks',
  layout: {
    'icon-image': [
      'case',
      ['==', ['get', 'type'], 'dpr'], 'dpr-ri',
      ['==', ['get', 'type'], 'monument'], 'monas',
      ['==', ['get', 'type'], 'temple'], 'borobudur',
      'marker' // fallback icon
    ],
    'icon-size': 1.5,
    'icon-allow-overlap': true
  }
});
```

### With Markers

Use custom icons with markers:

```typescript
// Create custom marker element
const el = document.createElement('div');
el.className = 'custom-marker';
el.style.backgroundImage = 'url(/assets/sprites/dpr-ri.png)';
el.style.width = '32px';
el.style.height = '32px';

new maplibregl.Marker(el)
  .setLngLat([106.8001, -6.2099])
  .addTo(map);
```

## Indonesian Landmark Examples

### Key Government Buildings
```javascript
const governmentLandmarks = [
  { name: 'DPR RI', type: 'dpr', coords: [106.8001, -6.2099] },
  { name: 'Istana Merdeka', type: 'palace', coords: [106.8242, -6.1702] },
  { name: 'Mahkamah Konstitusi', type: 'court', coords: [106.8335, -6.1693] }
];
```

### Cultural Sites
```javascript
const culturalSites = [
  { name: 'Monas', type: 'monument', coords: [106.8272, -6.1753] },
  { name: 'Borobudur', type: 'temple', coords: [110.2038, -7.6079] },
  { name: 'Prambanan', type: 'temple', coords: [110.4914, -7.7520] }
];
```

### Natural Landmarks
```javascript
const naturalLandmarks = [
  { name: 'Gunung Bromo', type: 'volcano', coords: [112.9530, -7.9425] },
  { name: 'Danau Toba', type: 'lake', coords: [98.8278, 2.6845] },
  { name: 'Raja Ampat', type: 'marine', coords: [130.8779, -0.2330] }
];
```

## Style Customization with Maputnik

[Maputnik](https://maputnik.github.io/) is a visual editor for map styles:

1. **Open Maputnik**: https://maputnik.github.io/
2. **Load OpenFreeMap Style**: Use style URL from our config
3. **Customize**:
   - Adjust colors, sizes, densities
   - Add/remove layers
   - Modify label styles
4. **Export**: Download modified style JSON
5. **Use**: Host the JSON and reference in your map config

### Quick Customizations

```javascript
// Remove all labels
style.layers = style.layers.filter(layer => 
  !layer.layout?.['text-field']
);

// Change all water to different color
style.layers.forEach(layer => {
  if (layer.id.includes('water')) {
    layer.paint['fill-color'] = '#004d99';
  }
});

// Increase POI icon sizes
style.layers.forEach(layer => {
  if (layer.type === 'symbol' && layer.layout?.['icon-image']) {
    layer.layout['icon-size'] = 1.5;
  }
});
```

## Performance Considerations

1. **Sprite Size**: Keep total sprite sheet under 512x512px for mobile
2. **Icon Count**: Limit to ~100 icons per sprite sheet
3. **Format**: Use PNG for sprites, optimize with tools like `pngquant`
4. **Caching**: Set proper cache headers for sprite files
5. **Loading**: Load custom sprites after map initialization

## Troubleshooting

### Missing Icons in Console

OpenFreeMap's Liberty style references some icons that aren't in the Maki sprite set. Common missing icons and their automatic mappings:

| Requested Icon | Mapped To | Description |
|---------------|-----------|-------------|
| `office` | `commercial` | Office buildings |
| `atm` | `bank` | ATM locations |
| `sports_centre` | `pitch` | Sports facilities |
| `swimming_pool` | `swimming` | Pool/aquatic centers |
| `gate` | `entrance` | Gates and barriers |
| `lift_gate` | `entrance` | Lift gates |
| `bollard` | `barrier` | Traffic bollards |
| `brownfield` | `industrial` | Brownfield sites |

These mappings are handled automatically in `use-map-instance.ts`. To add custom icons for these instead:

1. Create SVG icons for the missing items
2. Generate sprites using Spreet
3. Override the mappings in the sprite handler

### Icons Not Appearing
- Check browser console for 404 errors
- Verify CORS headers if hosting externally
- Ensure sprite JSON matches PNG file structure

### Blurry Icons
- Provide @2x versions for retina displays
- Check icon-size scaling in style

### Performance Issues
- Reduce sprite sheet size
- Use SDF (Signed Distance Field) for simple icons
- Consider icon clustering at low zoom levels

## Future Enhancements

1. **Dynamic Icon Loading**: Load sprites based on viewport
2. **Icon Theming**: Switch icon sets based on map style
3. **Localized Icons**: Different icons for different regions
4. **Animated Sprites**: For special landmarks or events

## Resources

- [Spreet Documentation](https://github.com/flother/spreet)
- [MapLibre Sprite Specification](https://maplibre.org/maplibre-style-spec/sprite/)
- [Maputnik Editor](https://maputnik.github.io/)
- [OpenFreeMap Styles Repository](https://github.com/hyperknot/openfreemap-styles)
- [SVG Icon Resources](https://thenounproject.com/)

## Implementation Status

- ✅ OpenFreeMap sprites integrated
- ✅ Fallback handling for missing icons
- ⏳ Custom Indonesian landmark sprites (future)
- ⏳ Runtime sprite merging (future)
- ⏳ Maputnik style customization (as needed)