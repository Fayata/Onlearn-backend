package domain

import (
	"time"
)

type Role string

const (
	RoleStudent    Role = "student"
	RoleInstructor Role = "instructor"
	RoleAdmin      Role = "admin"
)

type User struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	Name           string    `json:"name" gorm:"not null"`
	Email          string    `json:"email" gorm:"uniqueIndex;not null"`
	Password       string    `json:"-" gorm:"not null"`
	Role           Role      `json:"role" gorm:"type:varchar(20);default:'student'"`
	IsVerified     bool      `json:"is_verified" gorm:"default:false"`
	ProfilePicture string    `json:"profile_picture"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type Course struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Title        string    `json:"title" gorm:"not null"`
	Description  string    `json:"description" gorm:"type:text"`
	Thumbnail    string    `json:"thumbnail"`
	InstructorID uint      `json:"instructor_id" gorm:"not null"`
	IsPublished  bool      `json:"is_published" gorm:"default:false"` // Kursus tidak langsung tampil ke student
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	Instructor User `json:"instructor,omitempty" gorm:"foreignKey:InstructorID"`
}

type Lab struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description" gorm:"type:text"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Status      string    `json:"status" gorm:"type:varchar(20);default:'scheduled'"` // scheduled, open, closed
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Enrollment - Student mendaftar ke Course
type Enrollment struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"not null;index"`
	CourseID   uint      `json:"course_id" gorm:"not null;index"`
	Progress   float64   `json:"progress" gorm:"default:0"` // 0-100%
	IsFinished bool      `json:"is_finished" gorm:"default:false"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	User   User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Course Course `json:"course,omitempty" gorm:"foreignKey:CourseID"`
}

// ModuleProgress - Track progress student per module
type ModuleProgress struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"not null;index"`
	ModuleID   string    `json:"module_id" gorm:"not null;index"` // MongoDB ObjectID
	CourseID   uint      `json:"course_id" gorm:"not null;index"`
	IsComplete bool      `json:"is_complete" gorm:"default:false"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Assignment - Tugas yang di-upload student (untuk modul dengan Google Form)
type Assignment struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	UserID      uint       `json:"user_id" gorm:"not null;index"`
	ModuleID    string     `json:"module_id" gorm:"not null;index"` // MongoDB ObjectID
	CourseID    uint       `json:"course_id" gorm:"not null;index"`
	FileURL     string     `json:"file_url"` // Upload file tugas
	SubmittedAt time.Time  `json:"submitted_at" gorm:"autoCreateTime"`
	Grade       *float64   `json:"grade"` // nullable, dinilai oleh instructor
	Feedback    string     `json:"feedback" gorm:"type:text"`
	GradedAt    *time.Time `json:"graded_at"`
	GradedByID  *uint      `json:"graded_by_id"` // Instructor ID

	// Relations
	User     User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	GradedBy *User `json:"graded_by,omitempty" gorm:"foreignKey:GradedByID"`
}

// LabGrade - Nilai lab untuk student
type LabGrade struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	LabID     uint      `json:"lab_id" gorm:"not null;index"`
	Grade     string    `json:"grade" gorm:"type:varchar(10)"` // A, B, C, etc atau angka
	Feedback  string    `json:"feedback" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Lab  Lab  `json:"lab,omitempty" gorm:"foreignKey:LabID"`
}

// Certificate - Sertifikat untuk student
type Certificate struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	UserID     uint       `json:"user_id" gorm:"not null;index"`
	CourseID   *uint      `json:"course_id,omitempty" gorm:"index"`
	LabID      *uint      `json:"lab_id,omitempty" gorm:"index"`
	Title      string     `json:"title" gorm:"not null"`
	URL        string     `json:"url" gorm:"not null"`                              // Path ke file PDF sertifikat
	Status     string     `json:"status" gorm:"type:varchar(20);default:'pending'"` // pending, approved, rejected
	ApprovedBy *uint      `json:"approved_by" gorm:"index"`                         // Instructor/Admin ID
	ApprovedAt *time.Time `json:"approved_at"`
	IssueDate  time.Time  `json:"issue_date" gorm:"autoCreateTime"`

	// Relations
	User     User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Course   *Course `json:"course,omitempty" gorm:"foreignKey:CourseID"`
	Lab      *Lab    `json:"lab,omitempty" gorm:"foreignKey:LabID"`
	Approver *User   `json:"approver,omitempty" gorm:"foreignKey:ApprovedBy"`
}

// ========== MONGODB MODELS ==========

type ModuleType string

const (
	TypePDF ModuleType = "pdf"
	TypePPT ModuleType = "ppt"
)

