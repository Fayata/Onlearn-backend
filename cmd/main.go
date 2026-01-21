package main

import (
	"context"
	"log"
	"os"

	"onlearn-backend/config"
	"onlearn-backend/internal/domain"
	httpDelivery "onlearn-backend/internal/delivery/http"
	"onlearn-backend/internal/repository"
	"onlearn-backend/internal/usecase"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Connect to databases
	db := config.ConnectDB()
	postgres := db.PG
	mongo := db.Mongo

	// Auto migrate
	if err := config.AutoMigrate(postgres); err != nil {
		log.Fatal("Migration failed:", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(postgres)
	courseRepo := repository.NewCourseRepository(postgres)
	labRepo := repository.NewLabRepository(postgres)
	certRepo := repository.NewCertificateRepository(postgres)
	moduleRepo := repository.NewModuleRepository(mongo)

	// Initialize usecases
	authUsecase := usecase.NewAuthUsecase(userRepo)
	courseUsecase := usecase.NewCourseUsecase(courseRepo, moduleRepo, certRepo)
	labUsecase := usecase.NewLabUsecase(labRepo, userRepo)
	certUsecase := usecase.NewCertificateUsecase(certRepo, userRepo)

	// Seed demo users
	seedUsers(authUsecase)

	// Initialize handlers
	apiHandler := httpDelivery.NewHandler(authUsecase, courseUsecase, labUsecase, certUsecase)
	webHandler := httpDelivery.NewWebHandler(authUsecase, courseUsecase, labUsecase, certUsecase)

	// Initialize router with both API and Web handlers
	router := httpDelivery.InitRouter(apiHandler)
	httpDelivery.InitWebRouter(router, webHandler)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s", port)
	log.Printf("Web UI: http://localhost:%s", port)
	log.Printf("API: http://localhost:%s/api/v1", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func seedUsers(authUsecase domain.AuthUsecase) {
	// Student
	student := &domain.User{
		Name:     "Demo Student",
		Email:    "student@onlearn.com",
		Password: "password123",
		Role:     domain.RoleStudent,
	}
	err := authUsecase.Register(context.Background(), student)
	if err != nil && err.Error() != "email already registered" {
		log.Printf("Failed to seed student: %v", err)
	}

	// Instructor
	instructor := &domain.User{
		Name:     "Demo Instructor",
		Email:    "instructor@onlearn.com",
		Password: "password123",
		Role:     domain.RoleInstructor,
	}
	err = authUsecase.Register(context.Background(), instructor)
	if err != nil && err.Error() != "email already registered" {
		log.Printf("Failed to seed instructor: %v", err)
	}
}
