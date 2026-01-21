package http

import "github.com/gin-gonic/gin"

func InitWebRouter(router *gin.Engine, webHandler *WebHandler) {
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/**/*.html")

	// Web routes
	web := router.Group("/")
	{
		web.GET("/", webHandler.ShowLoginPage)
		web.GET("/student/dashboard", webHandler.StudentDashboard)
		web.GET("/instructor/dashboard", webHandler.InstructorDashboard)
	}
}
