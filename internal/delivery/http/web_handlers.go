package http

import (
	"net/http"
	"onlearn-backend/internal/domain"
	"onlearn-backend/pkg/utils"
	"strconv"
	"strings"

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
	token, err := c.Cookie("token")
	if err == nil && token != "" {
		claims, parseErr := utils.ValidateJWT(token)
		if parseErr == nil {
			switch claims.Role {
			case "instructor":
				c.Redirect(http.StatusFound, "/instructor/dashboard")
				return
			case "admin":
				c.Redirect(http.StatusFound, "/admin/dashboard")
				return
			default:
				c.Redirect(http.StatusFound, "/student/dashboard")
				return
			}
		}
	}

	data := gin.H{
		"title": "Login | OnLearn",
	}

	if err := c.Query("error"); err != "" {
		data["error"] = err
	}
	if reg := c.Query("registered"); reg == "true" {
		data["success"] = "Akun berhasil dibuat. Silakan login."
	}

	c.HTML(http.StatusOK, "auth/login.html", data)
}

func (h *WebHandler) LoginWeb(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	// Validasi Input
	if email == "" || password == "" {
		c.HTML(http.StatusOK, "auth/login.html", gin.H{
			"error": "Email dan password wajib diisi.",
			"email": email, // Keep input
		})
		return
	}

	// Call Usecase
	token, err := h.AuthUsecase.Login(c.Request.Context(), email, password)
	if err != nil {
		c.HTML(http.StatusOK, "auth/login.html", gin.H{
			"error": "Email atau password salah.",
			"email": email,
		})
		return
	}

	// Parse token to get user role
	claims, err := utils.ValidateJWT(token)
	if err != nil {
		c.HTML(http.StatusOK, "auth/login.html", gin.H{
			"error": "Gagal memproses token.",
			"email": email,
		})
		return
	}

	// Set Cookie (MaxAge 24 jam)
	c.SetCookie("token", token, 86400, "/", "", false, false)

	// Redirect based on role
	switch claims.Role {
	case "instructor":
		c.Redirect(http.StatusFound, "/instructor/dashboard")
	case "admin":
		c.Redirect(http.StatusFound, "/admin/dashboard")
	default:
		c.Redirect(http.StatusFound, "/student/dashboard")
	}
}

func (h *WebHandler) ShowRegisterPage(c *gin.Context) {
	if _, err := c.Cookie("token"); err == nil {
		c.Redirect(http.StatusFound, "/student/dashboard")
		return
	}
	c.HTML(http.StatusOK, "auth/register.html", gin.H{"title": "Register | OnLearn"})
}

func (h *WebHandler) RegisterWeb(c *gin.Context) {
	name := c.PostForm("name")
	email := c.PostForm("email")
	password := c.PostForm("password")

	if name == "" || email == "" || password == "" {
		c.HTML(http.StatusOK, "auth/register.html", gin.H{
			"error": "Semua kolom wajib diisi.",
			"title": "Register | OnLearn",
			"name":  name,
			"email": email,
		})
		return
	}

	user := &domain.User{
		Name:     name,
		Email:    email,
		Password: password,
		Role:     domain.RoleStudent,
	}

	// Call Usecase
	if err := h.AuthUsecase.Register(c.Request.Context(), user); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "duplicate") {
			errMsg = "Email sudah terdaftar."
		}
		c.HTML(http.StatusOK, "auth/register.html", gin.H{
			"error": errMsg,
			"title": "Register | OnLearn",
			"name":  name,
			"email": email,
		})
		return
	}

	c.Redirect(http.StatusFound, "/?registered=true")
}

func (h *WebHandler) LogoutWeb(c *gin.Context) {
	// Hapus cookie
	c.SetCookie("token", "", -1, "/", "", false, false)
	c.Redirect(http.StatusFound, "/")
}

// ========== STUDENT HANDLERS ==========

