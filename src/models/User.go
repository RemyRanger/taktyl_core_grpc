package models

import (
	"errors"
	"fmt"
	"html"
	"log"
	"strings"
	"time"

	"github.com/badoux/checkmail"
	"github.com/golang/protobuf/ptypes"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbUser "github.com/RemyRanger/taktyl_core_grpc/src/proto/user"
)

// User : user entity model database
type User struct {
	ID        uint32    `gorm:"primary_key;auto_increment" json:"id"`
	Nickname  string    `gorm:"size:255;not null;unique" json:"nickname"`
	Email     string    `gorm:"size:100;not null;unique" json:"email"`
	Password  string    `gorm:"size:100;not null;" json:"password"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// Hash : password hashing
func Hash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

// VerifyPassword : check password validity
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// BeforeSave : hash password before saving
func (u *User) BeforeSave() error {
	hashedPassword, err := Hash(u.Password)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// Prepare : prepare user entity
func (u *User) Prepare(Nickname, Email, Password string) {
	u.ID = 0
	u.Nickname = html.EscapeString(strings.TrimSpace(Nickname))
	u.Email = html.EscapeString(strings.TrimSpace(Email))
	u.Password = Password
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
}

// Validate : validate user fields
func (u *User) Validate(action string) error {
	switch strings.ToLower(action) {
	case "update":
		if u.Nickname == "" {
			return errors.New("Required Nickname")
		}
		if u.Password == "" {
			return errors.New("Required Password")
		}
		if u.Email == "" {
			return errors.New("Required Email")
		}
		if err := checkmail.ValidateFormat(u.Email); err != nil {
			return errors.New("Invalid Email")
		}

		return nil
	case "login":
		if u.Password == "" {
			return errors.New("Required Password")
		}
		if u.Email == "" {
			return errors.New("Required Email")
		}
		if err := checkmail.ValidateFormat(u.Email); err != nil {
			return errors.New("Invalid Email")
		}
		return nil

	default:
		if u.Nickname == "" {
			return errors.New("Required Nickname")
		}
		if u.Password == "" {
			return errors.New("Required Password")
		}
		if u.Email == "" {
			return errors.New("Required Email")
		}
		if err := checkmail.ValidateFormat(u.Email); err != nil {
			return errors.New("Invalid Email")
		}
		return nil
	}
}

// SaveUser : create user in database
func (u *User) SaveUser(db *gorm.DB) (*User, error) {
	var err error

	// To hash the password
	err = u.BeforeSave()
	if err != nil {
		log.Fatal(err)
	}

	err = db.Debug().Create(&u).Error
	if err != nil {
		return &User{}, err
	}
	return u, nil
}

// FindAllUsers : retrieve all users in database
func (u *User) FindAllUsers(db *gorm.DB, srv pbUser.UserService_ListUsersServer) error {

	var err error
	rows, err := db.Model(&User{}).Rows()
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var user User
		// ScanRows is a method of `gorm.DB`, it can be used to scan a row into a struct
		err := db.ScanRows(rows, &user)
		if err != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Error while reading database iteration: %v", err),
			)
		}

		// Convert timestamp
		createdAtTimesStamp, err := ptypes.TimestampProto(user.CreatedAt)
		updatedAtTimesStamp, err := ptypes.TimestampProto(user.UpdatedAt)

		// do something
		srv.Send(&pbUser.UserDTO{
			ID:        int32(user.ID),
			Nickname:  user.Nickname,
			Email:     user.Email,
			CreatedAt: createdAtTimesStamp,
			UpdatedAt: updatedAtTimesStamp,
		})
	}
	return nil
}

// FindUserByID : find one user by id
func (u *User) FindUserByID(db *gorm.DB, uid uint32) (*User, error) {
	var err error
	err = db.Debug().Model(User{}).Where("id = ?", uid).Take(&u).Error
	if err != nil {
		return &User{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return &User{}, errors.New("User Not Found")
	}
	return u, err
}

// UpdateAUser : update one user fields
func (u *User) UpdateAUser(db *gorm.DB, uid uint32) (*User, error) {

	// To hash the password
	err := u.BeforeSave()
	if err != nil {
		log.Fatal(err)
	}
	db = db.Debug().Model(&User{}).Where("id = ?", uid).Take(&User{}).UpdateColumns(
		map[string]interface{}{
			"password":   u.Password,
			"nickname":   u.Nickname,
			"email":      u.Email,
			"updated_at": time.Now(),
		},
	)
	if db.Error != nil {
		return &User{}, db.Error
	}
	// This is the display the updated user
	err = db.Debug().Model(&User{}).Where("id = ?", uid).Take(&u).Error
	if err != nil {
		return &User{}, err
	}
	return u, nil
}

// DeleteAUser : delete user
func (u *User) DeleteAUser(db *gorm.DB, uid uint32) (int64, error) {

	db = db.Debug().Model(&User{}).Where("id = ?", uid).Take(&User{}).Delete(&User{})

	if db.Error != nil {
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
