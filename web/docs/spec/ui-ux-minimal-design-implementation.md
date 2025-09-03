# Alunalun - Minimal UI/UX Design Implementation Specification

**Version**: 1.0.0  
**Date**: 2025-01-03  
**Status**: Implementation Ready

## Design Philosophy

Inspired by Queering the Map's ultra-minimal approach, Alunalun implements a **map-as-interface** design where:
- The map occupies 100% of the viewport
- UI elements appear only when needed
- No traditional web chrome (header, footer, sidebar)
- Icons over text, gestures over buttons
- Content (pins) drive the experience

## Core Design Principles

1. **Invisible Until Needed**: UI elements hidden by default
2. **Contextual Revelation**: Controls appear based on user intent
3. **Mobile-First**: Touch gestures primary, mouse secondary
4. **Geographic Navigation**: Location replaces traditional menus
5. **Minimal Cognitive Load**: One primary action at a time

---

## Viewport States & ASCII Representations

### 1. Initial Load - Anonymous User

```
DESKTOP (1920x1080)
┌────────────────────────────────────────────────────────────────────┐
│                                                                    │
│                                                                    │
│                         📍        📍                              │
│              📍                            📍                     │
│                                                                    │
│         📍                                      📍                 │
│                      📍         📍                                 │
│                                                                    │
│                                       📍                           │
│                                                                    │
│                                                                    │
│                                                                    │
│                                                                    │
│                                                                    │
│                                                                    │
│                                                                    │
│                                                          ⊕        │
│                                                          ◎        │
│                                                          ↗        │
└────────────────────────────────────────────────────────────────────┘

MOBILE (375x812)
┌─────────────────────┐
│                     │
│      📍      📍     │
│                     │
│  📍       📍         │
│       📍             │
│                     │
│            📍        │
│                     │
│                     │
│                     │
│                     │
│                     │
│                     │
│                     │
│                     │
│               ⊕     │
│               ◎     │
│               ↗     │
└─────────────────────┘

Legend:
📍 = Existing pins (last 24 hours)
⊕ = Zoom control (tap to expand)
◎ = Locate me
↗ = Sign in
```

### 2. HUD Hidden State (Center Tap)

```
DESKTOP - HUD HIDDEN
┌────────────────────────────────────────────────────────────────────┐
│                                                                    │
│                                                                    │
│                         📍        📍                              │
│              📍                            📍                     │
│                                                                    │
│         📍                                      📍                 │
│                      📍         📍                                 │
│                                                                    │
│                                       📍                           │
│                                                                    │
│                                                                    │
│                                                                    │
│                                                                    │
│                                                                    │
│                                                                    │
│                                                                    │
│                                                                    │
│                                                                    │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘

MOBILE - HUD HIDDEN
┌─────────────────────┐
│                     │
│      📍      📍     │
│                     │
│  📍       📍         │
│       📍             │
│                     │
│            📍        │
│                     │
│                     │
│                     │
│                     │
│                     │
│                     │
│                     │
│                     │
│                     │
│                     │
│                     │
└─────────────────────┘

Note: Tap anywhere except center to restore HUD
```

### 3. Zoom Expanded State

```
DESKTOP - ZOOM EXPANDED
┌────────────────────────────────────────────────────────────────────┐
│                                                                    │
│                                                                    │
│                         📍        📍                              │
│              📍                            📍                     │
│                                                                    │
│         📍                                      📍                 │
│                      📍         📍                                 │
│                                                                    │
│                                       📍                           │
│                                                                    │
│                                                                    │
│                                                                    │
│                                                                    │
│                                                                    │
│                                                          ┌───┐    │
│                                                          │ + │    │
│                                                          ├───┤    │
│                                                          │ - │    │
│                                                          └───┘    │
│                                                          ◎        │
│                                                          ↗        │
└────────────────────────────────────────────────────────────────────┐

Interaction: Tap ⊕ to expand zoom controls, tap elsewhere to collapse
```

### 4. Login Modal State

