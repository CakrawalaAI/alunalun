# Map Pins Integration Feature

## Overview
Integration layer between the posts system and the existing map feature. Renders posts as pins on the map, handles clustering at different zoom levels, and provides interactive popups for post details.

## Database Schema

### Migration: `008_map_optimizations.sql`

```sql
-- Map-specific optimizations for pin rendering
-- (Core tables already exist from posts feature)

-- Clustering cache for different zoom levels
CREATE TABLE map_pin_clusters (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  zoom_level INTEGER NOT NULL,
  cluster_geohash VARCHAR(8) NOT NULL,
  cluster_center GEOMETRY(POINT, 4326) NOT NULL,
  pin_count INTEGER NOT NULL DEFAULT 1,
  pin_ids UUID[] NOT NULL,
  
  -- Cluster metadata
  latest_post_at TIMESTAMP WITH TIME ZONE,
  predominant_author_id UUID,
  
  -- Cache control
  computed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  expires_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW() + INTERVAL '5 minutes',
  
  UNIQUE(zoom_level, cluster_geohash)
);

CREATE INDEX idx_pin_clusters_zoom ON map_pin_clusters(zoom_level, cluster_geohash);
CREATE INDEX idx_pin_clusters_expires ON map_pin_clusters(expires_at);

-- User map preferences
CREATE TABLE user_map_preferences (
  user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  
  -- Display preferences
  show_own_pins BOOLEAN DEFAULT true,
  show_anonymous_pins BOOLEAN DEFAULT true,
  cluster_pins BOOLEAN DEFAULT true,
  
  -- Interaction preferences
  auto_open_created_pin BOOLEAN DEFAULT true,
  follow_user_location BOOLEAN DEFAULT false,
  
  -- Visual preferences
  pin_size VARCHAR(10) DEFAULT 'medium', -- 'small', 'medium', 'large'
  cluster_style VARCHAR(20) DEFAULT 'circles', -- 'circles', 'hexagons', 'squares'
  
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Function to get pins for map bounds with clustering
CREATE OR REPLACE FUNCTION get_map_pins(
  p_bounds_north DOUBLE PRECISION,
  p_bounds_south DOUBLE PRECISION,
  p_bounds_east DOUBLE PRECISION,
  p_bounds_west DOUBLE PRECISION,
  p_zoom_level INTEGER,
  p_max_age_hours INTEGER DEFAULT 168, -- 1 week default
  p_cluster_threshold INTEGER DEFAULT 10
) RETURNS TABLE (
  pin_id UUID,
  pin_type VARCHAR(10), -- 'single' or 'cluster'
  latitude DOUBLE PRECISION,
  longitude DOUBLE PRECISION,
  cluster_count INTEGER,
  content_preview TEXT,
  author_name TEXT,
  created_at TIMESTAMP WITH TIME ZONE
) AS $$
DECLARE
  v_cluster_precision INTEGER;
BEGIN
  -- Determine clustering precision based on zoom level
  v_cluster_precision := CASE 
    WHEN p_zoom_level < 10 THEN 3  -- City level
    WHEN p_zoom_level < 13 THEN 5  -- Neighborhood level
    WHEN p_zoom_level < 16 THEN 7  -- Street level
    ELSE 9                          -- Building level
  END;
  
  -- Return clustered or individual pins based on density
  RETURN QUERY
  WITH pins_in_bounds AS (
    SELECT 
      p.id,
      pl.latitude,
      pl.longitude,
      LEFT(pl.geohash, v_cluster_precision) as cluster_key,
      p.content,
      COALESCE(u.username, 'Anonymous') as author,
      p.created_at
    FROM posts p
    JOIN post_locations pl ON p.id = pl.post_id
    LEFT JOIN users u ON p.author_user_id = u.id
    WHERE ST_Within(
      pl.location,
      ST_MakeEnvelope(p_bounds_west, p_bounds_south, p_bounds_east, p_bounds_north, 4326)
    )
    AND p.deleted_at IS NULL
    AND p.created_at > NOW() - (p_max_age_hours || ' hours')::INTERVAL
  ),
  cluster_groups AS (
    SELECT 
      cluster_key,
      COUNT(*) as pin_count,
      AVG(latitude) as avg_lat,
      AVG(longitude) as avg_lng,
      MAX(created_at) as latest_created,
      ARRAY_AGG(id) as pin_ids,
      ARRAY_AGG(content)[1] as sample_content,
      ARRAY_AGG(author)[1] as sample_author
    FROM pins_in_bounds
    GROUP BY cluster_key
  )
  -- Return clusters for groups above threshold
  SELECT 
    NULL::UUID as pin_id,
    'cluster'::VARCHAR as pin_type,
    avg_lat,
    avg_lng,
    pin_count::INTEGER,
    sample_content,
    sample_author,
    latest_created
  FROM cluster_groups
  WHERE pin_count >= p_cluster_threshold
  
  UNION ALL
  
  -- Return individual pins for groups below threshold
  SELECT 
    p.id,
    'single'::VARCHAR,
    p.latitude,
    p.longitude,
    1::INTEGER,
    LEFT(p.content, 100),
    p.author,
    p.created_at
  FROM pins_in_bounds p
  WHERE p.cluster_key IN (
    SELECT cluster_key 
    FROM cluster_groups 
    WHERE pin_count < p_cluster_threshold
  );
END;
$$ LANGUAGE plpgsql STABLE;
```

