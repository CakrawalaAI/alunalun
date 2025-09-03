# Frontend End-to-End Flow Documentation
**Date**: 2025-09-03  
**Status**: Implementation Guide

## Overview

This document provides comprehensive frontend flow documentation for the Alunalun collaborative map platform, detailing every user interaction, screen state, and data flow through the frontend architecture layers.

## Frontend Architecture Layers

```
┌─────────────────────────────────────────────────────────────────┐
│                    USER INTERACTION                              │
│              (Clicks, Taps, Form Inputs, Gestures)              │
├─────────────────────────────────────────────────────────────────┤
│                    UI COMPONENTS                                 │
│         (React Components, Forms, Modals, Drawers)              │
├─────────────────────────────────────────────────────────────────┤
│                   EVENT HANDLERS                                 │
│         (onClick, onSubmit, onChange, onBlur)                   │
├─────────────────────────────────────────────────────────────────┤
│                  CUSTOM HOOKS LAYER                              │
│     (Feature hooks, Form hooks, React Query hooks)              │
├─────────────────────────────────────────────────────────────────┤
│               STATE MANAGEMENT LAYER                             │
│     Local: Zustand Stores | Server: React Query Cache           │
├─────────────────────────────────────────────────────────────────┤
│                   API CLIENT LAYER                               │
│          (ConnectRPC Query/Mutation clients)                    │
├─────────────────────────────────────────────────────────────────┤
│                  NETWORK TRANSPORT                               │
│               (HTTP/ConnectRPC Protocol)                        │
└─────────────────────────────────────────────────────────────────┘
```

## Feature Module Structure

```
web/src/features/
├── auth/                    # Authentication & authorization
│   ├── components/         
│   │   ├── LoginButton.tsx
│   │   ├── AuthModal.tsx
│   │   ├── OAuthProviderButtons.tsx
│   │   ├── TurnstileWidget.tsx
│   │   ├── UsernamePickerForm.tsx
│   │   └── LogoutButton.tsx
│   ├── hooks/              
│   │   ├── useAuth.ts
│   │   ├── useOAuthLogin.ts
│   │   ├── useTokenRefresh.ts
│   │   └── useLogout.ts
│   ├── stores/             
│   │   └── authStore.ts
│   └── types/              
│       └── auth.types.ts
│
├── pins/                    # Pin management
│   ├── components/         
│   │   ├── PinCreationDrawer.tsx
│   │   ├── PinDetailsPanel.tsx
│   │   ├── PinPreviewTooltip.tsx
│   │   ├── PinList.tsx
│   │   ├── CommentSection.tsx
│   │   └── CommentForm.tsx
│   ├── hooks/              
│   │   ├── usePins.ts
│   │   ├── useCreatePin.ts
│   │   ├── useDeletePin.ts
│   │   └── useComments.ts
│   ├── stores/             
│   │   └── pinsStore.ts
│   └── types/              
│       └── pin.types.ts
│
├── user/                    # User profile management
│   ├── components/         
│   │   ├── UserProfileModal.tsx
│   │   ├── ProfileEditForm.tsx
│   │   ├── AvatarUploader.tsx
│   │   └── UserPinsList.tsx
│   ├── hooks/              
│   │   ├── useUserProfile.ts
│   │   ├── useUpdateProfile.ts
│   │   └── useUserPins.ts
│   └── types/              
│       └── user.types.ts
│
└── map/                     # Map interactions (existing)
    ├── components/         
    ├── hooks/              
    ├── stores/             
    └── lib/                
```

---

## User Journey Flows

### 1. Authentication Flow - Google OAuth with Username Selection

#### 1.1 Initial State - Anonymous User

```
┌─────────────────────────────────────────────────────────────┐
│  Header                                           [Sign In] │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│                      Map View                               │
│                   (Public pins visible)                     │
│                                                             │
└─────────────────────────────────────────────────────────────┘

USER SEES:
- Header with "Sign In" button (top right)
- Map showing public pins from last 24 hours
- Can view pin details but cannot create/comment
```

#### 1.2 Login Modal Opens

```
User clicks "Sign In" → Modal opens with overlay

┌─────────────────────────────────────────────────────────────┐
│  ╔═══════════════════════════════════════════════════╗     │
│  ║                                                [X] ║     │
│  ║                                                    ║     │
│  ║  Login dengan Google untuk mulai                   ║     │
│  ║                                                    ║     │
│  ║  ┌──────────────────────────────────────────┐     ║     │
│  ║  │   [Google Icon] Lanjutkan dengan Google  │     ║     │
│  ║  └──────────────────────────────────────────┘     ║     │
│  ║                                                    ║     │
│  ║  ┌──────────────────────────────────────────┐     ║     │
│  ║  │      Cloudflare Turnstile Widget          │     ║     │
│  ║  │         [Verification in progress]        │     ║     │
│  ║  └──────────────────────────────────────────┘     ║     │
│  ║                                                    ║     │
│  ╚═══════════════════════════════════════════════════╝     │
└─────────────────────────────────────────────────────────────┘
```

