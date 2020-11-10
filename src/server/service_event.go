package server

import (
	"context"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/RemyRanger/taktyl_core_grpc/src/models"
	pbEvent "github.com/RemyRanger/taktyl_core_grpc/src/proto/event"
)

// Event : event entity model database
type Event struct {
	ID        uint64      `gorm:"primary_key;auto_increment" json:"id"`
	Title     string      `gorm:"size:255;not null;unique" json:"title"`
	Content   string      `gorm:"size:255;not null;" json:"content"`
	Author    models.User `json:"author"`
	AuthorID  uint32      `gorm:"not null" json:"author_id"`
	CreatedAt time.Time   `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time   `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// Prepare : prepare Event entity
func (p *Event) Prepare(Title, Content string, AuthorID int32) {
	p.ID = 0
	p.Title = html.EscapeString(strings.TrimSpace(Title))
	p.Content = html.EscapeString(strings.TrimSpace(Content))
	p.Author = models.User{}
	p.AuthorID = uint32(AuthorID)
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
}

// GetEvent : get one event
func (b *Backend) GetEvent(ctx context.Context, req *pbEvent.GetEventRequest) (*pbEvent.Event, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var err error
	eventResult := Event{}
	err = b.DB.Debug().Model(&Event{}).Where("id = ?", req.UserId).Take(&eventResult).Error
	if err != nil {
		return &pbEvent.Event{}, err
	}

	// Convert timestamp
	createdAtTimesStamp, err := ptypes.TimestampProto(eventResult.CreatedAt)
	updatedAtTimesStamp, err := ptypes.TimestampProto(eventResult.UpdatedAt)

	return &pbEvent.Event{
		ID:        int64(eventResult.ID),
		Title:     eventResult.Title,
		Content:   eventResult.Content,
		AuthorID:  int32(eventResult.AuthorID),
		CreatedAt: createdAtTimesStamp,
		UpdatedAt: updatedAtTimesStamp,
	}, nil
}

// AddEvent : save one event
func (b *Backend) AddEvent(ctx context.Context, req *pbEvent.AddEventRequest) (*pbEvent.Event, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	var err error
	eventCreated := Event{}
	eventCreated.Prepare(req.Title, req.Content, req.AuthorID)
	err = b.DB.Debug().Model(&Event{}).Create(&eventCreated).Error
	if err != nil {
		return &pbEvent.Event{}, err
	}
	if eventCreated.ID != 0 {
		err = b.DB.Debug().Model(&models.User{}).Where("id = ?", req.AuthorID).Take(&eventCreated.Author).Error
		if err != nil {
			return &pbEvent.Event{}, err
		}
	}

	// Convert timestamp
	createdAtTimesStamp, err := ptypes.TimestampProto(eventCreated.CreatedAt)
	updatedAtTimesStamp, err := ptypes.TimestampProto(eventCreated.UpdatedAt)

	return &pbEvent.Event{
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
	rows, err := b.DB.Model(&Event{}).Rows()
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var event Event
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
		srv.Send(&pbEvent.Event{
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
