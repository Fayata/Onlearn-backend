package usecase

import (
	"context"
	"errors"
	"onlearn-backend/internal/domain"
	"time"
)

type courseUsecase struct {
	courseRepo     domain.CourseRepository
	moduleRepo     domain.ModuleRepository
	enrollmentRepo domain.EnrollmentRepository
	progressRepo   domain.ModuleProgressRepository
	assignmentRepo domain.AssignmentRepository
	certRepo       domain.CertificateRepository
}

func NewCourseUsecase(
	cr domain.CourseRepository,
	mr domain.ModuleRepository,
	er domain.EnrollmentRepository,
	pr domain.ModuleProgressRepository,
	ar domain.AssignmentRepository,
	certr domain.CertificateRepository,
) domain.CourseUsecase {
	return &courseUsecase{
		courseRepo:     cr,
		moduleRepo:     mr,
		enrollmentRepo: er,
		progressRepo:   pr,
		assignmentRepo: ar,
		certRepo:       certr,
	}
}

// ========== COURSE CRUD ==========

func (uc *courseUsecase) CreateCourse(ctx context.Context, course *domain.Course) error {
	return uc.courseRepo.Create(ctx, course)
}

func (uc *courseUsecase) UpdateCourse(ctx context.Context, course *domain.Course) error {
	existing, err := uc.courseRepo.GetByID(ctx, course.ID)
	if err != nil {
		return err
	}

	// Update only allowed fields
	existing.Title = course.Title
	existing.Description = course.Description
	if course.Thumbnail != "" {
		existing.Thumbnail = course.Thumbnail
	}

	return uc.courseRepo.Update(ctx, existing)
}

func (uc *courseUsecase) DeleteCourse(ctx context.Context, id uint) error {
	// Check if course has enrollments
	enrollments, _ := uc.enrollmentRepo.GetByCourseID(ctx, id)
	if len(enrollments) > 0 {
		return errors.New("cannot delete course with existing enrollments")
	}

	// Delete all modules (MongoDB)
	modules, _ := uc.moduleRepo.GetByCourseID(ctx, id)
	for _, module := range modules {
		uc.moduleRepo.Delete(ctx, module.ID)
	}

	return uc.courseRepo.Delete(ctx, id)
}

func (uc *courseUsecase) GetAllCourses(ctx context.Context) ([]domain.Course, error) {
	// Return only published courses (for student browse)
	// Instructors should use GetInstructorCourses to see all their courses
	return uc.courseRepo.GetPublished(ctx)
}

func (uc *courseUsecase) GetInstructorCourses(ctx context.Context, instructorID uint) ([]domain.Course, error) {
	return uc.courseRepo.GetByInstructorID(ctx, instructorID)
}

func (uc *courseUsecase) GetCourseDetails(ctx context.Context, courseID uint, userID *uint) (*domain.CourseDetail, error) {
	course, err := uc.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return nil, err
	}

	modules, err := uc.moduleRepo.GetByCourseID(ctx, courseID)
	if err != nil {
		return nil, err
	}

	enrolledCount, _ := uc.enrollmentRepo.CountByCourseID(ctx, courseID)

	isEnrolled := false
	if userID != nil {
		enrollment, _ := uc.enrollmentRepo.GetByUserAndCourse(ctx, *userID, courseID)
		isEnrolled = enrollment != nil
	}

	return &domain.CourseDetail{
		Course:           *course,
		Modules:          modules,
		EnrolledStudents: int(enrolledCount),
		IsEnrolled:       isEnrolled,
	}, nil
}

// ========== MODULE CRUD ==========

func (uc *courseUsecase) AddModule(ctx context.Context, module *domain.Module) error {
	// Verify course exists
	_, err := uc.courseRepo.GetByID(ctx, module.CourseID)
	if err != nil {
		return errors.New("course not found")
	}

	return uc.moduleRepo.Create(ctx, module)
}

func (uc *courseUsecase) GetModuleByID(ctx context.Context, moduleID string) (*domain.Module, error) {
	return uc.moduleRepo.GetByID(ctx, moduleID)
}

func (uc *courseUsecase) UpdateModule(ctx context.Context, module *domain.Module) error {
	return uc.moduleRepo.Update(ctx, module)
}

func (uc *courseUsecase) DeleteModule(ctx context.Context, moduleID string) error {
	// Check if module has submissions
	// We'll allow deletion for now, but in production you might want to prevent this
	return uc.moduleRepo.Delete(ctx, moduleID)
}

// ========== ENROLLMENT ==========

func (uc *courseUsecase) EnrollStudent(ctx context.Context, userID, courseID uint) error {
	// Check if already enrolled
	existing, _ := uc.enrollmentRepo.GetByUserAndCourse(ctx, userID, courseID)
	if existing != nil {
		return errors.New("already enrolled in this course")
	}

	// Verify course exists
	_, err := uc.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return errors.New("course not found")
	}

	enrollment := &domain.Enrollment{
		UserID:   userID,
		CourseID: courseID,
		Progress: 0,
	}

	return uc.enrollmentRepo.Create(ctx, enrollment)
}

func (uc *courseUsecase) GetStudentEnrollments(ctx context.Context, userID uint) ([]domain.EnrollmentWithCourse, error) {
	enrollments, err := uc.enrollmentRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var result []domain.EnrollmentWithCourse
	for _, e := range enrollments {
		modules, _ := uc.moduleRepo.GetByCourseID(ctx, e.CourseID)
		completedModules, _ := uc.progressRepo.CountCompletedByUserAndCourse(ctx, userID, e.CourseID)

		result = append(result, domain.EnrollmentWithCourse{
			Enrollment:       e,
			ModuleCount:      len(modules),
			CompletedModules: int(completedModules),
		})
	}

	return result, nil
}