#### 1.3 OAuth Flow Sequence

```
┌────────┐     ┌──────────┐     ┌─────────┐     ┌──────────┐     ┌──────────┐     ┌────────┐
│  User  │     │  React   │     │  Auth   │     │ Turnstile│     │   API    │     │ Google │
│        │     │Component │     │  Hook   │     │  Widget  │     │  Client  │     │  OAuth │
└───┬────┘     └────┬─────┘     └────┬────┘     └────┬─────┘     └────┬─────┘     └───┬────┘
    │               │                 │                │                │               │
    │ Click Google  │                 │                │                │               │
    │   button      │                 │                │                │               │
    ├──────────────►│                 │                │                │               │
    │               │                 │                │                │               │
    │               │ Check Turnstile │                │                │               │
    │               ├────────────────►│                │                │               │
    │               │                 │ Verify human   │                │               │
    │               │                 ├───────────────►│                │               │
    │               │                 │                │ Get token      │               │
    │               │                 │◄───────────────┤                │               │
    │               │                 │ Token valid    │                │               │
    │               │◄────────────────┤                │                │               │
    │               │                 │                │                │               │
    │               │ Initiate OAuth  │                │                │               │
    │               ├─────────────────────────────────►│                │               │
    │               │                 │                │ POST /auth/    │               │
    │               │                 │                │ oauth/google   │               │
    │               │                 │                ├───────────────►│               │
    │               │                 │                │                │ Build URL     │
    │               │                 │                │◄───────────────┤               │
    │               │                 │                │ Redirect URL   │               │
    │               │◄─────────────────────────────────┤                │               │
    │               │                 │                │                │               │
    │◄──────────────┤ window.location│                │                │               │
    │               │ = redirectUrl   │                │                │               │
    │                                                                                   │
    │ Redirect to Google OAuth consent screen                                          │
    ├──────────────────────────────────────────────────────────────────────────────────►│
    │                                                                                   │
    │ User approves permissions                                                        │
    │◄──────────────────────────────────────────────────────────────────────────────────┤
    │ Return with code                                                                 │
    │                                                                                   │
    │ GET /auth/oauth/google/callback?code=xxx&state=xxx                              │
    ├──────────────►│                 │                │                │               │
    │               │ Parse callback  │                │                │               │
    │               ├─────────────────────────────────►│                │               │
    │               │                 │                │ Exchange code  │               │
    │               │                 │                ├───────────────►│               │
    │               │                 │                │◄───────────────┤               │
    │               │                 │                │ JWT + User     │               │
    │               │◄─────────────────────────────────┤                │               │
    │               │ Store token     │                │                │               │
    │               │ Check username  │                │                │               │
```

#### 1.4 Username Selection Screen

```
If user.username is null → Show username picker

┌─────────────────────────────────────────────────────────────┐
│  ╔═══════════════════════════════════════════════════╗     │
│  ║         Choose Your Username                        ║     │
│  ║                                                    ║     │
│  ║  Pick a username to continue                       ║     │
│  ║                                                    ║     │
│  ║  Username *                                        ║     │
│  ║  ┌──────────────────────────────────────────┐     ║     │
│  ║  │ @johndoe123                              │     ║     │
│  ║  └──────────────────────────────────────────┘     ║     │
│  ║  ✓ Username is available                          ║     │
│  ║                                                    ║     │
│  ║  [Cancel]                    [Continue →]          ║     │
│  ╚═══════════════════════════════════════════════════╝     │
└─────────────────────────────────────────────────────────────┘

INTERACTIONS:
- Username field has real-time validation
- Shows "checking..." while validating
- Shows error if username taken
- Shows success checkmark if available
- Submit button disabled until valid username
- No display name field - just username
```

#### 1.5 Username Validation & Submission Flow

