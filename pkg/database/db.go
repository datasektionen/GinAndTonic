package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() (*gorm.DB, error) {
	var err error
	var dsn string

	if os.Getenv("ENV") == "dev" {
		if err = godotenv.Load(); err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}

		host := os.Getenv("DB_HOST")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")
		port := os.Getenv("DB_PORT")
		sslmode := os.Getenv("DB_SSLMODE")
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC", host, user, password, dbname, port, sslmode)
	} else if os.Getenv("ENV") == "prod" {
		dsn = os.Getenv("DATABASE_URL")
	} else {
		log.Fatalf("Error loading .env file: %v", err)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
		return nil, err
	}

	const IDLE_IN_TRANSACTION_SESSION_TIMEOUT = "300000" // 300000 milliseconds = 5 minutes

	// Directly interpolate the constant value into the SQL command
	err = db.Exec(fmt.Sprintf("SET idle_in_transaction_session_timeout = %s", IDLE_IN_TRANSACTION_SESSION_TIMEOUT)).Error
	if err != nil {
		log.Fatalf("failed to set idle_in_transaction_session_timeout, got error: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle
	// connection pool.
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(90)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Minute * 3)

	return db, nil
}
