package usecase

import (
	"context"
	"errors"
	"fmt"
	"onlearn-backend/internal/domain"
	"time"
)

// ========== CERTIFICATE USECASE ==========

type certificateUsecase struct {
	certRepo   domain.CertificateRepository
	userRepo   domain.UserRepository
	courseRepo domain.CourseRepository
	labRepo    domain.LabRepository
}

func NewCertificateUsecase(
	cr domain.CertificateRepository,
	ur domain.UserRepository,
	cour domain.CourseRepository,
	lr domain.LabRepository,
) domain.CertificateUsecase {
	return &certificateUsecase{
		certRepo:   cr,
		userRepo:   ur,
		courseRepo: cour,
		labRepo:    lr,
	}
}

func (uc *certificateUsecase) GenerateCertificate(ctx context.Context, userID uint, courseID *uint, labID *uint, title string) (*domain.Certificate, error) {
	// Verify user exists
	_, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Verify course or lab exists
	if courseID != nil {
		_, err := uc.courseRepo.GetByID(ctx, *courseID)
		if err != nil {
			return nil, errors.New("course not found")
		}
	}

	if labID != nil {
		_, err := uc.labRepo.GetByID(ctx, *labID)
		if err != nil {
			return nil, errors.New("lab not found")
		}
	}

	// In production, you would generate actual PDF here
	// For now, use dummy URL
	cert := &domain.Certificate{
		UserID:    userID,
		CourseID:  courseID,
		LabID:     labID,
		Title:     title,
		URL:       fmt.Sprintf("/uploads/certificates/%s-%d.pdf", title, time.Now().Unix()),
		Status:    "pending",
		IssueDate: time.Now(),
	}

	if err := uc.certRepo.Create(ctx, cert); err != nil {
		return nil, err
	}

	return cert, nil
}

func (uc *certificateUsecase) GetUserCertificates(ctx context.Context, userID uint) ([]domain.Certificate, error) {
	return uc.certRepo.GetByUserID(ctx, userID)
}

func (uc *certificateUsecase) GetRecentCertificates(ctx context.Context, userID uint, limit int) ([]domain.Certificate, error) {
	return uc.certRepo.GetRecentByUserID(ctx, userID, limit)
}

func (uc *certificateUsecase) GetPendingCertificates(ctx context.Context) ([]domain.Certificate, error) {
	return uc.certRepo.GetPending(ctx)
}

func (uc *certificateUsecase) ApproveCertificate(ctx context.Context, certID uint, approverID uint) error {
	cert, err := uc.certRepo.GetByID(ctx, certID)
	if err != nil {
		return err
	}

	// Verify approver
	approver, err := uc.userRepo.GetByID(ctx, approverID)
	if err != nil {
		return errors.New("approver not found")
	}

	if approver.Role != domain.RoleInstructor && approver.Role != domain.RoleAdmin {
		return errors.New("only instructors and admins can approve certificates")
	}

	cert.Status = "approved"
	cert.ApprovedBy = &approverID
	now := time.Now()
	cert.ApprovedAt = &now

	return uc.certRepo.Update(ctx, cert)
}

func (uc *certificateUsecase) RejectCertificate(ctx context.Context, certID uint, approverID uint) error {
	cert, err := uc.certRepo.GetByID(ctx, certID)
	if err != nil {
		return err
	}

	// Verify approver
	approver, err := uc.userRepo.GetByID(ctx, approverID)
	if err != nil {
		return errors.New("approver not found")
	}

	if approver.Role != domain.RoleInstructor && approver.Role != domain.RoleAdmin {
		return errors.New("only instructors and admins can reject certificates")
	}

	cert.Status = "rejected"
	cert.ApprovedBy = &approverID
	now := time.Now()
	cert.ApprovedAt = &now

	return uc.certRepo.Update(ctx, cert)
}

// ========== USER USECASE (for Admin CRUD) ==========

type userUsecase struct {
	userRepo domain.UserRepository
}

func NewUserUsecase(ur domain.UserRepository) domain.UserUsecase {
	return &userUsecase{
		userRepo: ur,
	}
}

func (uc *userUsecase) CreateUser(ctx context.Context, user *domain.User) error {
	// Check if email exists
	existing, _ := uc.userRepo.GetByEmail(ctx, user.Email)
	if existing != nil {
		return errors.New("email already exists")
	}

	return uc.userRepo.Create(ctx, user)
}

