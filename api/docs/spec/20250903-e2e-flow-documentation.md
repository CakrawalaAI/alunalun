# End-to-End Flow Documentation
**Date**: 2025-09-03  
**Status**: Implementation Complete

## Overview

This document details the complete request/response flows for the Alunalun collaborative map platform using ASCII sequence diagrams for each user journey.

## Architecture Layers

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLIENT                                   │
├─────────────────────────────────────────────────────────────────┤
│                    HTTP/ConnectRPC                               │
├─────────────────────────────────────────────────────────────────┤
│                     MIDDLEWARE                                   │
│                 (Auth Interceptor)                               │
├─────────────────────────────────────────────────────────────────┤
│                      SERVICES                                    │
│           (Auth, User, Pin Services)                             │
├─────────────────────────────────────────────────────────────────┤
│                     PROTOCONV                                    │
│              (Type Conversions)                                  │
├─────────────────────────────────────────────────────────────────┤
│                    REPOSITORY                                    │
│               (SQLC Generated)                                   │
├─────────────────────────────────────────────────────────────────┤
│                  PostgreSQL + PostGIS                            │
└─────────────────────────────────────────────────────────────────┘
```

## User Journey Slices

### 1. Public Map Viewing - Anonymous User Views Pins

```
┌────────┐     ┌────────┐     ┌──────────┐     ┌─────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│ Client │     │ Server │     │Middleware│     │   Pin   │     │ProtoConv │     │Repository│     │ Database │
│        │     │        │     │   Auth   │     │ Service │     │          │     │  (SQLC)  │     │(PostGIS) │
└───┬────┘     └───┬────┘     └────┬─────┘     └────┬────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
    │              │               │                │                │                │                │
    │ GET /pins    │               │                │                │                │                │
    │ ?lat=37.7749 │               │                │                │                │                │
    │ &lng=-122.41 │               │                │                │                │                │
    │ &zoom=15     │               │                │                │                │                │
    ├─────────────►│               │                │                │                │                │
    │              │               │                │                │                │                │
    │              │ Route to      │                │                │                │                │
    │              │ PinService    │                │                │                │                │
    │              ├──────────────►│                │                │                │                │
    │              │               │                │                │                │                │
    │              │               │ Check if       │                │                │                │
    │              │               │ public endpoint│                │                │                │
    │              │               │ (/ListPins)    │                │                │                │
    │              │               │                │                │                │                │
    │              │               │ ✓ Public      │                │                │                │
    │              │               │ Skip auth     │                │                │                │
    │              │               ├───────────────►│                │                │                │
    │              │               │                │                │                │                │
    │              │               │                │ Extract params │                │                │
    │              │               │                │ zoom=15        │                │                │
    │              │               │                │                │                │                │
    │              │               │                │ Calculate      │                │                │
    │              │               │                │ precision=5    │                │                │
    │              │               │                │ limit=100      │                │                │
    │              │               │                │ geohash="9q8yy"│                │                │
    │              │               │                │                │                │                │
    │              │               │                │ ListPinsByGeohash("9q8yy", 100)    │                │
    │              │               │                ├───────────────────────────────────►│                │
    │              │               │                │                │                │                │
    │              │               │                │                │                │ SELECT p.*, pl.*│
    │              │               │                │                │                │ WHERE          │
    │              │               │                │                │                │ p.type = 'pin' │
    │              │               │                │                │                │ AND pl.geohash │
    │              │               │                │                │                │ LIKE '9q8yy%'  │
    │              │               │                │                │                │ AND p.created_at│
    │              │               │                │                │                │ > NOW() - '24h'│
    │              │               │                │                │                ├───────────────►│
    │              │               │                │                │                │                │
    │              │               │                │                │                │◄───────────────┤
    │              │               │                │                │                │ rows[]         │
    │              │               │                │                │◄───────────────┤                │
    │              │               │                │                │ []Post,        │                │
    │              │               │                │                │ []Location     │                │
    │              │               │                │                │                │                │
    │              │               │                │ Convert to Proto                │                │
    │              │               │                ├───────────────►│                │                │
    │              │               │                │                │ PinToProto()   │                │
    │              │               │                │                │ LocationToProto│                │
    │              │               │                │                │ UserToProto()  │                │
    │              │               │                │◄───────────────┤                │                │
    │              │               │                │ []Pin entities │                │                │
    │              │               │◄───────────────┤                │                │                │
    │              │               │ ListPinsResponse               │                │                │
    │              │◄──────────────┤                │                │                │                │
    │              │ HTTP 200      │                │                │                │                │
    │◄─────────────┤ []Pin         │                │                │                │                │
    │              │               │                │                │                │                │
