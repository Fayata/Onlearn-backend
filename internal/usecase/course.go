package usecase

import (
	"context"
	"onlearn-backend/internal/domain"
)

type courseUsecase struct {
	courseRepo domain.CourseRepository
	moduleRepo domain.ModuleRepository
	certRepo   domain.CertificateRepository
}

func NewCourseUsecase(cr domain.CourseRepository, mr domain.ModuleRepository, certR domain.CertificateRepository) domain.CourseUsecase {
	return &courseUsecase{
		courseRepo: cr,
		moduleRepo: mr,
		certRepo:   certR,
	}
}

func (uc *courseUsecase) CreateCourse(ctx context.Context, course *domain.Course) error {
	return uc.courseRepo.Create(ctx, course)
}

func (uc *courseUsecase) AddModule(ctx context.Context, module *domain.Module) error {
	return uc.moduleRepo.Create(ctx, module)
}

func (uc *courseUsecase) GetCourseDetails(ctx context.Context, courseID uint) (*domain.Course, []domain.Module, error) {
	course, err := uc.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return nil, nil, err
	}

	modules, err := uc.moduleRepo.GetByCourseID(ctx, courseID)
	if err != nil {
		return nil, nil, err
	}

	return course, modules, nil
}

func (uc *courseUsecase) GetAllCourses(ctx context.Context) ([]domain.Course, error) {
	return uc.courseRepo.GetAll(ctx)
}