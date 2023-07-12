package test

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

const (
	UsernameTestDB = "postgres"
	PasswordTestDB = "password"
	HostTestDB     = "localhost"
	PortTestDB     = "5432"
	DBnameTestDB   = "lavka_test"
	SslmodeTestDB  = "disable"
	UpTestDBFile   = "migrations/up.sql"
	DownTestDBFile = "migrations/down.sql"
)

func OpenTestDatabase() (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		HostTestDB, PortTestDB, UsernameTestDB, DBnameTestDB, PasswordTestDB, SslmodeTestDB)), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database. %v", err)
	}

	db.Logger = logger.Default.LogMode(logger.Info)

	return db, nil
}

func PrepareTestDatabase(prefix string) (*gorm.DB, error) {
	db, err := OpenTestDatabase()
	if err != nil {
		log.Fatal(err)
	}

	down, err := os.ReadFile(prefix + DownTestDBFile)
	if err != nil {
		log.Fatal(err)
	}

	schema, err := os.ReadFile(prefix + UpTestDBFile)
	if err != nil {
		log.Fatal(err)
	}
	tx := db.Exec(string(down))
	if tx.Error != nil {
		log.Println(tx.Error.Error())
	}
	tx = db.Exec(string(schema))
	if tx.Error != nil {
		log.Println(tx.Error.Error())
	}

	return db, err
}