func (uc *userUsecase) GetUserByID(ctx context.Context, id uint) (*domain.User, error) {
	return uc.userRepo.GetByID(ctx, id)
}

func (uc *userUsecase) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	return uc.userRepo.GetAll(ctx)
}

func (uc *userUsecase) GetUsersByRole(ctx context.Context, role domain.Role) ([]domain.User, error) {
	return uc.userRepo.GetByRole(ctx, role)
}

func (uc *userUsecase) UpdateUser(ctx context.Context, user *domain.User) error {
	existing, err := uc.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return err
	}

	// Update allowed fields
	if user.Name != "" {
		existing.Name = user.Name
	}
	if user.Email != "" && user.Email != existing.Email {
		// Check if new email is taken
		emailCheck, _ := uc.userRepo.GetByEmail(ctx, user.Email)
		if emailCheck != nil && emailCheck.ID != user.ID {
			return errors.New("email already taken")
		}
		existing.Email = user.Email
	}
	if user.Role != "" {
		existing.Role = user.Role
	}
	if user.ProfilePicture != "" {
		existing.ProfilePicture = user.ProfilePicture
	}

	return uc.userRepo.Update(ctx, existing)
}

func (uc *userUsecase) DeleteUser(ctx context.Context, id uint) error {
	// In production, you might want to check if user has enrollments, etc.
	return uc.userRepo.Delete(ctx, id)
}

// ========== REPORT USECASE ==========

type reportUsecase struct {
	userRepo       domain.UserRepository
	enrollmentRepo domain.EnrollmentRepository
	assignmentRepo domain.AssignmentRepository
	certRepo       domain.CertificateRepository
}

func NewReportUsecase(
	ur domain.UserRepository,
	er domain.EnrollmentRepository,
	ar domain.AssignmentRepository,
	cr domain.CertificateRepository,
) domain.ReportUsecase {
	return &reportUsecase{
		userRepo:       ur,
		enrollmentRepo: er,
		assignmentRepo: ar,
		certRepo:       cr,
	}
}

func (uc *reportUsecase) GetStudentPerformance(ctx context.Context, userID uint) (*domain.StudentPerformance, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	enrollments, _ := uc.enrollmentRepo.GetByUserID(ctx, userID)

	completedCount := 0
	totalProgress := 0.0
	for _, e := range enrollments {
		if e.IsFinished {
			completedCount++
		}
		totalProgress += e.Progress
	}

	avgProgress := 0.0
	if len(enrollments) > 0 {
		avgProgress = totalProgress / float64(len(enrollments))
	}

	// Get assignment stats
	// This would need a custom repo method, but for now we'll use 0
	totalAssignments := 0
	gradedAssignments := 0
	avgGrade := 0.0

	certs, _ := uc.certRepo.GetByUserID(ctx, userID)

	return &domain.StudentPerformance{
		UserID:            user.ID,
		Name:              user.Name,
		Email:             user.Email,
		TotalEnrollments:  len(enrollments),
		CompletedCourses:  completedCount,
		AverageProgress:   avgProgress,
		TotalAssignments:  totalAssignments,
		GradedAssignments: gradedAssignments,
		AverageGrade:      avgGrade,
		TotalCertificates: len(certs),
	}, nil
}

func (uc *reportUsecase) GetAllStudentsPerformance(ctx context.Context) ([]domain.StudentPerformance, error) {
	students, err := uc.userRepo.GetByRole(ctx, domain.RoleStudent)
	if err != nil {
		return nil, err
	}

	var performances []domain.StudentPerformance
	for _, student := range students {
		perf, err := uc.GetStudentPerformance(ctx, student.ID)
		if err != nil {
			continue
		}
		performances = append(performances, *perf)
	}

	return performances, nil
}

func (uc *reportUsecase) GetCourseReport(ctx context.Context, courseID uint) (interface{}, error) {
	// This can be customized based on your needs
	enrollments, err := uc.enrollmentRepo.GetByCourseID(ctx, courseID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_students": len(enrollments),
		"completed_students": func() int {
			count := 0
			for _, e := range enrollments {
				if e.IsFinished {
					count++
				}
			}
			return count
		}(),
	}, nil
}
