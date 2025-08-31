# Map Service

**Owner:** feature/map-pins  
**Priority:** Phase 2 - Can be developed in parallel

## Dependencies
- Posts service (heavy dependency)
- Post locations data
- Feed service (shares location logic)

## Database Tables
- `map_pin_clusters` - Clustering cache
- `user_map_preferences` - User preferences

## Proto Files
- `proto/v1/service/map_pins.proto`

## Key Responsibilities
- Pin rendering on map
- Smart clustering at different zoom levels
- Popup details
- User preferences
- Performance optimization

## Merge Notes
Can be merged in any order after posts-core is merged.
Coordinates with feed service for location queries.