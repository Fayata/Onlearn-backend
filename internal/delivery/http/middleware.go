package http

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware untuk API (menggunakan Header Authorization)
func AuthMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid auth header format"})
			return
		}
		tokenString := parts[1]

		validateTokenAndSetContext(c, tokenString, roles, true)
	}
}

// WebAuthMiddleware untuk Website (menggunakan Cookie)
func WebAuthMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("token")
		if err != nil {
			// Jika tidak ada cookie, redirect ke halaman login
			c.Redirect(http.StatusFound, "/?error=Unauthorized")
			c.Abort()
			return
		}

		validateTokenAndSetContext(c, tokenString, roles, false)
	}
}

// Helper function untuk validasi token
func validateTokenAndSetContext(c *gin.Context, tokenString string, roles []string, isAPI bool) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		if isAPI {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		} else {
			c.Redirect(http.StatusFound, "/?error=Invalid+token")
			c.Abort()
		}
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		if isAPI {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
		} else {
			c.Redirect(http.StatusFound, "/?error=Invalid+claims")
			c.Abort()
		}
		return
	}

	userRole := claims["role"].(string)

	// Role Validation
	if len(roles) > 0 {
		roleAllowed := false
		for _, r := range roles {
			if r == userRole {
				roleAllowed = true
				break
			}
		}
		if !roleAllowed {
			if isAPI {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden access"})
			} else {
				// Redirect ke dashboard masing-masing jika salah role, atau logout
				c.Redirect(http.StatusFound, "/?error=Forbidden")
				c.Abort()
			}
			return
		}
	}

	// Simpan user_id dan role ke context
	userIDFloat := claims["user_id"].(float64)
	c.Set("user_id", uint(userIDFloat))
	c.Set("role", userRole)
	c.Next()
}
