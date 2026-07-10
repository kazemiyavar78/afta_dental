package session

import (
	"testing"
)

// TestSessionModel ساختار مدل نشست را بررسی می‌کند.
func TestSessionModel(t *testing.T) {
	s := Session{
		ID:                 NewMSSQLUUID(),
		Ip:                 "127.0.0.1",
		Browser:            "TestBrowser",
		PersonnelAccountID: 1,
	}
	if s.Ip != "127.0.0.1" {
		t.Fatal("IP باید 127.0.0.1 باشد")
	}
	if s.ID.String() == "" {
		t.Fatal("ID نباید خالی باشد")
	}
}
