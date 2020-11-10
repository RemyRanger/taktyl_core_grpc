package server

import (
	"fmt"
	"log"
	"sync"

	"github.com/RemyRanger/taktyl_core_grpc/src/models"
	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/postgres" //postgres database driver
)

// Backend implements the protobuf interface
type Backend struct {
	mu *sync.RWMutex
	DB *gorm.DB
}

// New initializes a new Backend struct.
func New() *Backend {
	return &Backend{
		mu: &sync.RWMutex{},
	}
}

// Initialize : init router and database link
func (b *Backend) Initialize(Dbdriver, DbUser, DbPassword, DbPort, DbHost, DbName string) *Backend {

	var err error

	if Dbdriver == "postgres" {
		DBURL := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", DbHost, DbPort, DbUser, DbName, DbPassword)
		b.DB, err = gorm.Open(Dbdriver, DBURL)
		if err != nil {
			fmt.Printf("Cannot connect to %s database", Dbdriver)
			log.Fatal("This is the error:", err)
		} else {
			fmt.Printf("We are connected to the %s database \n", Dbdriver)
		}
	}

	b.DB.Debug().AutoMigrate(&models.User{}) //, &models.Event{}) //database migration

	return b
}