func (h *WebHandler) StudentDashboard(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.Redirect(http.StatusFound, "/?error=Unauthorized")
		return
	}
	userID := userIDVal.(uint)

	// Get dashboard data
	dashboardData, err := h.DashboardUsecase.GetStudentDashboard(c.Request.Context(), userID)
	if err != nil {
		c.Redirect(http.StatusFound, "/?error=Gagal memuat dashboard")
		return
	}
	data := struct {
		*domain.StudentDashboardData
		ActiveMenu string
		Title      string
		PageTitle  string
	}{
		StudentDashboardData: dashboardData,
		ActiveMenu:           "dashboard",
		Title:                "Dashboard",
		PageTitle:            "Dashboard",
	}

	c.HTML(http.StatusOK, "student/dashboard.html", data)
}

func (h *WebHandler) StudentCourses(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.Redirect(http.StatusFound, "/?error=Unauthorized")
		return
	}
	userID := userIDVal.(uint)
	dashboardData, err := h.DashboardUsecase.GetStudentDashboard(c.Request.Context(), userID)
	if err != nil {
		c.Redirect(http.StatusFound, "/?error=Gagal memuat data user")
		return
	}
	user := dashboardData.User

	enrollments, err := h.CourseUsecase.GetStudentEnrollments(c.Request.Context(), userID)
	if err != nil {
		enrollments = []domain.EnrollmentWithCourse{}
	}

	var completed, inProgress int
	var totalProgress float64

	for _, e := range enrollments {
		totalProgress += e.Progress
		if e.Progress == 100 {
			completed++
		} else {
			inProgress++
		}
	}

	total := len(enrollments)
	var avgProgress float64
	if total > 0 {
		avgProgress = totalProgress / float64(total)
	}
	data := gin.H{
		"User":        user,
		"Enrollments": enrollments,
		"Stats": gin.H{
			"Total":       total,
			"Completed":   completed,
			"InProgress":  inProgress,
			"AvgProgress": int(avgProgress),
		},
		"ActiveMenu": "courses",
		"Title":      "Jalur Pembelajaran",
		"PageTitle":  "Jalur Pembelajaran",
	}

	c.HTML(http.StatusOK, "student/courses.html", data)
}

func (h *WebHandler) StudentCourseDetail(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.Redirect(http.StatusFound, "/?error=Unauthorized")
		return
	}
	userID := userIDVal.(uint)

	// Get course ID from URL
	courseIDStr := c.Param("id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 32)
	if err != nil {
		c.Redirect(http.StatusFound, "/student/courses?error=Invalid course ID")
		return
	}

	// Get user data
	user, err := h.AuthUsecase.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.Redirect(http.StatusFound, "/?error=Unauthorized")
		return
	}

	// Get course details with modules
	courseDetail, err := h.CourseUsecase.GetCourseDetails(c.Request.Context(), uint(courseID), &userID)
	if err != nil {
		c.Redirect(http.StatusFound, "/student/courses?error=Course not found")
		return
	}

	// Get modules with progress
	modulesWithProgress, err := h.CourseUsecase.GetModulesWithProgress(c.Request.Context(), userID, uint(courseID))
	if err != nil {
		modulesWithProgress = []domain.ModuleWithProgress{}
	}

	// Calculate progress
	var completedModules int
	for _, m := range modulesWithProgress {
		if m.IsComplete {
			completedModules++
		}
	}

	totalModules := len(modulesWithProgress)
	var progressPercent float64
	if totalModules > 0 {
		progressPercent = float64(completedModules) / float64(totalModules) * 100
	}

	data := gin.H{
		"User":             user,
		"Course":           courseDetail.Course,
		"Modules":          modulesWithProgress,
		"CompletedModules": completedModules,
		"TotalModules":     totalModules,
		"Progress":         int(progressPercent),
		"IsEnrolled":       courseDetail.IsEnrolled,
		"ActiveMenu":       "courses",
		"Title":            courseDetail.Course.Title,
		"PageTitle":        courseDetail.Course.Title,
	}

	c.HTML(http.StatusOK, "student/course_detail.html", data)
}

