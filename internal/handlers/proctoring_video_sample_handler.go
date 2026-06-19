package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"kslasbackend/internal/dto"
)

const randomVideoSampleStorageRoot = "storage/proctoring_review_clips"

func (h *ProctoringReviewHandler) RandomVideoSample(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	if err := r.ParseMultipartForm(96 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid random video sample upload")
		return
	}

	metadata := strings.TrimSpace(r.FormValue("metadata"))
	if metadata == "" {
		writeError(w, http.StatusBadRequest, "metadata is required")
		return
	}

	var req dto.RandomVideoSampleRequest
	if err := json.Unmarshal([]byte(metadata), &req); err != nil {
		writeError(w, http.StatusBadRequest, "metadata must be valid json")
		return
	}
	if strings.TrimSpace(req.AttemptID) == "" {
		req.AttemptID = strings.TrimSpace(r.FormValue("attempt_id"))
	}
	if strings.TrimSpace(req.ExamID) == "" {
		req.ExamID = strings.TrimSpace(r.FormValue("exam_id"))
	}
	if strings.TrimSpace(req.StudentID) == "" {
		req.StudentID = strings.TrimSpace(r.FormValue("student_id"))
	}
	if strings.TrimSpace(req.AttemptID) == "" || strings.TrimSpace(req.StudentID) == "" {
		writeError(w, http.StatusBadRequest, "student_id and attempt_id are required")
		return
	}

	files := r.MultipartForm.File["video_clip"]
	if len(files) == 0 {
		writeError(w, http.StatusBadRequest, "video_clip file is required")
		return
	}

	sampleID := fmt.Sprintf("review_clip_%d", time.Now().UTC().UnixNano())
	folder := filepath.Join(randomVideoSampleStorageRoot, safePath(req.AttemptID), sampleID)
	if err := os.MkdirAll(folder, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to prepare video storage")
		return
	}

	header := files[0]
	fileName := fmt.Sprintf("sample_%02d_%s", req.SampleNumber, filepath.Base(header.Filename))
	storedPath := filepath.Join(folder, fileName)
	if err := saveUploadedFile(header, storedPath); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save video sample")
		return
	}

	storedAt := time.Now().UTC().Format(time.RFC3339)
	response := dto.RandomVideoSampleResponse{
		SampleID:        sampleID,
		StudentID:       req.StudentID,
		ExamID:          req.ExamID,
		AttemptID:       req.AttemptID,
		SampleNumber:    req.SampleNumber,
		TotalSamples:    req.TotalSamples,
		DurationSeconds: req.DurationSeconds,
		CapturedAt:      req.CapturedAt,
		StoredAt:        storedAt,
		Purpose:         defaultString(req.Purpose, "invigilator_random_review"),
		ReviewTiming:    defaultString(req.ReviewTiming, "during_or_after_exam"),
		FileName:        filepath.Base(header.Filename),
		StoredPath:      storedPath,
		PlaybackURL:     fmt.Sprintf("/api/proctoring/random-video-samples/%s/file", sampleID),
		Status:          "stored",
		Message:         "Random review video sample stored for invigilator playback.",
	}
	if response.TotalSamples == 0 {
		response.TotalSamples = 5
	}
	if response.DurationSeconds == 0 {
		response.DurationSeconds = 10
	}

	if err := writeJSONFile(filepath.Join(folder, "metadata.json"), response); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save video metadata")
		return
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *ProctoringReviewHandler) RandomVideoSamples(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	studentFilter := strings.TrimSpace(r.URL.Query().Get("student_id"))
	attemptFilter := strings.TrimSpace(r.URL.Query().Get("attempt_id"))
	examFilter := strings.TrimSpace(r.URL.Query().Get("exam_id"))
	items := []dto.RandomVideoSampleResponse{}
	_ = filepath.WalkDir(randomVideoSampleStorageRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || entry.Name() != "metadata.json" {
			return nil
		}
		var item dto.RandomVideoSampleResponse
		if err := readJSONFile(path, &item); err != nil {
			return nil
		}
		if studentFilter != "" && item.StudentID != studentFilter {
			return nil
		}
		if attemptFilter != "" && item.AttemptID != attemptFilter {
			return nil
		}
		if examFilter != "" && item.ExamID != examFilter {
			return nil
		}
		items = append(items, item)
		return nil
	})

	writeJSON(w, http.StatusOK, dto.RandomVideoSampleListResponse{Items: items, Count: len(items)})
}

func (h *ProctoringReviewHandler) RandomVideoSampleByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	sampleID := strings.TrimSpace(r.PathValue("sampleID"))
	item, ok := findRandomVideoSample(sampleID)
	if !ok {
		writeError(w, http.StatusNotFound, "random video sample not found")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ProctoringReviewHandler) RandomVideoSampleFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	sampleID := strings.TrimSpace(r.PathValue("sampleID"))
	item, ok := findRandomVideoSample(sampleID)
	if !ok {
		writeError(w, http.StatusNotFound, "random video sample not found")
		return
	}
	http.ServeFile(w, r, item.StoredPath)
}

func findRandomVideoSample(sampleID string) (dto.RandomVideoSampleResponse, bool) {
	var found dto.RandomVideoSampleResponse
	if sampleID == "" {
		return found, false
	}
	_ = filepath.WalkDir(randomVideoSampleStorageRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || entry.Name() != "metadata.json" {
			return nil
		}
		var item dto.RandomVideoSampleResponse
		if err := readJSONFile(path, &item); err != nil {
			return nil
		}
		if item.SampleID == sampleID {
			found = item
			return filepath.SkipAll
		}
		return nil
	})
	return found, found.SampleID != ""
}
