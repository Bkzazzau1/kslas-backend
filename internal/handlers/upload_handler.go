package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type uploadResponse struct {
	FileName string `json:"file_name"`
	FileURL  string `json:"file_url"`
	Size     int64  `json:"size"`
}

func uploadRoot() string {
	root := os.Getenv("UPLOAD_ROOT")
	if root == "" {
		root = "uploads"
	}
	return root
}

func (h *AssessmentHandler) uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := r.ParseMultipartForm(25 << 20); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file field is required")
		return
	}
	defer file.Close()

	category := cleanPathPart(r.FormValue("category"))
	if category == "" {
		category = "general"
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = ".bin"
	}
	fileName := fmt.Sprintf("%s-%d%s", uuid.New().String(), time.Now().UnixNano(), ext)
	dir := filepath.Join(uploadRoot(), category)
	if err := os.MkdirAll(dir, 0755); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	destinationPath := filepath.Join(dir, fileName)
	destination, err := os.Create(destinationPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer destination.Close()

	size, err := io.Copy(destination, file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, uploadResponse{
		FileName: fileName,
		FileURL:  "/uploads/" + category + "/" + fileName,
		Size:     size,
	})
}

func cleanPathPart(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, "..", "")
	value = strings.ReplaceAll(value, "/", "-")
	value = strings.ReplaceAll(value, "\\", "-")
	value = strings.ReplaceAll(value, " ", "-")
	return value
}
