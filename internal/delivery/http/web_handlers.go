package http

import (
	"net/http"
	"onlearn-backend/internal/domain"

	"github.com/gin-gonic/gin"
)

type WebHandler struct {
	AuthUsecase        domain.AuthUsecase
	CourseUsecase      domain.CourseUsecase
	LabUsecase         domain.LabUsecase
	CertificateUsecase domain.CertificateUsecase
}

func NewWebHandler(au domain.AuthUsecase, cu domain.CourseUsecase, lu domain.LabUsecase, certUC domain.CertificateUsecase) *WebHandler {
	return &WebHandler{
		AuthUsecase:        au,
		CourseUsecase:      cu,
		LabUsecase:         lu,
		CertificateUsecase: certUC,
	}
}

func (h *WebHandler) ShowLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Login | Onlearn",
	})
}

func (h *WebHandler) StudentDashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title": "Student Dashboard | Onlearn",
	})
}

func (h *WebHandler) InstructorDashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "instructor_dashboard.html", gin.H{
		"title": "Instructor Dashboard | Onlearn",
	})
}