## SQLC Queries

### File: `sql/queries/map_pins.sql`

```sql
-- name: GetMapPins :many
-- Get pins for map display with smart clustering
SELECT * FROM get_map_pins($1, $2, $3, $4, $5, $6, $7);

-- name: GetPinDetails :one
-- Get full details for a single pin popup
SELECT 
  p.*,
  pl.longitude,
  pl.latitude,
  pl.place_name,
  pl.city,
  COALESCE(u.username, 'Anonymous') as author_name,
  (p.author_user_id IS NOT NULL) as is_verified_author,
  COUNT(DISTINCT pr.id) as reaction_count,
  COUNT(DISTINCT pc.id) as comment_count,
  EXISTS(
    SELECT 1 FROM post_reactions 
    WHERE post_id = p.id 
    AND (user_id = $2 OR session_id = $3)
  ) as user_has_reacted
FROM posts p
JOIN post_locations pl ON p.id = pl.post_id
LEFT JOIN users u ON p.author_user_id = u.id
LEFT JOIN post_reactions pr ON p.id = pr.post_id
LEFT JOIN post_comments pc ON p.id = pc.post_id AND pc.deleted_at IS NULL
WHERE p.id = $1
GROUP BY p.id, pl.longitude, pl.latitude, pl.place_name, pl.city, u.username;

-- name: GetClusterPins :many
-- Get individual pins within a cluster
SELECT 
  p.id,
  p.content,
  p.created_at,
  pl.longitude,
  pl.latitude,
  COALESCE(u.username, 'Anonymous') as author_name
FROM posts p
JOIN post_locations pl ON p.id = pl.post_id
LEFT JOIN users u ON p.author_user_id = u.id
WHERE 
  LEFT(pl.geohash, $1) = $2  -- precision, cluster_key
  AND p.deleted_at IS NULL
  AND p.created_at > NOW() - INTERVAL '168 hours'
ORDER BY p.created_at DESC
LIMIT 50;

-- name: GetUserMapPreferences :one
SELECT * FROM user_map_preferences WHERE user_id = $1;

-- name: UpdateUserMapPreferences :one
INSERT INTO user_map_preferences (
  user_id, show_own_pins, show_anonymous_pins, 
  cluster_pins, pin_size, cluster_style
) VALUES (
  $1, $2, $3, $4, $5, $6
)
ON CONFLICT (user_id) 
DO UPDATE SET
  show_own_pins = EXCLUDED.show_own_pins,
  show_anonymous_pins = EXCLUDED.show_anonymous_pins,
  cluster_pins = EXCLUDED.cluster_pins,
  pin_size = EXCLUDED.pin_size,
  cluster_style = EXCLUDED.cluster_style,
  updated_at = NOW()
RETURNING *;

-- name: PrecomputeClusters :exec
-- Background job to precompute clusters for popular areas
INSERT INTO map_pin_clusters (
  zoom_level, cluster_geohash, cluster_center,
  pin_count, pin_ids, latest_post_at, expires_at
)
SELECT 
  $1 as zoom_level,
  LEFT(pl.geohash, $2) as cluster_geohash,
  ST_Centroid(ST_Collect(pl.location)) as cluster_center,
  COUNT(*) as pin_count,
  ARRAY_AGG(p.id) as pin_ids,
  MAX(p.created_at) as latest_post_at,
  NOW() + INTERVAL '5 minutes' as expires_at
FROM posts p
JOIN post_locations pl ON p.id = pl.post_id
WHERE 
  pl.geohash LIKE $3 || '%'
  AND p.deleted_at IS NULL
  AND p.created_at > NOW() - INTERVAL '168 hours'
GROUP BY LEFT(pl.geohash, $2)
HAVING COUNT(*) > 1
ON CONFLICT (zoom_level, cluster_geohash)
DO UPDATE SET
  cluster_center = EXCLUDED.cluster_center,
  pin_count = EXCLUDED.pin_count,
  pin_ids = EXCLUDED.pin_ids,
  latest_post_at = EXCLUDED.latest_post_at,
  computed_at = NOW(),
  expires_at = EXCLUDED.expires_at;
```