```
┌────────┐     ┌──────────┐     ┌─────────┐     ┌──────────┐     ┌──────────┐     ┌────────┐
│  User  │     │   Form   │     │  Hook   │     │ Debounce │     │   API    │     │  Auth  │
│        │     │  State   │     │useForm  │     │  300ms   │     │  Client  │     │  Store │
└───┬────┘     └────┬─────┘     └────┬────┘     └────┬─────┘     └────┬─────┘     └───┬────┘
    │               │                 │                │                │               │
    │ Type username │                 │                │                │               │
    ├──────────────►│                 │                │                │               │
    │               │ onChange        │                │                │               │
    │               ├────────────────►│                │                │               │
    │               │                 │ Set loading    │                │               │
    │               │                 │ Start debounce │                │               │
    │               │                 ├───────────────►│                │               │
    │               │                 │                │ Wait 300ms     │               │
    │               │                 │                ├─┐              │               │
    │               │                 │                │ │              │               │
    │               │                 │                │◄┘              │               │
    │               │                 │◄───────────────┤                │               │
    │               │                 │ Check username │                │               │
    │               │                 ├─────────────────────────────────►               │
    │               │                 │                │ GET /users/    │               │
    │               │                 │                │ check-username │               │
    │               │                 │◄─────────────────────────────────               │
    │               │                 │ Available/Taken│                │               │
    │               │◄────────────────┤                │                │               │
    │               │ Update UI       │                │                │               │
    │◄──────────────┤ Show feedback   │                │                │               │
    │               │                 │                │                │               │
    │ Click Submit  │                 │                │                │               │
    ├──────────────►│                 │                │                │               │
    │               │ onSubmit        │                │                │               │
    │               ├────────────────►│                │                │               │
    │               │                 │ Validate form  │                │               │
    │               │                 │ Call mutation  │                │               │
    │               │                 ├─────────────────────────────────►               │
    │               │                 │                │ PUT /users/me  │               │
    │               │                 │◄─────────────────────────────────               │
    │               │                 │ Updated user   │                │               │
    │               │                 ├───────────────────────────────────────────────►│
    │               │                 │                │                │ Update store  │
    │               │                 │                │                │ Set authed    │
    │               │◄────────────────────────────────────────────────────────────────┤
    │               │ Close modal     │                │                │               │
    │◄──────────────┤ Show success    │                │                │               │
```

#### 1.6 Authenticated State - User Logged In

```
┌─────────────────────────────────────────────────────────────┐
│  Header                    [@johndoe123 ▼] [Drop Pin 📍]   │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│                      Map View                               │
│                  (Can now create pins)                      │
│                                                             │
└─────────────────────────────────────────────────────────────┘

HEADER CHANGES:
- "Sign In" replaced with username dropdown
- "Drop Pin" button appears
- Dropdown menu has: Profile, Settings, Logout
```

**Component Implementation Details:**

```typescript
// features/auth/components/AuthModal.tsx
interface AuthModalProps {
  isOpen: boolean
  onClose: () => void
  redirectTo?: string
}

// features/auth/components/UsernamePickerForm.tsx
interface UsernamePickerFormProps {
  user: { email: string, name: string, picture?: string }
  onComplete: (username: string) => void
  onCancel: () => void
}

// features/auth/hooks/useAuth.ts
interface UseAuthReturn {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  login: (provider: 'google' | 'github') => Promise<void>
  logout: () => Promise<void>
  refreshToken: () => Promise<void>
}

// features/auth/stores/authStore.ts
interface AuthStore {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  setAuth: (user: User, token: string) => void
  clearAuth: () => void
  updateUser: (user: Partial<User>) => void
}
```

---

### 2. Pin Creation Flow - From Map Click to Published Pin

#### 2.1 Initiating Pin Creation

```
USER ACTION: Click "Create Pin" button OR Click directly on map

┌─────────────────────────────────────────────────────────────┐
│  Header                    [@johndoe123 ▼] [Create Pin ●]   │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│                      Map View                               │
│                    [Click anywhere]                         │
│                         ⬇                                   │
│                    ╔════════════════════════════════════╗  │
│                    ║      Create New Pin                ║  │
│                    ║                                    ║  │
│                    ║  📍 Location                       ║  │
│                    ║  ┌────────────────────────────┐   ║  │
│                    ║  │  [Mini map preview]        │   ║  │
│                    ║  │   Lat: 37.7749             │   ║  │
│                    ║  │   Lng: -122.4194           │   ║  │
│                    ║  └────────────────────────────┘   ║  │
│                    ║                                    ║  │
│                    ║  ✏️ What's happening here?         ║  │
│                    ║  ┌────────────────────────────┐   ║  │
│                    ║  │                            │   ║  │
│                    ║  │                            │   ║  │
│                    ║  │                            │   ║  │
│                    ║  └────────────────────────────┘   ║  │
│                    ║              15/280 characters     ║  │
│                    ║                                    ║  │
│                    ║  📷 Add Photo (optional)           ║  │
│                    ║  ┌────────────────────────────┐   ║  │
│                    ║  │    [Drop zone or click]    │   ║  │
│                    ║  └────────────────────────────┘   ║  │
│                    ║                                    ║  │
│                    ║  [Cancel]        [Create Pin →]    ║  │
│                    ╚════════════════════════════════════╝  │
└─────────────────────────────────────────────────────────────┘

DRAWER BEHAVIOR:
- Slides up from bottom on mobile
- Modal on desktop
- Map dims but remains visible
- Can drag to adjust location
```

#### 2.2 Pin Creation Data Flow

