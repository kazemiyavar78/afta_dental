package audit

import "testing"

// TestEventCatalog_KnownTypesHaveLabel ????? ???? ? ????? ???? ????? ? ??? ????? ????? ?????.
func TestEventCatalog_KnownTypesHaveLabel(t *testing.T) {
	active := []string{
		EventSessionCreated, EventSessionTerminated, EventLogout, EventLoginFailed, EventLoginBlocked,
		EventIPLocked, EventAccountLocked, EventUserCreated, EventUserUpdated,
		EventPasswordChanged, EventRoleCreated, EventRoleUpdated, EventRoleDeleted,
		EventSettingChanged, EventSettingRemoved, EventAuditStarted, EventAuditStopped,
		EventAuditRead, EventAuditReadDenied, EventAuditChainBroken, EventEntityCreated,
		EventEntityUpdated, EventEntityDeleted, EventEntityRestored, EventUpload,
	}
	deprecated := []string{
		EventLoginSuccess, EventSecuritySettingChange, EventUserDataChange,
		EventLogCleanup, EventDataTampering, EventCSRFViolation, EventSessionRevoked,
	}
	for _, code := range active {
		spec, ok := LookupEvent(code)
		if !ok || spec.Label == "" || spec.Deprecated {
			t.Fatalf("??? ???? ???? ???: %s %+v", code, spec)
		}
		if spec.Severity != SeverityCritical && spec.Severity != SeverityHigh && spec.Severity != SeverityNormal {
			t.Fatalf("??? ????? %s ??????? ???: %s", code, spec.Severity)
		}
	}
	for _, code := range deprecated {
		spec, ok := LookupEvent(code)
		if !ok || spec.Label == "" || !spec.Deprecated {
			t.Fatalf("??? ????? ???? ???? ?????? ?????: %s %+v", code, spec)
		}
	}
	if _, ok := LookupEvent("DOWNLOAD"); ok {
		t.Fatal("DOWNLOAD ?? ??????? ??? ??? ???")
	}
	if _, ok := LookupEvent("ROLE_CHANGE"); ok {
		t.Fatal("ROLE_CHANGE ?? ??????? ??? ??? ???")
	}
}

// TestReasonLabel_AllCodesHavePersian ?? ?? ???? ?? ????? ????? ????.
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
			t.Fatalf("????? ???? ???? ???: %s", code)
		}
	}
	if ReasonLabel("") != "" {
		t.Fatal("???? ???? ???? ???? ?????")
	}
	if ReasonLabel("unknown_reason") != "unknown_reason" {
		t.Fatal("???? ???????? ???? ???? ?? ?????")
	}
}
