package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uint) (*User, error)
	Update(ctx context.Context, user *User) error
	UpdateVerified(ctx context.Context, email string) error
	GetByIDs(ctx context.Context, ids []uint) ([]User, error)
}

type CourseRepository interface {
	Create(ctx context.Context, course *Course) error
	GetAll(ctx context.Context) ([]Course, error)
	GetByID(ctx context.Context, id uint) (*Course, error)
	Update(ctx context.Context, course *Course) error
}

type ModuleRepository interface { // MongoDB
	Create(ctx context.Context, module *Module) error
	GetByCourseID(ctx context.Context, courseID uint) ([]Module, error)
}

type LabRepository interface {
	Create(ctx context.Context, lab *Lab) error
	Update(ctx context.Context, lab *Lab) error
	GetByID(ctx context.Context, id uint) (*Lab, error)
	GetAll(ctx context.Context) ([]Lab, error)
	CreateGrade(ctx context.Context, grade *LabGrade) error
	UpdateGrade(ctx context.Context, grade *LabGrade) error
	GetGrade(ctx context.Context, userID, labID uint) (*LabGrade, error)
	GetGradesByLabID(ctx context.Context, labID uint) ([]LabGrade, error)
}

type CertificateRepository interface {
	Create(ctx context.Context, cert *Certificate) error
	GetByUserID(ctx context.Context, userID uint) ([]Certificate, error)
}

type AuthUsecase interface {
	Register(ctx context.Context, user *User) error
	Login(ctx context.Context, email, password string) (string, error)
	UpdateUser(ctx context.Context, user *User) error
	VerifyEmail(ctx context.Context, email string, token string) error
	ForgotPassword(ctx context.Context, email string) error
}

type CourseUsecase interface {
	CreateCourse(ctx context.Context, course *Course) error
	AddModule(ctx context.Context, module *Module) error
	GetCourseDetails(ctx context.Context, courseID uint) (*Course, []Module, error)
	GetAllCourses(ctx context.Context) ([]Course, error)
}

type LabUsecase interface {
	CreateLab(ctx context.Context, lab *Lab) error
	UpdateLabStatus(ctx context.Context, labID uint, status string) error
	StudentEnroll(ctx context.Context, userID, labID uint) error
	SubmitGrade(ctx context.Context, instructorID, userID, labID uint, grade string) error
	GetUngradedStudents(ctx context.Context, labID uint) ([]User, error)
}

type CertificateUsecase interface {
	GenerateCertificate(ctx context.Context, userID uint, courseID *uint, labID *uint, title string) (*Certificate, error)
	GetUserCertificates(ctx context.Context, userID uint) ([]Certificate, error)
}
