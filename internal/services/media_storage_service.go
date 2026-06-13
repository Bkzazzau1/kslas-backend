package services

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type StoredFile struct {
	StoragePath      string
	OriginalFileName string
	StoredFileName   string
	MimeType         string
	SizeBytes        int64
}

type SaveFileOptions struct {
	SubDir            string
	AllowedExtensions map[string]struct{}
	RequireVideo      bool
}

type MediaStorageService struct {
	baseDir string
}

func NewMediaStorageService(baseDir string) *MediaStorageService {
	absBaseDir, err := filepath.Abs(strings.TrimSpace(baseDir))
	if err != nil || absBaseDir == "" {
		absBaseDir = strings.TrimSpace(baseDir)
	}

	return &MediaStorageService{baseDir: absBaseDir}
}

func (s *MediaStorageService) Save(originalFileName string, reader io.Reader, options SaveFileOptions) (*StoredFile, error) {
	originalFileName = strings.TrimSpace(originalFileName)
	if originalFileName == "" {
		return nil, errors.New("original file name is required")
	}

	ext := strings.ToLower(filepath.Ext(originalFileName))
	if ext == "" {
		return nil, errors.New("file extension is required")
	}

	if len(options.AllowedExtensions) > 0 {
		if _, ok := options.AllowedExtensions[ext]; !ok {
			return nil, fmt.Errorf("unsupported file type %q", ext)
		}
	}

	targetDir, err := s.targetDir(options.SubDir)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(targetDir, 0o750); err != nil {
		return nil, fmt.Errorf("create storage directory: %w", err)
	}

	storedFileName := uuid.NewString() + ext
	fullPath := filepath.Join(targetDir, storedFileName)

	file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o640)
	if err != nil {
		return nil, fmt.Errorf("create storage file: %w", err)
	}
	defer file.Close()

	header := make([]byte, 512)
	readBytes, readErr := io.ReadFull(reader, header)
	switch {
	case readErr == nil:
	case errors.Is(readErr, io.EOF), errors.Is(readErr, io.ErrUnexpectedEOF):
	default:
		return nil, fmt.Errorf("read file header: %w", readErr)
	}

	mimeType := http.DetectContentType(header[:readBytes])
	if options.RequireVideo && !isAllowedVideoContentType(mimeType, ext) {
		return nil, errors.New("uploaded file must be a supported video format")
	}

	written, err := io.Copy(file, io.MultiReader(strings.NewReader(string(header[:readBytes])), reader))
	if err != nil {
		_ = os.Remove(fullPath)
		return nil, fmt.Errorf("write storage file: %w", err)
	}

	if written == 0 {
		_ = os.Remove(fullPath)
		return nil, errors.New("uploaded file is empty")
	}

	relativePath, err := filepath.Rel(s.baseDir, fullPath)
	if err != nil {
		_ = os.Remove(fullPath)
		return nil, fmt.Errorf("resolve storage path: %w", err)
	}

	return &StoredFile{
		StoragePath:      filepath.ToSlash(relativePath),
		OriginalFileName: originalFileName,
		StoredFileName:   storedFileName,
		MimeType:         mimeType,
		SizeBytes:        written,
	}, nil
}

func (s *MediaStorageService) Open(storagePath string) (*os.File, os.FileInfo, error) {
	fullPath, err := s.resolve(storagePath)
	if err != nil {
		return nil, nil, err
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open storage file: %w", err)
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, nil, fmt.Errorf("stat storage file: %w", err)
	}

	return file, info, nil
}

func (s *MediaStorageService) targetDir(subDir string) (string, error) {
	parts := []string{s.baseDir}
	subDir = strings.TrimSpace(subDir)
	if subDir != "" {
		cleanSubDir := filepath.Clean(subDir)
		if cleanSubDir == "." || strings.HasPrefix(cleanSubDir, "..") {
			return "", errors.New("invalid storage directory")
		}
		parts = append(parts, cleanSubDir)
	}

	now := time.Now().UTC()
	parts = append(parts, now.Format("2006"), now.Format("01"))
	return filepath.Join(parts...), nil
}

func (s *MediaStorageService) resolve(storagePath string) (string, error) {
	if strings.TrimSpace(storagePath) == "" {
		return "", errors.New("storage path is required")
	}

	fullPath := filepath.Join(s.baseDir, filepath.Clean(storagePath))
	relPath, err := filepath.Rel(s.baseDir, fullPath)
	if err != nil {
		return "", fmt.Errorf("resolve storage file: %w", err)
	}

	if strings.HasPrefix(relPath, "..") {
		return "", errors.New("invalid storage path")
	}

	return fullPath, nil
}

func isAllowedVideoContentType(mimeType, ext string) bool {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(mimeType)), "video/") {
		return true
	}

	switch strings.ToLower(strings.TrimSpace(ext)) {
	case ".mp4", ".m4v", ".mov", ".webm", ".mkv":
		return mimeType == "application/octet-stream" || mimeType == "binary/octet-stream"
	default:
		return false
	}
}