```

**Data Transformations:**
- Client params → URL query params
- zoom=15 → geohash precision=5
- lat/lng → geohash="9q8yy"
- SQLC models → Protobuf entities
- PostGIS POINT → lat/lng floats

**Gradual Cleanup Strategy:**
- Only pins created within last 24 hours are shown
- Keeps map view clean and relevant
- Old pins automatically disappear from public view
- Prevents map overcrowding with stale content

---

### 2. OAuth Authentication Flow - Google Login

```
┌────────┐     ┌────────┐     ┌─────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐     ┌────────┐
│ Client │     │ Server │     │  OAuth  │     │  Google  │     │UserStore │     │ Database │     │ Google │
│        │     │        │     │ Handler │     │ Provider │     │          │     │          │     │  OAuth │
└───┬────┘     └───┬────┘     └────┬────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘     └───┬────┘
    │              │               │                │                │                │                │
    │ GET /auth/   │               │                │                │                │                │
    │ oauth/google │               │                │                │                │                │
    │ ?redirect=/map               │                │                │                │                │
    ├─────────────►│               │                │                │                │                │
    │              │               │                │                │                │                │
    │              │ Route to      │                │                │                │                │
    │              │ OAuth handler │                │                │                │                │
    │              ├──────────────►│                │                │                │                │
    │              │               │                │                │                │                │
    │              │               │ Generate       │                │                │                │
    │              │               │ encrypted state│                │                │                │
    │              │               │ (AES-256-GCM) │                │                │                │
    │              │               │                │                │                │                │
    │              │               │ Build OAuth URL│                │                │                │
    │              │               ├───────────────►│                │                │                │
    │              │               │                │ GetAuthURL()   │                │                │
    │              │               │◄───────────────┤                │                │                │
    │              │               │ OAuth URL      │                │                │                │
    │              │◄──────────────┤                │                │                │                │
    │              │ 302 Redirect  │                │                │                │                │
    │◄─────────────┤               │                │                │                │                │
    │              │               │                │                │                │                │
    │ Redirect to Google           │                │                │                │                │
    ├──────────────────────────────────────────────────────────────────────────────────────────────────►│
    │              │               │                │                │                │                │
    │              │               │                │                │                │                │ User
    │              │               │                │                │                │                │ Auth
    │◄──────────────────────────────────────────────────────────────────────────────────────────────────┤
    │              │               │                │                │                │                │ code
    │              │               │                │                │                │                │
    │ GET /auth/   │               │                │                │                │                │
    │ oauth/google/│               │                │                │                │                │
    │ callback     │               │                │                │                │                │
    │ ?code=xxx    │               │                │                │                │                │
    │ &state=xxx   │               │                │                │                │                │
    ├─────────────►│               │                │                │                │                │
    │              │               │                │                │                │                │
    │              │ Route to      │                │                │                │                │
    │              │ callback      │                │                │                │                │
    │              ├──────────────►│                │                │                │                │
    │              │               │                │                │                │                │
    │              │               │ Validate state │                │                │                │
    │              │               │ (decrypt)      │                │                │                │
    │              │               │                │                │                │                │
    │              │               │ Exchange code  │                │                │                │
    │              │               ├───────────────►│                │                │                │
    │              │               │                │ ExchangeCode() │                │                │
    │              │               │                ├───────────────────────────────────────────────────►│
    │              │               │                │                │                │                │ tokens
    │              │               │                │◄───────────────────────────────────────────────────┤
    │              │               │                │                │                │                │
    │              │               │                │ Verify ID token│                │                │
    │              │               │◄───────────────┤                │                │                │
    │              │               │ UserInfo       │                │                │                │
    │              │               │                │                │                │                │
    │              │               │ Find/Create user                │                │                │
    │              │               ├───────────────────────────────►│                │                │
    │              │               │                │                │ GetUserByEmail │                │
    │              │               │                │                ├───────────────►│                │
    │              │               │                │                │                │ SELECT * FROM  │
    │              │               │                │                │                │ users WHERE... │
    │              │               │                │                │◄───────────────┤                │
    │              │               │                │                │ User or NULL   │                │
    │              │               │                │                │                │                │
    │              │               │                │                │ [If not exists]│                │
    │              │               │                │                │ CreateUser()   │                │
    │              │               │                │                ├───────────────►│                │
    │              │               │                │                │                │ INSERT INTO    │
    │              │               │                │                │                │ users...       │
    │              │               │                │                │◄───────────────┤                │
    │              │               │◄───────────────────────────────┤                │                │
    │              │               │ User           │                │                │                │
    │              │               │                │                │                │                │
    │              │               │ Generate JWT   │                │                │                │
    │              │               │ (RS256, 1hr)   │                │                │                │
    │              │               │                │                │                │                │
    │              │◄──────────────┤                │                │                │                │
    │              │ 302 /map      │                │                │                │                │
    │◄─────────────┤ ?token=JWT    │                │                │                │                │
    │              │               │                │                │                │                │
