package repository

import (
	"context"
	"errors"
	"onlearn-backend/internal/domain"
	"time"

	"gorm.io/gorm"
)

// ========== USER REPOSITORY ==========

type userRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepo{db}
}
func (r *userRepo) UpdateLastLogin(ctx context.Context, userID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", userID).
		Update("last_login_at", now).Error
}
func (r *userRepo) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user not found")
	}
	return &user, err
}

func (r *userRepo) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user not found")
	}
	return &user, err
}

func (r *userRepo) GetByIDs(ctx context.Context, ids []uint) ([]domain.User, error) {
	var users []domain.User
	if len(ids) == 0 {
		return users, nil
	}
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error
	return users, err
}

func (r *userRepo) GetByRole(ctx context.Context, role domain.Role) ([]domain.User, error) {
	var users []domain.User
	err := r.db.WithContext(ctx).Where("role = ?", role).Find(&users).Error
	return users, err
}

func (r *userRepo) GetAll(ctx context.Context) ([]domain.User, error) {
	var users []domain.User
	err := r.db.WithContext(ctx).Find(&users).Error
	return users, err
}

func (r *userRepo) SearchStudents(ctx context.Context, searchTerm string) ([]domain.User, error) {
	var users []domain.User
	query := r.db.WithContext(ctx).Where("role = ?", domain.RoleStudent)
	
	if searchTerm != "" {
		searchPattern := "%" + searchTerm + "%"
		query = query.Where("name ILIKE ? OR email ILIKE ?", searchPattern, searchPattern)
	}
	
	err := query.Order("name ASC").Find(&users).Error
	return users, err
}

func (r *userRepo) Update(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepo) UpdateVerified(ctx context.Context, email string) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).Where("email = ?", email).Update("is_verified", true).Error
}

func (r *userRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.User{}, id).Error
}

func (r *userRepo) CountByRole(ctx context.Context, role domain.Role) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.User{}).Where("role = ?", role).Count(&count).Error
	return count, err
}

// ========== COURSE REPOSITORY ==========

type courseRepo struct {
	db *gorm.DB
}

func NewCourseRepository(db *gorm.DB) domain.CourseRepository {
	return &courseRepo{db}
}

func (r *courseRepo) Create(ctx context.Context, course *domain.Course) error {
	return r.db.WithContext(ctx).Create(course).Error
}

func (r *courseRepo) Update(ctx context.Context, course *domain.Course) error {
	return r.db.WithContext(ctx).Save(course).Error
}

func (r *courseRepo) GetAll(ctx context.Context) ([]domain.Course, error) {
	var courses []domain.Course
	err := r.db.WithContext(ctx).Preload("Instructor").Find(&courses).Error
	return courses, err
}

func (r *courseRepo) GetPublished(ctx context.Context) ([]domain.Course, error) {
	var courses []domain.Course
	err := r.db.WithContext(ctx).Where("is_published = ?", true).Preload("Instructor").Find(&courses).Error
	return courses, err
}

func (r *courseRepo) GetByID(ctx context.Context, id uint) (*domain.Course, error) {
	var course domain.Course
	err := r.db.WithContext(ctx).First(&course, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("course not found")
	}
	return &course, err
}

func (r *courseRepo) GetByInstructorID(ctx context.Context, instructorID uint) ([]domain.Course, error) {
	var courses []domain.Course
	err := r.db.WithContext(ctx).Where("instructor_id = ?", instructorID).Find(&courses).Error
	return courses, err
}

func (r *courseRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Course{}, id).Error
}

func (r *courseRepo) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Course{}).Count(&count).Error
	return count, err
}

// ========== ENROLLMENT REPOSITORY ==========

type enrollmentRepo struct {
	db *gorm.DB
}

func NewEnrollmentRepository(db *gorm.DB) domain.EnrollmentRepository {
	return &enrollmentRepo{db}
}

func (r *enrollmentRepo) Create(ctx context.Context, enrollment *domain.Enrollment) error {
	return r.db.WithContext(ctx).Create(enrollment).Error
}

func (r *enrollmentRepo) GetByUserAndCourse(ctx context.Context, userID, courseID uint) (*domain.Enrollment, error) {
	var enrollment domain.Enrollment
	err := r.db.WithContext(ctx).Where("user_id = ? AND course_id = ?", userID, courseID).First(&enrollment).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &enrollment, err
}

