package usecase

import (
	"context"
	"errors"
	"log"
	"onlearn-backend/internal/domain"
	"onlearn-backend/pkg/utils"
	"time"
)

type authUsecase struct {
	userRepo domain.UserRepository
}

func NewAuthUsecase(ur domain.UserRepository) domain.AuthUsecase {
	return &authUsecase{userRepo: ur}
}

func (uc *authUsecase) Register(ctx context.Context, user *domain.User) error {
	existing, _ := uc.userRepo.GetByEmail(ctx, user.Email)
	if existing != nil && existing.ID != 0 {
		return errors.New("email already exists")
	}

	hashed, err := utils.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashed
	user.CreatedAt = time.Now()

	err = uc.userRepo.Create(ctx, user)
	if err != nil {
		return err
	}

	go utils.SendEmail(user.Email, "Verify Your OnLearn Account", "Here is your verification code: FAKE-CODE")
	return nil
}

func (uc *authUsecase) Login(ctx context.Context, email, password string) (string, error) {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil || user.ID == 0 {
		return "", errors.New("invalid credentials")
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return "", errors.New("invalid credentials")
	}

	// Update last login timestamp
	if err := uc.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		log.Printf("Warning: Failed to update last login: %v", err)
	}

	return utils.GenerateJWT(user.ID, string(user.Role))
}

func (uc *authUsecase) UpdateUser(ctx context.Context, user *domain.User) error {
	existingUser, err := uc.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return errors.New("user not found")
	}

	if user.Name != "" {
		existingUser.Name = user.Name
	}
	if user.ProfilePicture != "" {
		existingUser.ProfilePicture = user.ProfilePicture
	}
	if user.Password != "" {
		hashedPassword, err := utils.HashPassword(user.Password)
		if err != nil {
			return err
		}
		existingUser.Password = hashedPassword
	}

	return uc.userRepo.Update(ctx, existingUser)
}

func (uc *authUsecase) VerifyEmail(ctx context.Context, email, code string) error {
	return uc.userRepo.UpdateVerified(ctx, email)
}

func (uc *authUsecase) ForgotPassword(ctx context.Context, email string) error {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil || user.ID == 0 {
		return nil
	}
	go utils.SendEmail(email, "Password Reset Request", "Here is your password reset link: FAKE-LINK")
	return nil
}

func (uc *authUsecase) GetUserByID(ctx context.Context, id uint) (*domain.User, error) {
	return uc.userRepo.GetByID(ctx, id)
}
