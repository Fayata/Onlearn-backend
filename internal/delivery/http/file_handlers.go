package http

import (
	"fmt"
	"io"
	"net/http"
	"onlearn-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

// FileHandler handles file operations with GridFS
type FileHandler struct {
	gridFS repository.GridFSRepository
}

// NewFileHandler creates a new FileHandler
func NewFileHandler(gridFS repository.GridFSRepository) *FileHandler {
	return &FileHandler{gridFS: gridFS}
}

// UploadFile handles file upload to GridFS
func (h *FileHandler) UploadFile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}
	defer file.Close()

	// Validate file size (50MB max)
	if header.Size > repository.MaxLargeFileSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Ukuran file terlalu besar. Maksimal %dMB", repository.MaxLargeFileSize/(1024*1024)),
		})
		return
	}

	metadata := repository.FileMetadata{
		UploadedBy: userID.(uint),
	}

	fileInfo, err := h.gridFS.Upload(c.Request.Context(), file, header, metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File uploaded successfully",
		"file": gin.H{
			"id":           fileInfo.ID,
			"filename":     fileInfo.Filename,
			"content_type": fileInfo.ContentType,
			"size":         fileInfo.Size,
			"upload_date":  fileInfo.UploadDate,
		},
	})
}

// StreamFile streams a file from GridFS
func (h *FileHandler) StreamFile(c *gin.Context) {
	fileID := c.Param("id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File ID is required"})
		return
	}

	stream, fileInfo, err := h.gridFS.Download(c.Request.Context(), fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	defer stream.Close()

	// Set headers for streaming
	c.Header("Content-Type", fileInfo.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size))
	
	// For PDF and PPT, allow inline viewing
	if fileInfo.ContentType == "application/pdf" ||
		fileInfo.ContentType == "application/vnd.ms-powerpoint" ||
		fileInfo.ContentType == "application/vnd.openxmlformats-officedocument.presentationml.presentation" {
		c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", fileInfo.Metadata.OriginalName))
	} else {
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileInfo.Metadata.OriginalName))
	}

	// Enable CORS for viewer
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Expose-Headers", "Content-Disposition, Content-Length")

	// Stream the file
	c.Status(http.StatusOK)
	_, err = io.Copy(c.Writer, stream)
	if err != nil {
		// Log error but don't send response since headers are already sent
		fmt.Printf("Error streaming file: %v\n", err)
	}
}

// GetFileInfo returns file metadata
func (h *FileHandler) GetFileInfo(c *gin.Context) {
	fileID := c.Param("id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File ID is required"})
		return
	}

	fileInfo, err := h.gridFS.GetFileInfo(c.Request.Context(), fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":            fileInfo.ID,
		"filename":      fileInfo.Filename,
		"original_name": fileInfo.Metadata.OriginalName,
		"content_type":  fileInfo.ContentType,
		"size":          fileInfo.Size,
		"upload_date":   fileInfo.UploadDate,
		"file_type":     fileInfo.Metadata.FileType,
	})
}

// DeleteFile deletes a file from GridFS
func (h *FileHandler) DeleteFile(c *gin.Context) {
	fileID := c.Param("id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File ID is required"})
		return
	}

	err := h.gridFS.Delete(c.Request.Context(), fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}

// UploadModuleFile handles file upload for modules and returns file ID
func (h *FileHandler) UploadModuleFile(c *gin.Context, formFieldName string, userID uint, courseID uint) (string, error) {
	file, header, err := c.Request.FormFile(formFieldName)
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil // No file uploaded, return empty
		}
		return "", err
	}
	defer file.Close()

	// Validate file size
	if header.Size > repository.MaxLargeFileSize {
		return "", fmt.Errorf("ukuran file terlalu besar. Maksimal %dMB", repository.MaxLargeFileSize/(1024*1024))
	}

	metadata := repository.FileMetadata{
		UploadedBy: userID,
		CourseID:   courseID,
	}

	fileInfo, err := h.gridFS.Upload(c.Request.Context(), file, header, metadata)
	if err != nil {
		return "", err
	}

	return fileInfo.ID, nil
}
