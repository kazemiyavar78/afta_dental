//go:build integration

package reception

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/tpdenta/afta-reception/internal/patient"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/encryption"
	"github.com/tpdenta/afta-reception/internal/platform/testutil"
	"gorm.io/gorm"
)

// TestIntegration_D13_ReceptionDraftRecordsActorAndTarget پیش‌نویس پذیرش را با فاعل و نام نمایشی بیمار ثبت می‌کند.
// مسیر ایجاد به سرویس بیمار و تراکنش پایگاه وصل است و بدون پایگاه قابل جداسازی نیست.
func TestIntegration_D13_ReceptionDraftRecordsActorAndTarget(t *testing.T) {
	database := testutil.OpenTestDB(t)
	if err := database.AutoMigrate(
		&patient.Patient{}, &Reception{}, &ReceptionService{}, &ReceptionPhoto{},
		&audit.SecurityEvent{}, &audit.AuditChainState{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	testutil.WithTestTx(t, func(tx *gorm.DB) {
		enc, err := encryption.NewEncryptor("12345678901234567890123456789012")
		if err != nil {
			t.Fatal(err)
		}
		mgr := audit.NewManager(tx, "integration-audit-key")
		patients := patient.NewService(tx, mgr, encryption.NewPatientEncryptionService(enc))
		svc := NewService(tx, mgr, patients, nil, nil, nil, nil, nil, nil, nil)
		actorID := 8
		file := fmt.Sprintf("IT-%d", time.Now().UnixNano())
		_, err = svc.CreateReception(UpsertReceptionRequest{
			Patient: PatientInput{
				FirstName: "علی", LastName: "رضایی", FileNumber: file,
				BirthDate: "1990-01-02", IsForeignNational: true,
			},
			ReceptionDate: "2026-01-01",
		}, context.Background(), audit.Actor{UserID: &actorID, Username: "reza", IP: "10.1.1.1"})
		if err != nil {
			t.Fatal(err)
		}
		var rows []audit.SecurityEvent
		if err := tx.Where("TargetType = ? AND EventType = ?", audit.TargetReception, audit.EventEntityCreated).Find(&rows).Error; err != nil {
			t.Fatal(err)
		}
		if len(rows) != 1 {
			t.Fatalf("یک رویداد ایجاد پذیرش انتظار می‌رود، دریافت %d", len(rows))
		}
		row := rows[0]
		if row.UserID == nil || *row.UserID != 8 || row.ActorUsername != "reza" || row.TargetName != "علی ر" {
			t.Fatalf("فاعل یا نام هدف نادرست است: %+v", row)
		}
		if row.TargetName == file {
			t.Fatal("نام هدف نباید شماره پرونده باشد")
		}
	})
}
