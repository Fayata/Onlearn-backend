package repository

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// MaxFileSize adalah batas maksimal ukuran file (16MB untuk GridFS chunk optimal)
	MaxFileSize = 16 * 1024 * 1024 // 16MB
	// MaxLargeFileSize untuk file yang lebih besar (50MB)
	MaxLargeFileSize = 50 * 1024 * 1024 // 50MB
)

// FileInfo menyimpan metadata file yang diupload
type FileInfo struct {
	ID          string    `json:"id" bson:"_id"`
	Filename    string    `json:"filename" bson:"filename"`
	ContentType string    `json:"content_type" bson:"contentType"`
	Size        int64     `json:"size" bson:"length"`
	UploadDate  time.Time `json:"upload_date" bson:"uploadDate"`
	Metadata    FileMetadata `json:"metadata" bson:"metadata"`
}

// FileMetadata menyimpan metadata tambahan
type FileMetadata struct {
	OriginalName string `json:"original_name" bson:"original_name"`
	UploadedBy   uint   `json:"uploaded_by" bson:"uploaded_by"`
	FileType     string `json:"file_type" bson:"file_type"` // pdf, ppt, pptx, image
	CourseID     uint   `json:"course_id,omitempty" bson:"course_id,omitempty"`
	ModuleID     string `json:"module_id,omitempty" bson:"module_id,omitempty"`
}

// GridFSRepository interface untuk operasi file
type GridFSRepository interface {
	Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader, metadata FileMetadata) (*FileInfo, error)
	Download(ctx context.Context, fileID string) (io.ReadCloser, *FileInfo, error)
	Delete(ctx context.Context, fileID string) error
	GetFileInfo(ctx context.Context, fileID string) (*FileInfo, error)
}

type gridFSRepo struct {
	db     *mongo.Database
	bucket *gridfs.Bucket
}

// NewGridFSRepository membuat instance baru GridFS repository
func NewGridFSRepository(db *mongo.Database) (GridFSRepository, error) {
	bucket, err := gridfs.NewBucket(db, options.GridFSBucket().SetName("uploads"))
	if err != nil {
		return nil, fmt.Errorf("failed to create GridFS bucket: %w", err)
	}
	return &gridFSRepo{db: db, bucket: bucket}, nil
}

// Upload mengupload file ke GridFS
func (r *gridFSRepo) Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader, metadata FileMetadata) (*FileInfo, error) {
	// Validasi ukuran file
	if header.Size > MaxLargeFileSize {
		return nil, fmt.Errorf("ukuran file terlalu besar. Maksimal %dMB", MaxLargeFileSize/(1024*1024))
	}

	// Deteksi content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = detectContentType(header.Filename)
	}

	// Validasi tipe file
	if !isAllowedFileType(contentType, header.Filename) {
		return nil, errors.New("tipe file tidak diizinkan. Hanya PDF, PPT, PPTX, dan gambar yang diperbolehkan")
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	uniqueFilename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), generateRandomString(8), ext)

	// Set metadata
	metadata.OriginalName = header.Filename
	metadata.FileType = getFileType(header.Filename)

	uploadOpts := options.GridFSUpload().SetMetadata(bson.M{
		"original_name": metadata.OriginalName,
		"uploaded_by":   metadata.UploadedBy,
		"file_type":     metadata.FileType,
		"course_id":     metadata.CourseID,
		"module_id":     metadata.ModuleID,
		"content_type":  contentType,
	})

	// Upload ke GridFS
	objectID, err := r.bucket.UploadFromStream(uniqueFilename, file, uploadOpts)
	if err != nil {
		return nil, fmt.Errorf("gagal upload file: %w", err)
	}

	return &FileInfo{
		ID:          objectID.Hex(),
		Filename:    uniqueFilename,
		ContentType: contentType,
		Size:        header.Size,
		UploadDate:  time.Now(),
		Metadata:    metadata,
	}, nil
}