```
┌────────┐     ┌──────────┐     ┌─────────┐     ┌──────────┐     ┌──────────┐     ┌────────┐
│  User  │     │   Form   │     │  Hook   │     │Optimistic│     │   API    │     │  Map   │
│        │     │Component │     │useCreate│     │  Update  │     │  Client  │     │  Store │
└───┬────┘     └────┬─────┘     └────┬────┘     └────┬─────┘     └────┬─────┘     └───┬────┘
    │               │                 │                │                │               │
    │ Type content  │                 │                │                │               │
    ├──────────────►│                 │                │                │               │
    │               │ Update state    │                │                │               │
    │               │ Count chars     │                │                │               │
    │◄──────────────┤ Show 45/280     │                │                │               │
    │               │                 │                │                │               │
    │ Upload photo  │                 │                │                │               │
    ├──────────────►│                 │                │                │               │
    │               │ Preview image   │                │                │               │
    │◄──────────────┤                 │                │                │               │
    │               │                 │                │                │               │
    │ Click Create  │                 │                │                │               │
    ├──────────────►│                 │                │                │               │
    │               │ Validate        │                │                │               │
    │               ├────────────────►│                │                │               │
    │               │                 │ Generate ID    │                │               │
    │               │                 │ (client-side)  │                │               │
    │               │                 │                │                │               │
    │               │                 │ Add to cache   │                │               │
    │               │                 ├───────────────►│                │               │
    │               │                 │                │ Update map     │               │
    │               │                 │                ├───────────────────────────────►│
    │               │                 │                │                │ Add temp pin  │
    │               │                 │                │                │               │
    │               │                 │ Call mutation  │                │               │
    │               │                 ├─────────────────────────────────►               │
    │               │                 │                │ POST /pins     │               │
    │               │                 │                │ {content,      │               │
    │               │                 │                │  location,     │               │
    │               │                 │                │  photo}        │               │
    │               │                 │◄─────────────────────────────────               │
    │               │                 │ Created pin    │                │               │
    │               │                 │                │                │               │
    │               │                 │ Update cache   │                │               │
    │               │                 ├───────────────►│                │               │
    │               │                 │                │ Replace temp   │               │
    │               │                 │                ├───────────────────────────────►│
    │               │                 │                │                │ Finalize pin  │
    │               │◄────────────────┤                │                │               │
    │               │ Close drawer    │                │                │               │
    │◄──────────────┤ Show success    │                │                │               │
    │               │ toast           │                │                │               │
```

#### 2.3 Error Handling States

```
ERROR SCENARIOS:

1. Network Error
┌────────────────────────────────────┐
│  ⚠️ Connection error                │
│  Could not create pin.              │
│  [Retry] [Cancel]                   │
└────────────────────────────────────┘

2. Validation Error
┌────────────────────────────────────┐
│  Content is required (min 10 chars) │
│  ________________________________   │
└────────────────────────────────────┘

3. Photo Upload Error
┌────────────────────────────────────┐
│  ❌ File too large (max 5MB)        │
│  Please choose a smaller image      │
└────────────────────────────────────┘
```

**Pin Creation Components:**

```typescript
// features/pins/components/PinCreationDrawer.tsx
interface PinCreationDrawerProps {
  isOpen: boolean
  location: { lat: number, lng: number }
  onClose: () => void
  onSuccess: (pin: Pin) => void
}

// features/pins/hooks/useCreatePin.ts
interface UseCreatePinReturn {
  createPin: (data: CreatePinInput) => Promise<Pin>
  isLoading: boolean
  error: Error | null
  reset: () => void
}

// Optimistic update pattern
const mutation = useMutation({
  mutationFn: createPin,
  onMutate: async (newPin) => {
    // Cancel outgoing refetches
    await queryClient.cancelQueries(['pins'])
    
    // Snapshot previous value
    const previous = queryClient.getQueryData(['pins'])
    
    // Optimistically update
    queryClient.setQueryData(['pins'], old => [...old, newPin])
    
    return { previous }
  },
  onError: (err, newPin, context) => {
    // Rollback on error
    queryClient.setQueryData(['pins'], context.previous)
  },
  onSettled: () => {
    // Always refetch after error or success
    queryClient.invalidateQueries(['pins'])
  }
})
```

---

### 3. Pin Viewing & Interaction Flow

#### 3.1 Pin Hover Preview

```
USER ACTION: Hover over pin marker on map

┌─────────────────────────────────────────────────────────────┐
│                                                             │
│                    ┌──────────────────┐                    │
│                    │ @johndoe123      │                    │
│                    │ "Great coffee!"  │                    │
│                    │ 2 hours ago      │                    │
│                    │ 💬 3 comments    │                    │
│                    └────────┬─────────┘                    │
│                             📍                              │
│                         Map View                            │
│                                                             │
└─────────────────────────────────────────────────────────────┘

TOOLTIP BEHAVIOR:
- Appears after 200ms hover
- Disappears on mouse leave
- Shows preview of content
- Click to open full details
```

#### 3.2 Pin Details Panel

