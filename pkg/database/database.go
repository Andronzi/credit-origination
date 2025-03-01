package database

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func readSecret(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read secret: %w", err)
	}
	return string(data), err
}

func ConnectPostgres() (*gorm.DB, error) {
	userPath := os.Getenv("DB_USER_FILE")
	passwordPath := os.Getenv("DB_PASSWORD_FILE")
	dbNamePath := os.Getenv("DB_NAME_FILE")

	user, err := readSecret(userPath)
	if err != nil {
		return nil, err
	}

	password, err := readSecret(passwordPath)
	if err != nil {
		return nil, err
	}

	dbname, err := readSecret(dbNamePath)
	if err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		user,
		password,
		dbname,
	)

	fmt.Println("Connecting with DSN:", dsn)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// sqlDB, err := db.DB()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get database instance: %w", err)
	// }

	// sqlDB.SetMaxIdleConns(10)
	// sqlDB.SetMaxOpenConns(100)
	// sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