## Protocol Buffer Definitions

### File: `proto/v1/service/map_pins.proto`

```protobuf
syntax = "proto3";

package service.v1;

import "entities/v1/post.proto";
import "google/protobuf/timestamp.proto";

service MapPinService {
  // Get pins for current map viewport
  rpc GetMapPins(GetMapPinsRequest) returns (GetMapPinsResponse);
  
  // Get details for a specific pin
  rpc GetPinDetails(GetPinDetailsRequest) returns (GetPinDetailsResponse);
  
  // Get pins within a cluster
  rpc GetClusterPins(GetClusterPinsRequest) returns (GetClusterPinsResponse);
  
  // User preferences
  rpc GetMapPreferences(GetMapPreferencesRequest) returns (GetMapPreferencesResponse);
  rpc UpdateMapPreferences(UpdateMapPreferencesRequest) returns (UpdateMapPreferencesResponse);
}

message GetMapPinsRequest {
  // Viewport bounds
  double north_lat = 1;
  double south_lat = 2;
  double east_lng = 3;
  double west_lng = 4;
  
  // Map context
  int32 zoom_level = 5;
  
  // Filters
  int32 max_age_hours = 6; // Default 168 (1 week)
  bool include_own_pins = 7;
  bool include_anonymous_pins = 8;
  
  // Clustering
  bool enable_clustering = 9;
  int32 cluster_threshold = 10; // Min pins to form cluster
}

message GetMapPinsResponse {
  repeated MapPin pins = 1;
  int32 total_in_bounds = 2;
  bool has_clusters = 3;
}

message MapPin {
  oneof pin_type {
    SinglePin single = 1;
    ClusterPin cluster = 2;
  }
  
  // Common fields
  double latitude = 3;
  double longitude = 4;
}

message SinglePin {
  string id = 1;
  string content_preview = 2;
  string author_name = 3;
  bool is_verified_author = 4;
  google.protobuf.Timestamp created_at = 5;
  
  // Quick stats
  int32 reaction_count = 6;
  int32 comment_count = 7;
}

message ClusterPin {
  string cluster_id = 1;
  int32 pin_count = 2;
  repeated string pin_ids = 3; // First few pin IDs
  string predominant_author = 4;
  google.protobuf.Timestamp latest_post_at = 5;
}

message GetPinDetailsRequest {
  string pin_id = 1;
}

message GetPinDetailsResponse {
  entities.v1.Post post = 1;
  float distance_from_center = 2; // Distance from map center
}

message GetClusterPinsRequest {
  string cluster_id = 1;
  int32 limit = 2;
}

message GetClusterPinsResponse {
  repeated entities.v1.Post posts = 1;
  int32 total_in_cluster = 2;
}

message MapPreferences {
  bool show_own_pins = 1;
  bool show_anonymous_pins = 2;
  bool cluster_pins = 3;
  bool auto_open_created_pin = 4;
  bool follow_user_location = 5;
  string pin_size = 6;
  string cluster_style = 7;
}
```

