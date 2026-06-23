package repository

import (
"context"
"strings"

"gorm.io/gorm"

"kslasbackend/internal/database/models"
)

func (r *TeachingRepository) AssignStaffRole(ctx context.Context, staffUserID uint, roleCode string, scopeType models.ScopeType, scopeID *uint, assignedBy uint) error {
return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
var user models.User
if err := tx.Where("id = ? AND user_type = ?", staffUserID, models.UserTypeStaff).First(&user).Error; err != nil {
return err
}

var role models.Role
if err := tx.Where("LOWER(code) = LOWER(?)", strings.TrimSpace(roleCode)).First(&role).Error; err != nil {
return err
}

if err := tx.Model(&models.UserRole{}).
Where("user_id = ? AND is_primary = ?", staffUserID, true).
UpdateColumn("is_primary", false).Error; err != nil {
return err
}

updateQuery := tx.Model(&models.UserRole{}).
Where("user_id = ? AND role_id = ? AND scope_type = ?", staffUserID, role.ID, scopeType)

if scopeID == nil {
updateQuery = updateQuery.Where("scope_id IS NULL")
} else {
updateQuery = updateQuery.Where("scope_id = ?", *scopeID)
}

result := updateQuery.UpdateColumns(map[string]any{
"is_primary":  true,
"assigned_by": assignedBy,
})
if result.Error != nil {
return result.Error
}
if result.RowsAffected > 0 {
return nil
}

return tx.Create(&models.UserRole{
UserID:     staffUserID,
RoleID:     role.ID,
ScopeType:  scopeType,
ScopeID:    scopeID,
IsPrimary:  true,
AssignedBy: &assignedBy,
}).Error
})
}
