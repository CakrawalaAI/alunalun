# UI Components Quick Reference

## Minimal Control Set

### Bottom-Right Floating Controls (40x40px circles)

```typescript
// All controls are icon-only, no text labels
interface ControlButton {
  size: 40 // Fixed 40x40px
  shape: 'circle'
  background: 'rgba(255, 255, 255, 0.9)'
  backdropFilter: 'blur(8px)'
  shadow: '0 2px 8px rgba(0,0,0,0.1)'
}
```

#### Control Icons
- **⊕** Zoom (expands to +/- on tap)
- **◎** Locate me (pulses when active)
- **↗** Sign in (anonymous state)
- **[A-Z]** User initial (authenticated state)

### HUD Toggle Zones

```
┌─────────────────────────┐
│  EDGE    CENTER   EDGE  │
│  ZONE    ZONE     ZONE  │
│  (35%)   (30%)    (35%) │
│                         │
│  Tap center: Hide HUD   │
│  Tap edge: Show HUD     │
└─────────────────────────┘
```

## Component Specifications

### 1. MinimalHUD
```tsx
// Bottom-right controls only
<div className="fixed bottom-4 right-4 flex flex-col gap-2">
  <ZoomControl />    // 40x40 circle
  <LocateControl />  // 40x40 circle  
  <AccountControl /> // 40x40 circle
</div>

// Top-left username (if authenticated)
<div className="fixed top-4 left-4 text-xs text-gray-500">
  {username}
</div>
```

### 2. PinCreationSheet
```tsx
// Desktop: Centered modal
<div className="fixed inset-0 flex items-end justify-center">
  <div className="bg-white rounded-t-2xl w-full max-w-md p-6">
    <textarea placeholder="What happened here?" />
    <div className="flex gap-2">
      <button>Cancel</button>
      <button>Drop Pin</button>
    </div>
  </div>
</div>

// Mobile: Bottom sheet
<div className="fixed bottom-0 left-0 right-0">
  {/* Same content, different position */}
</div>
```

### 3. PinDetailsView
```tsx
// Desktop: Floating panel
<div className="fixed bottom-8 right-8 w-96 bg-white rounded-xl shadow-xl">
  <div className="p-4">
    <div className="text-sm text-gray-600">@{author} · {time}</div>
    <div className="mt-2">{content}</div>
    <div className="mt-4 text-sm">💬 {commentCount} comments</div>
  </div>
</div>

// Mobile: Swipeable bottom sheet
<BottomSheet snapPoints={[0.1, 0.5, 0.9]}>
  {/* Same content structure */}
</BottomSheet>
```

## Animation Specifications

### Transitions
```css
/* HUD fade in/out */
.hud-transition {
  transition: opacity 300ms cubic-bezier(0.4, 0, 0.2, 1);
}

/* Bottom sheet slide */
.sheet-transition {
  transition: transform 300ms cubic-bezier(0.4, 0, 0.2, 1);
}

/* Pin creation indicator */
@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.6; transform: scale(1.1); }
}
```

### Interactive States
```css
/* Button hover (desktop only) */
.control-button:hover {
  background: rgba(255, 255, 255, 1);
  transform: scale(1.05);
}

/* Button active */
.control-button:active {
  transform: scale(0.95);
}

/* Locating pulse */
.locating {
  animation: pulse 2s infinite;
}
```

## Mobile Gesture Handlers

```typescript
// Bottom sheet swipe handler
const handleSwipe = (deltaY: number) => {
  const snapPoints = [0.1, 0.5, 0.9] // 10%, 50%, 90% of screen
  const currentHeight = sheetRef.current.offsetHeight
  const newHeight = currentHeight - deltaY
  
  // Snap to nearest point
  const nearestSnap = findNearestSnapPoint(newHeight)
  animateToHeight(nearestSnap)
}

// Map tap handler
const handleMapTap = (e: TapEvent) => {
  const center = isInCenterZone(e.x, e.y)
  
  if (center && hudVisible) {
    setHudVisible(false)
  } else if (!center && !hudVisible) {
    setHudVisible(true)
  } else if (user && !hasPin(e.x, e.y)) {
    startPinCreation(e.x, e.y)
  }
}
```

## Z-Index Hierarchy

```css
/* Layering order (bottom to top) */
.map-canvas { z-index: 0; }
.pin-markers { z-index: 1; }
.user-location { z-index: 2; }
.hud-controls { z-index: 10; }
.bottom-sheet { z-index: 20; }
.modal-backdrop { z-index: 30; }
.modal-content { z-index: 31; }
.toast-notifications { z-index: 40; }
```

## Responsive Breakpoints

```typescript
const breakpoints = {
  mobile: 0,     // 0-639px
  tablet: 640,   // 640-1023px
  desktop: 1024  // 1024px+
}

// Usage
const device = useDevice()

return device === 'mobile' 
  ? <BottomSheet /> 
  : <FloatingPanel />
```

## Color Palette

```css
:root {
  /* Minimal palette - mostly monochrome */
  --white: #ffffff;
  --white-90: rgba(255, 255, 255, 0.9);
  --black: #000000;
  --gray-500: #6b7280;
  --gray-100: #f3f4f6;
  
  /* Functional colors */
  --backdrop: rgba(0, 0, 0, 0.2);
  --error: #ef4444;
  --success: #10b981;
}
```

## Typography Scale

```css
.text-xs { font-size: 12px; }  /* Username indicator */
.text-sm { font-size: 14px; }  /* Metadata, timestamps */
.text-base { font-size: 16px; } /* Body text, forms */
.text-lg { font-size: 18px; }  /* Headers (rare) */

/* Single font family */
font-family: system-ui, -apple-system, sans-serif;
```

## Icon Set

Using Lucide React for consistency:
- `Plus` / `Minus` - Zoom controls
- `Locate` - GPS/location
- `LogIn` - Sign in arrow
- `X` - Close/dismiss
- `MessageCircle` - Comments
- `ChevronUp` / `ChevronDown` - Expand/collapse

## Loading States

```tsx
// Subtle inline loading
<span className="text-xs text-gray-500">
  {isLoading ? 'checking...' : null}
</span>

// Button loading
<button disabled={isLoading}>
  {isLoading ? <Spinner className="w-4 h-4 animate-spin" /> : <Icon />}
</button>

// Map loading
<div className="animate-pulse">
  Getting your location...
</div>
```

## Error Patterns

```tsx
// Inline validation
<input className={error && 'border-red-500'} />
<span className="text-xs text-red-500">{error}</span>

// Toast notifications (auto-dismiss)
<Toast message={error} duration={3000} />

// Location error
<button title={error ? `Error: ${error}` : 'Locate me'}>
  <Locate className={error && 'text-red-500'} />
</button>
```

## Accessibility Shortcuts

| Key | Action |
|-----|--------|
| `Esc` | Close modal/drawer |
| `Enter` | Submit form |
| `Tab` | Navigate controls |
| `Space` | Activate button |
| `+/-` | Zoom in/out |

## Testing Selectors

```tsx
// Data attributes for Playwright
<button data-testid="zoom-in" />
<button data-testid="locate-me" />
<button data-testid="sign-in" />
<button data-testid="create-pin" />
<div data-testid="pin-details" />
<form data-testid="pin-form" />
```