## Frontend Implementation

### File: `features/map-pins/hooks/use-map-pins.ts`

```typescript
import { useQuery } from '@tanstack/react-query';
import { useMapPinServiceClient } from '@/common/services/connectrpc';
import { useMapStore } from '@/features/map/stores/use-map-store';
import { useDebounce } from '@/common/hooks/use-debounce';

export function useMapPins(options?: {
  maxAgeHours?: number;
  enableClustering?: boolean;
}) {
  const client = useMapPinServiceClient();
  const bounds = useMapStore(state => state.bounds);
  const zoom = useMapStore(state => state.zoom);
  
  // Debounce bounds to avoid too many queries while panning
  const debouncedBounds = useDebounce(bounds, 300);
  
  return useQuery({
    queryKey: ['map-pins', debouncedBounds, zoom, options],
    queryFn: async () => {
      if (!debouncedBounds) return { pins: [], hasClusters: false };
      
      const response = await client.getMapPins({
        north_lat: debouncedBounds.getNorth(),
        south_lat: debouncedBounds.getSouth(),
        east_lng: debouncedBounds.getEast(),
        west_lng: debouncedBounds.getWest(),
        zoom_level: zoom,
        max_age_hours: options?.maxAgeHours ?? 168,
        enable_clustering: options?.enableClustering ?? true,
        cluster_threshold: zoom < 13 ? 5 : 10,
      });
      
      return {
        pins: response.pins,
        totalInBounds: response.total_in_bounds,
        hasClusters: response.has_clusters,
      };
    },
    enabled: !!debouncedBounds,
    staleTime: 30 * 1000, // 30 seconds
    refetchOnWindowFocus: false,
  });
}
```

### File: `features/map-pins/components/pin-layer.tsx`

