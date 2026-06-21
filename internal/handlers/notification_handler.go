package handlers

import (
	"net/http"
	"strings"

	"github.com/google/uuid"

	"kslasbackend/internal/middleware"
	"kslasbackend/internal/models"
)

type notificationCreatePayload struct {
	RecipientID  uuid.UUID  `json:"recipient_id"`
	SenderID     *uuid.UUID `json:"sender_id"`
	Title        string     `json:"title"`
	Message      string     `json:"message"`
	Category     string     `json:"category"`
	Priority     string     `json:"priority"`
	ActionURL    string     `json:"action_url"`
	ResourceType string     `json:"resource_type"`
	ResourceID   *uuid.UUID `json:"resource_id"`
}

type notificationCountResponse struct { Unread int64 `json:"unread"` }

func (h *AssessmentHandler) listNotifications(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.StaffClaimsFromRequest(r)
	if !ok { writeError(w, http.StatusUnauthorized, "staff authentication is required"); return }
	query := h.db.Where("recipient_id = ?", claims.ID).Order("created_at desc")
	if unread := r.URL.Query().Get("unread"); unread == "true" { query = query.Where("read_at IS NULL") }
	if category := r.URL.Query().Get("category"); category != "" { query = query.Where("category = ?", category) }
	var notifications []models.Notification
	if err := query.Limit(100).Find(&notifications).Error; err != nil { writeError(w, http.StatusInternalServerError, err.Error()); return }
	writeJSON(w, http.StatusOK, notifications)
}

func (h *AssessmentHandler) notificationUnreadCount(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.StaffClaimsFromRequest(r)
	if !ok { writeError(w, http.StatusUnauthorized, "staff authentication is required"); return }
	var count int64
	h.db.Model(&models.Notification{}).Where("recipient_id = ? AND read_at IS NULL", claims.ID).Count(&count)
	writeJSON(w, http.StatusOK, notificationCountResponse{Unread: count})
}

func (h *AssessmentHandler) markNotificationRead(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.StaffClaimsFromRequest(r)
	if !ok { writeError(w, http.StatusUnauthorized, "staff authentication is required"); return }
	idText := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/notifications/"), "/")
	idText = strings.TrimSuffix(idText, "/read")
	idText = strings.Trim(idText, "/")
	id, err := uuid.Parse(idText)
	if err != nil { writeError(w, http.StatusBadRequest, "invalid notification id"); return }
	var notification models.Notification
	if err := h.db.First(&notification, "id = ? AND recipient_id = ?", id, claims.ID).Error; err != nil { writeError(w, http.StatusNotFound, "notification not found"); return }
	notification.ReadAt = nowPtr()
	if err := h.db.Save(&notification).Error; err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	writeJSON(w, http.StatusOK, notification)
}

func (h *AssessmentHandler) markAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.StaffClaimsFromRequest(r)
	if !ok { writeError(w, http.StatusUnauthorized, "staff authentication is required"); return }
	now := nowPtr()
	if err := h.db.Model(&models.Notification{}).Where("recipient_id = ? AND read_at IS NULL", claims.ID).Update("read_at", now).Error; err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AssessmentHandler) createNotificationAPI(w http.ResponseWriter, r *http.Request) {
	if !h.requireAnyRole(w, r, "admin", "dlc_director", "hod", "exam_officer") { return }
	var payload notificationCreatePayload
	if err := decodeJSON(w, r, &payload); err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	if payload.RecipientID == uuid.Nil || strings.TrimSpace(payload.Title) == "" { writeError(w, http.StatusBadRequest, "recipient_id and title are required"); return }
	notification := models.Notification{RecipientID: payload.RecipientID, SenderID: payload.SenderID, Title: payload.Title, Message: payload.Message, Category: payload.Category, Priority: payload.Priority, ActionURL: payload.ActionURL, ResourceType: payload.ResourceType, ResourceID: payload.ResourceID}
	if err := h.db.Create(&notification).Error; err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	writeJSON(w, http.StatusCreated, notification)
}

func (h *AssessmentHandler) createNotification(recipientID uuid.UUID, senderID *uuid.UUID, title string, message string, category string, priority string, actionURL string, resourceType string, resourceID *uuid.UUID) {
	if recipientID == uuid.Nil || strings.TrimSpace(title) == "" { return }
	notification := models.Notification{RecipientID: recipientID, SenderID: senderID, Title: title, Message: message, Category: category, Priority: priority, ActionURL: actionURL, ResourceType: resourceType, ResourceID: resourceID}
	_ = h.db.Create(&notification).Error
}

func (h *AssessmentHandler) notifyStaffRole(role string, departmentID *uuid.UUID, senderID *uuid.UUID, title string, message string, category string, priority string, actionURL string, resourceType string, resourceID *uuid.UUID) {
	recipients := map[uuid.UUID]bool{}
	var staff []models.Staff
	staffQuery := h.db.Where("primary_role = ? AND is_active = ?", role, true)
	if departmentID != nil { staffQuery = staffQuery.Where("department_id = ?", *departmentID) }
	staffQuery.Find(&staff)
	for _, person := range staff { recipients[person.ID] = true }

	var roles []models.StaffRoleAssignment
	roleQuery := h.db.Where("role = ? AND is_active = ?", role, true)
	if departmentID != nil { roleQuery = roleQuery.Where("department_id = ? OR department_id IS NULL", *departmentID) }
	roleQuery.Find(&roles)
	for _, assignment := range roles { recipients[assignment.StaffID] = true }

	for recipientID := range recipients {
		h.createNotification(recipientID, senderID, title, message, category, priority, actionURL, resourceType, resourceID)
	}
}

func (h *AssessmentHandler) courseDepartmentID(courseID uuid.UUID) *uuid.UUID {
	var course models.Course
	if err := h.db.First(&course, "id = ?", courseID).Error; err != nil { return nil }
	if course.DepartmentID == uuid.Nil { return nil }
	return &course.DepartmentID
}