```

**Data Transformations:**
- redirect_uri → Encrypted state (AES-256-GCM)
- Google auth code → Access/ID tokens
- ID token → UserInfo (email, name, picture)
- UserInfo → auth.User → repository.User
- JSONB metadata stores provider details
- User → JWT claims → Signed token

---

### 3. Creating a Pin - Authenticated User

```
┌────────┐     ┌────────┐     ┌──────────┐     ┌─────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│ Client │     │ Server │     │Middleware│     │   Pin   │     │ProtoConv │     │Repository│     │ Database │
│        │     │        │     │   Auth   │     │ Service │     │          │     │  (SQLC)  │     │(PostGIS) │
└───┬────┘     └───┬────┘     └────┬─────┘     └────┬────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
    │              │               │                │                │                │                │
    │ POST /pins   │               │                │                │                │                │
    │ Auth: Bearer │               │                │                │                │                │
    │ <JWT>        │               │                │                │                │                │
    │ Body: {      │               │                │                │                │                │
    │  content:    │               │                │                │                │                │
    │  "Great!"    │               │                │                │                │                │
    │  location:{  │               │                │                │                │                │
    │   lat:37.77  │               │                │                │                │                │
    │   lng:-122.4 │               │                │                │                │                │
    │  }           │               │                │                │                │                │
    │ }            │               │                │                │                │                │
    ├─────────────►│               │                │                │                │                │
    │              │               │                │                │                │                │
    │              │ Route to      │                │                │                │                │
    │              │ PinService    │                │                │                │                │
    │              ├──────────────►│                │                │                │                │
    │              │               │                │                │                │                │
    │              │               │ Check if       │                │                │                │
    │              │               │ public endpoint│                │                │                │
    │              │               │ (/CreatePin)   │                │                │                │
    │              │               │                │                │                │                │
    │              │               │ ✗ Protected    │                │                │                │
    │              │               │                │                │                │                │
    │              │               │ Extract token  │                │                │                │
    │              │               │ from Bearer    │                │                │                │
    │              │               │                │                │                │                │
    │              │               │ ValidateToken()│                │                │                │
    │              │               │ - Verify RS256 │                │                │                │
    │              │               │ - Check expiry │                │                │                │
    │              │               │                │                │                │                │
    │              │               │ Add claims     │                │                │                │
    │              │               │ to context     │                │                │                │
    │              │               ├───────────────►│                │                │                │
    │              │               │                │                │                │                │
    │              │               │                │ Get claims     │                │                │
    │              │               │                │ from context   │                │                │
    │              │               │                │ UserID="user-123"                │                │
    │              │               │                │                │                │                │
    │              │               │                │ Begin Transaction                │                │
    │              │               │                ├─────────────────────────────────►│                │
    │              │               │                │                │                ├───────────────►│
    │              │               │                │                │                │ BEGIN          │
    │              │               │                │                │                │◄───────────────┤
    │              │               │                │                │                │                │
    │              │               │                │ Convert to params                │                │
    │              │               │                ├───────────────►│                │                │
    │              │               │                │                │ProtoToCreatePin│                │
    │              │               │                │                │Params()        │                │
    │              │               │                │                │- ID: pin-<ksuid>               │
    │              │               │                │                │- Type: "pin"   │                │
    │              │               │                │                │- POINT(-122.4  │                │
    │              │               │                │                │   37.77)       │                │
    │              │               │                │                │- geohash:      │                │
    │              │               │                │                │  "9q8yyk2g"    │                │
    │              │               │                │◄───────────────┤                │                │
    │              │               │                │ PostParams,    │                │                │
    │              │               │                │ LocationParams │                │                │
    │              │               │                │                │                │                │
    │              │               │                │ CreatePost()   │                │                │
    │              │               │                ├───────────────────────────────►│                │
    │              │               │                │                │                │ INSERT INTO    │
    │              │               │                │                │                │ posts...       │
    │              │               │                │                │                ├───────────────►│
    │              │               │                │                │                │◄───────────────┤
    │              │               │                │                │◄───────────────┤                │
    │              │               │                │                │ Post           │                │
    │              │               │                │                │                │                │
    │              │               │                │ CreatePostLocation()            │                │
    │              │               │                ├───────────────────────────────►│                │
    │              │               │                │                │                │ INSERT INTO    │
    │              │               │                │                │                │ posts_location │
    │              │               │                │                │                ├───────────────►│
    │              │               │                │                │                │◄───────────────┤
    │              │               │                │                │◄───────────────┤                │
    │              │               │                │                │ Location       │                │
    │              │               │                │                │                │                │
    │              │               │                │ Commit Transaction               │                │
    │              │               │                ├─────────────────────────────────►│                │
    │              │               │                │                │                ├───────────────►│
    │              │               │                │                │                │ COMMIT         │
    │              │               │                │                │                │◄───────────────┤
    │              │               │                │                │                │                │
    │              │               │                │ Get author info│                │                │
    │              │               │                ├───────────────────────────────►│                │
    │              │               │                │                │                │ SELECT * FROM  │
    │              │               │                │                │                │ users WHERE... │
    │              │               │                │                │◄───────────────┤                │
    │              │               │                │                │ User           │                │
    │              │               │                │                │                │                │
    │              │               │                │ Convert to Proto                │                │
    │              │               │                ├───────────────►│                │                │
    │              │               │                │                │ PinToProto()   │                │
    │              │               │                │◄───────────────┤                │                │
    │              │               │                │ Pin entity     │                │                │
    │              │               │◄───────────────┤                │                │                │
    │              │               │ CreatePinResponse               │                │                │
    │              │◄──────────────┤                │                │                │                │
    │              │ HTTP 200      │                │                │                │                │
    │◄─────────────┤ Pin created   │                │                │                │                │
    │              │               │                │                │                │                │