func (r *enrollmentRepo) GetByUserID(ctx context.Context, userID uint) ([]domain.Enrollment, error) {
	var enrollments []domain.Enrollment
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Preload("Course").Preload("Course.Instructor").Find(&enrollments).Error
	return enrollments, err
}

func (r *enrollmentRepo) GetByCourseID(ctx context.Context, courseID uint) ([]domain.Enrollment, error) {
	var enrollments []domain.Enrollment
	err := r.db.WithContext(ctx).Where("course_id = ?", courseID).Preload("User").Find(&enrollments).Error
	return enrollments, err
}

func (r *enrollmentRepo) Update(ctx context.Context, enrollment *domain.Enrollment) error {
	return r.db.WithContext(ctx).Save(enrollment).Error
}

func (r *enrollmentRepo) CountByCourseID(ctx context.Context, courseID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Enrollment{}).Where("course_id = ?", courseID).Count(&count).Error
	return count, err
}

func (r *enrollmentRepo) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Enrollment{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// ========== MODULE PROGRESS REPOSITORY ==========

type moduleProgressRepo struct {
	db *gorm.DB
}

func NewModuleProgressRepository(db *gorm.DB) domain.ModuleProgressRepository {
	return &moduleProgressRepo{db}
}

func (r *moduleProgressRepo) Create(ctx context.Context, progress *domain.ModuleProgress) error {
	return r.db.WithContext(ctx).Create(progress).Error
}

func (r *moduleProgressRepo) GetByUserAndModule(ctx context.Context, userID uint, moduleID string) (*domain.ModuleProgress, error) {
	var progress domain.ModuleProgress
	err := r.db.WithContext(ctx).Where("user_id = ? AND module_id = ?", userID, moduleID).First(&progress).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &progress, err
}

func (r *moduleProgressRepo) GetByUserAndCourse(ctx context.Context, userID uint, courseID uint) ([]domain.ModuleProgress, error) {
	var progress []domain.ModuleProgress
	err := r.db.WithContext(ctx).Where("user_id = ? AND course_id = ?", userID, courseID).Find(&progress).Error
	return progress, err
}

func (r *moduleProgressRepo) GetRecentByUser(ctx context.Context, userID uint, limit int) ([]domain.ModuleProgress, error) {
	var progress []domain.ModuleProgress
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_complete = ?", userID, true).
		Order("updated_at DESC").
		Limit(limit).
		Find(&progress).Error
	return progress, err
}

func (r *moduleProgressRepo) Update(ctx context.Context, progress *domain.ModuleProgress) error {
	return r.db.WithContext(ctx).Save(progress).Error
}

func (r *moduleProgressRepo) CountCompletedByUserAndCourse(ctx context.Context, userID uint, courseID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.ModuleProgress{}).
		Where("user_id = ? AND course_id = ? AND is_complete = ?", userID, courseID, true).
		Count(&count).Error
	return count, err
}

func (r *moduleProgressRepo) GetCompletedByModuleID(ctx context.Context, moduleID string) ([]domain.ModuleProgress, error) {
	var progress []domain.ModuleProgress
	err := r.db.WithContext(ctx).
		Where("module_id = ? AND is_complete = ?", moduleID, true).
		Order("updated_at DESC").
		Find(&progress).Error
	return progress, err
}

// ========== ASSIGNMENT REPOSITORY ==========

type assignmentRepo struct {
	db *gorm.DB
}

func NewAssignmentRepository(db *gorm.DB) domain.AssignmentRepository {
	return &assignmentRepo{db}
}

func (r *assignmentRepo) Create(ctx context.Context, assignment *domain.Assignment) error {
	return r.db.WithContext(ctx).Create(assignment).Error
}

func (r *assignmentRepo) GetByID(ctx context.Context, id uint) (*domain.Assignment, error) {
	var assignment domain.Assignment
	err := r.db.WithContext(ctx).Preload("User").Preload("GradedBy").First(&assignment, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("assignment not found")
	}
	return &assignment, err
}

func (r *assignmentRepo) GetByUserAndModule(ctx context.Context, userID uint, moduleID string) (*domain.Assignment, error) {
	var assignment domain.Assignment
	err := r.db.WithContext(ctx).Where("user_id = ? AND module_id = ?", userID, moduleID).First(&assignment).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &assignment, err
}