func (h *WebHandler) StudentModuleViewer(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.Redirect(http.StatusFound, "/?error=Unauthorized")
		return
	}
	userID := userIDVal.(uint)

	// Get course ID and module ID from URL
	courseIDStr := c.Param("id")
	moduleID := c.Param("module_id")

	courseID, err := strconv.ParseUint(courseIDStr, 10, 32)
	if err != nil {
		c.Redirect(http.StatusFound, "/student/courses?error=Invalid course ID")
		return
	}

	// Get user data
	user, err := h.AuthUsecase.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.Redirect(http.StatusFound, "/?error=Unauthorized")
		return
	}

	// Get course details
	courseDetail, err := h.CourseUsecase.GetCourseDetails(c.Request.Context(), uint(courseID), &userID)
	if err != nil {
		c.Redirect(http.StatusFound, "/student/courses?error=Course not found")
		return
	}

	// Verify user is enrolled
	if !courseDetail.IsEnrolled {
		c.Redirect(http.StatusFound, "/student/courses/"+courseIDStr+"?error=You are not enrolled in this course")
		return
	}

	// Get module by ID
	module, err := h.CourseUsecase.GetModuleByID(c.Request.Context(), moduleID)
	if err != nil {
		c.Redirect(http.StatusFound, "/student/courses/"+courseIDStr+"?error=Module not found")
		return
	}

	// Get module progress to check if complete
	modulesWithProgress, err := h.CourseUsecase.GetModulesWithProgress(c.Request.Context(), userID, uint(courseID))
	var isComplete bool
	for _, m := range modulesWithProgress {
		if m.ID == moduleID {
			isComplete = m.IsComplete
			break
		}
	}

	// Determine file URL (support both GridFS and legacy file paths)
	var fileURL string
	if module.FileID != "" {
		// File stored in GridFS - use protected endpoint
		fileURL = "/api/v1/files/" + module.FileID + "/stream"
	} else if module.ContentURL != "" {
		// Legacy file path
		fileURL = module.ContentURL
	}

	data := gin.H{
		"User":       user,
		"Course":     courseDetail.Course,
		"Module":     module,
		"FileURL":    fileURL,
		"IsComplete": isComplete,
		"ActiveMenu": "courses",
		"Title":      module.Title,
		"PageTitle":  module.Title,
	}

	c.HTML(http.StatusOK, "student/module_viewer.html", data)
}

// ========== INSTRUCTOR HANDLERS ==========

func (h *WebHandler) InstructorDashboard(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.Redirect(http.StatusFound, "/?error=Unauthorized")
		return
	}
	userID := userIDVal.(uint)

	// Get user data
	user, err := h.AuthUsecase.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.Redirect(http.StatusFound, "/?error=Gagal memuat data user")
		return
	}

	dashboardData, err := h.DashboardUsecase.GetInstructorDashboard(c.Request.Context(), userID)
	if err != nil {
		c.HTML(http.StatusOK, "instructor/dashboard.html", gin.H{"error": err.Error(), "User": user})
		return
	}

	data := gin.H{
		"User":                user,
		"TotalCourses":        dashboardData.TotalCourses,
		"TotalStudents":       dashboardData.TotalStudents,
		"PendingGrades":       dashboardData.PendingGrades,
		"PendingCertificates": dashboardData.PendingCertificates,
		"RecentSubmissions":   dashboardData.RecentSubmissions,
		"UngradedLabs":        dashboardData.UngradedLabs,
		"Title":               "Dashboard",
		"PageTitle":           "Dashboard",
		"ActiveMenu":          "dashboard",
	}

	c.HTML(http.StatusOK, "instructor/dashboard.html", data)
}

func (h *WebHandler) InstructorAllCourses(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.Redirect(http.StatusFound, "/?error=Unauthorized")
		return
	}
	userID := userIDVal.(uint)

	// Get user data
	user, err := h.AuthUsecase.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.Redirect(http.StatusFound, "/?error=Gagal memuat data user")
		return
	}

	// Get instructor's courses
	courses, err := h.CourseUsecase.GetInstructorCourses(c.Request.Context(), userID)
	if err != nil {
		courses = []domain.Course{}
	}

	data := gin.H{
		"User":       user,
		"Courses":    courses,
		"Title":      "Semua Kursus",
		"PageTitle":  "Semua Kursus",
		"ActiveMenu": "courses",
	}

	c.HTML(http.StatusOK, "instructor/all_courses.html", data)
}