```

**Data Transformations:**
- JWT Bearer token → Validated claims
- Claims → User context (ID, username)
- Request body → Protobuf CreatePinRequest
- lat/lng → PostGIS POINT geometry
- lat/lng → Geohash string (precision 8)
- Generate pin ID: "pin-" + ksuid
- Transaction ensures atomicity
- Repository models → Protobuf Pin entity

---

### 4. Viewing Pin Details - With Comments

```
┌────────┐     ┌────────┐     ┌──────────┐     ┌─────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│ Client │     │ Server │     │Middleware│     │   Pin   │     │ProtoConv │     │Repository│     │ Database │
│        │     │        │     │   Auth   │     │ Service │     │          │     │  (SQLC)  │     │(PostGIS) │
└───┬────┘     └───┬────┘     └────┬─────┘     └────┬────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
    │              │               │                │                │                │                │
    │ GET /pins/   │               │                │                │                │                │
    │ pin-k2j3h4   │               │                │                │                │                │
    ├─────────────►│               │                │                │                │                │
    │              │               │                │                │                │                │
    │              │ Route to      │                │                │                │                │
    │              │ PinService    │                │                │                │                │
    │              ├──────────────►│                │                │                │                │
    │              │               │                │                │                │                │
    │              │               │ Check if       │                │                │                │
    │              │               │ public endpoint│                │                │                │
    │              │               │ (/GetPin)      │                │                │                │
    │              │               │                │                │                │                │
    │              │               │ ✓ Public      │                │                │                │
    │              │               │ (optional auth)│                │                │                │
    │              │               ├───────────────►│                │                │                │
    │              │               │                │                │                │                │
    │              │               │                │ GetPinWithLocation()            │                │
    │              │               │                ├───────────────────────────────►│                │
    │              │               │                │                │                │ SELECT p.*, pl.*│
    │              │               │                │                │                │ FROM posts p   │
    │              │               │                │                │                │ JOIN posts_    │
    │              │               │                │                │                │ location pl    │
    │              │               │                │                │                │ WHERE p.id=?   │
    │              │               │                │                │                ├───────────────►│
    │              │               │                │                │                │◄───────────────┤
    │              │               │                │                │◄───────────────┤ Post, Location │
    │              │               │                │                │                │                │
    │              │               │                │ GetUserByID(pin.user_id)        │                │
    │              │               │                ├───────────────────────────────►│                │
    │              │               │                │                │                │ SELECT * FROM  │
    │              │               │                │                │                │ users WHERE    │
    │              │               │                │                │                │ id=?           │
    │              │               │                │                │                ├───────────────►│
    │              │               │                │                │                │◄───────────────┤
    │              │               │                │                │◄───────────────┤ User (author)  │
    │              │               │                │                │                │                │
    │              │               │                │ GetCommentsByPinID()            │                │
    │              │               │                ├───────────────────────────────►│                │
    │              │               │                │                │                │ SELECT * FROM  │
    │              │               │                │                │                │ posts WHERE    │
    │              │               │                │                │                │ parent_id=?    │
    │              │               │                │                │                │ AND type=      │
    │              │               │                │                │                │ 'comment'      │
    │              │               │                │                │                ├───────────────►│
    │              │               │                │                │                │◄───────────────┤
    │              │               │                │                │◄───────────────┤ []Comment      │
    │              │               │                │                │                │                │
    │              │               │                │ [For each comment]              │                │
    │              │               │                │ GetUserByID(comment.user_id)    │                │
    │              │               │                ├───────────────────────────────►│                │
    │              │               │                │                │◄───────────────┤                │
    │              │               │                │                │ Comment authors│                │
    │              │               │                │                │                │                │
    │              │               │                │ Convert all to Proto            │                │
    │              │               │                ├───────────────►│                │                │
    │              │               │                │                │ PinToProto()   │                │
    │              │               │                │                │ CommentToProto()               │
    │              │               │                │◄───────────────┤                │                │
    │              │               │                │ Pin with       │                │                │
    │              │               │                │ []Comment      │                │                │
    │              │               │◄───────────────┤                │                │                │
    │              │               │ GetPinResponse │                │                │                │
    │              │◄──────────────┤                │                │                │                │
    │              │ HTTP 200      │                │                │                │                │
    │◄─────────────┤ Pin + Comments│                │                │                │                │
    │              │               │                │                │                │                │
