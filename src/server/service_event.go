package server

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/RemyRanger/taktyl_core_grpc/src/models"
	pbEvent "github.com/RemyRanger/taktyl_core_grpc/src/proto/event"
)

// GetEvent : get one event
func (b *Backend) GetEvent(ctx context.Context, req *pbEvent.GetEventRequest) (*pbEvent.EventDTO, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var err error
	eventResult := models.Event{}
	err = b.DB.Debug().Model(&models.Event{}).Where("id = ?", req.EventId).Take(&eventResult).Error
	if err != nil {
		return &pbEvent.EventDTO{}, err
	}

	// Convert timestamp
	createdAtTimesStamp, err := ptypes.TimestampProto(eventResult.CreatedAt)
	updatedAtTimesStamp, err := ptypes.TimestampProto(eventResult.UpdatedAt)

	return &pbEvent.EventDTO{
		ID:        int64(eventResult.ID),
		Title:     eventResult.Title,
		Content:   eventResult.Content,
		AuthorID:  int32(eventResult.AuthorID),
		CreatedAt: createdAtTimesStamp,
		UpdatedAt: updatedAtTimesStamp,
	}, nil
}

// UpdateEvent adds a event to the database
func (b *Backend) UpdateEvent(ctx context.Context, req *pbEvent.UpdateEventRequest) (*pbEvent.EventDTO, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	event := models.Event{}

	event.Prepare(req.Title, req.Content, req.AuthorID)
	err := event.Validate()
	if err != nil {
		return &pbEvent.EventDTO{}, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Error while updating user in database: %v", err),
		)
	}
	eventUpdated, err := event.UpdateAEvent(b.DB, uint64(req.ID))
	if err != nil {
		return &pbEvent.EventDTO{}, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Error while updating user in database: %v", err),
		)
	}

	// Convert timestamp
	createdAtTimesStamp, err := ptypes.TimestampProto(eventUpdated.CreatedAt)
	updatedAtTimesStamp, err := ptypes.TimestampProto(eventUpdated.UpdatedAt)

	return &pbEvent.EventDTO{
		ID:        int64(eventUpdated.ID),
		Title:     eventUpdated.Title,
		Content:   eventUpdated.Content,
		AuthorID:  int32(eventUpdated.AuthorID),
		CreatedAt: createdAtTimesStamp,
		UpdatedAt: updatedAtTimesStamp,
	}, nil
}

// AddEvent : save one event
func (b *Backend) AddEvent(ctx context.Context, req *pbEvent.AddEventRequest) (*pbEvent.EventDTO, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	var err error
	eventCreated := models.Event{}
	eventCreated.Prepare(req.Title, req.Content, req.AuthorID)
	err = b.DB.Debug().Model(&models.Event{}).Create(&eventCreated).Error
	if err != nil {
		return &pbEvent.EventDTO{}, err
	}
	if eventCreated.ID != 0 {
		err = b.DB.Debug().Model(&models.User{}).Where("id = ?", req.AuthorID).Take(&eventCreated.Author).Error
		if err != nil {
			return &pbEvent.EventDTO{}, err
		}
	}

	// Convert timestamp
	createdAtTimesStamp, err := ptypes.TimestampProto(eventCreated.CreatedAt)
	updatedAtTimesStamp, err := ptypes.TimestampProto(eventCreated.UpdatedAt)

	return &pbEvent.EventDTO{
		ID:        int64(eventCreated.ID),
		Title:     eventCreated.Title,
		Content:   eventCreated.Content,
		AuthorID:  int32(eventCreated.AuthorID),
		CreatedAt: createdAtTimesStamp,
		UpdatedAt: updatedAtTimesStamp,
	}, nil
}

// ListEvents lists all events in the store.
func (b *Backend) ListEvents(_ *pbEvent.ListEventsRequest, srv pbEvent.EventService_ListEventsServer) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var err error
	rows, err := b.DB.Model(&models.Event{}).Rows()
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var event models.Event
		// ScanRows is a method of `gorm.DB`, it can be used to scan a row into a struct
		err := b.DB.ScanRows(rows, &event)
		if err != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Error while reading database iteration: %v", err),
			)
		}

		// Convert timestamp
		createdAtTimesStamp, err := ptypes.TimestampProto(event.CreatedAt)
		updatedAtTimesStamp, err := ptypes.TimestampProto(event.UpdatedAt)

		// do something
		srv.Send(&pbEvent.EventDTO{
			ID:        int64(event.ID),
			Title:     event.Title,
			Content:   event.Content,
			AuthorID:  int32(event.AuthorID),
			CreatedAt: createdAtTimesStamp,
			UpdatedAt: updatedAtTimesStamp,
		})
	}
	return nil
}

// DeleteEvent delete one event in the database.
func (b *Backend) DeleteEvent(ctx context.Context, req *pbEvent.DeleteEventRequest) (*pbEvent.DeleteEventRequest, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	event := models.Event{}

	rowAffected, err := event.DeleteAEvent(b.DB, uint64(req.EventId), uint32(req.AuthorId))
	if err != nil {
		return &pbEvent.DeleteEventRequest{}, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Error unable to delete user: %v", err),
		)
	}

	return &pbEvent.DeleteEventRequest{
		EventId:  rowAffected,
		AuthorId: req.AuthorId,
	}, nil
}
