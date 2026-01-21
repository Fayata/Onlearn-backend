package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- PostgreSQL Models (User, Lab, Course Metadata) ---

type Role string

const (
	RoleStudent    Role = "student"
	RoleInstructor Role = "instructor"
	RoleAdmin      Role = "admin"
)

type User struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	Name             string    `json:"name"`
	Email            string    `gorm:"uniqueIndex" json:"email"`
	Password         string    `json:"-"` // Hide password in JSON
	Role             Role      `json:"role"`
	IsVerified       bool      `json:"is_verified" gorm:"default:false"`
	VerificationCode string    `json:"-"`
	Avatar           string    `json:"avatar"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Course struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	InstructorID uint      `json:"instructor_id"`
	Thumbnail    string    `json:"thumbnail"`
	ModulesRefID string    `json:"modules_ref_id"` // ID referensi ke MongoDB (opsional, bisa query by course_id)
	CreatedAt    time.Time `json:"created_at"`
	Instructor   User      `gorm:"foreignKey:InstructorID" json:"instructor,omitempty"`
}

type Enrollment struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	StudentID   uint       `json:"student_id"`
	CourseID    uint       `json:"course_id"`
	Progress    float64    `json:"progress" gorm:"default:0"`       // 0-100%
	Status      string     `json:"status" gorm:"default:'ongoing'"` // ongoing, completed
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	Course      Course     `gorm:"foreignKey:CourseID" json:"course"`
}

type Lab struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Status       string    `json:"status"`
	InstructorID uint      `json:"instructor_id"`
	CreatedAt    time.Time `json:"created_at"`
}

type LabSubmission struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	LabID     uint      `json:"lab_id"`
	StudentID uint      `json:"student_id"`
	Grade     *float64  `json:"grade"` 
	Feedback  string    `json:"feedback"`
	Attended  bool      `json:"attended" gorm:"default:false"`
	UpdatedAt time.Time `json:"updated_at"`
	Student   User      `gorm:"foreignKey:StudentID" json:"student"`
}

type Certificate struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	UserID   uint      `json:"user_id"`
	CourseID *uint     `json:"course_id"`
	LabID    *uint     `json:"lab_id"`
	Code     string    `json:"code" gorm:"unique"`
	IssuedAt time.Time `json:"issued_at"`
	Status   string    `json:"status" gorm:"default:'pending'"` 
}

// --- MongoDB Models (Course Modules & Content) ---

type ModuleType string

const (
	TypePDF   ModuleType = "pdf"
	TypePPT   ModuleType = "ppt"
	TypeVideo ModuleType = "video"
)

// Module disimpan di MongoDB karena strukturnya dinamis & hirarkis
type Module struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CourseID       uint               `bson:"course_id" json:"course_id"` // Relasi ke PG Course ID
	Title          string             `bson:"title" json:"title"`
	Order          int                `bson:"order" json:"order"`
	Type           ModuleType         `bson:"type" json:"type"`
	ContentURL     string             `bson:"content_url" json:"content_url"`                   // Link ke file PDF/PPT
	AssignmentLink string             `bson:"assignment_link,omitempty" json:"assignment_link"` // GForm Link
	IsCompleted    bool               `bson:"-" json:"is_completed"`                            // Field virtual untuk response API student
}

// --- API Request/Response Structs ---

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     Role   `json:"role" binding:"required"` // Sebaiknya dibatasi di BE, tapi utk demo kita buka
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type CreateCourseRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

type CreateModuleRequest struct {
	CourseID       uint       `json:"course_id" binding:"required"`
	Title          string     `json:"title" binding:"required"`
	Type           ModuleType `json:"type" binding:"required"`
	ContentURL     string     `json:"content_url"`
	AssignmentLink string     `json:"assignment_link"`
	Order          int        `json:"order"`
}