func (r *assignmentRepo) GetByCourseID(ctx context.Context, courseID uint) ([]domain.Assignment, error) {
	var assignments []domain.Assignment
	err := r.db.WithContext(ctx).Where("course_id = ?", courseID).
		Preload("User").Preload("GradedBy").
		Order("submitted_at DESC").
		Find(&assignments).Error
	return assignments, err
}

func (r *assignmentRepo) GetByModuleID(ctx context.Context, moduleID string) ([]domain.Assignment, error) {
	var assignments []domain.Assignment
	err := r.db.WithContext(ctx).Where("module_id = ?", moduleID).
		Preload("User").Preload("GradedBy").
		Order("submitted_at DESC").
		Find(&assignments).Error
	return assignments, err
}

func (r *assignmentRepo) GetUngradedByCourseID(ctx context.Context, courseID uint) ([]domain.Assignment, error) {
	var assignments []domain.Assignment
	err := r.db.WithContext(ctx).Where("course_id = ? AND grade IS NULL", courseID).
		Preload("User").
		Order("submitted_at ASC").
		Find(&assignments).Error
	return assignments, err
}

func (r *assignmentRepo) GetRecentSubmissionsByUserID(ctx context.Context, userID uint, limit int) ([]domain.Assignment, error) {
	var assignments []domain.Assignment
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("User").
		Order("submitted_at DESC").
		Limit(limit).
		Find(&assignments).Error
	return assignments, err
}

func (r *assignmentRepo) GetRecentSubmissions(ctx context.Context, limit int) ([]domain.Assignment, error) {
	var assignments []domain.Assignment
	err := r.db.WithContext(ctx).
		Preload("User").
		Order("submitted_at DESC").
		Limit(limit).
		Find(&assignments).Error
	return assignments, err
}

func (r *assignmentRepo) Update(ctx context.Context, assignment *domain.Assignment) error {
	return r.db.WithContext(ctx).Save(assignment).Error
}

func (r *assignmentRepo) CountUngradedByInstructor(ctx context.Context, instructorID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Assignment{}).
		Joins("JOIN courses ON assignments.course_id = courses.id").
		Where("courses.instructor_id = ? AND assignments.grade IS NULL", instructorID).
		Count(&count).Error
	return count, err
}

func (r *assignmentRepo) GetStudentsByModuleID(ctx context.Context, moduleID string) ([]domain.UserWithAssignment, error) {
	type TempResult struct {
		domain.User
		AssignmentID   uint      `gorm:"column:assignment_id"`
		FileURL        string
		Grade          *float64
		Feedback       string
		SubmittedAt    time.Time
		GradedAt       *time.Time
	}

	var tempResults []TempResult
	err := r.db.WithContext(ctx).
		Table("users").
		Select("users.*, assignments.id as assignment_id, assignments.file_url, assignments.grade, assignments.feedback, assignments.submitted_at, assignments.graded_at").
		Joins("LEFT JOIN assignments ON users.id = assignments.user_id AND assignments.module_id = ?", moduleID).
		Where("users.id IN (SELECT user_id FROM module_progresses WHERE module_id = ? AND is_complete = ?)", moduleID, true).
		Scan(&tempResults).Error

	if err != nil {
		return nil, err
	}

	var results []domain.UserWithAssignment
	for _, temp := range tempResults {
		userWithAssignment := domain.UserWithAssignment{
			User: temp.User,
		}
		if temp.AssignmentID != 0 {
			userWithAssignment.Assignment = &domain.Assignment{
				ID:          temp.AssignmentID,
				UserID:      temp.User.ID,
				ModuleID:    moduleID,
				FileURL:     temp.FileURL,
				Grade:       temp.Grade,
				Feedback:    temp.Feedback,
				SubmittedAt: temp.SubmittedAt,
				GradedAt:    temp.GradedAt,
			}
		}
		results = append(results, userWithAssignment)
	}

	return results, nil
}

// ========== LAB REPOSITORY ==========

type labRepo struct {
	db *gorm.DB
}

func NewLabRepository(db *gorm.DB) domain.LabRepository {
	return &labRepo{db}
}

func (r *labRepo) Create(ctx context.Context, lab *domain.Lab) error {
	return r.db.WithContext(ctx).Create(lab).Error
}

