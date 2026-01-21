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
	Description  string    `json:"description"`
	Thumbnail    string    `json:"thumbnail"`
	InstructorID uint      `json:"instructor_id" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type Lab struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" gorm:"not null"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status" gorm:"type:varchar(20);default:'scheduled'"` // scheduled, open, closed
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type Enrollment struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"not null"`
	CourseID   uint      `json:"course_id" gorm:"not null"`
	Progress   float64   `json:"progress" gorm:"default:0"`
	IsFinished bool      `json:"is_finished" gorm:"default:false"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type LabGrade struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	LabID     uint      `json:"lab_id" gorm:"not null"`
	Grade     string    `json:"grade" gorm:"type:varchar(10)"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type Certificate struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	CourseID  *uint     `json:"course_id,omitempty"`
	LabID     *uint     `json:"lab_id,omitempty"`
	Title     string    `json:"title" gorm:"not null"`
	URL       string    `json:"url" gorm:"not null"`
	IssueDate time.Time `json:"issue_date" gorm:"autoCreateTime"`
}

type ModuleType string

const (
	TypePDF   ModuleType = "pdf"
	TypePPT   ModuleType = "ppt"
	TypeVideo ModuleType = "video"
)

type Module struct {
	ID          string     `json:"id" bson:"_id,omitempty"`
	CourseID    uint       `json:"course_id" bson:"course_id"`
	Title       string     `json:"title" bson:"title"`
	Type        ModuleType `json:"type" bson:"type"`
	ContentURL  string     `json:"content_url" bson:"content_url"`
	QuizLink    string     `json:"quiz_link,omitempty" bson:"quiz_link,omitempty"`
	Description string     `json:"description" bson:"description"`
	Order       int        `json:"order" bson:"order"`
}

// Response DTOs
type StudentDashboard struct {
	Enrollments  []Enrollment  `json:"enrollments"`
	Certificates []Certificate `json:"certificates"`
}

type InstructorDashboard struct {
	Courses         []Course     `json:"courses"`
	Labs            []Lab        `json:"labs"`
	UngradedSummary map[uint]int `json:"ungraded_summary"`
}

type UngradedStudent struct {
	UserID uint   `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}
