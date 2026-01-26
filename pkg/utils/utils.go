package utils

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	UploadDirectory = "./uploads"
)

// getJWTSecret gets JWT secret from environment variable
func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-key-change-in-production"
	}
	return []byte(secret)
}

// HashPassword hashes a password using bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash compares a password with its hash.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Claims defines the JWT claims.
type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a new JWT token.
func GenerateJWT(userID uint, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

// ValidateJWT validates a JWT token.
func ValidateJWT(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return getJWTSecret(), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func HandleUpload(c *gin.Context, formFile string) (string, error) {
	file, err := c.FormFile(formFile)
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil
		}
		return "", err
	}

	if err := os.MkdirAll(UploadDirectory, os.ModePerm); err != nil {
		return "", err
	}

	extension := filepath.Ext(file.Filename)
	uniqueFileName := fmt.Sprintf("%d-%s%s", time.Now().Unix(), randString(10), extension)
	filePath := filepath.Join(UploadDirectory, uniqueFileName)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		return "", err
	}

	return "/" + strings.ReplaceAll(filePath, "\\", "/"), nil
}

// randString generates a random string of a given length.
func randString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// SendEmail is a placeholder for email functionality
func SendEmail(to, subject, body string) error {
	fmt.Printf("--- EMAIL SENT ---\nTo: %s\nSubject: %s\nBody: %s\n-----------------\n", to, subject, body)
	return nil
}