// Module - Disimpan di MongoDB karena struktur dinamis
type Module struct {
	ID          string     `json:"id" bson:"_id,omitempty"`
	CourseID    uint       `json:"course_id" bson:"course_id"`
	Title       string     `json:"title" bson:"title"`
	Type        ModuleType `json:"type" bson:"type"`                               // pdf atau ppt
	ContentURL  string     `json:"content_url" bson:"content_url"`                 // Path ke file PDF/PPT (legacy)
	FileID      string     `json:"file_id,omitempty" bson:"file_id,omitempty"`     // GridFS File ID
	QuizLink    string     `json:"quiz_link,omitempty" bson:"quiz_link,omitempty"` // Google Form (optional)
	Description string     `json:"description" bson:"description"`
	Order       int        `json:"order" bson:"order"` // Urutan modul dalam course
	CreatedAt   time.Time  `json:"created_at" bson:"created_at"`
}

// ========== RESPONSE DTOs ==========

// StudentDashboardData - Data untuk dashboard student
type StudentDashboardData struct {
	User                *User                  `json:"user"`
	TotalEnrollments    int                    `json:"total_enrollments"`
	CompletedCourses    int                    `json:"completed_courses"`
	InProgressCourses   int                    `json:"in_progress_courses"`
	TotalCertificates   int                    `json:"total_certificates"`
	LeaderboardRank     int                    `json:"leaderboard_rank"`
	CompletedLabs       int                    `json:"completed_labs"`
	AssignedLabs        int                    `json:"assigned_labs"`
	WeeklyGoalProgress  int                    `json:"weekly_goal_progress"`
	RecentActivities    []RecentActivity       `json:"recent_activities"`
	RecentCertificates  []Certificate          `json:"recent_certificates"`
	OngoingEnrollments  []EnrollmentWithCourse `json:"ongoing_enrollments"`
	UpcomingLabs        []Lab                  `json:"upcoming_labs"`
	UpcomingLabsSorted  []Lab                  `json:"upcoming_labs_sorted"`
}

type ActivityType string

const (
	ActivityModuleCompleted ActivityType = "module_completed"
	ActivityTaskSubmitted   ActivityType = "task_submitted"
	ActivityCertificateNew  ActivityType = "certificate_new"
)

type RecentActivity struct {
	Type      ActivityType `json:"type"`
	Title     string       `json:"title"`
	Subtitle  string       `json:"subtitle"`
	Timestamp time.Time    `json:"timestamp"`
	Icon      string       `json:"icon"`
	BgColor   string       `json:"bg_color"`
	IconColor string       `json:"icon_color"`
}

// EnrollmentWithCourse - Enrollment dengan detail course
type EnrollmentWithCourse struct {
	Enrollment
	ModuleCount      int `json:"module_count"`
	CompletedModules int `json:"completed_modules"`
}

// InstructorDashboardData - Data untuk dashboard instructor
type InstructorDashboardData struct {
	TotalCourses        int                    `json:"total_courses"`
	TotalStudents       int                    `json:"total_students"`
	PendingGrades       int                    `json:"pending_grades"`
	PendingCertificates int                    `json:"pending_certificates"`
	RecentSubmissions   []Assignment           `json:"recent_submissions"`
	UngradedLabs        []LabWithUngradedCount `json:"ungraded_labs"`
}

// LabWithUngradedCount - Lab dengan jumlah student yang belum dinilai
type LabWithUngradedCount struct {
	Lab
	UngradedCount int `json:"ungraded_count"`
}

// AdminDashboardData - Data untuk dashboard admin
type AdminDashboardData struct {
	TotalUsers          int `json:"total_users"`
	TotalStudents       int `json:"total_students"`
	TotalInstructors    int `json:"total_instructors"`
	TotalCourses        int `json:"total_courses"`
	TotalLabs           int `json:"total_labs"`
	TotalCertificates   int `json:"total_certificates"`
	PendingCertificates int `json:"pending_certificates"`
}

// UngradedStudent - Student yang belum dinilai
type UngradedStudent struct {
	UserID uint   `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

// CourseDetail - Detail course dengan modules
type CourseDetail struct {
	Course
	Modules          []Module `json:"modules"`
	EnrolledStudents int      `json:"enrolled_students"`
	IsEnrolled       bool     `json:"is_enrolled"` // Untuk student
}

// ModuleWithProgress - Module dengan progress tracking untuk student
type ModuleWithProgress struct {
	Module
	IsComplete      bool     `json:"is_complete"`
	HasSubmission   bool     `json:"has_submission"`
	SubmissionGrade *float64 `json:"submission_grade,omitempty"`
}

// StudentPerformance - Performa student untuk laporan
type StudentPerformance struct {
	UserID            uint    `json:"user_id"`
	Name              string  `json:"name"`
	Email             string  `json:"email"`
	TotalEnrollments  int     `json:"total_enrollments"`
	CompletedCourses  int     `json:"completed_courses"`
	AverageProgress   float64 `json:"average_progress"`
	TotalAssignments  int     `json:"total_assignments"`
	GradedAssignments int     `json:"graded_assignments"`
	AverageGrade      float64 `json:"average_grade"`
	TotalCertificates int     `json:"total_certificates"`
}

// UserWithAssignment - Student dengan data assignment untuk penilaian modul
type UserWithAssignment struct {
	User      User       `json:"user"`
	Assignment *Assignment `json:"assignment,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}
