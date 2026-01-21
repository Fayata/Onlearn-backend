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
	AuthUsecase      domain.AuthUsecase
	UserUsecase      domain.UserUsecase
	CourseUsecase    domain.CourseUsecase
	LabUsecase       domain.LabUsecase
	CertUsecase      domain.CertificateUsecase
	DashboardUsecase domain.DashboardUsecase
	ReportUsecase    domain.ReportUsecase
}

func NewHandler(
	au domain.AuthUsecase,
	uu domain.UserUsecase,
	cu domain.CourseUsecase,
	lu domain.LabUsecase,
	certu domain.CertificateUsecase,
	du domain.DashboardUsecase,
	ru domain.ReportUsecase,
) *Handler {
	return &Handler{
		AuthUsecase:      au,
		UserUsecase:      uu,
		CourseUsecase:    cu,
		LabUsecase:       lu,
		CertUsecase:      certu,
		DashboardUsecase: du,
		ReportUsecase:    ru,
	}
}

// ========== UTILITY FUNCTIONS ==========

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

func getUserID(c *gin.Context) (uint, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, errors.New("user ID not found in token")
	}
	return userID.(uint), nil
}

func getUserRole(c *gin.Context) (string, error) {
	role, exists := c.Get("role")
	if !exists {
		return "", errors.New("role not found in token")
	}
	return role.(string), nil
}

// ========== AUTH HANDLERS ==========

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

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
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

// ========== DASHBOARD HANDLERS ==========

func (h *Handler) GetStudentDashboard(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	data, err := h.DashboardUsecase.GetStudentDashboard(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

func (h *Handler) GetInstructorDashboard(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	data, err := h.DashboardUsecase.GetInstructorDashboard(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

func (h *Handler) GetAdminDashboard(c *gin.Context) {
	data, err := h.DashboardUsecase.GetAdminDashboard(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

// ========== COURSE HANDLERS ==========

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

func (h *Handler) GetAllCourses(c *gin.Context) {
	courses, err := h.CourseUsecase.GetAllCourses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"courses": courses,
		"count":   len(courses),
	})
}

func (h *Handler) GetCourseDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	// Get user ID if authenticated
	var userIDPtr *uint
	if userID, err := getUserID(c); err == nil {
		userIDPtr = &userID
	}

	detail, err := h.CourseUsecase.GetCourseDetails(c.Request.Context(), uint(id), userIDPtr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, detail)
}

func (h *Handler) EnrollCourse(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	idStr := c.Param("id")
	courseID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	if err := h.CourseUsecase.EnrollStudent(c.Request.Context(), userID, uint(courseID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully enrolled in course"})
}

func (h *Handler) GetMyEnrollments(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	enrollments, err := h.CourseUsecase.GetStudentEnrollments(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enrollments": enrollments,
		"count":       len(enrollments),
	})
}

// ========== MODULE HANDLERS ==========

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

func (h *Handler) GetModulesWithProgress(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	courseIDStr := c.Param("id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	modules, err := h.CourseUsecase.GetModulesWithProgress(c.Request.Context(), userID, uint(courseID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"modules": modules,
		"count":   len(modules),
	})
}

func (h *Handler) MarkModuleComplete(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		ModuleID string `json:"module_id" binding:"required"`
		CourseID uint   `json:"course_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, formatValidationErrors(err))
		return
	}

	if err := h.CourseUsecase.MarkModuleComplete(c.Request.Context(), userID, req.ModuleID, req.CourseID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Module marked as complete"})
}

// ========== ASSIGNMENT HANDLERS ==========

func (h *Handler) SubmitAssignment(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	moduleID := c.PostForm("module_id")
	courseIDStr := c.PostForm("course_id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course_id"})
		return
	}

	filePath, err := utils.HandleUpload(c, "file")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file: " + err.Error()})
		return
	}

	assignment := &domain.Assignment{
		UserID:   userID,
		ModuleID: moduleID,
		CourseID: uint(courseID),
		FileURL:  filePath,
	}

	if err := h.CourseUsecase.SubmitAssignment(c.Request.Context(), assignment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Assignment submitted successfully"})
}

func (h *Handler) GradeAssignment(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		AssignmentID uint    `json:"assignment_id" binding:"required"`
		Grade        float64 `json:"grade" binding:"required"`
		Feedback     string  `json:"feedback"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, formatValidationErrors(err))
		return
	}

	if err := h.CourseUsecase.GradeAssignment(c.Request.Context(), req.AssignmentID, req.Grade, req.Feedback, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Assignment graded successfully"})
}

// ========== LAB HANDLERS ==========

func (h *Handler) CreateLab(c *gin.Context) {
	var lab domain.Lab
	if err := c.ShouldBindJSON(&lab); err != nil {
		c.JSON(http.StatusBadRequest, formatValidationErrors(err))
		return
	}

	if err := h.LabUsecase.CreateLab(c.Request.Context(), &lab); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, lab)
}

func (h *Handler) GetAllLabs(c *gin.Context) {
	labs, err := h.LabUsecase.GetAllLabs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"labs":  labs,
		"count": len(labs),
	})
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

func (h *Handler) SubmitLabGrade(c *gin.Context) {
	instructorID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		LabID     uint   `json:"lab_id" binding:"required"`
		StudentID uint   `json:"student_id" binding:"required"`
		Grade     string `json:"grade" binding:"required"`
		Feedback  string `json:"feedback"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, formatValidationErrors(err))
		return
	}

	err = h.LabUsecase.SubmitGrade(c.Request.Context(), instructorID, req.StudentID, req.LabID, req.Grade, req.Feedback)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Grade submitted successfully"})
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

	c.JSON(http.StatusOK, gin.H{
		"students": students,
		"count":    len(students),
	})
}

// ========== CERTIFICATE HANDLERS ==========

func (h *Handler) GetUserCertificates(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	certs, err := h.CertUsecase.GetUserCertificates(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"certificates": certs,
		"count":        len(certs),
	})
}

func (h *Handler) GetPendingCertificates(c *gin.Context) {
	certs, err := h.CertUsecase.GetPendingCertificates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"certificates": certs,
		"count":        len(certs),
	})
}

func (h *Handler) ApproveCertificate(c *gin.Context) {
	approverID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	idStr := c.Param("id")
	certID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid certificate ID"})
		return
	}

	if err := h.CertUsecase.ApproveCertificate(c.Request.Context(), uint(certID), approverID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Certificate approved successfully"})
}

// ========== USER MANAGEMENT (ADMIN) ==========

func (h *Handler) GetAllUsers(c *gin.Context) {
	users, err := h.UserUsecase.GetAllUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"count": len(users),
	})
}

func (h *Handler) CreateUser(c *gin.Context) {
	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, formatValidationErrors(err))
		return
	}

	if err := h.UserUsecase.CreateUser(c.Request.Context(), &user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "user": user})
}

// ========== REPORTS ==========

func (h *Handler) GetStudentPerformance(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	performance, err := h.ReportUsecase.GetStudentPerformance(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, performance)
}

func (h *Handler) GetAllStudentsPerformance(c *gin.Context) {
	performances, err := h.ReportUsecase.GetAllStudentsPerformance(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"performances": performances,
		"count":        len(performances),
	})
}
