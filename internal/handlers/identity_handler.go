package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"kslasbackend/internal/dto"
)

type IdentityHandler struct {
	storageRoot string
}

func NewIdentityHandler(storageRoot string) *IdentityHandler {
	root := strings.TrimSpace(storageRoot)
	if root == "" {
		root = "storage/identity_enrollments"
	}
	return &IdentityHandler{storageRoot: root}
}

func (h *IdentityHandler) FaceEnrollment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	if err := r.ParseMultipartForm(96 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid identity enrollment upload")
		return
	}

	manifest := strings.TrimSpace(r.FormValue("manifest"))
	if manifest == "" {
		writeError(w, http.StatusBadRequest, "manifest is required")
		return
	}

	var req dto.FaceEnrollmentRequest
	if err := json.Unmarshal([]byte(manifest), &req); err != nil {
		writeError(w, http.StatusBadRequest, "manifest must be valid json")
		return
	}

	studentID := strings.TrimSpace(req.StudentID)
	if studentID == "" {
		studentID = strings.TrimSpace(r.FormValue("student_id"))
	}
	if studentID == "" {
		writeError(w, http.StatusBadRequest, "student_id is required")
		return
	}

	if len(req.Images) < 5 {
		writeError(w, http.StatusBadRequest, "at least five guided identity images are required")
		return
	}

	enrollmentID := fmt.Sprintf("face_%d", time.Now().UTC().UnixNano())
	enrollmentDir := filepath.Join(h.storageRoot, safePath(studentID), enrollmentID)
	if err := os.MkdirAll(enrollmentDir, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to prepare identity storage")
		return
	}

	storedImages, err := h.storeEnrollmentImages(enrollmentID, enrollmentDir, req.Images, r.MultipartForm)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if len(storedImages) < 5 {
		writeError(w, http.StatusBadRequest, "at least five identity image files must be uploaded")
		return
	}

	storedAt := time.Now().UTC().Format(time.RFC3339)
	response := dto.FaceEnrollmentResponse{
		EnrollmentID:            enrollmentID,
		StudentID:               studentID,
		Status:                  "submitted",
		Message:                 "Identity images submitted and stored for invigilator review.",
		CapturedAt:              req.CapturedAt,
		StoredAt:                storedAt,
		RequiredImages:          req.RequiredImages,
		UploadedImages:          len(storedImages),
		Purpose:                 defaultString(req.Purpose, "exam_identity_reference"),
		ReviewableByInvigilator: req.ReviewableByInvigilator,
		ManifestPath:            filepath.Join(enrollmentDir, "manifest.json"),
		Images:                  storedImages,
	}

	if response.RequiredImages == 0 {
		response.RequiredImages = len(req.Images)
	}
	if !response.ReviewableByInvigilator {
		response.ReviewableByInvigilator = true
	}

	if err := writeJSONFile(filepath.Join(enrollmentDir, "manifest.json"), req); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save identity manifest")
		return
	}
	if err := writeJSONFile(filepath.Join(enrollmentDir, "enrollment.json"), response); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save enrollment record")
		return
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *IdentityHandler) FaceEnrollments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	studentFilter := strings.TrimSpace(r.URL.Query().Get("student_id"))
	items := []dto.FaceEnrollmentResponse{}
	_ = filepath.WalkDir(h.storageRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || entry.Name() != "enrollment.json" {
			return nil
		}
		var item dto.FaceEnrollmentResponse
		if err := readJSONFile(path, &item); err != nil {
			return nil
		}
		if studentFilter != "" && item.StudentID != studentFilter {
			return nil
		}
		items = append(items, item)
		return nil
	})

	writeJSON(w, http.StatusOK, dto.FaceEnrollmentListResponse{Items: items, Count: len(items)})
}

func (h *IdentityHandler) FaceEnrollmentByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	enrollmentID := strings.TrimSpace(r.PathValue("enrollmentID"))
	item, ok := h.findEnrollment(enrollmentID)
	if !ok {
		writeError(w, http.StatusNotFound, "identity enrollment not found")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *IdentityHandler) FaceEnrollmentImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	enrollmentID := strings.TrimSpace(r.PathValue("enrollmentID"))
	fileName := filepath.Base(strings.TrimSpace(r.PathValue("fileName")))
	item, ok := h.findEnrollment(enrollmentID)
	if !ok {
		writeError(w, http.StatusNotFound, "identity enrollment not found")
		return
	}

	for _, image := range item.Images {
		if filepath.Base(image.StoredPath) == fileName {
			http.ServeFile(w, r, image.StoredPath)
			return
		}
	}
	writeError(w, http.StatusNotFound, "identity image not found")
}

func (h *IdentityHandler) storeEnrollmentImages(enrollmentID, enrollmentDir string, images []dto.FaceEnrollmentImageRequest, form *multipart.Form) ([]dto.FaceEnrollmentImageResponse, error) {
	stored := []dto.FaceEnrollmentImageResponse{}
	if form == nil {
		return stored, nil
	}

	for index, image := range images {
		field := strings.TrimSpace(image.Field)
		if field == "" {
			field = fmt.Sprintf("identity_image_%d", index+1)
		}
		headers := form.File[field]
		if len(headers) == 0 {
			continue
		}
		header := headers[0]
		storedName := fmt.Sprintf("%02d_%s_%s", index+1, safePath(image.PoseCode), filepath.Base(header.Filename))
		storedPath := filepath.Join(enrollmentDir, storedName)
		if err := saveUploadedFile(header, storedPath); err != nil {
			return nil, fmt.Errorf("failed to save identity image: %w", err)
		}
		stored = append(stored, dto.FaceEnrollmentImageResponse{
			Field:        field,
			PoseCode:     image.PoseCode,
			Title:        image.Title,
			Instruction:  image.Instruction,
			QualityScore: image.QualityScore,
			FileName:     filepath.Base(header.Filename),
			StoredPath:   storedPath,
			ViewURL:      fmt.Sprintf("/api/identity/face-enrollments/%s/images/%s", enrollmentID, storedName),
		})
	}
	return stored, nil
}

func (h *IdentityHandler) findEnrollment(enrollmentID string) (dto.FaceEnrollmentResponse, bool) {
	var found dto.FaceEnrollmentResponse
	if enrollmentID == "" {
		return found, false
	}
	_ = filepath.WalkDir(h.storageRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || entry.Name() != "enrollment.json" {
			return nil
		}
		var item dto.FaceEnrollmentResponse
		if err := readJSONFile(path, &item); err != nil {
			return nil
		}
		if item.EnrollmentID == enrollmentID {
			found = item
			return filepath.SkipAll
		}
		return nil
	})
	return found, found.EnrollmentID != ""
}

func saveUploadedFile(header *multipart.FileHeader, destination string) error {
	source, err := header.Open()
	if err != nil {
		return err
	}
	defer source.Close()

	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}

	target, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer target.Close()

	_, err = io.Copy(target, source)
	return err
}

func writeJSONFile(path string, payload any) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(payload)
}

func readJSONFile(path string, target any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewDecoder(file).Decode(target)
}

func safePath(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "/", "_")
	value = strings.ReplaceAll(value, "\\", "_")
	value = strings.ReplaceAll(value, " ", "_")
	cleaned := strings.Builder{}
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' || char == '-' {
			cleaned.WriteRune(char)
		}
	}
	if cleaned.Len() == 0 {
		return "unknown"
	}
	return cleaned.String()
}

func defaultString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