func (h *WebHandler) InstructorCourseDetail(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.Redirect(http.StatusFound, "/?error=Unauthorized")
		return
	}
	userID := userIDVal.(uint)

	// Get user data
	user, err := h.AuthUsecase.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.Redirect(http.StatusFound, "/?error=Gagal memuat data user")
		return
	}

	// Get course ID
	courseIDStr := c.Param("id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 32)
	if err != nil {
		c.Redirect(http.StatusFound, "/instructor/courses?error=Invalid course ID")
		return
	}

	// Get course detail
	courseDetail, err := h.CourseUsecase.GetCourseDetails(c.Request.Context(), uint(courseID), nil)
	if err != nil {
		c.Redirect(http.StatusFound, "/instructor/courses?error=Course not found")
		return
	}

	// Verify ownership
	if courseDetail.InstructorID != userID {
		c.Redirect(http.StatusFound, "/instructor/courses?error=Unauthorized")
		return
	}

	// Get enrolled students
	students, _ := h.CourseUsecase.GetCourseStudents(c.Request.Context(), uint(courseID))

	// CourseDetail embeds Course struct, so we can access Course fields directly
	// In Go templates, embedded fields are accessible directly
	data := gin.H{
		"User":             user,
		"Course":           courseDetail, // CourseDetail embeds Course, template can access .Course.IsPublished
		"Modules":          courseDetail.Modules,
		"Students":         students,
		"EnrolledStudents": courseDetail.EnrolledStudents,
		"Title":            courseDetail.Title,
		"PageTitle":        courseDetail.Title,
		"ActiveMenu":       "courses",
	}

	c.HTML(http.StatusOK, "instructor/course_detail.html", data)
}

func (h *WebHandler) InstructorLabs(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.Redirect(http.StatusFound, "/?error=Unauthorized")
		return
	}
	userID := userIDVal.(uint)

	// Get user data
	user, err := h.AuthUsecase.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.Redirect(http.StatusFound, "/?error=Gagal memuat data user")
		return
	}

	// Get all labs
	labs, err := h.LabUsecase.GetAllLabs(c.Request.Context())
	if err != nil {
		labs = []domain.Lab{}
	}

	// Get ungraded count for each lab
	type LabWithUngraded struct {
		domain.Lab
		UngradedCount int64
	}
	var labsWithUngraded []LabWithUngraded
	for _, lab := range labs {
		ungradedCount, _ := h.LabUsecase.GetUngradedCountByLabID(c.Request.Context(), lab.ID)
		labsWithUngraded = append(labsWithUngraded, LabWithUngraded{
			Lab:           lab,
			UngradedCount: ungradedCount,
		})
	}

	data := gin.H{
		"User":       user,
		"Labs":       labsWithUngraded,
		"Title":      "Lab Management",
		"PageTitle":  "Lab Management",
		"ActiveMenu": "labs",
	}

	c.HTML(http.StatusOK, "instructor/labs.html", data)
}

func (h *WebHandler) InstructorCertificates(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.Redirect(http.StatusFound, "/?error=Unauthorized")
		return
	}
	userID := userIDVal.(uint)

	// Get user data
	user, err := h.AuthUsecase.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.Redirect(http.StatusFound, "/?error=Gagal memuat data user")
		return
	}

	// Get pending certificates
	pendingCerts, err := h.CertUsecase.GetPendingCertificates(c.Request.Context())
	if err != nil {
		pendingCerts = []domain.Certificate{}
	}

	// Get instructor's courses to filter certificates
	courses, _ := h.CourseUsecase.GetInstructorCourses(c.Request.Context(), userID)
	courseIDs := make(map[uint]bool)
	for _, course := range courses {
		courseIDs[course.ID] = true
	}

	// Filter only certificates for instructor's courses
	var filteredPending []domain.Certificate
	for _, cert := range pendingCerts {
		if cert.CourseID != nil && courseIDs[*cert.CourseID] {
			filteredPending = append(filteredPending, cert)
		} else if cert.LabID != nil {
			// Lab certificates - include all for now
			filteredPending = append(filteredPending, cert)
		}
	}

	data := gin.H{
		"User":         user,
		"PendingCerts": filteredPending,
		"Title":        "Sertifikat Management",
		"PageTitle":    "Sertifikat Management",
		"ActiveMenu":   "certificates",
	}

	c.HTML(http.StatusOK, "instructor/certificates.html", data)
}

