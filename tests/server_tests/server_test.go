package servertests

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"

	"github.com/RemyRanger/taktyl_core_grpc/src/models"
	"github.com/RemyRanger/taktyl_core_grpc/src/server"
)

var backend = server.New()
var userInstance = models.User{}
var eventInstance = models.Event{}

func TestMain(m *testing.M) {
	var err error
	err = godotenv.Load(os.ExpandEnv("../../.env"))
	if err != nil {
		log.Fatalf("Error getting env %v\n", err)
	}
	Database()

	os.Exit(m.Run())
}

func Database() {

	var err error

	TestDbDriver := os.Getenv("TestDbDriver")

	if TestDbDriver == "postgres" {
		DBURL := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", os.Getenv("TestDbHost"), os.Getenv("TestDbPort"), os.Getenv("TestDbUser"), os.Getenv("TestDbName"), os.Getenv("TestDbPassword"))
		backend.DB, err = gorm.Open(TestDbDriver, DBURL)
		if err != nil {
			fmt.Printf("Cannot connect to %s database\n", TestDbDriver)
			log.Fatal("This is the error:", err)
		} else {
			fmt.Printf("We are connected to the %s database\n", TestDbDriver)
		}
	}
}

func refreshUserTable() error {
	err := backend.DB.DropTableIfExists(&models.User{}).Error
	if err != nil {
		return err
	}
	err = backend.DB.AutoMigrate(&models.User{}).Error
	if err != nil {
		return err
	}
	log.Printf("Successfully refreshed table")
	return nil
}

func seedOneUser() (models.User, error) {

	err := refreshUserTable()
	if err != nil {
		log.Fatal(err)
	}

	user := models.User{
		Nickname: "Pet",
		Email:    "pet@gmail.com",
		Password: "password",
	}

	err = backend.DB.Model(&models.User{}).Create(&user).Error
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func seedUsers() ([]models.User, error) {

	var err error
	if err != nil {
		return nil, err
	}
	users := []models.User{
		models.User{
			Nickname: "Steven victor",
			Email:    "steven@gmail.com",
			Password: "password",
		},
		models.User{
			Nickname: "Kenny Morris",
			Email:    "kenny@gmail.com",
			Password: "password",
		},
	}
	for i := range users {
		err := backend.DB.Model(&models.User{}).Create(&users[i]).Error
		if err != nil {
			return []models.User{}, err
		}
	}
	return users, nil
}

func refreshUserAndEventTable() error {

	err := backend.DB.DropTableIfExists(&models.User{}, &models.Event{}).Error
	if err != nil {
		return err
	}
	err = backend.DB.AutoMigrate(&models.User{}, &models.Event{}).Error
	if err != nil {
		return err
	}
	log.Printf("Successfully refreshed tables")
	return nil
}

func seedOneUserAndOneEvent() (models.Event, error) {

	err := refreshUserAndEventTable()
	if err != nil {
		return models.Event{}, err
	}
	user := models.User{
		Nickname: "Sam Phil",
		Email:    "sam@gmail.com",
		Password: "password",
	}
	err = backend.DB.Model(&models.User{}).Create(&user).Error
	if err != nil {
		return models.Event{}, err
	}
	event := models.Event{
		Title:    "This is the title sam",
		Content:  "This is the content sam",
		AuthorID: user.ID,
	}
	err = backend.DB.Model(&models.Event{}).Create(&event).Error
	if err != nil {
		return models.Event{}, err
	}
	return event, nil
}

func seedUsersAndEvents() ([]models.User, []models.Event, error) {

	var err error

	if err != nil {
		return []models.User{}, []models.Event{}, err
	}
	var users = []models.User{
		models.User{
			Nickname: "Steven victor",
			Email:    "steven@gmail.com",
			Password: "password",
		},
		models.User{
			Nickname: "Magu Frank",
			Email:    "magu@gmail.com",
			Password: "password",
		},
	}
	var events = []models.Event{
		models.Event{
			Title:   "Title 1",
			Content: "Hello world 1",
		},
		models.Event{
			Title:   "Title 2",
			Content: "Hello world 2",
		},
	}

	for i := range users {
		err = backend.DB.Model(&models.User{}).Create(&users[i]).Error
		if err != nil {
			log.Fatalf("cannot seed users table: %v", err)
		}
		events[i].AuthorID = users[i].ID

		err = backend.DB.Model(&models.Event{}).Create(&events[i]).Error
		if err != nil {
			log.Fatalf("cannot seed events table: %v", err)
		}
	}
	return users, events, nil
}