```tsx
import { useEffect, useRef } from 'react';
import maplibregl from 'maplibre-gl';
import { useMapPins } from '../hooks/use-map-pins';
import { PinPopup } from './pin-popup';
import { createPinMarkers } from '../lib/pin-markers';

interface PinLayerProps {
  map: maplibregl.Map;
}

export function PinLayer({ map }: PinLayerProps) {
  const { data: pinsData } = useMapPins();
  const markersRef = useRef<maplibregl.Marker[]>([]);
  const popupRef = useRef<maplibregl.Popup | null>(null);
  
  useEffect(() => {
    if (!map || !pinsData) return;
    
    // Clear existing markers
    markersRef.current.forEach(marker => marker.remove());
    markersRef.current = [];
    
    // Add source if not exists
    if (!map.getSource('pins')) {
      map.addSource('pins', {
        type: 'geojson',
        data: {
          type: 'FeatureCollection',
          features: [],
        },
        cluster: true,
        clusterMaxZoom: 14,
        clusterRadius: 50,
      });
      
      // Cluster layer
      map.addLayer({
        id: 'clusters',
        type: 'circle',
        source: 'pins',
        filter: ['has', 'point_count'],
        paint: {
          'circle-color': [
            'step',
            ['get', 'point_count'],
            '#51bbd6', 10,
            '#f1f075', 30,
            '#f28cb1'
          ],
          'circle-radius': [
            'step',
            ['get', 'point_count'],
            20, 10,
            30, 30,
            40
          ],
        },
      });
      
      // Cluster count layer
      map.addLayer({
        id: 'cluster-count',
        type: 'symbol',
        source: 'pins',
        filter: ['has', 'point_count'],
        layout: {
          'text-field': '{point_count_abbreviated}',
          'text-font': ['DIN Offc Pro Medium', 'Arial Unicode MS Bold'],
          'text-size': 12,
        },
        paint: {
          'text-color': '#ffffff',
        },
      });
      
      // Individual pins layer
      map.addLayer({
        id: 'unclustered-point',
        type: 'circle',
        source: 'pins',
        filter: ['!', ['has', 'point_count']],
        paint: {
          'circle-color': '#3b82f6',
          'circle-radius': 8,
          'circle-stroke-width': 2,
          'circle-stroke-color': '#fff',
        },
      });
    }
    
    // Convert pins to GeoJSON
    const features = pinsData.pins.map(pin => {
      if (pin.single) {
        return {
          type: 'Feature' as const,
          geometry: {
            type: 'Point' as const,
            coordinates: [pin.longitude, pin.latitude],
          },
          properties: {
            id: pin.single.id,
            content: pin.single.content_preview,
            author: pin.single.author_name,
            isVerified: pin.single.is_verified_author,
            reactions: pin.single.reaction_count,
            comments: pin.single.comment_count,
          },
        };
      } else {
        // Cluster pins are handled by MapLibre clustering
        return null;
      }
    }).filter(Boolean);
    
    // Update source data
    const source = map.getSource('pins') as maplibregl.GeoJSONSource;
    source.setData({
      type: 'FeatureCollection',
      features,
    });
    
    // Handle click on clusters
    map.on('click', 'clusters', (e) => {
      const features = map.queryRenderedFeatures(e.point, {
        layers: ['clusters'],
      });
      
      const clusterId = features[0].properties.cluster_id;
      const source = map.getSource('pins') as maplibregl.GeoJSONSource;
      
      source.getClusterExpansionZoom(clusterId, (err, zoom) => {
        if (err) return;
        
        map.easeTo({
          center: features[0].geometry.coordinates as [number, number],
          zoom: zoom,
        });
      });
    });
    
    // Handle click on individual pins
    map.on('click', 'unclustered-point', (e) => {
      const features = map.queryRenderedFeatures(e.point, {
        layers: ['unclustered-point'],
      });
      
      if (features.length === 0) return;
      
      const coordinates = features[0].geometry.coordinates.slice();
      const properties = features[0].properties;
      
      // Create and show popup
      if (popupRef.current) {
        popupRef.current.remove();
      }
      
      popupRef.current = new maplibregl.Popup({
        closeButton: true,
        closeOnClick: false,
        maxWidth: '300px',
      })
        .setLngLat(coordinates as [number, number])
        .setHTML(`<div id="pin-popup-${properties.id}"></div>`)
        .addTo(map);
      
      // Render React component into popup
      const container = document.getElementById(`pin-popup-${properties.id}`);
      if (container) {
        ReactDOM.render(
          <PinPopup pinId={properties.id} />,
          container
        );
      }
    });
    
    // Change cursor on hover
    map.on('mouseenter', 'clusters', () => {
      map.getCanvas().style.cursor = 'pointer';
    });
    
    map.on('mouseleave', 'clusters', () => {
      map.getCanvas().style.cursor = '';
    });
    
    map.on('mouseenter', 'unclustered-point', () => {
      map.getCanvas().style.cursor = 'pointer';
    });
    
    map.on('mouseleave', 'unclustered-point', () => {
      map.getCanvas().style.cursor = '';
    });
    
    return () => {
      // Cleanup
      markersRef.current.forEach(marker => marker.remove());
      if (popupRef.current) {
        popupRef.current.remove();
      }
    };
  }, [map, pinsData]);
  
  return null;
}
```