func (h *WebHandler) InstructorStudents(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.Redirect(http.StatusFound, "/?error=Unauthorized")
		return
	}
	userID := userIDVal.(uint)

	// Get user data
	user, err := h.AuthUsecase.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.Redirect(http.StatusFound, "/?error=Gagal memuat data user")
		return
	}

	// Get instructor's courses
	courses, err := h.CourseUsecase.GetInstructorCourses(c.Request.Context(), userID)
	if err != nil {
		courses = []domain.Course{}
	}

	// Get all students from all courses
	type StudentWithCourses struct {
		domain.User
		EnrolledCourses []domain.Course
		TotalProgress   float64
		PendingTasks    int
	}

	allStudentsMap := make(map[uint]*StudentWithCourses)

	for _, course := range courses {
		students, _ := h.CourseUsecase.GetCourseStudents(c.Request.Context(), course.ID)
		for _, student := range students {
			if _, exists := allStudentsMap[student.ID]; !exists {
				allStudentsMap[student.ID] = &StudentWithCourses{
					User:            student,
					EnrolledCourses: []domain.Course{},
					TotalProgress:   0,
					PendingTasks:    0,
				}
			}
			allStudentsMap[student.ID].EnrolledCourses = append(allStudentsMap[student.ID].EnrolledCourses, course)
		}
	}

	// Calculate progress and pending tasks for each student
	for studentID, studentData := range allStudentsMap {
		enrollments, _ := h.CourseUsecase.GetStudentEnrollments(c.Request.Context(), studentID)
		var totalProgress float64
		for _, e := range enrollments {
			totalProgress += e.Progress
		}
		if len(enrollments) > 0 {
			studentData.TotalProgress = totalProgress / float64(len(enrollments))
		}

		// Count pending assignments (ungraded)
		// This would require additional API call, for now set to 0
		studentData.PendingTasks = 0
	}

	var allStudents []StudentWithCourses
	for _, student := range allStudentsMap {
		allStudents = append(allStudents, *student)
	}

	data := gin.H{
		"User":       user,
		"Students":   allStudents,
		"Courses":    courses,
		"Title":      "Daftar Student",
		"PageTitle":  "Daftar Student",
		"ActiveMenu": "students",
	}

	c.HTML(http.StatusOK, "instructor/students.html", data)
}

// ========== ADMIN HANDLERS ==========

func (h *WebHandler) AdminDashboard(c *gin.Context) {
	dashboardData, err := h.DashboardUsecase.GetAdminDashboard(c.Request.Context())
	if err != nil {
		c.HTML(http.StatusOK, "admin_dashboard.html", gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "admin_dashboard.html", dashboardData)
}
func (h *WebHandler) StudentBrowseCourses(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.Redirect(http.StatusFound, "/?error=Unauthorized")
		return
	}
	userID := userIDVal.(uint)

	dashboardData, err := h.DashboardUsecase.GetStudentDashboard(c.Request.Context(), userID)
	if err != nil {
		c.Redirect(http.StatusFound, "/?error=Gagal memuat data user")
		return
	}

	// Get all published courses (GetAllCourses already filters published)
	allCourses, err := h.CourseUsecase.GetAllCourses(c.Request.Context())
	if err != nil {
		allCourses = []domain.Course{}
	}

	// Get user's enrollments to filter out enrolled courses
	enrollments, _ := h.CourseUsecase.GetStudentEnrollments(c.Request.Context(), userID)
	enrolledCourseIDs := make(map[uint]bool)
	for _, e := range enrollments {
		enrolledCourseIDs[e.CourseID] = true
	}

	// Filter out courses that user already enrolled
	var availableCourses []domain.Course
	for _, course := range allCourses {
		if !enrolledCourseIDs[course.ID] {
			availableCourses = append(availableCourses, course)
		}
	}

	data := gin.H{
		"User":       dashboardData.User,
		"Courses":    availableCourses,
		"ActiveMenu": "browse",
		"Title":      "Browse Kursus",
		"PageTitle":  "Browse Kursus",
	}

	c.HTML(http.StatusOK, "student/browse_courses.html", data)
}

