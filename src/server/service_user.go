package server

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/RemyRanger/taktyl_core_grpc/src/models"
	pbUser "github.com/RemyRanger/taktyl_core_grpc/src/proto/user"
	"github.com/golang/protobuf/ptypes"
)

// AddUser adds a user to the database
func (b *Backend) AddUser(ctx context.Context, req *pbUser.AddUserRequest) (*pbUser.UserDTO, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	user := models.User{}

	user.Prepare(req.Nickname, req.Email, req.Password)
	err := user.Validate("")
	if err != nil {
		return &pbUser.UserDTO{}, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Error while creating user in database: %v", err),
		)
	}
	userCreated, err := user.SaveUser(b.DB)
	if err != nil {
		return &pbUser.UserDTO{}, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Error while creating user in database: %v", err),
		)
	}

	// Convert timestamp
	createdAtTimesStamp, err := ptypes.TimestampProto(userCreated.CreatedAt)
	updatedAtTimesStamp, err := ptypes.TimestampProto(userCreated.UpdatedAt)

	return &pbUser.UserDTO{
		ID:        int32(userCreated.ID),
		Nickname:  userCreated.Nickname,
		Email:     userCreated.Email,
		CreatedAt: createdAtTimesStamp,
		UpdatedAt: updatedAtTimesStamp,
	}, nil
}

// UpdateUser adds a user to the database
func (b *Backend) UpdateUser(ctx context.Context, req *pbUser.UpdateUserRequest) (*pbUser.UserDTO, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	user := models.User{}

	user.Prepare(req.Nickname, req.Email, req.Password)
	err := user.Validate("update")
	if err != nil {
		return &pbUser.UserDTO{}, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Error while updating user in database: %v", err),
		)
	}
	userUpdated, err := user.UpdateAUser(b.DB, uint32(req.ID))
	if err != nil {
		return &pbUser.UserDTO{}, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Error while updating user in database: %v", err),
		)
	}

	// Convert timestamp
	createdAtTimesStamp, err := ptypes.TimestampProto(userUpdated.CreatedAt)
	updatedAtTimesStamp, err := ptypes.TimestampProto(userUpdated.UpdatedAt)

	return &pbUser.UserDTO{
		ID:        int32(userUpdated.ID),
		Nickname:  userUpdated.Nickname,
		Email:     userUpdated.Email,
		CreatedAt: createdAtTimesStamp,
		UpdatedAt: updatedAtTimesStamp,
	}, nil
}

// ListUsers lists all users in the database.
func (b *Backend) ListUsers(_ *pbUser.ListUsersRequest, srv pbUser.UserService_ListUsersServer) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	user := models.User{}

	err := user.FindAllUsers(b.DB, srv)
	if err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Error while streaming users from database: %v", err),
		)
	}
	return nil
}

// GetUser get one user in the database.
func (b *Backend) GetUser(ctx context.Context, req *pbUser.GetUserRequest) (*pbUser.UserDTO, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	user := models.User{}

	userResult, err := user.FindUserByID(b.DB, uint32(req.UserId))
	if err != nil {
		return &pbUser.UserDTO{}, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Error unable to get user: %v", err),
		)
	}

	// Convert timestamp
	createdAtTimesStamp, err := ptypes.TimestampProto(userResult.CreatedAt)
	updatedAtTimesStamp, err := ptypes.TimestampProto(userResult.UpdatedAt)

	return &pbUser.UserDTO{
		ID:        int32(userResult.ID),
		Nickname:  userResult.Nickname,
		Email:     userResult.Email,
		CreatedAt: createdAtTimesStamp,
		UpdatedAt: updatedAtTimesStamp,
	}, nil
}

// DeleteUser delete one user in the database.
func (b *Backend) DeleteUser(ctx context.Context, req *pbUser.DeleteUserRequest) (*pbUser.DeleteUserRequest, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	user := models.User{}

	rowAffected, err := user.DeleteAUser(b.DB, uint32(req.UserId))
	if err != nil {
		return &pbUser.DeleteUserRequest{}, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Error unable to delete user: %v", err),
		)
	}

	return &pbUser.DeleteUserRequest{
		UserId: int32(rowAffected),
	}, nil
}