```

**Data Transformations:**
- Pin ID from URL path
- Multiple database queries (pin, author, comments)
- PostGIS POINT → lat/lng coordinates
- JSONB metadata → structured user info
- Repository models → Protobuf entities
- Hierarchical data assembly (pin → comments)

---

### 5. Adding a Comment - Authenticated User

```
┌────────┐     ┌────────┐     ┌──────────┐     ┌─────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│ Client │     │ Server │     │Middleware│     │   Pin   │     │ProtoConv │     │Repository│     │ Database │
│        │     │        │     │   Auth   │     │ Service │     │          │     │  (SQLC)  │     │(PostGIS) │
└───┬────┘     └───┬────┘     └────┬─────┘     └────┬────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
    │              │               │                │                │                │                │
    │ POST /pins/  │               │                │                │                │                │
    │ pin-k2j3h4/  │               │                │                │                │                │
    │ comments     │               │                │                │                │                │
    │ Auth: Bearer │               │                │                │                │                │
    │ <JWT>        │               │                │                │                │                │
    │ Body: {      │               │                │                │                │                │
    │  content:    │               │                │                │                │                │
    │  "Nice spot!"│               │                │                │                │                │
    │ }            │               │                │                │                │                │
    ├─────────────►│               │                │                │                │                │
    │              │               │                │                │                │                │
    │              │ Route to      │                │                │                │                │
    │              │ PinService    │                │                │                │                │
    │              ├──────────────►│                │                │                │                │
    │              │               │                │                │                │                │
    │              │               │ ✗ Protected    │                │                │                │
    │              │               │ Validate JWT   │                │                │                │
    │              │               │                │                │                │                │
    │              │               │ Add claims     │                │                │                │
    │              │               │ to context     │                │                │                │
    │              │               ├───────────────►│                │                │                │
    │              │               │                │                │                │                │
    │              │               │                │ Get claims     │                │                │
    │              │               │                │ UserID="user-456"                │                │
    │              │               │                │                │                │                │
    │              │               │                │ Verify pin exists               │                │
    │              │               │                ├───────────────────────────────►│                │
    │              │               │                │                │                │ SELECT * FROM  │
    │              │               │                │                │                │ posts WHERE    │
    │              │               │                │                │                │ id=? AND       │
    │              │               │                │                │                │ type='pin'     │
    │              │               │                │                │                ├───────────────►│
    │              │               │                │                │                │◄───────────────┤
    │              │               │                │                │◄───────────────┤ Pin exists     │
    │              │               │                │                │                │                │
    │              │               │                │ Convert to params                │                │
    │              │               │                ├───────────────►│                │                │
    │              │               │                │                │ProtoToCreateComment            │
    │              │               │                │                │Params()        │                │
    │              │               │                │                │- ID: comment-<ksuid>           │
    │              │               │                │                │- Type: "comment"               │
    │              │               │                │                │- parent_id:    │                │
    │              │               │                │                │  "pin-k2j3h4"  │                │
    │              │               │                │◄───────────────┤                │                │
    │              │               │                │ CommentParams  │                │                │
    │              │               │                │                │                │                │
    │              │               │                │ CreatePost()   │                │                │
    │              │               │                ├───────────────────────────────►│                │
    │              │               │                │                │                │ INSERT INTO    │
    │              │               │                │                │                │ posts(id,      │
    │              │               │                │                │                │ user_id, type, │
    │              │               │                │                │                │ parent_id...)  │
    │              │               │                │                │                ├───────────────►│
    │              │               │                │                │                │◄───────────────┤
    │              │               │                │                │◄───────────────┤ Comment        │
    │              │               │                │                │                │                │
    │              │               │                │ Get commenter info              │                │
    │              │               │                ├───────────────────────────────►│                │
    │              │               │                │                │◄───────────────┤ User           │
    │              │               │                │                │                │                │
    │              │               │                │ Convert to Proto                │                │
    │              │               │                ├───────────────►│                │                │
    │              │               │                │                │CommentToProto()│                │
    │              │               │                │◄───────────────┤                │                │
    │              │               │                │ Comment entity │                │                │
    │              │               │◄───────────────┤                │                │                │
    │              │               │ CreateCommentResponse           │                │                │
    │              │◄──────────────┤                │                │                │                │
    │              │ HTTP 200      │                │                │                │                │
    │◄─────────────┤ Comment added │                │                │                │                │
    │              │               │                │                │                │                │
