package domain

import "context"

// ========== REPOSITORIES ==========

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uint) (*User, error)
	Update(ctx context.Context, user *User) error
	UpdateVerified(ctx context.Context, email string) error
	GetByIDs(ctx context.Context, ids []uint) ([]User, error)
	GetByRole(ctx context.Context, role Role) ([]User, error)
	GetAll(ctx context.Context) ([]User, error)
	Delete(ctx context.Context, id uint) error
	CountByRole(ctx context.Context, role Role) (int64, error)
}

type CourseRepository interface {
	Create(ctx context.Context, course *Course) error
	GetAll(ctx context.Context) ([]Course, error)
	GetByID(ctx context.Context, id uint) (*Course, error)
	GetByInstructorID(ctx context.Context, instructorID uint) ([]Course, error)
	Update(ctx context.Context, course *Course) error
	Delete(ctx context.Context, id uint) error
	Count(ctx context.Context) (int64, error)
}

type ModuleRepository interface {
	Create(ctx context.Context, module *Module) error
	GetByCourseID(ctx context.Context, courseID uint) ([]Module, error)
	GetByID(ctx context.Context, id string) (*Module, error)
	Update(ctx context.Context, module *Module) error
	Delete(ctx context.Context, id string) error
}

type EnrollmentRepository interface {
	Create(ctx context.Context, enrollment *Enrollment) error
	GetByUserAndCourse(ctx context.Context, userID, courseID uint) (*Enrollment, error)
	GetByUserID(ctx context.Context, userID uint) ([]Enrollment, error)
	GetByCourseID(ctx context.Context, courseID uint) ([]Enrollment, error)
	Update(ctx context.Context, enrollment *Enrollment) error
	CountByCourseID(ctx context.Context, courseID uint) (int64, error)
	CountByUserID(ctx context.Context, userID uint) (int64, error)
}

type ModuleProgressRepository interface {
	Create(ctx context.Context, progress *ModuleProgress) error
	GetByUserAndModule(ctx context.Context, userID uint, moduleID string) (*ModuleProgress, error)
	GetByUserAndCourse(ctx context.Context, userID uint, courseID uint) ([]ModuleProgress, error)
	GetRecentByUser(ctx context.Context, userID uint, limit int) ([]ModuleProgress, error)
	Update(ctx context.Context, progress *ModuleProgress) error
	CountCompletedByUserAndCourse(ctx context.Context, userID uint, courseID uint) (int64, error)
}

type AssignmentRepository interface {
	Create(ctx context.Context, assignment *Assignment) error
	GetByID(ctx context.Context, id uint) (*Assignment, error)
	GetByUserAndModule(ctx context.Context, userID uint, moduleID string) (*Assignment, error)
	GetByCourseID(ctx context.Context, courseID uint) ([]Assignment, error)
	GetUngradedByCourseID(ctx context.Context, courseID uint) ([]Assignment, error)
	GetRecentSubmissions(ctx context.Context, limit int) ([]Assignment, error)
	GetRecentSubmissionsByUserID(ctx context.Context, userID uint, limit int) ([]Assignment, error)
	Update(ctx context.Context, assignment *Assignment) error
	CountUngradedByInstructor(ctx context.Context, instructorID uint) (int64, error)
}

type LabRepository interface {
	Create(ctx context.Context, lab *Lab) error
	Update(ctx context.Context, lab *Lab) error
	GetByID(ctx context.Context, id uint) (*Lab, error)
	GetAll(ctx context.Context) ([]Lab, error)
	GetUpcoming(ctx context.Context) ([]Lab, error)
	Delete(ctx context.Context, id uint) error
	Count(ctx context.Context) (int64, error)

	// Lab Grades
	CreateGrade(ctx context.Context, grade *LabGrade) error
	UpdateGrade(ctx context.Context, grade *LabGrade) error
	GetGrade(ctx context.Context, userID, labID uint) (*LabGrade, error)
	GetGradesByUserID(ctx context.Context, userID uint) ([]LabGrade, error)
	GetGradesByLabID(ctx context.Context, labID uint) ([]LabGrade, error)
	CountUngradedByLabID(ctx context.Context, labID uint) (int64, error)
}

