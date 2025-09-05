package protoconv

import (
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

	protoUser := &entitiesv1.User{
		Id:        user.ID.String(),
		Email:     &user.Email, // Email is string, convert to pointer
		CreatedAt: user.CreatedAt.Time.Unix(),
		UpdatedAt: user.UpdatedAt.Time.Unix(),
	}

	// Set display name as username (since we don't have username field)
	if user.DisplayName != nil {
		protoUser.Username = *user.DisplayName
	} else {
		protoUser.Username = "Anonymous"
	}

	return protoUser
}

// ProtoToCreateUserParams converts protobuf User to repository CreateUserParams
func ProtoToCreateUserParams(user *entitiesv1.User) (*repository.CreateUserParams, error) {
	// Generate ID if not provided
	idStr := user.Id
	if idStr == "" {
		idStr = ksuid.New().String()
	}

	// Parse UUID
	var userID pgtype.UUID
	err := userID.Scan(idStr)
	if err != nil {
		return nil, err
	}

	// Set timestamps
	now := time.Now()
	if user.CreatedAt > 0 {
		now = time.Unix(user.CreatedAt, 0)
	}

	params := &repository.CreateUserParams{
		ID: userID,
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
		params.Email = *user.Email // Email is string in DB
	}

	// Set display name from username
	if user.Username != "" && user.Username != "Anonymous" {
		params.DisplayName = &user.Username
	}

	return params, nil
}

// ProtoToUpdateUserParams converts protobuf User to repository UpdateUserParams
func ProtoToUpdateUserParams(user *entitiesv1.User) (*repository.UpdateUserParams, error) {
	// Parse UUID
	var userID pgtype.UUID
	err := userID.Scan(user.Id)
	if err != nil {
		return nil, err
	}

	params := &repository.UpdateUserParams{
		ID: userID,
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
	}

	// Set email if provided
	if user.Email != nil && *user.Email != "" {
		params.Email = *user.Email // Email is string in DB
	}

	// Set display name from username
	if user.Username != "" && user.Username != "Anonymous" {
		params.DisplayName = &user.Username
	}

	return params, nil
}

// CreateAnonymousUser creates proto User for anonymous session
func CreateAnonymousUser(fingerprint string) *entitiesv1.User {
	id := ksuid.New().String()
	now := time.Now().Unix()

	return &entitiesv1.User{
		Id:        id,
		Username:  "Anonymous", 
		Email:     nil, // No email for anonymous
		CreatedAt: now,
		UpdatedAt: now,
	}
}