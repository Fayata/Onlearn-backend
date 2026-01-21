package http

import (
	"errors"
	"fmt"
	"net/http"
	"onlearn-backend/internal/domain"
	"onlearn-backend/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	AuthUsecase        domain.AuthUsecase
	CourseUsecase      domain.CourseUsecase
	LabUsecase         domain.LabUsecase
	CertificateUsecase domain.CertificateUsecase
}

func NewHandler(au domain.AuthUsecase, cu domain.CourseUsecase, lu domain.LabUsecase, certUC domain.CertificateUsecase) *Handler {
	return &Handler{
		AuthUsecase:        au,
		CourseUsecase:      cu,
		LabUsecase:         lu,
		CertificateUsecase: certUC,
	}
}

// formatValidationErrors formats validation errors into a readable format.
func formatValidationErrors(err error) gin.H {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		errors := make(map[string]string)
		for _, f := range ve {
			errors[f.Field()] = fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", f.Field(), f.Tag())
		}
		return gin.H{"error": "Validation failed", "details": errors}
	}
	return gin.H{"error": "Invalid request: " + err.Error()}
}

// getUserID extracts user ID from context
func getUserID(c *gin.Context) (uint, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, errors.New("user ID not found in token")
	}
	return userID.(uint), nil
}

// --- Auth Handlers ---

func (h *Handler) Register(c *gin.Context) {
	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, formatValidationErrors(err))
		return
	}

	if user.Role == "" {
		user.Role = domain.RoleStudent
	}

	if err := h.AuthUsecase.Register(c.Request.Context(), &user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully. Please check your email for verification."})
}

func (h *Handler) Login(c *gin.Context) {
	var creds struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, formatValidationErrors(err))
		return
	}

	token, err := h.AuthUsecase.Login(c.Request.Context(), creds.Email, creds.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var user domain.User
	user.ID = userID
	user.Name = c.PostForm("name")
	user.Password = c.PostForm("password")

	filePath, err := utils.HandleUpload(c, "profile_picture")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file: " + err.Error()})
		return
	}
	if filePath != "" {
		user.ProfilePicture = filePath
	}

	if err := h.AuthUsecase.UpdateUser(c.Request.Context(), &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

func (h *Handler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, formatValidationErrors(err))
		return
	}

	h.AuthUsecase.ForgotPassword(c.Request.Context(), req.Email)
	c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a password reset link has been sent."})
}

// --- Course Handlers ---

func (h *Handler) CreateCourse(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var course domain.Course
	course.Title = c.PostForm("title")
	course.Description = c.PostForm("description")
	course.InstructorID = userID

	if course.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is required"})
		return
	}

	filePath, err := utils.HandleUpload(c, "thumbnail")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload thumbnail: " + err.Error()})
		return
	}
	course.Thumbnail = filePath

	if err := h.CourseUsecase.CreateCourse(c.Request.Context(), &course); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, course)
}

func (h *Handler) AddModule(c *gin.Context) {
	courseIDStr := c.PostForm("course_id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course_id"})
		return
	}

	orderStr := c.PostForm("order")
	order, err := strconv.Atoi(orderStr)
	if err != nil {
		order = 0
	}

	var module domain.Module
	module.CourseID = uint(courseID)
	module.Title = c.PostForm("title")
	module.Type = domain.ModuleType(c.PostForm("type"))
	module.Description = c.PostForm("description")
	module.QuizLink = c.PostForm("quiz_link")
	module.Order = order

	if module.Title == "" || module.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title and type are required"})
		return
	}

	filePath, err := utils.HandleUpload(c, "content_url")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload content: " + err.Error()})
		return
	}
	if filePath != "" {
		module.ContentURL = filePath
	}

	if err := h.CourseUsecase.AddModule(c.Request.Context(), &module); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, module)
}

func (h *Handler) GetCourseDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	course, modules, err := h.CourseUsecase.GetCourseDetails(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"course":  course,
		"modules": modules,
	})
}

// --- Lab Handlers ---

func (h *Handler) CreateLab(c *gin.Context) {
	var lab domain.Lab
	if err := c.ShouldBindJSON(&lab); err != nil {
		c.JSON(http.StatusBadRequest, formatValidationErrors(err))
		return
	}

	if lab.Status == "" {
		lab.Status = "scheduled"
	}

	if err := h.LabUsecase.CreateLab(c.Request.Context(), &lab); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, lab)
}

func (h *Handler) UpdateLabStatus(c *gin.Context) {
	idStr := c.Param("id")
	labID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lab ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, formatValidationErrors(err))
		return
	}

	validStatuses := map[string]bool{"scheduled": true, "open": true, "closed": true}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status. Must be: scheduled, open, or closed"})
		return
	}

	if err := h.LabUsecase.UpdateLabStatus(c.Request.Context(), uint(labID), req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Lab status updated successfully"})
}

func (h *Handler) StudentEnrollInLab(c *gin.Context) {
	idStr := c.Param("id")
	labID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lab ID"})
		return
	}

	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if err := h.LabUsecase.StudentEnroll(c.Request.Context(), userID, uint(labID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Successfully enrolled in lab"})
}

func (h *Handler) SubmitGrade(c *gin.Context) {
	instructorID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		LabID     uint   `json:"lab_id" binding:"required"`
		StudentID uint   `json:"student_id" binding:"required"`
		Grade     string `json:"grade" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, formatValidationErrors(err))
		return
	}

	err = h.LabUsecase.SubmitGrade(c.Request.Context(), instructorID, req.StudentID, req.LabID, req.Grade)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Grade submitted successfully",
		"lab_id":     req.LabID,
		"student_id": req.StudentID,
		"grade":      req.Grade,
	})
}

func (h *Handler) GetUngradedStudents(c *gin.Context) {
	idStr := c.Param("id")
	labID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lab ID"})
		return
	}

	students, err := h.LabUsecase.GetUngradedStudents(c.Request.Context(), uint(labID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Transform to simpler response
	var ungradedStudents []domain.UngradedStudent
	for _, student := range students {
		ungradedStudents = append(ungradedStudents, domain.UngradedStudent{
			UserID: student.ID,
			Name:   student.Name,
			Email:  student.Email,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"lab_id":   labID,
		"students": ungradedStudents,
		"count":    len(ungradedStudents),
	})
}

// --- Certificate Handlers ---

func (h *Handler) GetUserCertificates(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	certs, err := h.CertificateUsecase.GetUserCertificates(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"certificates": certs,
		"count":        len(certs),
	})
}