```
DESKTOP - LOGIN MODAL
┌────────────────────────────────────────────────────────────────────┐
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░┌──────────────────────────────────┐░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░│                                  │░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░│     Continue with Google         │░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░│                                  │░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░│  ┌────────────────────────┐     │░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░│  │    🔍 Google Sign In    │     │░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░│  └────────────────────────┘     │░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░│                                  │░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░│  [Turnstile Widget]              │░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░│                                  │░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░└──────────────────────────────────┘░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
└────────────────────────────────────────────────────────────────────┘

MOBILE - LOGIN MODAL
┌─────────────────────┐
│░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░│
│░░┌───────────────┐░░│
│░░│               │░░│
│░░│  Continue w/  │░░│
│░░│    Google     │░░│
│░░│               │░░│
│░░│ ┌───────────┐ │░░│
│░░│ │ 🔍 Sign In│ │░░│
│░░│ └───────────┘ │░░│
│░░│               │░░│
│░░│ [Turnstile]   │░░│
│░░│               │░░│
│░░└───────────────┘░░│
│░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░│
└─────────────────────┘

░ = Backdrop blur overlay
```

### 5. Username Selection (First-time Login)

```
DESKTOP - USERNAME PICKER
┌────────────────────────────────────────────────────────────────────┐
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░┌──────────────────────────────────┐░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░│                                  │░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░│     Choose your username         │░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░│                                  │░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░│  ┌────────────────────────┐     │░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░│  │ @_                     │     │░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░│  └────────────────────────┘     │░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░│                                  │░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░│  ✓ Username available            │░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░│                                  │░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░│  [Cancel]      [Continue →]      │░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░│                                  │░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░└──────────────────────────────────┘░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
└────────────────────────────────────────────────────────────────────┘

Real-time validation states:
- Typing: "checking..." (gray)
- Available: "✓ Username available" (green)
- Taken: "✗ Username taken" (red)
```

### 6. Authenticated User - Main View

```
DESKTOP - LOGGED IN
┌────────────────────────────────────────────────────────────────────┐
│john                                                               │ ← Subtle username
│                                                                    │
│                         📍        📍                              │
│              📍                            📍                     │
│                                                                    │
│         📍                                      📍                 │
│                      📍         📍                                 │
│                                                                    │
│                                       📍                           │
│                                                                    │
│                                                                    │
│                                                                    │
│                                                                    │
│                                                                    │
│                                                                    │
│                                                                    │
│                                                          ⊕        │
│                                                          ◎        │
│                                                          J        │ ← User initial
└────────────────────────────────────────────────────────────────────┘

MOBILE - LOGGED IN
┌─────────────────────┐
│j                    │ ← Tiny username
│                     │
│      📍      📍     │
│                     │
│  📍       📍         │
│       📍             │
│                     │
│            📍        │
│                     │
│                     │
│                     │
│                     │
│                     │
│                     │
│                     │
│               ⊕     │
│               ◎     │
│               J     │
└─────────────────────┘
```

### 7. Pin Creation Flow - Location Selected

```
DESKTOP - PIN CREATION
┌────────────────────────────────────────────────────────────────────┐
│john                                                               │
│                                                                    │
│                         📍        📍                              │
│              📍                            📍                     │
│                                                                    │
│         📍              ⊗                      📍                 │ ← Creating here
│                      📍         📍                                 │
│                                                                    │
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░┌────────────────────────────────────────────────────────────┐░░░│
│░░│                                                            │░░░│
│░░│  What happened here?                                       │░░░│
│░░│  ┌──────────────────────────────────────────────────┐     │░░░│
│░░│  │                                                  │     │░░░│
│░░│  │                                                  │     │░░░│
│░░│  └──────────────────────────────────────────────────┘     │░░░│
│░░│                                            15/280          │░░░│
│░░│                                                            │░░░│
│░░│  [Cancel]                           [Drop Pin →]           │░░░│
│░░└────────────────────────────────────────────────────────────┘░░░│
└────────────────────────────────────────────────────────────────────┘

MOBILE - PIN CREATION (Bottom Sheet)
┌─────────────────────┐
│john                 │
│                     │
│      📍      📍     │
│                     │
│  📍   ⊗   📍         │ ← Creating
│       📍             │
│░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░│
├─────────────────────┤
│ What happened here? │
│ ┌─────────────────┐ │
│ │                 │ │
│ │                 │ │
│ │                 │ │
│ └─────────────────┘ │
│             15/280  │
│                     │
│ [Cancel] [Drop Pin] │
└─────────────────────┘

⊗ = Pin being created (pulsing indicator)
```

### 8. Pin Details View

