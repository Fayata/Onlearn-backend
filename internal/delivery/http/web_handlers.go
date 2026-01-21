package http

import (
	"net/http"
	"onlearn-backend/internal/domain"

	"github.com/gin-gonic/gin"
)

type WebHandler struct {
	AuthUsecase      domain.AuthUsecase
	CourseUsecase    domain.CourseUsecase
	LabUsecase       domain.LabUsecase
	CertUsecase      domain.CertificateUsecase
	DashboardUsecase domain.DashboardUsecase
}

func NewWebHandler(
	au domain.AuthUsecase,
	cu domain.CourseUsecase,
	lu domain.LabUsecase,
	certu domain.CertificateUsecase,
	du domain.DashboardUsecase,
) *WebHandler {
	return &WebHandler{
		AuthUsecase:      au,
		CourseUsecase:    cu,
		LabUsecase:       lu,
		CertUsecase:      certu,
		DashboardUsecase: du,
	}
}

// ========== AUTH PAGES ==========

func (h *WebHandler) ShowLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Login | OnLearn",
	})
}

// ========== STUDENT PAGES ==========

func (h *WebHandler) StudentDashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title": "Student Dashboard | OnLearn",
	})
}

// ========== INSTRUCTOR PAGES ==========

func (h *WebHandler) InstructorDashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "instructor_dashboard.html", gin.H{
		"title": "Instructor Dashboard | OnLearn",
	})
}

// ========== ADMIN PAGES ==========

func (h *WebHandler) AdminDashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "admin_dashboard.html", gin.H{
		"title": "Admin Dashboard | OnLearn",
	})
}