```

**Data Transformations:**
- JWT → User context
- Pin ID from URL path
- Generate comment ID: "comment-" + ksuid
- Type field = "comment" (polymorphic)
- parent_id links to pin
- No location data for comments
- Repository model → Protobuf Comment

---

### 6. Deleting a Pin - Owner Only

```
┌────────┐     ┌────────┐     ┌──────────┐     ┌─────────┐     ┌──────────┐     ┌──────────┐
│ Client │     │ Server │     │Middleware│     │   Pin   │     │Repository│     │ Database │
│        │     │        │     │   Auth   │     │ Service │     │  (SQLC)  │     │(PostGIS) │
└───┬────┘     └───┬────┘     └────┬─────┘     └────┬────┘     └────┬─────┘     └────┬─────┘
    │              │               │                │                │                │
    │ DELETE /pins/│               │                │                │                │
    │ pin-k2j3h4   │               │                │                │                │
    │ Auth: Bearer │               │                │                │                │
    │ <JWT>        │               │                │                │                │
    ├─────────────►│               │                │                │                │
    │              │               │                │                │                │
    │              │ Route to      │                │                │                │
    │              │ PinService    │                │                │                │
    │              ├──────────────►│                │                │                │
    │              │               │                │                │                │
    │              │               │ ✗ Protected    │                │                │
    │              │               │ Validate JWT   │                │                │
    │              │               │                │                │                │
    │              │               │ Add claims     │                │                │
    │              │               │ to context     │                │                │
    │              │               ├───────────────►│                │                │
    │              │               │                │                │                │
    │              │               │                │ Get claims     │                │
    │              │               │                │ UserID="user-123"              │
    │              │               │                │                │                │
    │              │               │                │ Get pin details│                │
    │              │               │                ├───────────────►│                │
    │              │               │                │                │ SELECT * FROM  │
    │              │               │                │                │ posts WHERE    │
    │              │               │                │                │ id=? AND       │
    │              │               │                │                │ type='pin'     │
    │              │               │                │                ├───────────────►│
    │              │               │                │                │◄───────────────┤
    │              │               │                │◄───────────────┤ Pin           │
    │              │               │                │ Pin{user_id:   │                │
    │              │               │                │  "user-123"}   │                │
    │              │               │                │                │                │
    │              │               │                │ Check ownership│                │
    │              │               │                │ pin.user_id == │                │
    │              │               │                │ claims.user_id │                │
    │              │               │                │ ✓ Match        │                │
    │              │               │                │                │                │
    │              │               │                │ DeletePost()   │                │
    │              │               │                ├───────────────►│                │
    │              │               │                │                │ DELETE FROM    │
    │              │               │                │                │ posts WHERE    │
    │              │               │                │                │ id=?           │
    │              │               │                │                ├───────────────►│
    │              │               │                │                │                │ CASCADE:
    │              │               │                │                │                │ - posts_location
    │              │               │                │                │                │ - comments
    │              │               │                │                │◄───────────────┤
    │              │               │                │◄───────────────┤ Deleted       │
    │              │               │                │                │                │
    │              │               │◄───────────────┤                │                │
    │              │               │ DeletePinResponse               │                │
    │              │◄──────────────┤                │                │                │
    │              │ HTTP 200      │                │                │                │
    │◄─────────────┤ Pin deleted   │                │                │                │
    │              │               │                │                │                │
    
    ┌──────────────────────────────────────────────────────────┐
    │ If ownership check fails:                               │
    │                                                          │
    │ pin.user_id="user-999" != claims.user_id="user-123"     │
    │ ✗ No match                                               │
    │                                                          │
    │ Return 403 Forbidden                                     │
    │ "You can only delete your own pins"                     │
    └──────────────────────────────────────────────────────────┘
```

**Security Checks:**
- JWT validation required
- Ownership verification: pin.user_id == claims.user_id
- 403 Forbidden if not owner
- CASCADE deletes handle cleanup

---

### 7. Getting User Profile - Public

```
┌────────┐     ┌────────┐     ┌──────────┐     ┌─────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│ Client │     │ Server │     │Middleware│     │  User   │     │ProtoConv │     │Repository│     │ Database │
│        │     │        │     │   Auth   │     │ Service │     │          │     │  (SQLC)  │     │          │
└───┬────┘     └───┬────┘     └────┬─────┘     └────┬────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
    │              │               │                │                │                │                │
    │ GET /users/  │               │                │                │                │                │
    │ user-123     │               │                │                │                │                │
    ├─────────────►│               │                │                │                │                │
    │              │               │                │                │                │                │
    │              │ Route to      │                │                │                │                │
    │              │ UserService   │                │                │                │                │
    │              ├──────────────►│                │                │                │                │
    │              │               │                │                │                │                │
    │              │               │ Check if       │                │                │                │
    │              │               │ public endpoint│                │                │                │
    │              │               │ (/GetUser)     │                │                │                │
    │              │               │                │                │                │                │
    │              │               │ ✓ Public      │                │                │                │
    │              │               ├───────────────►│                │                │                │
    │              │               │                │                │                │                │
    │              │               │                │ GetUserByID()  │                │                │
    │              │               │                ├───────────────────────────────►│                │
    │              │               │                │                │                │ SELECT * FROM  │
    │              │               │                │                │                │ users WHERE    │
    │              │               │                │                │                │ id=?           │
    │              │               │                │                │                ├───────────────►│
    │              │               │                │                │                │◄───────────────┤
    │              │               │                │                │◄───────────────┤ User           │
    │              │               │                │                │                │                │
    │              │               │                │ GetUserStats() │                │                │
    │              │               │                ├───────────────────────────────►│                │
    │              │               │                │                │                │ SELECT         │
    │              │               │                │                │                │ COUNT(*) pins, │
    │              │               │                │                │                │ COUNT(*) comments
    │              │               │                │                │                ├───────────────►│
    │              │               │                │                │                │◄───────────────┤
    │              │               │                │                │◄───────────────┤ Stats         │
    │              │               │                │                │                │                │
    │              │               │                │ Convert to Proto                │                │
    │              │               │                ├───────────────►│                │                │
    │              │               │                │                │ UserToProto()  │                │
    │              │               │                │                │ Extract metadata               │
    │              │               │                │                │ from JSONB     │                │
    │              │               │                │◄───────────────┤                │                │
    │              │               │                │ User entity    │                │
    │              │               │◄───────────────┤                │                │                │
    │              │               │ GetUserResponse                │                │                │
    │              │◄──────────────┤                │                │                │                │
    │              │ HTTP 200      │                │                │                │                │
    │◄─────────────┤ User profile  │                │                │                │                │
    │              │               │                │                │                │                │
