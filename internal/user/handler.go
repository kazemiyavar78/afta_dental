package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
	"github.com/tpdenta/afta-reception/internal/platform/security/session"
)

// Handler کنترلر HTTP ماژول کاربران.
type Handler struct {
	service    *Service
	sessionSvc sessionDirectory
	csrfKey    string
	secure     bool
}

// NewHandler نمونه Handler می‌سازد.
func NewHandler(service *Service, sessionSvc *session.Service, csrfKey string, secure bool) *Handler {
	return &Handler{
		service:    service,
		sessionSvc: sessionSvc,
		csrfKey:    csrfKey,
		secure:     secure,
	}
}

// Login ورود کاربر.
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	user, sess, err := h.service.HandleUserLogin(c.Request.Context(), req, middleware.ActorFromContext(c))
	if err != nil {
		middleware.WriteError(c, err)
		return
	}

	middleware.SetSessionCookie(c, sess.ID.UUID(), h.secure)
	middleware.SetCSRFCookieOnLogin(c, sess.ID.String(), h.csrfKey, h.secure)

	c.JSON(http.StatusOK, LoginResponse{
		User:      *h.service.toResponse(user),
		SessionID: sess.ID.String(),
	})
}

// Logout خروج کاربر.
func (h *Handler) Logout(c *gin.Context) {
	sessionID, _ := c.Get(middleware.ContextKeySessionID)
	if sid, ok := sessionID.(uuid.UUID); ok {
		if err := h.sessionSvc.DeleteSession(c.Request.Context(), sid, middleware.ActorFromContext(c), true); err != nil {
			middleware.WriteError(c, err)
			return
		}
	}

	middleware.ClearSessionCookie(c, h.secure)
	c.SetCookie(middleware.CSRFCookieName, "", -1, "/", "", h.secure, false)

	c.JSON(http.StatusOK, gin.H{"message": "خروج موفق"})
}

// CreateUser ایجاد کاربر جدید.
func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	resp, err := h.service.CreateUser(c.Request.Context(), req, middleware.ActorFromContext(c))
	if err != nil {
		middleware.WriteError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetUser دریافت اطلاعات کاربر.
func (h *Handler) GetUser(c *gin.Context) {
	var uri struct {
		ID int `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		middleware.WriteError(c, err)
		return
	}

	canManage := middleware.IsAdmin(c) || HasPermission(middleware.GetPermissions(c), "users.read")

	resp, err := h.service.GetUserByID(c.Request.Context(), uri.ID, middleware.ActorFromContext(c), canManage)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListUsers لیست کاربران.
func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.service.ListUsers()
	if err != nil {
		middleware.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, users)
}

// ListDoctors لیست پزشکان فعال را برمی‌گرداند.
func (h *Handler) ListDoctors(c *gin.Context) {
	users, err := h.service.ListDoctors()
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, users)
}

// ListAssistants لیست دستیاران را برمی‌گرداند.
func (h *Handler) ListAssistants(c *gin.Context) {
	users, err := h.service.ListAssistants()
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, users)
}

// UpdateUser به‌روزرسانی کاربر.
func (h *Handler) UpdateUser(c *gin.Context) {
	var uri struct {
		ID int `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		middleware.WriteError(c, err)
		return
	}

	var req UpdateUserRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	canManage := middleware.IsAdmin(c) || HasPermission(middleware.GetPermissions(c), "users.update")

	resp, err := h.service.UpdateUser(c.Request.Context(), uri.ID, req, middleware.ActorFromContext(c), canManage)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ChangePassword تغییر رمز عبور.
func (h *Handler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	if err := h.service.ChangePassword(c.Request.Context(), middleware.ActorFromContext(c), req); err != nil {
		middleware.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "رمز عبور با موفقیت تغییر یافت"})
}

// GetMe اطلاعات کاربر فعلی.
func (h *Handler) GetMe(c *gin.Context) {
	userID, _ := c.Get(middleware.ContextKeyUserID)
	uid, _ := userID.(int)

	resp, err := h.service.GetMe(uid)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListRoles لیست کامل نقش‌ها (مدیریت).
func (h *Handler) ListRoles(c *gin.Context) {
	roles, err := h.service.ListRoles()
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, roles)
}

// ListAssignableRoles نقش‌های قابل انتصاب را برمی‌گرداند.
func (h *Handler) ListAssignableRoles(c *gin.Context) {
	roles, err := h.service.ListAssignableRoles()
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, roles)
}

