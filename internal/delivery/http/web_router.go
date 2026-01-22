package http

import (
	"fmt"
	"html/template"
	"onlearn-backend/internal/domain"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func InitWebRouter(router *gin.Engine, webHandler *WebHandler) {
	router.SetFuncMap(template.FuncMap{
		// String & Slice utils
		"upper": strings.ToUpper,
		"slice": func(s string, start, end int) string {
			if start < 0 || end > len(s) || start > end {
				return s
			}
			return s[start:end]
		},
		"len": func(arr interface{}) int {
			v := reflect.ValueOf(arr)
			if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
				return v.Len()
			}
			return 0
		},
		"seq": func(start, end int) []int {
			var result []int
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
		},

		// Logic utils
		"eq": func(a, b interface{}) bool {
			return a == b
		},

		// Math utils
		"div": func(b, a interface{}) float64 {
			valA := toFloat(a)
			valB := toFloat(b)
			if valB == 0 {
				return 0
			}
			return valA / valB
		},
		"mul": func(b, a interface{}) float64 { return toFloat(a) * toFloat(b) },
		"sub": func(b, a interface{}) float64 { return toFloat(a) - toFloat(b) },
		"add": func(a interface{}, b interface{}) interface{} {
			valA := toFloat(a)
			valB := toFloat(b)
			if valA == float64(int(valA)) && valB == float64(int(valB)) {
				return int(valA) + int(valB)
			}
			return valA + valB
		},
		"mod": func(i, j int) int { return i % j },

		"date": func(t interface{}) string {
			if t == nil {
				return ""
			}
			switch v := t.(type) {
			case time.Time:
				return v.Format("02 Jan 2006, 15:04")
			case *time.Time:
				if v == nil {
					return ""
				}
				return v.Format("02 Jan 2006, 15:04")
			}
			return ""
		},
		"timeSince": func(t time.Time) string {
			now := time.Now()
			diff := now.Sub(t)
			days := int(diff.Hours() / 24)
			hours := int(diff.Hours())
			minutes := int(diff.Minutes())
			if days > 0 {
				return fmt.Sprintf("%d hari lalu", days)
			}
			if hours > 0 {
				return fmt.Sprintf("%d jam lalu", hours)
			}
			if minutes > 0 {
				return fmt.Sprintf("%d menit lalu", minutes)
			}
			return "Baru saja"
		},
		"timeago": func(t time.Time) string {
			return time.Since(t).String()
		},
		"countdown": func(t time.Time) string {
			duration := time.Until(t)
			if duration < 0 {
				return "Sedang berlangsung"
			}
			days := int(duration.Hours() / 24)
			hours := int(duration.Hours()) % 24
			minutes := int(duration.Minutes()) % 60
			if days > 0 {
				return fmt.Sprintf("%d hari %d jam", days, hours)
			}
			return fmt.Sprintf("%d jam %d menit", hours, minutes)
		},

		// Helper
		"limit": func(n int, v interface{}) interface{} {
			s := reflect.ValueOf(v)
			if s.Kind() == reflect.Slice {
				if s.Len() > n {
					return s.Slice(0, n).Interface()
				}
			}
			return v
		},
	})

	router.Static("/static", "./static")
	
	// Load templates using glob pattern - this will use path as template name
	// e.g., templates/auth/login.html -> auth/login.html
	router.LoadHTMLGlob("templates/*/*")
	router.LoadHTMLGlob("../templates/*/*")

	web := router.Group("/")
	{
		// Public Routes (Login/Register)
		web.GET("/", webHandler.ShowLoginPage)
		web.POST("/login", webHandler.LoginWeb)

		web.GET("/register", webHandler.ShowRegisterPage)
		web.POST("/register", webHandler.RegisterWeb)

		web.GET("/logout", webHandler.LogoutWeb)

		// Student Routes
		student := web.Group("/student")
		student.Use(WebAuthMiddleware(string(domain.RoleStudent)))
		{
			student.GET("/dashboard", webHandler.StudentDashboard)
			student.GET("/courses", webHandler.StudentCourses)
			student.GET("/browse", webHandler.StudentBrowseCourses)
			student.GET("/labs", webHandler.StudentLabs)
			student.GET("/certificates", webHandler.StudentCertificates)
			student.GET("/profile", webHandler.StudentProfile)
			student.GET("/profile/edit", webHandler.StudentProfileEdit)
			student.POST("/profile/edit", webHandler.StudentProfileEdit)
		}

		// Instructor Routes
		instructor := web.Group("/instructor")
		instructor.Use(WebAuthMiddleware(string(domain.RoleInstructor)))
		{
			instructor.GET("/dashboard", webHandler.InstructorDashboard)
			instructor.GET("/courses", webHandler.InstructorAllCourses)
			instructor.GET("/courses/:id", webHandler.InstructorCourseDetail)
			instructor.GET("/labs", webHandler.InstructorLabs)
			instructor.GET("/certificates", webHandler.InstructorCertificates)
			instructor.GET("/students", webHandler.InstructorStudents)
		}

		// Admin Routes
		admin := web.Group("/admin")
		admin.Use(WebAuthMiddleware(string(domain.RoleAdmin)))
		{
			admin.GET("/dashboard", webHandler.AdminDashboard)
		}
	}
}

// Helper local
func toFloat(v interface{}) float64 {
	switch i := v.(type) {
	case int:
		return float64(i)
	case int64:
		return float64(i)
	case float64:
		return i
	default:
		return 0
	}
}
