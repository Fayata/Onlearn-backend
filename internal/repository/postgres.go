package repository

import (
	"context"
	"errors"
	"onlearn-backend/internal/domain"

	"gorm.io/gorm"
)

// --- User Repo ---
type userRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepo{db}
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

func (r *userRepo) Update(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepo) UpdateVerified(ctx context.Context, email string) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).Where("email = ?", email).Update("is_verified", true).Error
}

// --- Course Repo ---
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
	err := r.db.WithContext(ctx).Find(&courses).Error
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

// --- Lab Repo ---
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
	err := r.db.WithContext(ctx).Find(&labs).Error
	return labs, err
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

func (r *labRepo) GetGradesByLabID(ctx context.Context, labID uint) ([]domain.LabGrade, error) {
	var grades []domain.LabGrade
	err := r.db.WithContext(ctx).Where("lab_id = ?", labID).Find(&grades).Error
	return grades, err
}

// --- Certificate Repo ---
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
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&certs).Error
	return certs, err
}
