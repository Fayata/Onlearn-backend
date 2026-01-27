package main

import (
	"context"
	"log"
	"os"

	"onlearn-backend/config"
	httpDelivery "onlearn-backend/internal/delivery/http"
	"onlearn-backend/internal/domain"
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

	// ========== Initialize Repositories ==========
	userRepo := repository.NewUserRepository(postgres)
	courseRepo := repository.NewCourseRepository(postgres)
	enrollmentRepo := repository.NewEnrollmentRepository(postgres)
	moduleProgressRepo := repository.NewModuleProgressRepository(postgres)
	assignmentRepo := repository.NewAssignmentRepository(postgres)
	labRepo := repository.NewLabRepository(postgres)
	certRepo := repository.NewCertificateRepository(postgres)
	moduleRepo := repository.NewModuleRepository(mongo)

	// Initialize GridFS Repository for file storage
	gridFSRepo, err := repository.NewGridFSRepository(mongo)
	if err != nil {
		log.Fatal("Failed to initialize GridFS:", err)
	}
	log.Println("✅ GridFS initialized successfully")

	// ========== Initialize Usecases ==========
	authUsecase := usecase.NewAuthUsecase(userRepo)

	userUsecase := usecase.NewUserUsecase(userRepo)

	courseUsecase := usecase.NewCourseUsecase(
		courseRepo,
		moduleRepo,
		enrollmentRepo,
		moduleProgressRepo,
		assignmentRepo,
		certRepo,
		userRepo,
	)

	labUsecase := usecase.NewLabUsecase(
		labRepo,
		userRepo,
		certRepo,
	)

	certUsecase := usecase.NewCertificateUsecase(
		certRepo,
		userRepo,
		courseRepo,
		labRepo,
	)

	dashboardUsecase := usecase.NewDashboardUsecase(
		userRepo,
		courseRepo,
		enrollmentRepo,
		moduleRepo,
		moduleProgressRepo,
		assignmentRepo,
		labRepo,
		certRepo,
	)

	reportUsecase := usecase.NewReportUsecase(
		userRepo,
		enrollmentRepo,
		assignmentRepo,
		certRepo,
	)

	// Seed demo users
	seedUsers(authUsecase)

	// ========== Initialize Handlers ==========
	apiHandler := httpDelivery.NewHandler(
		authUsecase,
		userUsecase,
		courseUsecase,
		labUsecase,
		certUsecase,
		dashboardUsecase,
		reportUsecase,
	)

	webHandler := httpDelivery.NewWebHandler(
		authUsecase,
		courseUsecase,
		labUsecase,
		certUsecase,
		dashboardUsecase,
	)

	fileHandler := httpDelivery.NewFileHandler(gridFSRepo)
	fileHandler.SetCourseUsecase(courseUsecase)

	// ========== Initialize Router ==========
	router := httpDelivery.InitRouter(apiHandler)
	httpDelivery.InitWebRouter(router, webHandler)
	httpDelivery.InitFileRouter(router, fileHandler)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("=================================================")
	log.Printf(" OnLearn Backend Server Starting...")
	log.Printf("=================================================")
	log.Printf(" Server running on port %s", port)
	log.Printf(" Web UI: http://localhost:%s", port)
	log.Printf(" API: http://localhost:%s/api/v1", port)
	log.Printf("=================================================")
	log.Printf(" Demo Accounts:")
	log.Printf("   Student: student@onlearn.com / password123")
	log.Printf("   Instructor: instructor@onlearn.com / password123")
	log.Printf("=================================================")

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func seedUsers(authUsecase domain.AuthUsecase) {
	ctx := context.Background()

	// Student
	student := &domain.User{
		Name:     "Demo Student",
		Email:    "student@onlearn.com",
		Password: "password123",
		Role:     domain.RoleStudent,
	}
	err := authUsecase.Register(ctx, student)
	if err != nil && err.Error() != "email already exists" {
		log.Printf("Failed to seed student: %v", err)
	} else if err == nil {
		log.Println("✅ Demo student account created")
	}

	// Instructor
	instructor := &domain.User{
		Name:     "Demo Instructor",
		Email:    "instructor@onlearn.com",
		Password: "password123",
		Role:     domain.RoleInstructor,
	}
	err = authUsecase.Register(ctx, instructor)
	if err != nil && err.Error() != "email already exists" {
		log.Printf("Failed to seed instructor: %v", err)
	} else if err == nil {
		log.Println("✅ Demo instructor account created")
	}

	// Admin
	admin := &domain.User{
		Name:     "Demo Admin",
		Email:    "admin@onlearn.com",
		Password: "password123",
		Role:     domain.RoleAdmin,
	}
	err = authUsecase.Register(ctx, admin)
	if err != nil && err.Error() != "email already exists" {
		log.Printf("Failed to seed admin: %v", err)
	} else if err == nil {
		log.Println("✅ Demo admin account created")
	}
}