```
DESKTOP - PIN DETAILS
┌────────────────────────────────────────────────────────────────────┐
│john                                                               │
│                                                                    │
│                         📍        📍                              │
│              📍                            📌                     │ ← Selected pin
│                                                                    │
│         📍                                      📍                 │
│                      📍         📍                                 │
│                                                                    │
│░░░░░░░░░░░░░░░░░░░░░░░░░┌─────────────────────────────────┐░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░░░░░│ @sarah · 2 hours ago           │░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░░░░░│                                 │░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░░░░░│ "This bench has the best sunset │░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░░░░░│  view in the entire city. I     │░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░░░░░│  come here when I need to       │░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░░░░░│  think."                         │░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░░░░░│                                 │░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░░░░░│ 💬 3 comments                    │░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░░░░░│                                 │░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░░░░░│ [Add comment...]          [×]   │░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░░░░░└─────────────────────────────────┘░░░░░░░│
└────────────────────────────────────────────────────────────────────┘

MOBILE - PIN DETAILS (Bottom Sheet - Swipeable)
┌─────────────────────┐
│john                 │
│                     │
│      📍      📍     │
│                     │
│  📍       📌         │ ← Selected
│       📍             │
│░░░░░░░░░░░░░░░░░░░░░│
├─────────────────────┤
│   ━━━━━━━━━━━━━    │ ← Drag handle
│ @sarah · 2 hours   │
│                     │
│ "This bench has the │
│  best sunset view   │
│  in the entire      │
│  city..."           │
│                     │
│ 💬 3 comments       │
│                     │
│ [Add comment...]    │
└─────────────────────┘

Swipe down to minimize, swipe up to expand
```

### 9. User Menu Expanded

```
DESKTOP - USER MENU
┌────────────────────────────────────────────────────────────────────┐
│john                                                               │
│                                                                    │
│                         📍        📍                              │
│              📍                            📍                     │
│                                                                    │
│         📍                                      📍                 │
│                      📍         📍                                 │
│                                                                    │
│                                       📍                           │
│                                                                    │
│                                                                    │
│                                                                    │
│                                                       ┌────────┐   │
│                                                       │ @john  │   │
│                                                       │ ────── │   │
│                                                       │ Logout │   │
│                                                       └────────┘   │
│                                                          ⊕        │
│                                                          ◎        │
│                                                          J        │
└────────────────────────────────────────────────────────────────────┘

Minimal menu - just username and logout
```

### 10. Loading/Locating State

```
DESKTOP - LOCATING
┌────────────────────────────────────────────────────────────────────┐
│john                                                               │
│                                                                    │
│              Getting your location...                             │ ← Subtle message
│                                                                    │
│                         📍        📍                              │
│              📍                            📍                     │
│                                                                    │
│         📍                                      📍                 │
│                      📍         📍                                 │
│                                                                    │
│                                       📍                           │
│                                                                    │
│                         ◈                                         │ ← Pulsing location
│                                                                    │
│                                                                    │
│                                                                    │
│                                                          ⊕        │
│                                                          ◉        │ ← Active/pulsing
│                                                          J        │
└────────────────────────────────────────────────────────────────────┘

◈ = User location (blue pulsing dot)
◉ = Locate button active/pulsing
```

---

## Interaction Patterns

### Touch/Click Zones

```
┌─────────────────────────────────────┐
│                                     │
│         TAP TO CREATE PIN           │ ← Any empty map area
│                                     │
│    ┌─────────────────────────┐      │
│    │                         │      │ 
│    │    TAP CENTER TO        │      │ ← Center 30% hides HUD
│    │      HIDE HUD           │      │
│    │                         │      │
│    └─────────────────────────┘      │
│                                     │
│                                     │
│         TAP EDGES TO                │ ← Edges restore HUD
│          SHOW HUD                   │
│                                     │
│                          [Controls] │ ← Bottom-right controls
└─────────────────────────────────────┘
```

### Gesture Support

| Gesture | Action | Context |
|---------|--------|---------|
| Tap | Create pin / Select pin | Map area |
| Pinch | Zoom in/out | Map |
| Pan | Move map | Map |
| Swipe up | Expand bottom sheet | Mobile pin details |
| Swipe down | Minimize bottom sheet | Mobile pin details |
| Long press | (Reserved) | Future: Quick actions |
| Double tap | Zoom in | Map |

---

## Component Hierarchy