func (h *WebHandler) StudentLabs(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.Redirect(http.StatusFound, "/?error=Unauthorized")
		return
	}
	userID := userIDVal.(uint)

	dashboardData, err := h.DashboardUsecase.GetStudentDashboard(c.Request.Context(), userID)
	if err != nil {
		c.Redirect(http.StatusFound, "/?error=Gagal memuat data user")
		return
	}

	labs, err := h.LabUsecase.GetAllLabs(c.Request.Context())
	if err != nil {
		labs = []domain.Lab{}
	}

	data := gin.H{
		"User":       dashboardData.User,
		"Labs":       labs,
		"ActiveMenu": "labs",
		"Title":      "Lab Praktikum",
		"PageTitle":  "Lab Praktikum",
	}

	c.HTML(http.StatusOK, "student/labs.html", data)
}

func (h *WebHandler) StudentCertificates(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.Redirect(http.StatusFound, "/?error=Unauthorized")
		return
	}
	userID := userIDVal.(uint)

	// 1. Ambil Data User (untuk Sidebar)
	dashboardData, err := h.DashboardUsecase.GetStudentDashboard(c.Request.Context(), userID)
	if err != nil {
		c.Redirect(http.StatusFound, "/?error=Gagal memuat data user")
		return
	}

	// 2. Ambil Daftar Sertifikat User
	certs, err := h.CertUsecase.GetUserCertificates(c.Request.Context(), userID)
	if err != nil {
		certs = []domain.Certificate{}
	}

	data := gin.H{
		"User":         dashboardData.User,
		"Certificates": certs,
		"ActiveMenu":   "certificates",
		"Title":        "Sertifikat Saya",
		"PageTitle":    "Sertifikat Saya",
	}

	c.HTML(http.StatusOK, "student/certificates.html", data)
}

func (h *WebHandler) StudentProfile(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.Redirect(http.StatusFound, "/?error=Unauthorized")
		return
	}
	userID := userIDVal.(uint)

	// 1. Ambil Data User
	dashboardData, err := h.DashboardUsecase.GetStudentDashboard(c.Request.Context(), userID)
	if err != nil {
		c.Redirect(http.StatusFound, "/?error=Gagal memuat data user")
		return
	}

	// 2. Ambil Enrollments (Learning Paths)
	enrollments, err := h.CourseUsecase.GetStudentEnrollments(c.Request.Context(), userID)
	if err != nil {
		enrollments = []domain.EnrollmentWithCourse{}
	}

	// 3. Ambil Completed Labs (labs dengan grade yang sudah diisi)
	completedLabGrades, _ := h.LabUsecase.GetCompletedLabsByUserID(c.Request.Context(), userID)

	data := gin.H{
		"User":          dashboardData.User,
		"Enrollments":   enrollments,
		"CompletedLabs": completedLabGrades,
		"Title":         "Profile",
		"PageTitle":     "Profile",
		"ActiveMenu":    "profile",
	}

	c.HTML(http.StatusOK, "student/profile.html", data)
}

func (h *WebHandler) StudentProfileEdit(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.Redirect(http.StatusFound, "/?error=Unauthorized")
		return
	}
	userID := userIDVal.(uint)

	// Get user data
	dashboardData, err := h.DashboardUsecase.GetStudentDashboard(c.Request.Context(), userID)
	if err != nil {
		c.Redirect(http.StatusFound, "/?error=Gagal memuat data user")
		return
	}

	data := gin.H{
		"User":       dashboardData.User,
		"Title":      "Edit Profile",
		"PageTitle":  "Edit Profile",
		"ActiveMenu": "profile",
	}

	if c.Request.Method == "POST" {
		// Handle form submission
		name := c.PostForm("name")
		password := c.PostForm("password")

		user := &domain.User{
			ID:       userID,
			Name:     name,
			Password: password,
		}

		// Handle profile picture upload
		filePath, err := utils.HandleUpload(c, "profile_picture")
		if err == nil && filePath != "" {
			user.ProfilePicture = filePath
		}

		if err := h.AuthUsecase.UpdateUser(c.Request.Context(), user); err != nil {
			data["error"] = err.Error()
			c.HTML(http.StatusOK, "student/profile_edit.html", data)
			return
		}

		c.Redirect(http.StatusFound, "/student/profile?success=Profile updated successfully")
		return
	}

	c.HTML(http.StatusOK, "student/profile_edit.html", data)
}