```
USER ACTION: Click on pin marker

┌─────────────────────────────────────────────────────────────┐
│  Map View (dimmed)          ║  Pin Details                  │
│                             ║                               │
│                             ║  [@johndoe123]                │
│            📍               ║  2 hours ago · Marina District│
│                             ║                               │
│                             ║  "Found an amazing coffee     │
│                             ║   spot with outdoor seating  │
│                             ║   and great vibes!"           │
│                             ║                               │
│                             ║  [📷 Photo if exists]         │
│                             ║                               │
│                             ║  ❤️ 12  💬 3                  │
│                             ║                               │
│                             ║  ─────────────────────        │
│                             ║                               │
│                             ║  Comments                     │
│                             ║                               │
│                             ║  [@user456] 1 hour ago       │
│                             ║  "Love this place!"           │
│                             ║                               │
│                             ║  [@user789] 30 min ago       │
│                             ║  "What's the wifi password?" │
│                             ║                               │
│                             ║  ┌─────────────────────┐      │
│                             ║  │ Add a comment...   │      │
│                             ║  └─────────────────────┘      │
│                             ║  [Post]                       │
│                             ║                               │
│                             ║  [🗑️ Delete] (if owner)      │
└─────────────────────────────────────────────────────────────┘

PANEL BEHAVIOR:
- Slides in from right on desktop
- Full screen modal on mobile
- Escape key or click outside to close
- Smooth transitions
```

#### 3.3 Comment Interaction Flow

```
┌────────┐     ┌──────────┐     ┌─────────┐     ┌──────────┐     ┌──────────┐
│  User  │     │  Comment │     │  Hook   │     │   API    │     │  Cache   │
│        │     │   Form   │     │useComment│    │  Client  │     │  Update  │
└───┬────┘     └────┬─────┘     └────┬────┘     └────┬─────┘     └────┬─────┘
    │               │                 │                │                │
    │ Type comment  │                 │                │                │
    ├──────────────►│                 │                │                │
    │               │ Update state    │                │                │
    │               │                 │                │                │
    │ Press Enter   │                 │                │                │
    │ or Post       │                 │                │                │
    ├──────────────►│                 │                │                │
    │               │ onSubmit        │                │                │
    │               ├────────────────►│                │                │
    │               │                 │ Optimistic add │                │
    │               │                 ├───────────────────────────────►│
    │               │                 │                │                │ Add to UI
    │               │                 │ POST /pins/    │                │
    │               │                 │ {id}/comments │                │
    │               │                 ├───────────────►│                │
    │               │                 │◄───────────────┤                │
    │               │                 │ Created comment│                │
    │               │                 ├───────────────────────────────►│
    │               │                 │                │                │ Replace temp
    │               │◄────────────────┤                │                │
    │               │ Clear input     │                │                │
    │◄──────────────┤ Show new comment│                │                │
```

#### 3.4 Pin Deletion Flow (Owner Only)

```
USER ACTION: Click Delete button (only visible to pin owner)

┌──────────────────────────────────────┐
│  ⚠️ Delete Pin?                       │
│                                      │
│  This action cannot be undone.       │
│  The pin and all comments will be    │
│  permanently removed.                │
│                                      │
│  [Cancel]      [Delete Pin]          │
└──────────────────────────────────────┘

DELETION FLOW:
1. Show confirmation modal
2. On confirm → Call delete mutation
3. Optimistically remove from map
4. Close details panel
5. Show success toast
6. On error → Restore pin, show error
```

---

### 4. User Profile Management Flow

#### 4.1 Profile Modal View

```
USER ACTION: Click on any username

┌─────────────────────────────────────────────────────────────┐
│  ╔═══════════════════════════════════════════════════╗     │
│  ║  User Profile                                 [X] ║     │
│  ║                                                    ║     │
│  ║     [Avatar]                                      ║     │
│  ║       ⬤        @johndoe123                        ║     │
│  ║              John Doe                             ║     │
│  ║              Member since Dec 2024                ║     │
│  ║                                                    ║     │
│  ║  ─────────────────────────────────────           ║     │
│  ║                                                    ║     │
│  ║  📍 Pins (24)      💬 Comments (156)              ║     │
│  ║                                                    ║     │
│  ║  Recent Pins:                                     ║     │
│  ║  ┌──────────┐ ┌──────────┐ ┌──────────┐         ║     │
│  ║  │          │ │          │ │          │         ║     │
│  ║  │  Pin 1   │ │  Pin 2   │ │  Pin 3   │         ║     │
│  ║  │          │ │          │ │          │         ║     │
│  ║  └──────────┘ └──────────┘ └──────────┘         ║     │
│  ║                                                    ║     │
│  ║  [Edit Profile] (only for own profile)            ║     │
│  ╚═══════════════════════════════════════════════════╝     │
└─────────────────────────────────────────────────────────────┘
```

#### 4.2 Profile Edit Mode