```
HomePage
├── MapView (100% viewport)
│   ├── MapLibre Canvas
│   ├── PinMarkers Layer
│   └── UserLocation Layer
├── MinimalHUD
│   ├── SubtleUsername (top-left)
│   └── FloatingControls (bottom-right)
│       ├── ZoomButton (expandable)
│       ├── LocateButton
│       └── AccountButton
├── PinCreationFlow
│   ├── LocationIndicator
│   └── CreationSheet (modal/bottom-sheet)
└── PinDetailsView
    └── DetailsSheet (overlay/bottom-sheet)
```

---

## State Management

### ViewportStore
```typescript
interface ViewportState {
  hudVisible: boolean       // Toggle for HUD visibility
  activeFlow: 'idle' | 'creating' | 'viewing' | 'authenticating'
  selectedPin: string | null
  creatingLocation: Coords | null
  device: 'mobile' | 'tablet' | 'desktop'
}
```

### User Flows

#### 1. First-Time User Journey
```
Land → Explore Pins → Tap Sign In → OAuth → Pick Username → Create First Pin
```

#### 2. Returning User Journey
```
Land → Auto-signin → Explore → Tap Map → Create Pin → View Others → Comment
```

#### 3. Discovery Flow
```
Land → Browse by Location → Read Stories → Get Inspired → Sign Up → Contribute
```

---

## Mobile Responsiveness Strategy

### Breakpoints
- Mobile: < 640px (bottom sheets, single column)
- Tablet: 640-1024px (side panels possible)
- Desktop: > 1024px (floating panels)

### Mobile Adaptations
1. **Bottom sheets** replace all modals
2. **Swipe gestures** for panel control
3. **Larger tap targets** (min 44px)
4. **Single-column** layouts
5. **Full-screen** overlays

### Desktop Enhancements
1. **Hover states** on pins
2. **Keyboard shortcuts** (Esc, Enter, etc.)
3. **Floating panels** with backdrop
4. **Multi-select** capabilities
5. **Right-click** context menus

---

## Technical Implementation Notes

### Performance Optimizations
- Virtual scrolling for pin lists
- Cluster pins at low zoom levels
- Lazy load pin details
- Debounce map interactions
- Progressive image loading

### Accessibility
- Keyboard navigation support
- Screen reader announcements
- High contrast mode support
- Focus trap in modals
- ARIA labels on all controls

### Browser Support
- Modern browsers (Chrome, Firefox, Safari, Edge)
- Mobile Safari iOS 14+
- Chrome Android 90+
- Progressive enhancement for older browsers

---

## Design Tokens

### Colors
```css
--color-background: rgba(255, 255, 255, 0.9)
--color-backdrop: rgba(0, 0, 0, 0.2)
--color-text: #1a1a1a
--color-subtle: #6b7280
--color-accent: #000000
--color-error: #ef4444
--color-success: #10b981
```

### Spacing
```css
--space-xs: 4px
--space-sm: 8px
--space-md: 16px
--space-lg: 24px
--space-xl: 32px
```

### Typography
```css
--font-sans: system-ui, -apple-system, sans-serif
--text-xs: 12px
--text-sm: 14px
--text-base: 16px
--text-lg: 18px
```

### Animations
```css
--duration-fast: 150ms
--duration-normal: 300ms
--duration-slow: 500ms
--easing-default: cubic-bezier(0.4, 0, 0.2, 1)
```

---

## Implementation Checklist

### Phase 1: Core Infrastructure
- [ ] Viewport state management
- [ ] HUD toggle system
- [ ] Device detection
- [ ] Map container setup

### Phase 2: Anonymous Experience
- [ ] Read-only map view
- [ ] Pin browsing
- [ ] Sign-in flow
- [ ] OAuth integration

### Phase 3: Authenticated Features
- [ ] Username selection
- [ ] Pin creation flow
- [ ] Pin details view
- [ ] Comment system

### Phase 4: Mobile Optimization
- [ ] Bottom sheet component
- [ ] Touch gesture support
- [ ] Responsive breakpoints
- [ ] Mobile-specific interactions

### Phase 5: Polish
- [ ] Animations & transitions
- [ ] Loading states
- [ ] Error handling
- [ ] Performance optimization

---

This specification provides a complete blueprint for implementing Alunalun's ultra-minimal UI, ensuring a clean, focused user experience that prioritizes content over chrome.