// Download mengunduh file dari GridFS
func (r *gridFSRepo) Download(ctx context.Context, fileID string) (io.ReadCloser, *FileInfo, error) {
	objectID, err := primitive.ObjectIDFromHex(fileID)
	if err != nil {
		return nil, nil, errors.New("invalid file ID")
	}

	// Get file info first
	fileInfo, err := r.GetFileInfo(ctx, fileID)
	if err != nil {
		return nil, nil, err
	}

	// Open download stream
	stream, err := r.bucket.OpenDownloadStream(objectID)
	if err != nil {
		return nil, nil, fmt.Errorf("file tidak ditemukan: %w", err)
	}

	return stream, fileInfo, nil
}

// Delete menghapus file dari GridFS
func (r *gridFSRepo) Delete(ctx context.Context, fileID string) error {
	objectID, err := primitive.ObjectIDFromHex(fileID)
	if err != nil {
		return errors.New("invalid file ID")
	}

	err = r.bucket.Delete(objectID)
	if err != nil {
		return fmt.Errorf("gagal menghapus file: %w", err)
	}

	return nil
}

// GetFileInfo mendapatkan informasi file
func (r *gridFSRepo) GetFileInfo(ctx context.Context, fileID string) (*FileInfo, error) {
	objectID, err := primitive.ObjectIDFromHex(fileID)
	if err != nil {
		return nil, errors.New("invalid file ID")
	}

	// Query files collection
	collection := r.db.Collection("uploads.files")
	
	var result struct {
		ID         primitive.ObjectID `bson:"_id"`
		Filename   string             `bson:"filename"`
		Length     int64              `bson:"length"`
		UploadDate time.Time          `bson:"uploadDate"`
		Metadata   bson.M             `bson:"metadata"`
	}

	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("file tidak ditemukan")
		}
		return nil, err
	}

	// Extract metadata
	metadata := FileMetadata{}
	if result.Metadata != nil {
		if v, ok := result.Metadata["original_name"].(string); ok {
			metadata.OriginalName = v
		}
		if v, ok := result.Metadata["uploaded_by"].(int64); ok {
			metadata.UploadedBy = uint(v)
		} else if v, ok := result.Metadata["uploaded_by"].(int32); ok {
			metadata.UploadedBy = uint(v)
		}
		if v, ok := result.Metadata["file_type"].(string); ok {
			metadata.FileType = v
		}
		if v, ok := result.Metadata["course_id"].(int64); ok {
			metadata.CourseID = uint(v)
		} else if v, ok := result.Metadata["course_id"].(int32); ok {
			metadata.CourseID = uint(v)
		}
		if v, ok := result.Metadata["module_id"].(string); ok {
			metadata.ModuleID = v
		}
	}

	contentType := ""
	if result.Metadata != nil {
		if v, ok := result.Metadata["content_type"].(string); ok {
			contentType = v
		}
	}
	if contentType == "" {
		contentType = detectContentType(result.Filename)
	}

	return &FileInfo{
		ID:          result.ID.Hex(),
		Filename:    result.Filename,
		ContentType: contentType,
		Size:        result.Length,
		UploadDate:  result.UploadDate,
		Metadata:    metadata,
	}, nil
}

// Helper functions

func detectContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".ppt":
		return "application/vnd.ms-powerpoint"
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

func isAllowedFileType(contentType, filename string) bool {
	allowedTypes := map[string]bool{
		"application/pdf": true,
		"application/vnd.ms-powerpoint": true,
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	if allowedTypes[contentType] {
		return true
	}

	// Check by extension as fallback
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExts := map[string]bool{
		".pdf":  true,
		".ppt":  true,
		".pptx": true,
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}

	return allowedExts[ext]
}

func getFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".pdf":
		return "pdf"
	case ".ppt", ".pptx":
		return "ppt"
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return "image"
	default:
		return "other"
	}
}

func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(1 * time.Nanosecond)
	}
	return string(b)
}
