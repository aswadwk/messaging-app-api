package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func DBConnect() *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		Cfg.DBHost, Cfg.DBPort, Cfg.DBUserName, Cfg.DBPassword, Cfg.DBName,
	)
	dbUrl := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		Cfg.DBUserName,
		Cfg.DBPassword,
		Cfg.DBHost,
		Cfg.DBPort,
		Cfg.DBName,
	)

	if Cfg.Debug {
		fmt.Println("Connecting to main database with URL:", dbUrl)
	}

	// Create GORM logger
	dbLogger := createGormLogger(Cfg.LogLevel)

	// Open a new database connection using GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: dbLogger,
	})

	if err != nil {
		log.Fatal("Failed to connect to main database:", err)
	}

	sqlDB, err := db.Debug().DB()
	if err != nil {
		log.Fatal("Failed to get DB instance:", err)
	}

	//  DB connection pooling
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db
}

// Helper function to create a GORM logger with the specified log level
func createGormLogger(logLevel string) logger.Interface {
	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:        time.Second,             // Slow SQL threshold
			LogLevel:             parseLogLevel(logLevel), // Log level
			ParameterizedQueries: true,                    // Don't include params in the SQL log
			Colorful:             true,                    // Enable color
		},
	)
}

func parseLogLevel(level string) logger.LogLevel {
	switch level {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Info // fallback
	}
}
