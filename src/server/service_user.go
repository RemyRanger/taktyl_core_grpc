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

// AddUser adds a user to the in-memory store.
func (b *Backend) AddUser(ctx context.Context, req *pbUser.AddUserRequest) (*pbUser.User, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	user := models.User{}

	user.Prepare(req.Nickname, req.Email, req.Password)
	err := user.Validate("")
	if err != nil {
		return &pbUser.User{}, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Error while creating user in database: %v", err),
		)
	}
	userCreated, err := user.SaveUser(b.DB)

	// Convert timestamp
	createdAtTimesStamp, err := ptypes.TimestampProto(userCreated.CreatedAt)
	updatedAtTimesStamp, err := ptypes.TimestampProto(userCreated.UpdatedAt)

	return &pbUser.User{
		ID:        int32(userCreated.ID),
		Nickname:  userCreated.Nickname,
		Email:     userCreated.Email,
		CreatedAt: createdAtTimesStamp,
		UpdatedAt: updatedAtTimesStamp,
	}, nil
}

// ListUsers lists all users in the store.
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
