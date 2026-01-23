package http

import (
	"net/http"

	"github.com/gin-gonic/gin/render"
)

// NoOpRenderer is a renderer that does nothing, useful for tests.
type NoOpRenderer struct{}

// Instance returns a gin.Render instance that does nothing.
func (r *NoOpRenderer) Instance(string, interface{}) render.Render {
	return &noOpRender{}
}

type noOpRender struct{}

// Render does nothing.
func (r *noOpRender) Render(http.ResponseWriter) error {
	return nil
}

// WriteContentType does nothing.
func (r *noOpRender) WriteContentType(w http.ResponseWriter) {
	// do nothing
}
