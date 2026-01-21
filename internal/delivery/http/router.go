package http

import (
	"github.com/gin-gonic/gin"
)

func InitRouter(handler *Handler) *gin.Engine {
	r := gin.Default()

	// Serve static files from the "uploads" directory
	r.Static("/uploads", "./uploads")

	// Public Routes
	api := r.Group("/api/v1")
	{
		api.POST("/register", handler.Register)
		api.POST("/login", handler.Login)
		api.POST("/forgot-password", handler.ForgotPassword)
	}

	// Protected Routes (Student, Instructor, Admin)
	protected := api.Group("/")
	protected.Use(AuthMiddleware())
	{
		protected.GET("/courses/:id", handler.GetCourseDetail)
		protected.PUT("/profile", handler.UpdateProfile)
		protected.POST("/labs/:id/enroll", handler.StudentEnrollInLab)
		protected.GET("/certificates", handler.GetUserCertificates)
	}

	// Instructor & Admin Only
	instructor := api.Group("/instructor")
	instructor.Use(AuthMiddleware("instructor", "admin"))
	{
		// Course routes
		instructor.POST("/courses", handler.CreateCourse)
		instructor.POST("/modules", handler.AddModule)

		// Lab routes
		instructor.POST("/labs", handler.CreateLab)
		instructor.PATCH("/labs/:id/status", handler.UpdateLabStatus)
		instructor.GET("/labs/:id/ungraded", handler.GetUngradedStudents)

		instructor.POST("/grades", handler.SubmitGrade)
	}

	return r
}