```

**Data Transformations:**
- User ID from URL path
- JSONB metadata → structured fields (picture, first_name, last_name)
- Aggregate stats (pin_count, comment_count)
- Public fields only (no email unless own profile)

---

### 8. Updating User Profile - Self Only

```
┌────────┐     ┌────────┐     ┌──────────┐     ┌─────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│ Client │     │ Server │     │Middleware│     │  User   │     │ProtoConv │     │Repository│     │ Database │
│        │     │        │     │   Auth   │     │ Service │     │          │     │  (SQLC)  │     │          │
└───┬────┘     └───┬────┘     └────┬─────┘     └────┬────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
    │              │               │                │                │                │                │
    │ PUT /users/  │               │                │                │                │                │
    │ user-123     │               │                │                │                │                │
    │ Auth: Bearer │               │                │                │                │                │
    │ <JWT>        │               │                │                │                │                │
    │ Body: {      │               │                │                │                │                │
    │  username:   │               │                │                │                │                │
    │  "newname"   │               │                │                │                │                │
    │ }            │               │                │                │                │                │
    ├─────────────►│               │                │                │                │                │
    │              │               │                │                │                │                │
    │              │ Route to      │                │                │                │                │
    │              │ UserService   │                │                │                │                │
    │              ├──────────────►│                │                │                │                │
    │              │               │                │                │                │                │
    │              │               │ ✗ Protected    │                │                │                │
    │              │               │ Validate JWT   │                │                │                │
    │              │               │                │                │                │                │
    │              │               │ Add claims     │                │                │                │
    │              │               │ to context     │                │                │                │
    │              │               ├───────────────►│                │                │                │
    │              │               │                │                │                │                │
    │              │               │                │ Get claims     │                │                │
    │              │               │                │ UserID="user-123"              │                │
    │              │               │                │                │                │                │
    │              │               │                │ Verify self    │                │                │
    │              │               │                │ req.user_id == │                │                │
    │              │               │                │ claims.user_id │                │                │
    │              │               │                │ ✓ Match        │                │                │
    │              │               │                │                │                │                │
    │              │               │                │ Check username │                │                │
    │              │               │                │ availability   │                │                │
    │              │               │                ├───────────────────────────────►│                │
    │              │               │                │                │                │ SELECT * FROM  │
    │              │               │                │                │                │ users WHERE    │
    │              │               │                │                │                │ username=?     │
    │              │               │                │                │                ├───────────────►│
    │              │               │                │                │                │◄───────────────┤
    │              │               │                │                │◄───────────────┤ Available     │
    │              │               │                │                │                │                │
    │              │               │                │ Convert to params                │                │
    │              │               │                ├───────────────►│                │                │
    │              │               │                │                │ProtoToUpdateUser               │
    │              │               │                │                │Params()        │                │
    │              │               │                │                │ Update metadata│                │
    │              │               │                │                │ JSONB          │                │
    │              │               │                │◄───────────────┤                │                │
    │              │               │                │ UpdateParams   │                │                │
    │              │               │                │                │                │                │
    │              │               │                │ UpdateUser()   │                │                │
    │              │               │                ├───────────────────────────────►│                │
    │              │               │                │                │                │ UPDATE users   │
    │              │               │                │                │                │ SET username=?,│
    │              │               │                │                │                │ metadata=?,    │
    │              │               │                │                │                │ updated_at=NOW()
    │              │               │                │                │                ├───────────────►│
    │              │               │                │                │                │◄───────────────┤
    │              │               │                │                │◄───────────────┤ Updated User  │
    │              │               │                │                │                │                │
    │              │               │                │ Convert to Proto                │                │
    │              │               │                ├───────────────►│                │                │
    │              │               │                │                │ UserToProto()  │                │
    │              │               │                │◄───────────────┤                │                │
    │              │               │                │ User entity    │                │                │
    │              │               │◄───────────────┤                │                │                │
    │              │               │ UpdateUserResponse              │                │                │
    │              │◄──────────────┤                │                │                │                │
    │              │ HTTP 200      │                │                │                │                │
    │◄─────────────┤ User updated  │                │                │                │                │
    │              │               │                │                │                │                │
    
    ┌──────────────────────────────────────────────────────────┐
    │ If self check fails:                                     │
    │                                                          │
    │ req.user_id="user-999" != claims.user_id="user-123"     │
    │ ✗ No match                                               │
    │                                                          │
    │ Return 403 Forbidden                                     │
    │ "You can only update your own profile"                  │
    └──────────────────────────────────────────────────────────┘
