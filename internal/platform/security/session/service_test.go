package session

import (
	"testing"

	"github.com/google/uuid"
)

// TestSessionModel ساختار مدل نشست را بررسی می‌کند.
func TestSessionModel(t *testing.T) {
	s := Session{
		ID:                 uuid.New(),
		Ip:                 "127.0.0.1",
		Browser:            "TestBrowser",
		PersonnelAccountID: 1,
	}
	if s.Ip != "127.0.0.1" {
		t.Fatal("IP باید 127.0.0.1 باشد")
	}
}
