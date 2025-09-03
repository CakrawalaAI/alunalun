package protoconv

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	entitiesv1 "github.com/radjathaher/alunalun/api/internal/protocgen/v1/entities"
	"github.com/radjathaher/alunalun/api/internal/repository"
	"github.com/segmentio/ksuid"
)

// UserToProto converts repository.User to protobuf User entity
func UserToProto(user *repository.User) *entitiesv1.User {
	if user == nil {
		return nil
	}

	// Parse metadata JSONB into map
	metadata := make(map[string]string)
	if len(user.Metadata) > 0 {
		var jsonMap map[string]interface{}
		if err := json.Unmarshal(user.Metadata, &jsonMap); err == nil {
			// Convert to string map for proto
			for k, v := range jsonMap {
				if str, ok := v.(string); ok {
					metadata[k] = str
				} else if v != nil {
					// Convert non-string values to JSON strings
					if bytes, err := json.Marshal(v); err == nil {
						metadata[k] = string(bytes)
					}
				}
			}
		}
	}

	protoUser := &entitiesv1.User{
		Id:        user.ID,
		Username:  user.Username,
		Metadata:  metadata,
		CreatedAt: user.CreatedAt.Time.Unix(),
		UpdatedAt: user.UpdatedAt.Time.Unix(),
	}

	// Set email if not null
	if user.Email != nil {
		protoUser.Email = user.Email
	}

	return protoUser
}

// ProtoToCreateUserParams converts protobuf User to repository CreateUserParams
func ProtoToCreateUserParams(user *entitiesv1.User) (*repository.CreateUserParams, error) {
	// Generate ID if not provided
	id := user.Id
	if id == "" {
		id = "user-" + ksuid.New().String()
	}

	// Convert metadata map to JSON
	metadata, err := MetadataToJSON(user.Metadata)
	if err != nil {
		return nil, err
	}

	// Set timestamps
	now := time.Now()
	if user.CreatedAt > 0 {
		now = time.Unix(user.CreatedAt, 0)
	}

	params := &repository.CreateUserParams{
		ID:       id,
		Username: user.Username,
		Metadata: metadata,
		CreatedAt: pgtype.Timestamptz{
			Time:  now,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  now,
			Valid: true,
		},
	}

	// Set email if provided
	if user.Email != nil && *user.Email != "" {
		params.Email = user.Email
	}

	return params, nil
}

// ProtoToUpdateUserParams converts protobuf User to repository UpdateUserParams
func ProtoToUpdateUserParams(user *entitiesv1.User) (*repository.UpdateUserParams, error) {
	metadata, err := MetadataToJSON(user.Metadata)
	if err != nil {
		return nil, err
	}

	params := &repository.UpdateUserParams{
		ID:       user.Id,
		Username: user.Username,
		Metadata: metadata,
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
	}

	// Set email if provided
	if user.Email != nil && *user.Email != "" {
		params.Email = user.Email
	}

	return params, nil
}

// MetadataToJSON converts a string map to JSON bytes for JSONB storage
func MetadataToJSON(metadata map[string]string) ([]byte, error) {
	if len(metadata) == 0 {
		return []byte("{}"), nil
	}

	// Convert string values that are JSON back to proper types
	jsonMap := make(map[string]interface{})
	for k, v := range metadata {
		// Try to parse as JSON first (for nested objects/arrays)
		var parsed interface{}
		if err := json.Unmarshal([]byte(v), &parsed); err == nil {
			jsonMap[k] = parsed
		} else {
			// Keep as string if not valid JSON
			jsonMap[k] = v
		}
	}

	return json.Marshal(jsonMap)
}

// JSONToMetadata converts JSONB bytes to a string map for proto
func JSONToMetadata(data []byte) (map[string]string, error) {
	if len(data) == 0 {
		return make(map[string]string), nil
	}

	var jsonMap map[string]interface{}
	if err := json.Unmarshal(data, &jsonMap); err != nil {
		return nil, err
	}

	metadata := make(map[string]string)
	for k, v := range jsonMap {
		if str, ok := v.(string); ok {
			metadata[k] = str
		} else if v != nil {
			// Convert non-string values to JSON strings
			if bytes, err := json.Marshal(v); err == nil {
				metadata[k] = string(bytes)
			}
		}
	}

	return metadata, nil
}

// CreateAnonymousUser creates proto User for anonymous session
func CreateAnonymousUser(fingerprint string) *entitiesv1.User {
	id := "anon-" + ksuid.New().String()
	now := time.Now().Unix()

	return &entitiesv1.User{
		Id:       id,
		Username: id, // Username same as ID for anonymous
		Email:    nil, // No email for anonymous
		Metadata: map[string]string{
			"type":        "anonymous",
			"fingerprint": fingerprint,
			"created_via": "fallback",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}