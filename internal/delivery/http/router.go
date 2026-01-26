package http

import (
	"github.com/gin-gonic/gin"
)

func InitRouter(handler *Handler) *gin.Engine {
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Serve static files
	r.Static("/uploads", "./uploads")

	// API v1
	api := r.Group("/api/v1")
	{
		// ========== PUBLIC ROUTES ==========
		auth := api.Group("/auth")
		{
			auth.POST("/register", handler.Register)
			auth.POST("/login", handler.Login)
			auth.POST("/forgot-password", handler.ForgotPassword)
		}

		// ========== STUDENT ROUTES ==========
		student := api.Group("/student")
		student.Use(AuthMiddleware("student"))
		{
			// Dashboard
			student.GET("/dashboard", handler.GetStudentDashboard)

			// Profile
			student.PUT("/profile", handler.UpdateProfile)

			// Courses (Browse & Enroll)
			student.GET("/courses", handler.GetAllCourses)
			student.GET("/courses/:id", handler.GetCourseDetail)
			student.POST("/courses/:id/enroll", handler.EnrollCourse)

			// Enrollments (Jalur Pembelajaran)
			student.GET("/enrollments", handler.GetMyEnrollments)
			student.GET("/courses/:id/modules", handler.GetModulesWithProgress)
			student.POST("/modules/complete", handler.MarkModuleComplete)

			// Assignments
			student.POST("/assignments/submit", handler.SubmitAssignment)

			// Labs
			student.GET("/labs", handler.GetAllLabs)
			student.POST("/labs/:id/enroll", handler.StudentEnrollInLab)

			// Certificates
			student.GET("/certificates", handler.GetUserCertificates)

			// Reports
			student.GET("/performance", handler.GetStudentPerformance)
		}

		// ========== INSTRUCTOR ROUTES ==========
		instructor := api.Group("/instructor")
		instructor.Use(AuthMiddleware("instructor", "admin"))
		{
			// Dashboard
			instructor.GET("/dashboard", handler.GetInstructorDashboard)

			// Profile
			instructor.PUT("/profile", handler.UpdateProfile)

			// Courses Management
			instructor.POST("/courses", handler.CreateCourse)
			instructor.GET("/courses", handler.GetAllCourses)
			instructor.GET("/courses/:id", handler.GetCourseDetail)
			instructor.PUT("/courses/:id", handler.UpdateCourse)
			instructor.DELETE("/courses/:id", handler.DeleteCourse)
			instructor.POST("/courses/:id/publish", handler.PublishCourse)
			instructor.POST("/courses/:id/unpublish", handler.UnpublishCourse)
			
			// Modules Management
			instructor.POST("/modules", handler.AddModule)
			instructor.PUT("/modules/:id", handler.UpdateModule)
			instructor.DELETE("/modules/:id", handler.DeleteModule)

			// Grading
			instructor.POST("/assignments/grade", handler.GradeAssignment)

			// Labs Management
			instructor.POST("/labs", handler.CreateLab)
			instructor.GET("/labs", handler.GetAllLabs)
			instructor.GET("/labs/:id", handler.GetLabByID)
			instructor.PUT("/labs/:id", handler.UpdateLab)
			instructor.PATCH("/labs/:id/status", handler.UpdateLabStatus)
			instructor.DELETE("/labs/:id", handler.DeleteLab)
			instructor.POST("/labs/grade", handler.SubmitLabGrade)
			instructor.GET("/labs/:id/ungraded", handler.GetUngradedStudents)
			instructor.GET("/labs/:id/students", handler.GetLabStudents)

			// Certificates
			instructor.GET("/certificates/pending", handler.GetPendingCertificates)
			instructor.POST("/certificates/:id/approve", handler.ApproveCertificate)

			// Reports
			instructor.GET("/students/performance", handler.GetAllStudentsPerformance)
		}

		// ========== ADMIN ROUTES ==========
		admin := api.Group("/admin")
		admin.Use(AuthMiddleware("admin"))
		{
			// Dashboard
			admin.GET("/dashboard", handler.GetAdminDashboard)

			// User Management (Pendaftaran)
			admin.GET("/users", handler.GetAllUsers)
			admin.POST("/users", handler.CreateUser)

			// Inherits all instructor routes
			admin.POST("/courses", handler.CreateCourse)
			admin.GET("/courses", handler.GetAllCourses)
			admin.GET("/courses/:id", handler.GetCourseDetail)
			admin.PUT("/courses/:id", handler.UpdateCourse)
			admin.DELETE("/courses/:id", handler.DeleteCourse)
			admin.POST("/courses/:id/publish", handler.PublishCourse)
			admin.POST("/courses/:id/unpublish", handler.UnpublishCourse)
			admin.POST("/modules", handler.AddModule)
			admin.PUT("/modules/:id", handler.UpdateModule)
			admin.DELETE("/modules/:id", handler.DeleteModule)
			admin.POST("/labs", handler.CreateLab)
			admin.GET("/labs", handler.GetAllLabs)
			admin.GET("/labs/:id", handler.GetLabByID)
			admin.PUT("/labs/:id", handler.UpdateLab)
			admin.PATCH("/labs/:id/status", handler.UpdateLabStatus)
			admin.DELETE("/labs/:id", handler.DeleteLab)
			admin.POST("/labs/grade", handler.SubmitLabGrade)
			admin.GET("/labs/:id/ungraded", handler.GetUngradedStudents)
			admin.GET("/certificates/pending", handler.GetPendingCertificates)
			admin.POST("/certificates/:id/approve", handler.ApproveCertificate)
		}
	}

	return r
}

// InitFileRouter initializes file-related routes for GridFS
func InitFileRouter(r *gin.Engine, fileHandler *FileHandler) {
	// Protected file streaming with enrollment verification
	api := r.Group("/api/v1")
	{
		files := api.Group("/files")
		files.Use(AuthMiddleware("student", "instructor", "admin"))
		{
			// Protected file streaming (requires auth and enrollment check)
			files.GET("/:id/stream", fileHandler.StreamFileProtected)
			files.GET("/:id/info", fileHandler.GetFileInfo)
			files.POST("/upload", fileHandler.UploadFile)
			files.DELETE("/:id", fileHandler.DeleteFile)
		}
	}

	// Legacy public endpoint (deprecated, kept for backward compatibility)
	// Should be removed in production for security
	r.GET("/files/:id", fileHandler.StreamFile)
}
