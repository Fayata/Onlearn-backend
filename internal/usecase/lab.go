package usecase

import (
	"context"
	"errors"
	"onlearn-backend/internal/domain"
)

type labUsecase struct {
	labRepo  domain.LabRepository
	userRepo domain.UserRepository
	certRepo domain.CertificateRepository
}

func NewLabUsecase(
	lr domain.LabRepository,
	ur domain.UserRepository,
	cr domain.CertificateRepository,
) domain.LabUsecase {
	return &labUsecase{
		labRepo:  lr,
		userRepo: ur,
		certRepo: cr,
	}
}

// ========== LAB CRUD ==========

func (uc *labUsecase) CreateLab(ctx context.Context, lab *domain.Lab) error {
	if lab.Status == "" {
		lab.Status = "scheduled"
	}
	return uc.labRepo.Create(ctx, lab)
}

func (uc *labUsecase) UpdateLab(ctx context.Context, lab *domain.Lab) error {
	existing, err := uc.labRepo.GetByID(ctx, lab.ID)
	if err != nil {
		return err
	}

	// Update fields
	existing.Title = lab.Title
	existing.Description = lab.Description
	existing.StartTime = lab.StartTime
	existing.EndTime = lab.EndTime
	if lab.Status != "" {
		existing.Status = lab.Status
	}

	return uc.labRepo.Update(ctx, existing)
}

func (uc *labUsecase) UpdateLabStatus(ctx context.Context, labID uint, status string) error {
	lab, err := uc.labRepo.GetByID(ctx, labID)
	if err != nil {
		return errors.New("lab not found")
	}

	// Validate status
	validStatuses := map[string]bool{
		"scheduled": true,
		"open":      true,
		"closed":    true,
	}

	if !validStatuses[status] {
		return errors.New("invalid status")
	}

	lab.Status = status
	return uc.labRepo.Update(ctx, lab)
}

func (uc *labUsecase) GetLabByID(ctx context.Context, labID uint) (*domain.Lab, error) {
	return uc.labRepo.GetByID(ctx, labID)
}

func (uc *labUsecase) GetAllLabs(ctx context.Context) ([]domain.Lab, error) {
	return uc.labRepo.GetAll(ctx)
}

func (uc *labUsecase) GetUpcomingLabs(ctx context.Context) ([]domain.Lab, error) {
	return uc.labRepo.GetUpcoming(ctx)
}

func (uc *labUsecase) DeleteLab(ctx context.Context, labID uint) error {
	// Check if lab has grades
	grades, _ := uc.labRepo.GetGradesByLabID(ctx, labID)
	if len(grades) > 0 {
		return errors.New("cannot delete lab with existing grades")
	}

	return uc.labRepo.Delete(ctx, labID)
}

// ========== LAB GRADING ==========

func (uc *labUsecase) StudentEnroll(ctx context.Context, userID, labID uint) error {
	// Check if already enrolled
	existing, err := uc.labRepo.GetGrade(ctx, userID, labID)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.New("already enrolled in this lab")
	}

	// Verify lab exists
	_, err = uc.labRepo.GetByID(ctx, labID)
	if err != nil {
		return errors.New("lab not found")
	}

	// Create grade entry (ungraded initially)
	grade := &domain.LabGrade{
		UserID: userID,
		LabID:  labID,
		Grade:  "", // Empty = not graded yet
	}

	return uc.labRepo.CreateGrade(ctx, grade)
}

func (uc *labUsecase) SubmitGrade(ctx context.Context, instructorID, userID, labID uint, grade string, feedback string) error {
	// Verify instructor exists
	instructor, err := uc.userRepo.GetByID(ctx, instructorID)
	if err != nil {
		return errors.New("instructor not found")
	}

	if instructor.Role != domain.RoleInstructor && instructor.Role != domain.RoleAdmin {
		return errors.New("only instructors and admins can grade labs")
	}

	// Get existing grade record
	labGrade, err := uc.labRepo.GetGrade(ctx, userID, labID)
	if err != nil {
		return err
	}

	if labGrade == nil {
		return errors.New("student not enrolled in this lab")
	}

	// Update grade
	labGrade.Grade = grade
	labGrade.Feedback = feedback

	err = uc.labRepo.UpdateGrade(ctx, labGrade)
	if err != nil {
		return err
	}

	// Auto-generate certificate if grade is good (e.g., A or B)
	// You can customize this logic
	if grade == "A" || grade == "B" {
		uc.certRepo.Create(ctx, &domain.Certificate{
			UserID: userID,
			LabID:  &labID,
			Title:  "Lab Completion Certificate",
			URL:    "/certificates/lab-auto-generated.pdf",
			Status: "pending",
		})
	}

	return nil
}

func (uc *labUsecase) GetUngradedStudents(ctx context.Context, labID uint) ([]domain.User, error) {
	grades, err := uc.labRepo.GetGradesByLabID(ctx, labID)
	if err != nil {
		return nil, err
	}

	var ungradedUserIDs []uint
	for _, g := range grades {
		if g.Grade == "" {
			ungradedUserIDs = append(ungradedUserIDs, g.UserID)
		}
	}

	if len(ungradedUserIDs) == 0 {
		return []domain.User{}, nil
	}

	return uc.userRepo.GetByIDs(ctx, ungradedUserIDs)
}

func (uc *labUsecase) GetLabsWithUngradedCount(ctx context.Context) ([]domain.LabWithUngradedCount, error) {
	labs, err := uc.labRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var result []domain.LabWithUngradedCount
	for _, lab := range labs {
		ungradedCount, _ := uc.labRepo.CountUngradedByLabID(ctx, lab.ID)

		result = append(result, domain.LabWithUngradedCount{
			Lab:           lab,
			UngradedCount: int(ungradedCount),
		})
	}

	return result, nil
}
