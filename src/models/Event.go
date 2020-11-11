package models

import (
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/jinzhu/gorm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbEvent "github.com/RemyRanger/taktyl_core_grpc/src/proto/event"
)

// Event : event entity
type Event struct {
	ID        uint64    `gorm:"primary_key;auto_increment" json:"id"`
	Title     string    `gorm:"size:255;not null;unique" json:"title"`
	Content   string    `gorm:"size:255;not null;" json:"content"`
	Author    User      `json:"author"`
	AuthorID  uint32    `gorm:"not null" json:"author_id"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// Prepare : prepare Event entity
func (p *Event) Prepare(Title, Content string, AuthorID int32) {
	p.ID = 0
	p.Title = html.EscapeString(strings.TrimSpace(Title))
	p.Content = html.EscapeString(strings.TrimSpace(Content))
	p.Author = User{}
	p.AuthorID = uint32(AuthorID)
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
}

// Validate : validate field for one event
func (p *Event) Validate() error {

	if p.Title == "" {
		return errors.New("Required Title")
	}
	if p.Content == "" {
		return errors.New("Required Content")
	}
	if p.AuthorID < 1 {
		return errors.New("Required Author")
	}
	return nil
}

// SaveEvent : save one event
func (p *Event) SaveEvent(db *gorm.DB) (*Event, error) {
	var err error
	err = db.Debug().Model(&Event{}).Create(&p).Error
	if err != nil {
		return &Event{}, err
	}
	if p.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", p.AuthorID).Take(&p.Author).Error
		if err != nil {
			return &Event{}, err
		}
	}
	return p, nil
}

// FindAllEvents : retrieve all events in database
func (u *User) FindAllEvents(db *gorm.DB, srv pbEvent.EventService_ListEventsServer) error {

	var err error
	rows, err := db.Model(&Event{}).Rows()
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var event Event
		// ScanRows is a method of `gorm.DB`, it can be used to scan a row into a struct
		err := db.ScanRows(rows, &event)
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

// FindEventByID : find one event by id
func (p *Event) FindEventByID(db *gorm.DB, eid uint64) (*Event, error) {
	var err error
	err = db.Debug().Model(&Event{}).Where("id = ?", eid).Take(&p).Error
	if err != nil {
		return &Event{}, err
	}
	if p.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", p.AuthorID).Take(&p.Author).Error
		if err != nil {
			return &Event{}, err
		}
	}
	return p, nil
}

// UpdateAEvent : update one event
func (p *Event) UpdateAEvent(db *gorm.DB, eid uint64) (*Event, error) {

	var err error

	err = db.Debug().Model(&Event{}).Where("id = ?", eid).Updates(Event{Title: p.Title, Content: p.Content, UpdatedAt: time.Now()}).Error
	if err != nil {
		return &Event{}, err
	}
	if p.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", p.AuthorID).Take(&p.Author).Error
		if err != nil {
			return &Event{}, err
		}
	}
	return p, nil
}

// DeleteAEvent : delete one event
func (p *Event) DeleteAEvent(db *gorm.DB, eid uint64, uid uint32) (int64, error) {

	db = db.Debug().Model(&Event{}).Where("id = ? and author_id = ?", eid, uid).Take(&Event{}).Delete(&Event{})

	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Event not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