### File: `features/map-pins/components/pin-popup.tsx`

```tsx
import { useQuery } from '@tanstack/react-query';
import { useMapPinServiceClient } from '@/common/services/connectrpc';
import { PostCard } from '@/features/posts/components/post-card';
import { formatTimeAgo } from '@/common/lib/time';

interface PinPopupProps {
  pinId: string;
}

export function PinPopup({ pinId }: PinPopupProps) {
  const client = useMapPinServiceClient();
  
  const { data: post, isLoading } = useQuery({
    queryKey: ['pin-details', pinId],
    queryFn: async () => {
      const response = await client.getPinDetails({
        pin_id: pinId,
      });
      return response.post;
    },
  });
  
  if (isLoading) {
    return (
      <div className="p-3">
        <div className="animate-pulse space-y-2">
          <div className="h-4 bg-gray-200 rounded w-3/4"></div>
          <div className="h-3 bg-gray-200 rounded w-1/2"></div>
        </div>
      </div>
    );
  }
  
  if (!post) return null;
  
  return (
    <div className="p-3 space-y-2 max-w-xs">
      <div className="flex items-center gap-2">
        <div className="w-8 h-8 rounded-full bg-gray-200" />
        <div>
          <p className="font-medium text-sm">
            {post.author_username}
            {post.is_verified_author && (
              <span className="ml-1 text-blue-500">✓</span>
            )}
          </p>
          <p className="text-xs text-gray-500">
            {formatTimeAgo(post.created_at)}
          </p>
        </div>
      </div>
      
      <p className="text-sm text-gray-900 line-clamp-3">
        {post.content}
      </p>
      
      {post.location?.place_name && (
        <p className="text-xs text-gray-500 flex items-center gap-1">
          <MapPin size={12} />
          {post.location.place_name}
        </p>
      )}
      
      <div className="flex items-center gap-3 text-xs text-gray-500">
        <span>{post.reaction_count} reactions</span>
        <span>{post.comment_count} comments</span>
      </div>
      
      <button
        onClick={() => {
          // Open full post view
          window.location.href = `/post/${post.id}`;
        }}
        className="text-xs text-blue-500 hover:underline"
      >
        View full post →
      </button>
    </div>
  );
}
```

### File: `features/map-pins/components/cluster-popup.tsx`

```tsx
import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useMapPinServiceClient } from '@/common/services/connectrpc';

interface ClusterPopupProps {
  clusterId: string;
  pinCount: number;
}

export function ClusterPopup({ clusterId, pinCount }: ClusterPopupProps) {
  const client = useMapPinServiceClient();
  const [isExpanded, setIsExpanded] = useState(false);
  
  const { data: posts, isLoading } = useQuery({
    queryKey: ['cluster-pins', clusterId],
    queryFn: async () => {
      const response = await client.getClusterPins({
        cluster_id: clusterId,
        limit: 5,
      });
      return response.posts;
    },
    enabled: isExpanded,
  });
  
  return (
    <div className="p-3 max-w-xs">
      <div className="flex items-center justify-between mb-2">
        <span className="font-medium text-sm">
          {pinCount} posts in this area
        </span>
        <button
          onClick={() => setIsExpanded(!isExpanded)}
          className="text-xs text-blue-500 hover:underline"
        >
          {isExpanded ? 'Hide' : 'Show'}
        </button>
      </div>
      
      {isExpanded && (
        <div className="space-y-2 max-h-60 overflow-y-auto">
          {isLoading ? (
            <div className="text-xs text-gray-500">Loading...</div>
          ) : (
            posts?.map(post => (
              <div key={post.id} className="border-t pt-2">
                <p className="text-xs font-medium">{post.author_username}</p>
                <p className="text-xs text-gray-600 line-clamp-2">
                  {post.content}
                </p>
              </div>
            ))
          )}
        </div>
      )}
      
      <button
        onClick={() => {
          // Zoom in to expand cluster
          // Implementation depends on map instance access
        }}
        className="mt-2 text-xs text-blue-500 hover:underline"
      >
        Zoom in to see individual posts
      </button>
    </div>
  );
}
```

