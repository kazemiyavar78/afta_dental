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
	sessionSvc *session.Service
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
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err)
		return
	}

	user, sess, err := h.service.HandleUserLogin(req, c.ClientIP(), c.GetHeader("User-Agent"))
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
	userID, _ := c.Get(middleware.ContextKeyUserID)
	uid, _ := userID.(int)

	sessionID, _ := c.Get(middleware.ContextKeySessionID)
	if sid, ok := sessionID.(uuid.UUID); ok {
		_ = h.sessionSvc.DeleteSession(sid, uid, "", c.ClientIP())
	}

	_ = h.service.HandleLogout("", uid, c.ClientIP())
	middleware.ClearSessionCookie(c, h.secure)
	c.SetCookie(middleware.CSRFCookieName, "", -1, "/", "", h.secure, false)

	c.JSON(http.StatusOK, gin.H{"message": "خروج موفق"})
}

// CreateUser ایجاد کاربر جدید.
func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err)
		return
	}

	actorID, _ := c.Get(middleware.ContextKeyUserID)
	uid, _ := actorID.(int)

	resp, err := h.service.CreateUser(req, uid, c.ClientIP())
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

	actorID, _ := c.Get(middleware.ContextKeyUserID)
	uid, _ := actorID.(int)
	role, _ := c.Get(middleware.ContextKeyRoleName)
	roleName, _ := role.(string)

	resp, err := h.service.GetUserByID(uri.ID, uid, roleName)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListUsers لیست کاربران.
func (h *Handler) ListUsers(c *gin.Context) {
	role, _ := c.Get(middleware.ContextKeyRoleName)
	roleName, _ := role.(string)

	users, err := h.service.ListUsers(roleName)
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
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err)
		return
	}

	actorID, _ := c.Get(middleware.ContextKeyUserID)
	uid, _ := actorID.(int)
	role, _ := c.Get(middleware.ContextKeyRoleName)
	roleName, _ := role.(string)

	resp, err := h.service.UpdateUser(uri.ID, req, uid, roleName, c.ClientIP())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ChangePassword تغییر رمز عبور.
func (h *Handler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err)
		return
	}

	userID, _ := c.Get(middleware.ContextKeyUserID)
	uid, _ := userID.(int)

	if err := h.service.ChangePassword(uid, req, c.ClientIP()); err != nil {
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

// ListRoles لیست نقش‌ها.
func (h *Handler) ListRoles(c *gin.Context) {
	roles, err := h.service.ListRoles()
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, roles)
}

// UpdateSecuritySetting تغییر تنظیم امنیتی.
func (h *Handler) UpdateSecuritySetting(c *gin.Context) {
	var req UpdateSecuritySettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err)
		return
	}

	actorID, _ := c.Get(middleware.ContextKeyUserID)
	uid, _ := actorID.(int)
	role, _ := c.Get(middleware.ContextKeyRoleName)
	roleName, _ := role.(string)

	if err := h.service.UpdateSecuritySetting(req.Name, req.Value, uid, roleName, c.ClientIP()); err != nil {
		middleware.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تنظیم با موفقیت به‌روزرسانی شد"})
}

// ListSecuritySettings لیست تنظیمات امنیتی.
func (h *Handler) ListSecuritySettings(c *gin.Context) {
	role, _ := c.Get(middleware.ContextKeyRoleName)
	roleName, _ := role.(string)

	settings, err := h.service.GetSecuritySettings(roleName)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}

	var resp []SecuritySettingResponse
	for _, s := range settings {
		resp = append(resp, SecuritySettingResponse{Name: s.SettingName, Value: s.SettingValue})
	}

	c.JSON(http.StatusOK, resp)
}

// GetUserProfile اطلاعات پروفایل کاربر لاگین‌شده و نشست‌های فعال را برمی‌گرداند.
func (h *Handler) GetUserProfile(c *gin.Context) {
	userID, _ := c.Get(middleware.ContextKeyUserID)
	uid, _ := userID.(int)

	userData, err := h.service.GetUserByID(uid, uid, "")
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

// ListSessions لیست نشست‌ها.
func (h *Handler) ListSessions(c *gin.Context) {
	actorID, _ := c.Get(middleware.ContextKeyUserID)
	uid, _ := actorID.(int)
	role, _ := c.Get(middleware.ContextKeyRoleName)
	roleName, _ := role.(string)

	sessions, err := h.sessionSvc.ListSessions(uid, roleName)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}


	var resp []SessionResponse
	for _, s := range sessions {
		resp = append(resp, SessionResponse{
			ID:           s.ID.String(),
			Ip:           s.Ip,
			Browser:      s.Browser,
			CreationTime: s.CreationTime.Format("2006-01-02T15:04:05Z"),
			UserID:       s.PersonnelAccountID,
		})
	}

	c.JSON(http.StatusOK, resp)
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

	actorID, _ := c.Get(middleware.ContextKeyUserID)
	uid, _ := actorID.(int)
	role, _ := c.Get(middleware.ContextKeyRoleName)
	roleName, _ := role.(string)

	if err := h.sessionSvc.DeleteSession(sessionID, uid, roleName, c.ClientIP()); err != nil {
		middleware.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "نشست حذف شد"})
}
