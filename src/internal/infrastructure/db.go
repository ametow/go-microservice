package infrastructure

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

func ConnectDb(cfg Config) (*gorm.DB, error) {
	maxAttempt := 3
	tryCount := 0

TryConnect:
	db, err := gorm.Open(postgres.Open(fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.DBName, cfg.Password, cfg.SSLMode)), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	for tryCount < maxAttempt && err != nil {
		tryCount++
		time.Sleep(time.Second)
		goto TryConnect
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts. %v", tryCount, err)
	}

	raw, err := os.ReadFile("migrations/up.sql")
	if err != nil {
		return nil, err
	}

	tx := db.Exec(string(raw))

	return db, tx.Error
}
