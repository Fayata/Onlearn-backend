package usecase

import (
	"context"
	"onlearn-backend/internal/domain"
	"sort"
	"time"
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
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	enrollments, err := uc.enrollmentRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	completedCount := 0
	inProgressCount := 0
	var ongoingEnrollments []domain.EnrollmentWithCourse
	for _, e := range enrollments {
		if e.IsFinished {
			completedCount++
		} else {
			inProgressCount++
			modules, _ := uc.moduleRepo.GetByCourseID(ctx, e.CourseID)
			completedModules, _ := uc.progressRepo.CountCompletedByUserAndCourse(ctx, userID, e.CourseID)
			ongoingEnrollments = append(ongoingEnrollments, domain.EnrollmentWithCourse{
				Enrollment:       e,
				ModuleCount:      len(modules),
				CompletedModules: int(completedModules),
			})
		}
	}

	totalCerts, _ := uc.certRepo.GetByUserID(ctx, userID)
	recentCerts, _ := uc.certRepo.GetRecentByUserID(ctx, userID, 3)

	allLabs, _ := uc.labRepo.GetAll(ctx)
	userLabGrades, _ := uc.labRepo.GetGradesByUserID(ctx, userID)
	upcomingLabs, _ := uc.labRepo.GetUpcoming(ctx)
	upcomingLabsSorted := sortAndClassifyLabs(upcomingLabs)

	recentActivities, _ := uc.getRecentActivities(ctx, userID)

	return &domain.StudentDashboardData{
		User:                user,
		TotalEnrollments:    len(enrollments),
		CompletedCourses:    completedCount,
		InProgressCourses:   inProgressCount,
		TotalCertificates:   len(totalCerts),
		RecentCertificates:  recentCerts,
		OngoingEnrollments:  ongoingEnrollments,
		UpcomingLabs:        upcomingLabs,
		UpcomingLabsSorted:  upcomingLabsSorted,
		LeaderboardRank:     5, // Mock data
		CompletedLabs:       len(userLabGrades),
		AssignedLabs:        len(allLabs),     // Assuming all labs are assigned to all students for now
		WeeklyGoalProgress:  80,               // Mock data
		RecentActivities:    recentActivities,
	}, nil
}

// getRecentActivities aggregates different types of activities for the user.
func (uc *dashboardUsecase) getRecentActivities(ctx context.Context, userID uint) ([]domain.RecentActivity, error) {
	var activities []domain.RecentActivity

	// Completed Modules
	completedModules, err := uc.progressRepo.GetRecentByUser(ctx, userID, 3)
	if err == nil {
		for _, p := range completedModules {
			// To get module title, we'd need to fetch the module from Mongo
			// For now, let's use a placeholder
			activities = append(activities, domain.RecentActivity{
				Type:      domain.ActivityModuleCompleted,
				Title:     "Module Selesai",
				Subtitle:  "Anda menyelesaikan sebuah modul",
				Timestamp: p.UpdatedAt,
				Icon:      "fas fa-check",
				BgColor:   "bg-green-100",
				IconColor: "text-green-600",
			})
		}
	}

	// Submitted Assignments
	submittedAssignments, err := uc.assignmentRepo.GetRecentSubmissionsByUserID(ctx, userID, 3)
	if err == nil {
		for _, a := range submittedAssignments {
			// To get assignment title, we'd need module title from Mongo
			activities = append(activities, domain.RecentActivity{
				Type:      domain.ActivityTaskSubmitted,
				Title:     "Tugas Dikirim",
				Subtitle:  "Menunggu penilaian instruktur",
				Timestamp: a.SubmittedAt,
				Icon:      "fas fa-upload",
				BgColor:   "bg-blue-100",
				IconColor: "text-blue-600",
			})
		}
	}

	// New Certificates
	newCerts, err := uc.certRepo.GetRecentByUserID(ctx, userID, 3)
	if err == nil {
		for _, c := range newCerts {
			activities = append(activities, domain.RecentActivity{
				Type:      domain.ActivityCertificateNew,
				Title:     "Sertifikat Baru Diterima",
				Subtitle:  c.Title,
				Timestamp: c.IssueDate,
				Icon:      "fas fa-trophy",
				BgColor:   "bg-purple-100",
				IconColor: "text-purple-600",
			})
		}
	}

	// Sort activities by timestamp descending
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].Timestamp.After(activities[j].Timestamp)
	})

	// Limit to latest 3-5 activities overall
	if len(activities) > 3 {
		activities = activities[:3]
	}

	return activities, nil
}

// sortAndClassifyLabs sorts labs into open, scheduled, and closed.
func sortAndClassifyLabs(labs []domain.Lab) []domain.Lab {
	now := time.Now()
	
	// Separate labs by status
	var open, scheduled, closed []domain.Lab
	for _, lab := range labs {
		if lab.Status == "open" || (now.After(lab.StartTime) && now.Before(lab.EndTime)) {
			lab.Status = "Open"
			open = append(open, lab)
		} else if lab.Status == "scheduled" && now.Before(lab.StartTime) {
			lab.Status = "Scheduled"
			scheduled = append(scheduled, lab)
		} else {
			lab.Status = "Closed"
			closed = append(closed, lab)
		}
	}

	// Sort each category
	sort.Slice(open, func(i, j int) bool { return open[i].StartTime.Before(open[j].StartTime) })
	sort.Slice(scheduled, func(i, j int) bool { return scheduled[i].StartTime.Before(scheduled[j].StartTime) })
	sort.Slice(closed, func(i, j int) bool { return closed[i].EndTime.After(closed[j].EndTime) })

	// Combine them in order: Open, Scheduled, Closed
	return append(append(open, scheduled...), closed...)
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