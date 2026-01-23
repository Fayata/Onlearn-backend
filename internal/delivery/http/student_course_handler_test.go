package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"onlearn-backend/internal/domain"
	"onlearn-backend/internal/delivery/http"
	"onlearn-backend/internal/usecase"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCourseUsecase struct {
	mock.Mock
}

func (m *MockCourseUsecase) GetCourseDetails(ctx context.Context, courseID uint, userID *uint) (*domain.CourseDetail, error) {
	args := m.Called(ctx, courseID, userID)
	return args.Get(0).(*domain.CourseDetail), args.Error(1)
}

func (m *MockCourseUsecase) GetModulesWithProgress(ctx context.Context, userID uint, courseID uint) ([]domain.ModuleWithProgress, error) {
	args := m.Called(ctx, userID, courseID)
	return args.Get(0).([]domain.ModuleWithProgress), args.Error(1)
}

func (m *MockCourseUsecase) EnrollStudent(ctx context.Context, userID, courseID uint) error {
	args := m.Called(ctx, userID, courseID)
	return args.Error(0)
}

func (m *MockCourseUsecase) GetStudentEnrollments(ctx context.Context, userID uint) ([]domain.EnrollmentWithCourse, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.EnrollmentWithCourse), args.Error(1)
}

func (m *MockCourseUsecase) GetModuleByID(ctx context.Context, moduleID string) (*domain.Module, error) {
	args := m.Called(ctx, moduleID)
	return args.Get(0).(*domain.Module), args.Error(1)
}

func TestStudentCourseRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUsecase := new(MockCourseUsecase)
	
	// Setup router
	router := gin.Default()
	// No-op renderer for testing
	router.HTMLRender = http.NewNoOpRenderer()
	
	// Create a dummy user for context
	user := &domain.User{
		ID: 1,
		Role: "student",
	}

	// Middleware to set user in context
	router.Use(func(c *gin.Context) {
		c.Set("user", user)
		c.Next()
	})

	// Register routes
	http.NewStudentCourseHandler(router.Group("/student"), mockUsecase)

	t.Run("View Course Detail - Not Enrolled", func(t *testing.T) {
		// Mocking
		courseDetail := &domain.CourseDetail{
			Course: domain.Course{ID: 1, Title: "Test Course"},
			IsEnrolled: false,
		}
		mockUsecase.On("GetCourseDetails", mock.Anything, uint(1), &user.ID).Return(courseDetail, nil).Once()

		// Request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/student/courses/1", nil)
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		// Check that the response contains the course details, even if not enrolled
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		// This depends on how your handler formats the response. Let's assume it renders a template.
		// For API tests, you'd check the JSON body.
		// For HTML, we can check for non-redirect.
	})

	t.Run("Access Module Viewer - Not Enrolled", func(t *testing.T) {
		// Mocking
		// Simulate the use case check for enrollment.
		// In the real implementation, this check happens before calling GetModuleByID.
		// So, GetCourseDetails is the key.
		courseDetail := &domain.CourseDetail{
			Course: domain.Course{ID: 1, Title: "Test Course"},
			IsEnrolled: false, // Explicitly not enrolled
		}
		module := &domain.Module{ID: "module1", CourseID: 1}

		mockUsecase.On("GetCourseDetails", mock.Anything, uint(1), &user.ID).Return(courseDetail, nil).Once()
		// GetModuleByID should NOT be called if the logic is correct.
		// We'll set it up just in case the current logic calls it.
		mockUsecase.On("GetModuleByID", mock.Anything, "module1").Return(module, nil)


		// Request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/student/courses/1/modules/module1", nil)
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Contains(t, w.Header().Get("Location"), "/student/courses/1")
		// Ensure GetModuleByID was not called because the user is not enrolled.
		mockUsecase.AssertNotCalled(t, "GetModuleByID", mock.Anything, "module1")
	})

}

// Add empty methods to satisfy the interface
func (m *MockCourseUsecase) CreateCourse(ctx context.Context, course *domain.Course) error { return nil }
func (m *MockCourseUsecase) UpdateCourse(ctx context.Context, course *domain.Course) error { return nil }
func (m *MockCourseUsecase) DeleteCourse(ctx context.Context, id uint) error { return nil }
func (m *MockCourseUsecase) GetAllCourses(ctx context.Context) ([]domain.Course, error) { return nil, nil }
func (m *MockCourseUsecase) GetInstructorCourses(ctx context.Context, instructorID uint) ([]domain.Course, error) { return nil, nil }
func (m *MockCourseUsecase) AddModule(ctx context.Context, module *domain.Module) error { return nil }
func (m *MockCourseUsecase) UpdateModule(ctx context.Context, module *domain.Module) error { return nil }
func (m *MockCourseUsecase) DeleteModule(ctx context.Context, moduleID string) error { return nil }
func (m *MockCourseUsecase) GetCourseStudents(ctx context.Context, courseID uint) ([]domain.User, error) { return nil, nil }
func (m *MockCourseUsecase) MarkModuleComplete(ctx context.Context, userID uint, moduleID string, courseID uint) error { return nil }
func (m *MockCourseUsecase) SubmitAssignment(ctx context.Context, assignment *domain.Assignment) error { return nil }
func (m *MockCourseUsecase) GradeAssignment(ctx context.Context, assignmentID uint, grade float64, feedback string, gradedByID uint) error { return nil }
func (m *MockCourseUsecase) GetCourseAssignments(ctx context.Context, courseID uint) ([]domain.Assignment, error) { return nil, nil }