// GetRole نقش را با شناسه برمی‌گرداند.
func (h *Handler) GetRole(c *gin.Context) {
	var uri struct {
		ID int `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		middleware.WriteError(c, err)
		return
	}
	resp, err := h.service.GetRole(uri.ID)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CreateRole نقش جدید ایجاد می‌کند.
func (h *Handler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if !middleware.BindJSON(c, &req) {
		return
	}
	resp, err := h.service.CreateRole(c.Request.Context(), req, middleware.ActorFromContext(c))
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// UpdateRole نقش را بروزرسانی می‌کند.
func (h *Handler) UpdateRole(c *gin.Context) {
	var uri struct {
		ID int `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		middleware.WriteError(c, err)
		return
	}
	var req UpdateRoleRequest
	if !middleware.BindJSON(c, &req) {
		return
	}
	resp, err := h.service.UpdateRole(c.Request.Context(), uri.ID, req, middleware.ActorFromContext(c))
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DeleteRole نقش را حذف می‌کند.
func (h *Handler) DeleteRole(c *gin.Context) {
	var uri struct {
		ID int `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		middleware.WriteError(c, err)
		return
	}
	if err := h.service.DeleteRole(c.Request.Context(), uri.ID, middleware.ActorFromContext(c)); err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListPermissions لیست مجوزهای سیستم را برمی‌گرداند.
func (h *Handler) ListPermissions(c *gin.Context) {
	perms, err := h.service.ListPermissions()
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, perms)
}

// UpdateSecuritySetting تغییر تنظیم امنیتی.
func (h *Handler) UpdateSecuritySetting(c *gin.Context) {
	var req UpdateSecuritySettingRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	if err := h.service.UpdateSecuritySetting(c.Request.Context(), req.ID, req.Value, middleware.ActorFromContext(c)); err != nil {
		middleware.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تنظیم با موفقیت به‌روزرسانی شد"})
}

// ListSecuritySettings لیست تنظیمات امنیتی.
func (h *Handler) ListSecuritySettings(c *gin.Context) {
	settings, err := h.service.GetSecuritySettings()
	if err != nil {
		middleware.WriteError(c, err)
		return
	}

	var resp []SecuritySettingResponse
	for _, s := range settings {
		resp = append(resp, SecuritySettingResponse{ID: s.ID, Name: s.SettingName, Value: s.SettingValue})
	}

	c.JSON(http.StatusOK, resp)
}

// GetUserProfile اطلاعات پروفایل کاربر لاگین‌شده و نشست‌های فعال را برمی‌گرداند.
func (h *Handler) GetUserProfile(c *gin.Context) {
	userID, _ := c.Get(middleware.ContextKeyUserID)
	uid, _ := userID.(int)

	userData, err := h.service.GetUserByID(c.Request.Context(), uid, middleware.ActorFromContext(c), true)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	sessions, err := h.service.GetUserSessions(uid)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}

	sessionResponses := make([]SessionResponse, 0, len(sessions))
	for _, s := range sessions {
		sessionResponses = append(sessionResponses, SessionResponse{
			ID:           s.ID.String(),
			Ip:           s.Ip,
			Browser:      s.Browser,
			CreationTime: s.CreationTime.Format("2006-01-02T15:04:05Z"),
			UserID:       s.PersonnelAccountID,
		})
	}

	c.JSON(http.StatusOK, UserProfileResponse{
		User:     *userData,
		Sessions: sessionResponses,
	})
}

// DeleteSession حذف نشست.
func (h *Handler) DeleteSession(c *gin.Context) {
	var uri struct {
		ID string `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		middleware.WriteError(c, err)
		return
	}

	sessionID, err := uuid.Parse(uri.ID)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}

	canManageOthers := middleware.IsAdmin(c) || HasPermission(middleware.GetPermissions(c), "users.listSessions")

	if err := h.sessionSvc.DeleteSession(c.Request.Context(), sessionID, middleware.ActorFromContext(c), canManageOthers); err != nil {
		middleware.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "نشست حذف شد"})
}