```
USER ACTION: Click "Edit Profile" (own profile only)

┌─────────────────────────────────────────────────────────────┐
│  ╔═══════════════════════════════════════════════════╗     │
│  ║  Edit Profile                              [Save] ║     │
│  ║                                            [Cancel]║     │
│  ║                                                    ║     │
│  ║  Avatar                                           ║     │
│  ║  ┌──────────┐                                     ║     │
│  ║  │          │ [Change Photo]                      ║     │
│  ║  │  Current │ [Remove]                            ║     │
│  ║  │          │                                     ║     │
│  ║  └──────────┘                                     ║     │
│  ║                                                    ║     │
│  ║  Username                                         ║     │
│  ║  ┌──────────────────────────────────────────┐    ║     │
│  ║  │ @johndoe123                              │    ║     │
│  ║  └──────────────────────────────────────────┘    ║     │
│  ║  ✓ Username is available                         ║     │
│  ║                                                    ║     │
│  ║  Display Name                                     ║     │
│  ║  ┌──────────────────────────────────────────┐    ║     │
│  ║  │ John Doe                                 │    ║     │
│  ║  └──────────────────────────────────────────┘    ║     │
│  ║                                                    ║     │
│  ║  Bio                                              ║     │
│  ║  ┌──────────────────────────────────────────┐    ║     │
│  ║  │ Coffee enthusiast and urban explorer     │    ║     │
│  ║  │                                          │    ║     │
│  ║  └──────────────────────────────────────────┘    ║     │
│  ║                                     50/150        ║     │
│  ╚═══════════════════════════════════════════════════╝     │
└─────────────────────────────────────────────────────────────┘
```

#### 4.3 Avatar Upload Flow

```
┌────────┐     ┌──────────┐     ┌─────────┐     ┌──────────┐     ┌──────────┐
│  User  │     │  Upload  │     │  Hook   │     │  Image   │     │   API    │
│        │     │Component │     │useAvatar│     │ Process  │     │  Client  │
└───┬────┘     └────┬─────┘     └────┬────┘     └────┬─────┘     └────┬─────┘
    │               │                 │                │                │
    │ Select file   │                 │                │                │
    ├──────────────►│                 │                │                │
    │               │ Validate        │                │                │
    │               │ - Size < 5MB    │                │                │
    │               │ - Type: jpg/png │                │                │
    │               ├────────────────►│                │                │
    │               │                 │ Create preview │                │
    │               │                 ├───────────────►│                │
    │               │                 │                │ Resize to     │
    │               │                 │                │ 200x200       │
    │               │                 │◄───────────────┤                │
    │               │◄────────────────┤ Preview URL    │                │
    │◄──────────────┤ Show preview    │                │                │
    │               │                 │                │                │
    │ Confirm       │                 │                │                │
    ├──────────────►│                 │                │                │
    │               │ Upload          │                │                │
    │               ├────────────────►│                │                │
    │               │                 │ Convert to     │                │
    │               │                 │ base64/blob    │                │
    │               │                 ├─────────────────────────────────►
    │               │                 │                │ PUT /users/me  │
    │               │                 │                │ /avatar        │
    │               │                 │◄─────────────────────────────────
    │               │                 │ Updated URL    │                │
    │               │◄────────────────┤                │                │
    │◄──────────────┤ Update avatar   │                │                │
```

---

### 5. Map Interaction Flow

#### 5.1 Initial Map Load with Geolocation

```
┌────────┐     ┌──────────┐     ┌─────────┐     ┌──────────┐     ┌──────────┐
│  User  │     │   Map    │     │Geolocation│    │  Pins    │     │   API    │
│        │     │Component │     │  Hook    │     │  Query   │     │  Client  │
└───┬────┘     └────┬─────┘     └────┬────┘     └────┬─────┘     └────┬─────┘
    │               │                 │                │                │
    │ Load page     │                 │                │                │
    ├──────────────►│                 │                │                │
    │               │ Initialize map  │                │                │
    │               │ Request location│                │                │
    │               ├────────────────►│                │                │
    │               │                 │ Show prompt    │                │
    │◄──────────────────────────────────────────────────                │
    │ [Browser location permission prompt]             │                │
    │                                                   │                │
    │ Allow location│                 │                │                │
    ├──────────────────────────────────►                │                │
    │               │                 │ Get coords     │                │
    │               │◄────────────────┤ lat/lng        │                │
    │               │ Center map      │                │                │
    │               │ Load pins       │                │                │
    │               ├─────────────────────────────────►│                │
    │               │                 │                │ Query by       │
    │               │                 │                │ geohash        │
    │               │                 │                ├───────────────►│
    │               │                 │                │ GET /pins      │
    │               │                 │                │ ?geohash=xxx   │
    │               │                 │                │◄───────────────┤
    │               │                 │                │ Pins (24h old) │
    │               │◄─────────────────────────────────┤                │
    │               │ Render markers  │                │                │
    │◄──────────────┤                 │                │                │
```

#### 5.2 Pan/Zoom Triggering New Queries

