package usecase

import (
	"context"
	"onlearn-backend/internal/domain"
)

type dashboardUsecase struct {
	userRepo       domain.UserRepository
	courseRepo     domain.CourseRepository
	enrollmentRepo domain.EnrollmentRepository
	moduleRepo     domain.ModuleRepository
	progressRepo   domain.ModuleProgressRepository
	assignmentRepo domain.AssignmentRepository
	labRepo        domain.LabRepository
	certRepo       domain.CertificateRepository
}

func NewDashboardUsecase(
	ur domain.UserRepository,
	cr domain.CourseRepository,
	er domain.EnrollmentRepository,
	mr domain.ModuleRepository,
	pr domain.ModuleProgressRepository,
	ar domain.AssignmentRepository,
	lr domain.LabRepository,
	certr domain.CertificateRepository,
) domain.DashboardUsecase {
	return &dashboardUsecase{
		userRepo:       ur,
		courseRepo:     cr,
		enrollmentRepo: er,
		moduleRepo:     mr,
		progressRepo:   pr,
		assignmentRepo: ar,
		labRepo:        lr,
		certRepo:       certr,
	}
}

func (uc *dashboardUsecase) GetStudentDashboard(ctx context.Context, userID uint) (*domain.StudentDashboardData, error) {
	// Get enrollments
	enrollments, err := uc.enrollmentRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Count completed courses
	completedCount := 0
	inProgressCount := 0
	var ongoingEnrollments []domain.EnrollmentWithCourse

	for _, e := range enrollments {
		if e.IsFinished {
			completedCount++
		} else {
			inProgressCount++

			// Get module count
			modules, _ := uc.moduleRepo.GetByCourseID(ctx, e.CourseID)
			completedModules, _ := uc.progressRepo.CountCompletedByUserAndCourse(ctx, userID, e.CourseID)

			ongoingEnrollments = append(ongoingEnrollments, domain.EnrollmentWithCourse{
				Enrollment:       e,
				ModuleCount:      len(modules),
				CompletedModules: int(completedModules),
			})
		}
	}

	// Get certificates
	totalCerts, _ := uc.certRepo.GetByUserID(ctx, userID)
	recentCerts, _ := uc.certRepo.GetRecentByUserID(ctx, userID, 5)

	// Get upcoming labs
	upcomingLabs, _ := uc.labRepo.GetUpcoming(ctx)

	return &domain.StudentDashboardData{
		TotalEnrollments:   len(enrollments),
		CompletedCourses:   completedCount,
		InProgressCourses:  inProgressCount,
		TotalCertificates:  len(totalCerts),
		RecentCertificates: recentCerts,
		OngoingEnrollments: ongoingEnrollments,
		UpcomingLabs:       upcomingLabs,
	}, nil
}

func (uc *dashboardUsecase) GetInstructorDashboard(ctx context.Context, instructorID uint) (*domain.InstructorDashboardData, error) {
	// Get courses by instructor
	courses, err := uc.courseRepo.GetByInstructorID(ctx, instructorID)
	if err != nil {
		return nil, err
	}

	// Count total students across all courses
	totalStudents := 0
	for _, course := range courses {
		count, _ := uc.enrollmentRepo.CountByCourseID(ctx, course.ID)
		totalStudents += int(count)
	}

	// Count pending grades (assignments)
	pendingGrades, _ := uc.assignmentRepo.CountUngradedByInstructor(ctx, instructorID)

	// Get pending certificates
	allPendingCerts, _ := uc.certRepo.GetPending(ctx)
	pendingCertsCount := 0
	for _, cert := range allPendingCerts {
		// Check if this certificate is for instructor's course
		if cert.CourseID != nil {
			for _, course := range courses {
				if *cert.CourseID == course.ID {
					pendingCertsCount++
					break
				}
			}
		}
	}

	// Get recent submissions
	recentSubmissions, _ := uc.assignmentRepo.GetRecentSubmissions(ctx, 10)

	// Get labs with ungraded count
	allLabs, _ := uc.labRepo.GetAll(ctx)
	var ungradedLabs []domain.LabWithUngradedCount
	for _, lab := range allLabs {
		ungradedCount, _ := uc.labRepo.CountUngradedByLabID(ctx, lab.ID)
		if ungradedCount > 0 {
			ungradedLabs = append(ungradedLabs, domain.LabWithUngradedCount{
				Lab:           lab,
				UngradedCount: int(ungradedCount),
			})
		}
	}

	return &domain.InstructorDashboardData{
		TotalCourses:        len(courses),
		TotalStudents:       totalStudents,
		PendingGrades:       int(pendingGrades),
		PendingCertificates: pendingCertsCount,
		RecentSubmissions:   recentSubmissions,
		UngradedLabs:        ungradedLabs,
	}, nil
}

func (uc *dashboardUsecase) GetAdminDashboard(ctx context.Context) (*domain.AdminDashboardData, error) {
	// Count users by role
	totalStudents, _ := uc.userRepo.CountByRole(ctx, domain.RoleStudent)
	totalInstructors, _ := uc.userRepo.CountByRole(ctx, domain.RoleInstructor)
	totalAdmins, _ := uc.userRepo.CountByRole(ctx, domain.RoleAdmin)

	// Count courses
	totalCourses, _ := uc.courseRepo.Count(ctx)

	// Count labs
	totalLabs, _ := uc.labRepo.Count(ctx)

	// Count certificates
	totalCerts, _ := uc.certRepo.Count(ctx)
	pendingCerts, _ := uc.certRepo.CountByStatus(ctx, "pending")

	return &domain.AdminDashboardData{
		TotalUsers:          int(totalStudents + totalInstructors + totalAdmins),
		TotalStudents:       int(totalStudents),
		TotalInstructors:    int(totalInstructors),
		TotalCourses:        int(totalCourses),
		TotalLabs:           int(totalLabs),
		TotalCertificates:   int(totalCerts),
		PendingCertificates: int(pendingCerts),
	}, nil
}