### File: `features/map-pins/lib/pin-markers.ts`

```typescript
import maplibregl from 'maplibre-gl';

export function createPinMarker(
  pin: any,
  options: {
    size?: 'small' | 'medium' | 'large';
    color?: string;
  } = {}
): maplibregl.Marker {
  const el = document.createElement('div');
  el.className = 'pin-marker';
  
  // Size classes
  const sizeClasses = {
    small: 'w-6 h-6',
    medium: 'w-8 h-8',
    large: 'w-10 h-10',
  };
  
  el.classList.add(sizeClasses[options.size || 'medium']);
  
  // Style
  el.style.backgroundColor = options.color || '#3b82f6';
  el.style.borderRadius = '50%';
  el.style.border = '2px solid white';
  el.style.cursor = 'pointer';
  el.style.boxShadow = '0 2px 4px rgba(0,0,0,0.2)';
  
  return new maplibregl.Marker(el)
    .setLngLat([pin.longitude, pin.latitude]);
}

export function createClusterMarker(
  cluster: any,
  options: {
    style?: 'circles' | 'hexagons' | 'squares';
  } = {}
): maplibregl.Marker {
  const el = document.createElement('div');
  el.className = 'cluster-marker';
  
  // Size based on count
  const size = Math.min(60, 20 + Math.sqrt(cluster.pin_count) * 5);
  el.style.width = `${size}px`;
  el.style.height = `${size}px`;
  
  // Style based on preference
  if (options.style === 'hexagons') {
    el.style.clipPath = 'polygon(50% 0%, 100% 25%, 100% 75%, 50% 100%, 0% 75%, 0% 25%)';
  } else if (options.style === 'squares') {
    el.style.borderRadius = '4px';
  } else {
    el.style.borderRadius = '50%';
  }
  
  // Gradient based on density
  const intensity = Math.min(1, cluster.pin_count / 50);
  el.style.background = `radial-gradient(circle, 
    rgba(59, 130, 246, ${0.8 + intensity * 0.2}), 
    rgba(59, 130, 246, ${0.4 + intensity * 0.3})
  )`;
  
  // Count label
  el.innerHTML = `
    <span style="
      position: absolute;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
      color: white;
      font-weight: bold;
      font-size: ${Math.min(16, 10 + size / 10)}px;
    ">
      ${cluster.pin_count}
    </span>
  `;
  
  return new maplibregl.Marker(el)
    .setLngLat([cluster.latitude, cluster.longitude]);
}
```

## Key Implementation Notes

### Clustering Strategy
- Use database-level clustering for performance
- Adjust cluster precision based on zoom level
- Precompute clusters for hot areas
- Client-side clustering as fallback

### Performance Optimizations
- Debounce map pan/zoom events
- Cache pin data for 30 seconds
- Lazy load popup details
- Use WebGL for large pin counts

### Visual Design
- Different pin colors for verified/anonymous
- Animate pin additions/removals
- Hover effects for better UX
- Custom cluster styles per user preference

### Interaction Patterns
- Click pin to show popup
- Click cluster to zoom in
- Long press for quick actions
- Swipe between posts in cluster

### Mobile Optimizations
- Larger touch targets on mobile
- Bottom sheet for popups
- Gesture handling for map interactions
- Reduced pin density on small screens

### Accessibility
- Keyboard navigation for pins
- Screen reader descriptions
- High contrast mode support
- Focus management in popups

### Future Enhancements
- Heat map visualization
- Time-lapse animation
- 3D pin extrusion
- AR mode for mobile
- Pin filtering by category