```
USER ACTION: Pan or zoom the map

┌────────┐     ┌──────────┐     ┌─────────┐     ┌──────────┐     ┌──────────┐
│  User  │     │   Map    │     │ Viewport│     │  Pins    │     │   API    │
│        │     │  Events  │     │  Hook   │     │  Query   │     │  Client  │
└───┬────┘     └────┬─────┘     └────┬────┘     └────┬─────┘     └────┬─────┘
    │               │                 │                │                │
    │ Pan/Zoom map  │                 │                │                │
    ├──────────────►│                 │                │                │
    │               │ onMoveEnd       │                │                │
    │               ├────────────────►│                │                │
    │               │                 │ Get bounds     │                │
    │               │                 │ Calculate      │                │
    │               │                 │ geohash        │                │
    │               │                 │ precision      │                │
    │               │                 │                │                │
    │               │                 │ Debounce 500ms │                │
    │               │                 ├─┐              │                │
    │               │                 │ │ Wait         │                │
    │               │                 │◄┘              │                │
    │               │                 │                │                │
    │               │                 │ Trigger query  │                │
    │               │                 ├───────────────►│                │
    │               │                 │                │ Check cache    │
    │               │                 │                │ If stale →     │
    │               │                 │                ├───────────────►│
    │               │                 │                │ GET /pins      │
    │               │                 │                │ ?geohash=new   │
    │               │                 │                │◄───────────────┤
    │               │                 │                │ Updated pins   │
    │               │                 │◄───────────────┤                │
    │               │◄────────────────┤ New markers    │                │
    │◄──────────────┤ Update view     │                │                │
```

#### 5.3 Pin Clustering at Different Zoom Levels

```
ZOOM LEVEL 10-12: City view - Heavy clustering
┌─────────────────────────────────────┐
│         (42)    (18)                │
│           ⭕      ⭕                  │
│                         (7)          │
│    (23)                 ⭕           │
│     ⭕                               │
└─────────────────────────────────────┘

ZOOM LEVEL 13-15: Neighborhood - Moderate clustering
┌─────────────────────────────────────┐
│    📍📍  (8)    📍                   │
│      📍   ⭕     📍📍                 │
│                    📍  (3)           │
│    📍  📍          📍  ⭕            │
│     📍                               │
└─────────────────────────────────────┘

ZOOM LEVEL 16+: Street level - Individual pins
┌─────────────────────────────────────┐
│    📍  📍  📍    📍                   │
│      📍      📍    📍                 │
│         📍      📍                    │
│    📍  📍     📍    📍                │
│     📍          📍                    │
└─────────────────────────────────────┘

CLUSTERING LOGIC:
- Supercluster library for performance
- Dynamic cluster radius based on zoom
- Click cluster to zoom in
- Smooth transitions between levels
```

---

## State Management Patterns

### Zustand Store Structure

```typescript
// features/auth/stores/authStore.ts
interface AuthState {
  // State
  user: User | null
  token: string | null
  isAuthenticated: boolean
  isLoading: boolean
  
  // Actions
  login: (provider: 'google' | 'github') => Promise<void>
  logout: () => Promise<void>
  setAuth: (user: User, token: string) => void
  clearAuth: () => void
  refreshToken: () => Promise<void>
  updateUser: (updates: Partial<User>) => void
}

// features/pins/stores/pinsStore.ts
interface PinsState {
  // State
  selectedPin: Pin | null
  isCreating: boolean
  optimisticPins: Pin[]
  
  // Actions
  selectPin: (pin: Pin | null) => void
  setCreating: (creating: boolean) => void
  addOptimisticPin: (pin: Pin) => void
  removeOptimisticPin: (id: string) => void
  clearOptimisticPins: () => void
}

// features/map/stores/mapStore.ts
interface MapState {
  // State
  viewport: {
    center: [number, number]
    zoom: number
    bearing: number
    pitch: number
  }
  userLocation: [number, number] | null
  isTracking: boolean
  
  // Actions
  setViewport: (viewport: Partial<Viewport>) => void
  setUserLocation: (location: [number, number]) => void
  setTracking: (tracking: boolean) => void
  flyTo: (location: [number, number], zoom?: number) => void
}
```

### React Query Patterns

```typescript
// features/pins/hooks/usePins.ts
export const usePins = (geohash: string) => {
  return useQuery({
    queryKey: ['pins', geohash],
    queryFn: () => pinService.listPins({ geohash }),
    staleTime: 30 * 1000, // 30 seconds
    cacheTime: 5 * 60 * 1000, // 5 minutes
    refetchOnWindowFocus: true,
    refetchInterval: 60 * 1000, // Refetch every minute
  })
}

// features/pins/hooks/useCreatePin.ts
export const useCreatePin = () => {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: pinService.createPin,
    
    // Optimistic update
    onMutate: async (newPin) => {
      await queryClient.cancelQueries({ queryKey: ['pins'] })
      
      const previousPins = queryClient.getQueryData(['pins'])
      
      queryClient.setQueryData(['pins'], (old: Pin[]) => {
        return [...old, { ...newPin, id: tempId(), isPending: true }]
      })
      
      return { previousPins }
    },
    
    // Rollback on error
    onError: (err, newPin, context) => {
      queryClient.setQueryData(['pins'], context?.previousPins)
      toast.error('Failed to create pin')
    },
    
    // Always refetch
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['pins'] })
    },
    
    onSuccess: (data) => {
      toast.success('Pin created successfully!')
    }
  })
}
```