type CertificateRepository interface {
	Create(ctx context.Context, cert *Certificate) error
	GetByUserID(ctx context.Context, userID uint) ([]Certificate, error)
	GetByID(ctx context.Context, id uint) (*Certificate, error)
	GetPending(ctx context.Context) ([]Certificate, error)
	GetRecentByUserID(ctx context.Context, userID uint, limit int) ([]Certificate, error)
	Update(ctx context.Context, cert *Certificate) error
	Count(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
}

// ========== USECASES ==========

type AuthUsecase interface {
	Register(ctx context.Context, user *User) error
	Login(ctx context.Context, email, password string) (string, error)
	UpdateUser(ctx context.Context, user *User) error
	VerifyEmail(ctx context.Context, email string, token string) error
	ForgotPassword(ctx context.Context, email string) error
}

type UserUsecase interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, id uint) (*User, error)
	GetAllUsers(ctx context.Context) ([]User, error)
	GetUsersByRole(ctx context.Context, role Role) ([]User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id uint) error
}

type CourseUsecase interface {
	CreateCourse(ctx context.Context, course *Course) error
	AddModule(ctx context.Context, module *Module) error
	UpdateModule(ctx context.Context, module *Module) error
	DeleteModule(ctx context.Context, moduleID string) error
	GetCourseDetails(ctx context.Context, courseID uint, userID *uint) (*CourseDetail, error)
	GetAllCourses(ctx context.Context) ([]Course, error)
	UpdateCourse(ctx context.Context, course *Course) error
	DeleteCourse(ctx context.Context, id uint) error

	// Enrollment
	EnrollStudent(ctx context.Context, userID, courseID uint) error
	GetStudentEnrollments(ctx context.Context, userID uint) ([]EnrollmentWithCourse, error)
	GetCourseStudents(ctx context.Context, courseID uint) ([]User, error)

	// Module Progress
	MarkModuleComplete(ctx context.Context, userID uint, moduleID string, courseID uint) error
	GetModulesWithProgress(ctx context.Context, userID uint, courseID uint) ([]ModuleWithProgress, error)

	// Assignments
	SubmitAssignment(ctx context.Context, assignment *Assignment) error
	GradeAssignment(ctx context.Context, assignmentID uint, grade float64, feedback string, gradedByID uint) error
	GetCourseAssignments(ctx context.Context, courseID uint) ([]Assignment, error)
}

type LabUsecase interface {
	CreateLab(ctx context.Context, lab *Lab) error
	UpdateLab(ctx context.Context, lab *Lab) error
	UpdateLabStatus(ctx context.Context, labID uint, status string) error
	GetLabByID(ctx context.Context, labID uint) (*Lab, error)
	GetAllLabs(ctx context.Context) ([]Lab, error)
	GetUpcomingLabs(ctx context.Context) ([]Lab, error)
	DeleteLab(ctx context.Context, labID uint) error

	// Grading
	StudentEnroll(ctx context.Context, userID, labID uint) error
	SubmitGrade(ctx context.Context, instructorID, userID, labID uint, grade string, feedback string) error
	GetUngradedStudents(ctx context.Context, labID uint) ([]User, error)
	GetLabsWithUngradedCount(ctx context.Context) ([]LabWithUngradedCount, error)
}

type CertificateUsecase interface {
	GenerateCertificate(ctx context.Context, userID uint, courseID *uint, labID *uint, title string) (*Certificate, error)
	GetUserCertificates(ctx context.Context, userID uint) ([]Certificate, error)
	GetRecentCertificates(ctx context.Context, userID uint, limit int) ([]Certificate, error)
	GetPendingCertificates(ctx context.Context) ([]Certificate, error)
	ApproveCertificate(ctx context.Context, certID uint, approverID uint) error
	RejectCertificate(ctx context.Context, certID uint, approverID uint) error
}

type DashboardUsecase interface {
	GetStudentDashboard(ctx context.Context, userID uint) (*StudentDashboardData, error)
	GetInstructorDashboard(ctx context.Context, instructorID uint) (*InstructorDashboardData, error)
	GetAdminDashboard(ctx context.Context) (*AdminDashboardData, error)
}

type ReportUsecase interface {
	GetStudentPerformance(ctx context.Context, userID uint) (*StudentPerformance, error)
	GetAllStudentsPerformance(ctx context.Context) ([]StudentPerformance, error)
	GetCourseReport(ctx context.Context, courseID uint) (interface{}, error)
}
