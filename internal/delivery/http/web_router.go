package http

import (
	"fmt"
	"html/template"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
)

// Helper to safely convert various number types to float64
func toFloat(v interface{}) float64 {
	switch i := v.(type) {
	case int:
		return float64(i)
	case int8:
		return float64(i)
	case int16:
		return float64(i)
	case int32:
		return float64(i)
	case int64:
		return float64(i)
	case float32:
		return float64(i)
	case float64:
		return i
	default:
		return 0
	}
}

func InitWebRouter(router *gin.Engine, webHandler *WebHandler) {
	// Register custom template functions BEFORE loading HTML glob
	router.SetFuncMap(template.FuncMap{
		// Arithmetic functions (Pipeline: val | div arg => div(arg, val))
		"div": func(b, a interface{}) float64 {
			valA := toFloat(a)
			valB := toFloat(b)
			if valB == 0 {
				return 0
			}
			return valA / valB
		},
		"mul": func(b, a interface{}) float64 {
			return toFloat(a) * toFloat(b)
		},
		"sub": func(b, a interface{}) float64 {
			return toFloat(a) - toFloat(b)
		},
		"add": func(b, a interface{}) float64 {
			return toFloat(a) + toFloat(b)
		},

		// Date/Time formatting
		"timeago": func(t time.Time) string {
			duration := time.Since(t)
			if duration.Seconds() < 60 {
				return "baru saja"
			}
			if duration.Minutes() < 60 {
				return fmt.Sprintf("%.0f menit yang lalu", duration.Minutes())
			}
			if duration.Hours() < 24 {
				return fmt.Sprintf("%.0f jam yang lalu", duration.Hours())
			}
			if duration.Hours() < 48 {
				return "kemarin"
			}
			return t.Format("2 Jan 2006")
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

		// Slice utilities
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
	router.LoadHTMLGlob("templates/**/*.html")

	// Web routes
	web := router.Group("/")
	{
		web.GET("/", webHandler.ShowLoginPage)
		web.GET("/register", func(c *gin.Context) {
			// Basic register page rendering if needed,
			// otherwise you can remove this or point to a register.html
			c.HTML(200, "login.html", gin.H{"title": "Register"})
		})
		web.GET("/student/dashboard", webHandler.StudentDashboard)
		web.GET("/instructor/dashboard", webHandler.InstructorDashboard)
		web.GET("/admin/dashboard", webHandler.AdminDashboard)
	}
}