func (r *labRepo) Update(ctx context.Context, lab *domain.Lab) error {
	return r.db.WithContext(ctx).Save(lab).Error
}

func (r *labRepo) GetByID(ctx context.Context, id uint) (*domain.Lab, error) {
	var lab domain.Lab
	err := r.db.WithContext(ctx).First(&lab, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("lab not found")
	}
	return &lab, err
}

func (r *labRepo) GetAll(ctx context.Context) ([]domain.Lab, error) {
	var labs []domain.Lab
	err := r.db.WithContext(ctx).Order("start_time DESC").Find(&labs).Error
	return labs, err
}

func (r *labRepo) GetUpcoming(ctx context.Context) ([]domain.Lab, error) {
	var labs []domain.Lab
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("start_time > ? OR status = ?", now, "open").
		Order("start_time ASC").
		Find(&labs).Error
	return labs, err
}

func (r *labRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Lab{}, id).Error
}

func (r *labRepo) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Lab{}).Count(&count).Error
	return count, err
}

func (r *labRepo) CreateGrade(ctx context.Context, grade *domain.LabGrade) error {
	return r.db.WithContext(ctx).Create(grade).Error
}

func (r *labRepo) UpdateGrade(ctx context.Context, grade *domain.LabGrade) error {
	return r.db.WithContext(ctx).Save(grade).Error
}

func (r *labRepo) GetGrade(ctx context.Context, userID, labID uint) (*domain.LabGrade, error) {
	var grade domain.LabGrade
	err := r.db.WithContext(ctx).Where("user_id = ? AND lab_id = ?", userID, labID).First(&grade).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &grade, err
}

func (r *labRepo) GetGradesByUserID(ctx context.Context, userID uint) ([]domain.LabGrade, error) {
	var grades []domain.LabGrade
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Preload("Lab").Find(&grades).Error
	return grades, err
}

func (r *labRepo) GetGradesByLabID(ctx context.Context, labID uint) ([]domain.LabGrade, error) {
	var grades []domain.LabGrade
	err := r.db.WithContext(ctx).Where("lab_id = ?", labID).Preload("User").Find(&grades).Error
	return grades, err
}

func (r *labRepo) CountUngradedByLabID(ctx context.Context, labID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.LabGrade{}).
		Where("lab_id = ? AND grade = ?", labID, "").
		Count(&count).Error
	return count, err
}

// ========== CERTIFICATE REPOSITORY ==========

type certRepo struct {
	db *gorm.DB
}

func NewCertificateRepository(db *gorm.DB) domain.CertificateRepository {
	return &certRepo{db}
}

func (r *certRepo) Create(ctx context.Context, cert *domain.Certificate) error {
	return r.db.WithContext(ctx).Create(cert).Error
}

func (r *certRepo) GetByUserID(ctx context.Context, userID uint) ([]domain.Certificate, error) {
	var certs []domain.Certificate
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("Course").
		Preload("Lab").
		Order("issue_date DESC").
		Find(&certs).Error
	return certs, err
}

func (r *certRepo) GetByID(ctx context.Context, id uint) (*domain.Certificate, error) {
	var cert domain.Certificate
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Course").
		Preload("Lab").
		First(&cert, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("certificate not found")
	}
	return &cert, err
}

func (r *certRepo) GetPending(ctx context.Context) ([]domain.Certificate, error) {
	var certs []domain.Certificate
	err := r.db.WithContext(ctx).
		Where("status = ?", "pending").
		Preload("User").
		Preload("Course").
		Preload("Lab").
		Order("issue_date DESC").
		Find(&certs).Error
	return certs, err
}

func (r *certRepo) GetRecentByUserID(ctx context.Context, userID uint, limit int) ([]domain.Certificate, error) {
	var certs []domain.Certificate
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, "approved").
		Preload("Course").
		Preload("Lab").
		Order("issue_date DESC").
		Limit(limit).
		Find(&certs).Error
	return certs, err
}

func (r *certRepo) Update(ctx context.Context, cert *domain.Certificate) error {
	return r.db.WithContext(ctx).Save(cert).Error
}

func (r *certRepo) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Certificate{}).Count(&count).Error
	return count, err
}

func (r *certRepo) CountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Certificate{}).Where("status = ?", status).Count(&count).Error
	return count, err
}
