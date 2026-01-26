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