---

## Error Handling & Loading States

### Loading Patterns

```typescript
// Skeleton loading for pin list
const PinListSkeleton = () => (
  <div className="space-y-4">
    {[...Array(3)].map((_, i) => (
      <div key={i} className="animate-pulse">
        <div className="h-4 bg-gray-200 rounded w-3/4 mb-2" />
        <div className="h-3 bg-gray-200 rounded w-1/2" />
      </div>
    ))}
  </div>
)

// Loading states in components
const PinList = () => {
  const { data: pins, isLoading, error } = usePins()
  
  if (isLoading) return <PinListSkeleton />
  if (error) return <ErrorMessage error={error} />
  if (!pins?.length) return <EmptyState />
  
  return <PinGrid pins={pins} />
}
```

### Error Boundaries

```typescript
// Global error boundary
class ErrorBoundary extends Component {
  state = { hasError: false, error: null }
  
  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error }
  }
  
  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Error boundary caught:', error, errorInfo)
    // Send to error tracking service
  }
  
  render() {
    if (this.state.hasError) {
      return (
        <ErrorFallback
          error={this.state.error}
          resetError={() => this.setState({ hasError: false })}
        />
      )
    }
    
    return this.props.children
  }
}
```

---

## Implementation Checklist

### Phase 1: Authentication Foundation
- [ ] Create `/features/auth/` folder structure
- [ ] Implement AuthModal component with OAuth buttons
- [ ] Add Turnstile widget integration
- [ ] Build username picker form with validation
- [ ] Create useAuth hook with login/logout
- [ ] Set up authStore with Zustand
- [ ] Implement JWT token persistence
- [ ] Add token refresh mechanism
- [ ] Create protected route wrapper

### Phase 2: Pin Management
- [ ] Create `/features/pins/` folder structure
- [ ] Build PinCreationDrawer component
- [ ] Implement form validation and character counter
- [ ] Add photo upload with preview
- [ ] Create PinDetailsPanel with comments
- [ ] Build PinPreviewTooltip component
- [ ] Implement useCreatePin with optimistic updates
- [ ] Add useDeletePin with confirmation
- [ ] Create comment system with real-time updates

### Phase 3: User Profiles
- [ ] Create `/features/user/` folder structure
- [ ] Build UserProfileModal component
- [ ] Implement ProfileEditForm with inline editing
- [ ] Add AvatarUploader with image processing
- [ ] Create useUserProfile hook
- [ ] Implement useUpdateProfile with validation
- [ ] Add user pins list component

### Phase 4: Map Integration
- [ ] Connect pin creation to map clicks
- [ ] Implement viewport-based pin queries
- [ ] Add clustering at different zoom levels
- [ ] Create smooth transitions between states
- [ ] Add loading states for pin fetching
- [ ] Implement error handling for failed queries

### Phase 5: Polish & Optimization
- [ ] Add smooth animations and transitions
- [ ] Implement proper error boundaries
- [ ] Add loading skeletons for all data fetching
- [ ] Optimize bundle size with code splitting
- [ ] Add proper TypeScript types everywhere
- [ ] Implement comprehensive error handling
- [ ] Add success/error toast notifications
- [ ] Test all user flows end-to-end

---

## Technical Specifications

### API Client Configuration

```typescript
// common/services/api/client.ts
import { createConnectTransport } from "@connectrpc/connect-web"

export const transport = createConnectTransport({
  baseUrl: import.meta.env.VITE_API_URL || "http://localhost:8080",
  interceptors: [
    (next) => async (req) => {
      const token = authStore.getState().token
      if (token) {
        req.header.set("Authorization", `Bearer ${token}`)
      }
      return await next(req)
    }
  ],
})
```

### Environment Variables

```env
VITE_API_URL=http://localhost:8080
VITE_TURNSTILE_SITE_KEY=xxx
VITE_MAPBOX_TOKEN=xxx
VITE_ENABLE_DEVTOOLS=true
```

### Performance Considerations

1. **Code Splitting**: Lazy load feature modules
2. **Image Optimization**: Resize and compress before upload
3. **Debouncing**: Map movements, search inputs, validations
4. **Caching**: Aggressive React Query caching for pins
5. **Optimistic Updates**: Instant UI feedback for mutations
6. **Virtual Scrolling**: For long lists of pins/comments
7. **Service Worker**: For offline support and caching

---

## Conclusion

This documentation provides a complete blueprint for implementing the frontend of the Alunalun platform. Each user interaction is detailed with visual representations, data flows, and technical specifications. Developers can use this as a step-by-step guide to build each feature with confidence, knowing exactly what components to create, what hooks to implement, and how data should flow through the application.

The 24-hour pin visibility window ensures the map stays fresh and relevant, while the comprehensive state management patterns ensure a smooth, responsive user experience with proper optimistic updates and error handling.