```

**Security Checks:**
- JWT validation required
- Self-update only: req.user_id == claims.user_id
- Username uniqueness check
- 403 Forbidden if not self

---

### 9. Refreshing JWT Token

```
┌────────┐     ┌────────┐     ┌──────────┐     ┌─────────┐     ┌──────────┐     ┌──────────┐
│ Client │     │ Server │     │Middleware│     │  Auth   │     │  Token   │     │ Database │
│        │     │        │     │   Auth   │     │ Service │     │ Manager  │     │          │
└───┬────┘     └───┬────┘     └────┬─────┘     └────┬────┘     └────┬─────┘     └────┬─────┘
    │              │               │                │                │                │
    │ POST /auth/  │               │                │                │                │
    │ refresh      │               │                │                │                │
    │ Body: {      │               │                │                │                │
    │  token:      │               │                │                │                │
    │  "<old_jwt>" │               │                │                │                │
    │ }            │               │                │                │                │
    ├─────────────►│               │                │                │                │
    │              │               │                │                │                │
    │              │ Route to      │                │                │                │
    │              │ AuthService   │                │                │                │
    │              ├──────────────►│                │                │                │
    │              │               │                │                │                │
    │              │               │ Check if       │                │                │                │
    │              │               │ public endpoint│                │                │                │
    │              │               │ (/RefreshToken)│                │                │                │
    │              │               │                │                │                │
    │              │               │ ✓ Public      │                │                │                │
    │              │               ├───────────────►│                │                │
    │              │               │                │                │                │
    │              │               │                │ Validate old   │                │
    │              │               │                │ token          │                │
    │              │               │                ├───────────────►│                │
    │              │               │                │                │ ValidateToken()│
    │              │               │                │                │ - Check sig    │
    │              │               │                │                │ - Allow expired│
    │              │               │                │                │   within grace │
    │              │               │                │◄───────────────┤ period (7d)    │
    │              │               │                │ Claims         │                │
    │              │               │                │                │                │
    │              │               │                │ Get fresh user │                │
    │              │               │                │ data           │                │
    │              │               │                ├───────────────────────────────►│
    │              │               │                │                │                │ SELECT * FROM
    │              │               │                │                │                │ users WHERE
    │              │               │                │                │                │ id=?
    │              │               │                │◄───────────────────────────────┤
    │              │               │                │ User           │                │
    │              │               │                │                │                │
    │              │               │                │ Generate new   │                │
    │              │               │                │ token          │                │
    │              │               │                ├───────────────►│                │
    │              │               │                │                │GenerateToken() │
    │              │               │                │                │ - New expiry   │
    │              │               │                │                │   (1hr)        │
    │              │               │                │                │ - Same claims  │
    │              │               │                │                │ - New signature│
    │              │               │                │◄───────────────┤                │
    │              │               │                │ New JWT        │                │
    │              │               │◄───────────────┤                │                │
    │              │               │ RefreshTokenResponse           │                │
    │              │◄──────────────┤                │                │                │
    │              │ HTTP 200      │                │                │                │
    │◄─────────────┤ {token: new}  │                │                │                │
    │              │               │                │                │                │
    
    ┌──────────────────────────────────────────────────────────┐
    │ If token is too old (>7 days expired):                   │
    │                                                          │
    │ Return 401 Unauthorized                                  │
    │ "Token expired. Please login again"                      │
    └──────────────────────────────────────────────────────────┘
```

**Token Refresh Logic:**
- Accept expired tokens within grace period (7 days)
- Fetch fresh user data from database
- Generate new token with updated expiry
- Same user claims, new signature
- 401 if beyond grace period

---

## Key Design Patterns

### Authentication Flow
- **Public endpoints**: Skip auth or optional auth
- **Protected endpoints**: Require valid JWT
- **Owner-only actions**: Additional ownership check

### Data Transformation Pipeline
1. **Client → Server**: HTTP/Protobuf request
2. **Middleware**: Auth validation, context injection
3. **Service**: Business logic, orchestration
4. **ProtoConv**: Type conversions
5. **Repository**: SQLC database operations
6. **Database**: PostGIS spatial queries
7. **Response**: Protobuf → HTTP response

### Error Handling
- **401 Unauthorized**: Invalid/missing token
- **403 Forbidden**: Valid token but insufficient permissions
- **404 Not Found**: Resource doesn't exist
- **409 Conflict**: Username taken, etc.

### Transaction Boundaries
- Pin creation: Transaction for posts + posts_location
- Comment creation: Single insert (no location)
- Deletions: CASCADE handles related records

### Spatial Optimization
- Zoom level determines geohash precision
- Dynamic result limits based on zoom
- PostGIS indexes for efficient queries
- Composite index on (type, created_at, geohash) for optimal query performance

### Temporal Filtering
- **24-hour visibility window**: Pins automatically expire from public view
- **Gradual content cleanup**: Prevents map overcrowding
- **Fresh content focus**: Ensures relevance and engagement
- **Index optimization**: created_at DESC index for efficient filtering

## Performance Considerations

### Database Queries
- **N+1 prevention**: Batch user lookups for comments
- **Index usage**: Geohash prefix matching, user_id lookups, created_at for 24h filter
- **Connection pooling**: 25 max, 5 min connections
- **24-hour filter**: Reduces result set size, improves query performance

### Caching Opportunities
- User profiles (short TTL)
- Pin lists by geohash (invalidate on create/delete)
- JWT validation results (until expiry)

### Scalability Points
- Stateless JWT authentication
- Read replicas for public endpoints
- Geohash sharding for spatial distribution
- CDN for static pin images (future)