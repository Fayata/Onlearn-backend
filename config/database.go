package config

import (
	"context"
	"fmt"
	"log"
	"onlearn-backend/internal/domain"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	PG    *gorm.DB
	Mongo *mongo.Database
}

func ConnectDB() *Database {
	err := godotenv.Load()
	if err != nil {
		log.Println("Note: .env file not found, using system environment variables")
	}

	// 1. PostgreSQL Connection
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)
	pgDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}

	// 2. MongoDB Connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoURI := os.Getenv("MONGO_URI")
	clientOptions := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	mongoDB := mongoClient.Database(os.Getenv("MONGO_DB_NAME"))

	log.Println("Connected to PostgreSQL and MongoDB successfully!")

	return &Database{
		PG:    pgDB,
		Mongo: mongoDB,
	}
}

func AutoMigrate(db *gorm.DB) error {
	// Pre-migration: Convert existing letter grades to numeric values
	// Check if lab_grades table exists and has grade column as varchar
	var columnType string
	db.Raw("SELECT data_type FROM information_schema.columns WHERE table_name = 'lab_grades' AND column_name = 'grade'").Scan(&columnType)
	
	if columnType == "character varying" || columnType == "varchar" {
		log.Println("Converting letter grades to numeric values...")
		// Convert A=90, B=80, C=70, D=60, E=50, empty/null=null
		db.Exec(`
			UPDATE lab_grades SET grade = CASE 
				WHEN grade = 'A' THEN '90'
				WHEN grade = 'B' THEN '80'
				WHEN grade = 'C' THEN '70'
				WHEN grade = 'D' THEN '60'
				WHEN grade = 'E' THEN '50'
				WHEN grade = '' OR grade IS NULL THEN NULL
				ELSE grade
			END
		`)
	}
	
	err := db.AutoMigrate(
		&domain.User{},
		&domain.Course{},
		&domain.Lab{},
		&domain.Enrollment{},
		&domain.LabGrade{},
		&domain.Certificate{},
		&domain.ModuleProgress{},
		&domain.Assignment{},
	)
	if err != nil {
		return err
	}
	log.Println("Database migration completed!")
	return nil
}
