package usecase

import (
	"context"
	"errors"
	"onlearn-backend/internal/domain"
)

type labUsecase struct {
	labRepo  domain.LabRepository
	userRepo domain.UserRepository
}

func NewLabUsecase(lr domain.LabRepository, ur domain.UserRepository) domain.LabUsecase {
	return &labUsecase{
		labRepo:  lr,
		userRepo: ur,
	}
}

func (uc *labUsecase) CreateLab(ctx context.Context, lab *domain.Lab) error {
	return uc.labRepo.Create(ctx, lab)
}

func (uc *labUsecase) UpdateLabStatus(ctx context.Context, labID uint, status string) error {
	lab, err := uc.labRepo.GetByID(ctx, labID)
	if err != nil {
		return errors.New("lab not found")
	}

	lab.Status = status
	return uc.labRepo.Update(ctx, lab)
}

func (uc *labUsecase) StudentEnroll(ctx context.Context, userID, labID uint) error {
	existing, err := uc.labRepo.GetGrade(ctx, userID, labID)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.New("user already enrolled in this lab")
	}

	grade := &domain.LabGrade{
		UserID: userID,
		LabID:  labID,
		Grade:  "", // Initially ungraded
	}
	return uc.labRepo.CreateGrade(ctx, grade)
}

func (uc *labUsecase) SubmitGrade(ctx context.Context, instructorID, userID, labID uint, grade string) error {
	// Here you might add logic to verify the instructorID has authority over the lab
	labGrade, err := uc.labRepo.GetGrade(ctx, userID, labID)
	if err != nil {
		return err
	}
	if labGrade == nil {
		return errors.New("student not enrolled in this lab")
	}

	labGrade.Grade = grade
	return uc.labRepo.UpdateGrade(ctx, labGrade)
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