func (uc *courseUsecase) GetCourseStudents(ctx context.Context, courseID uint) ([]domain.User, error) {
	enrollments, err := uc.enrollmentRepo.GetByCourseID(ctx, courseID)
	if err != nil {
		return nil, err
	}

	var students []domain.User
	for _, e := range enrollments {
		students = append(students, e.User)
	}

	return students, nil
}

// ========== MODULE PROGRESS ==========

func (uc *courseUsecase) MarkModuleComplete(ctx context.Context, userID uint, moduleID string, courseID uint) error {
	// Check if already marked
	existing, _ := uc.progressRepo.GetByUserAndModule(ctx, userID, moduleID)
	if existing != nil {
		if existing.IsComplete {
			return nil // Already complete
		}
		existing.IsComplete = true
		if err := uc.progressRepo.Update(ctx, existing); err != nil {
			return err
		}
	} else {
		// Create new progress
		progress := &domain.ModuleProgress{
			UserID:     userID,
			ModuleID:   moduleID,
			CourseID:   courseID,
			IsComplete: true,
		}
		if err := uc.progressRepo.Create(ctx, progress); err != nil {
			return err
		}
	}

	// Update enrollment progress
	return uc.updateEnrollmentProgress(ctx, userID, courseID)
}

func (uc *courseUsecase) updateEnrollmentProgress(ctx context.Context, userID uint, courseID uint) error {
	enrollment, err := uc.enrollmentRepo.GetByUserAndCourse(ctx, userID, courseID)
	if err != nil {
		return err
	}

	// Calculate progress percentage
	modules, _ := uc.moduleRepo.GetByCourseID(ctx, courseID)
	totalModules := len(modules)

	if totalModules == 0 {
		return nil
	}

	completedModules, _ := uc.progressRepo.CountCompletedByUserAndCourse(ctx, userID, courseID)
	progress := (float64(completedModules) / float64(totalModules)) * 100

	enrollment.Progress = progress

	// Mark as finished if 100%
	if progress >= 100 {
		enrollment.IsFinished = true

		// Auto-generate certificate
		uc.certRepo.Create(ctx, &domain.Certificate{
			UserID:   userID,
			CourseID: &courseID,
			Title:    "Course Completion Certificate",
			URL:      "/certificates/auto-generated.pdf", // TODO: Generate actual PDF
			Status:   "pending",
		})
	}

	return uc.enrollmentRepo.Update(ctx, enrollment)
}

// ========== PUBLISH/UNPUBLISH COURSE ==========

func (uc *courseUsecase) PublishCourse(ctx context.Context, courseID uint, instructorID uint) error {
	// Verify course exists and belongs to instructor
	course, err := uc.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return errors.New("course not found")
	}

	// Verify ownership
	if course.InstructorID != instructorID {
		return errors.New("unauthorized: course does not belong to instructor")
	}

	// Publish course
	course.IsPublished = true
	return uc.courseRepo.Update(ctx, course)
}

func (uc *courseUsecase) UnpublishCourse(ctx context.Context, courseID uint, instructorID uint) error {
	// Verify course exists and belongs to instructor
	course, err := uc.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return errors.New("course not found")
	}

	// Verify ownership
	if course.InstructorID != instructorID {
		return errors.New("unauthorized: course does not belong to instructor")
	}

	// Unpublish course
	course.IsPublished = false
	return uc.courseRepo.Update(ctx, course)
}

func (uc *courseUsecase) GetModulesWithProgress(ctx context.Context, userID uint, courseID uint) ([]domain.ModuleWithProgress, error) {
	modules, err := uc.moduleRepo.GetByCourseID(ctx, courseID)
	if err != nil {
		return nil, err
	}

	var result []domain.ModuleWithProgress
	for _, module := range modules {
		// Check progress
		progress, _ := uc.progressRepo.GetByUserAndModule(ctx, userID, module.ID)
		isComplete := progress != nil && progress.IsComplete

		// Check if has submission (for modules with quiz/assignment)
		var hasSubmission bool
		var grade *float64
		if module.QuizLink != "" {
			assignment, _ := uc.assignmentRepo.GetByUserAndModule(ctx, userID, module.ID)
			hasSubmission = assignment != nil
			if assignment != nil {
				grade = assignment.Grade
			}
		}

		result = append(result, domain.ModuleWithProgress{
			Module:          module,
			IsComplete:      isComplete,
			HasSubmission:   hasSubmission,
			SubmissionGrade: grade,
		})
	}

	return result, nil
}

// ========== ASSIGNMENTS ==========

func (uc *courseUsecase) SubmitAssignment(ctx context.Context, assignment *domain.Assignment) error {
	// Check if already submitted
	existing, _ := uc.assignmentRepo.GetByUserAndModule(ctx, assignment.UserID, assignment.ModuleID)
	if existing != nil {
		return errors.New("assignment already submitted")
	}

	assignment.SubmittedAt = time.Now()
	return uc.assignmentRepo.Create(ctx, assignment)
}

func (uc *courseUsecase) GradeAssignment(ctx context.Context, assignmentID uint, grade float64, feedback string, gradedByID uint) error {
	assignment, err := uc.assignmentRepo.GetByID(ctx, assignmentID)
	if err != nil {
		return err
	}

	assignment.Grade = &grade
	assignment.Feedback = feedback
	assignment.GradedByID = &gradedByID
	now := time.Now()
	assignment.GradedAt = &now

	return uc.assignmentRepo.Update(ctx, assignment)
}

func (uc *courseUsecase) GetCourseAssignments(ctx context.Context, courseID uint) ([]domain.Assignment, error) {
	return uc.assignmentRepo.GetByCourseID(ctx, courseID)
}
