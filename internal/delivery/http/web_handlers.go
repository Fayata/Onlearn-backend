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
	// Get user from context (set by auth middleware)
	userValue, exists := c.Get("user")
	if !exists {
		c.Redirect(http.StatusFound, "/login?error=Unauthorized")
		return
	}

	user, ok := userValue.(domain.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login?error=Invalid+user+data")
		return
	}

	// Get dashboard data
	dashboardData, err := h.DashboardUsecase.GetStudentDashboard(c.Request.Context(), user.ID)
	if err != nil {
		// For simplicity, redirecting to login. A proper error page would be better.
		c.Redirect(http.StatusFound, "/login?error=Could+not+load+dashboard")
		return
	}

	c.HTML(http.StatusOK, "dashboard.html", dashboardData)
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
