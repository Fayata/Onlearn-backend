package usecase

import (
	"context"
	"errors"
	"fmt"
	"onlearn-backend/internal/domain"
	"time"
)

type certificateUsecase struct {
	certRepo domain.CertificateRepository
	userRepo domain.UserRepository
}

func NewCertificateUsecase(cr domain.CertificateRepository, ur domain.UserRepository) domain.CertificateUsecase {
	return &certificateUsecase{
		certRepo: cr,
		userRepo: ur,
	}
}

func (uc *certificateUsecase) GenerateCertificate(ctx context.Context, userID uint, courseID *uint, labID *uint, title string) (*domain.Certificate, error) {
	_, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// In a real app, you would use a PDF generation library to create the certificate
	// and upload it somewhere, then save the URL.
	// For now, we will just create a record with a dummy URL.
	cert := &domain.Certificate{
		UserID:    userID,
		CourseID:  courseID,
		LabID:     labID,
		Title:     title,
		URL:       fmt.Sprintf("/uploads/certificates/%s-%d.pdf", title, time.Now().Unix()),
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
