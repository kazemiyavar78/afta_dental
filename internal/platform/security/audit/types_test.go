package audit

import "testing"

// TestEventCatalog_KnownTypesHaveLabel انواع جاری و منسوخ باید برچسب و سطح اهمیت داشته باشند.
func TestEventCatalog_KnownTypesHaveLabel(t *testing.T) {
	active := []string{
		EventSessionCreated, EventSessionTerminated, EventLogout, EventLoginBlocked,
		EventIPLocked, EventAccountLocked, EventUserCreated, EventUserUpdated,
		EventPasswordChanged, EventRoleCreated, EventRoleUpdated, EventRoleDeleted,
		EventSettingChanged, EventSettingRemoved, EventAuditStarted, EventAuditStopped,
		EventAuditRead, EventAuditReadDenied, EventAuditChainBroken, EventEntityCreated,
		EventEntityUpdated, EventEntityDeleted, EventEntityRestored, EventUpload,
	}
	deprecated := []string{
		EventLoginSuccess, EventLoginFailed, EventSecuritySettingChange, EventUserDataChange,
		EventLogCleanup, EventDataTampering, EventCSRFViolation, EventSessionRevoked,
	}
	for _, code := range active {
		spec, ok := LookupEvent(code)
		if !ok || spec.Label == "" || spec.Deprecated {
			t.Fatalf("نوع جاری ناقص است: %s %+v", code, spec)
		}
		if spec.Severity != SeverityCritical && spec.Severity != SeverityHigh && spec.Severity != SeverityNormal {
			t.Fatalf("سطح اهمیت %s نامعتبر است: %s", code, spec.Severity)
		}
	}
	for _, code := range deprecated {
		spec, ok := LookupEvent(code)
		if !ok || spec.Label == "" || !spec.Deprecated {
			t.Fatalf("نوع منسوخ باید برای خواندن بماند: %s %+v", code, spec)
		}
	}
	if _, ok := LookupEvent("DOWNLOAD"); ok {
		t.Fatal("DOWNLOAD از کاتالوگ حذف شده است")
	}
	if _, ok := LookupEvent("ROLE_CHANGE"); ok {
		t.Fatal("ROLE_CHANGE از کاتالوگ حذف شده است")
	}
}

// TestReasonLabel_AllCodesHavePersian هر کد دلیل یک برچسب فارسی دارد.
func TestReasonLabel_AllCodesHavePersian(t *testing.T) {
	codes := []string{
		ReasonUserLogout, ReasonAdminTerminated, ReasonMaxSessions, ReasonSecurityCodeFailed,
		ReasonIdleTimeout, ReasonSettingsChanged, ReasonRoleChanged, ReasonIPChanged,
		ReasonWrongPassword, ReasonUserNotFound, ReasonAccountLocked, ReasonAccountInactive,
		ReasonIPLocked, ReasonOutsideWorkHours, ReasonIntegrityViolation,
	}
	for _, code := range codes {
		label := ReasonLabel(code)
		if label == "" || label == code {
			t.Fatalf("برچسب دلیل خالی است: %s", code)
		}
	}
	if ReasonLabel("") != "" {
		t.Fatal("دلیل خالی باید خالی بماند")
	}
	if ReasonLabel("unknown_reason") != "unknown_reason" {
		t.Fatal("دلیل ناشناخته باید همان کد بماند")
	}
}
