package tests

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type dbProxy struct {
	db       *gorm.DB
	username string
	password string
	host     string
	port     string
}

var (
	instance *dbProxy
	once     sync.Once
)

func ConnectDB() {
	once.Do(func() {
		instance = &dbProxy{
			username: os.Getenv("DB_USER"),
			password: os.Getenv("DB_PASSWORD"),
			port:     os.Getenv("DB_PORT"),
			host:     os.Getenv("GATEWAY"), //172.17.0.1
		}
		instance.connect()
	})
}

func (d *dbProxy) connect() {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		instance.host,
		instance.port, //5434
		instance.username,
		instance.password,
		"postgres",
	)

	maxRetries := 4
	retryDelay := 4 * time.Second

	var err error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		instance.db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break // Success, exit retry loop
		}

		log.Printf("Database connection attempt %d failed: %v", attempt, err)

		if attempt < maxRetries {
			log.Printf("Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		log.Fatal("Failed to connect to database after", maxRetries, "attempts:", err)
		panic("Dropping tests")
	}
}
