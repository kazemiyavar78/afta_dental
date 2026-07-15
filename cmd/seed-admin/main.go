// ابزار یک‌بار مصرف برای ایجاد کاربر Admin اولیه.
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/tpdenta/afta-reception/internal/platform/config"
	"github.com/tpdenta/afta-reception/internal/platform/db"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/encryption"
	"github.com/tpdenta/afta-reception/internal/platform/security/integrity"
	"github.com/tpdenta/afta-reception/internal/platform/security/settings"
	"github.com/tpdenta/afta-reception/internal/user"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("تنظیمات: %v", err)
	}

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("دیتابیس: %v", err)
	}

	username := envOr("SEED_ADMIN_USERNAME", "admin")
	password := envOr("SEED_ADMIN_PASSWORD", "Admin@1234")

	signer := integrity.NewSigner(cfg.IntegrityHMACKey)
	auditMgr := audit.NewManager(database, cfg.AuditHMACKey)
	encryptor, _ := encryption.NewEncryptor(cfg.AESKey)
	userEncrypt := encryption.NewUserEncryptionService(encryptor)

	repo := user.NewRepository(database)
	service := user.NewService(database, signer, userEncrypt, nil, auditMgr, nil, nil)





	existing, err := repo.FindByUsername(username)
	if err == nil && existing != nil {
		fmt.Printf("کاربر %s از قبل وجود دارد (ID=%d)\n", username, existing.ID)
		return
	}

	permissions, err := repo.FindAllPermissions()
	if err != nil {
		log.Fatalf("مجوزها یافت نشد: %v", err)
	}
	permissionIDs := make([]int, len(permissions))
	for i, permission := range permissions {
		permissionIDs[i] = permission.ID
	}

	service.CreateRole(user.CreateRoleRequest{
		Name:        "Admin",
		Description: "مدیر سیستم",
		PermissionIDs: permissionIDs,
	}, 0, "Admin")

	adminRole, err := repo.FindRoleByName("Admin")
	if err != nil {
		log.Fatalf("نقش Admin یافت نشد — ابتدا migration را اجرا کنید: %v", err)
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	now := time.Now().UTC()

	u := &user.User{
		Username:          username,
		PasswordHash:      string(hash),
		Name:              "مدیر",
		Family:            "سیستم",
		RoleID:            adminRole.ID,
		IsActive:          true,
		PasswordChangedAt: now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	secCode, err := userEncrypt.CreateSecurityCode(u.ToSensitiveData())
	if err != nil {
		log.Fatalf("SecurityCode: %v", err)
	}
	u.SecurityCode = secCode

	if err := repo.Create(u); err != nil {
		log.Fatalf("ایجاد کاربر: %v", err)
	}

	created, _ := repo.FindByID(u.ID)
	created.IntegrityHash = user.SignUserIntegrityHash(signer, created)
	_ = repo.Update(created)

	// اصلاح هش نقش‌ها
	settingsSvc := settings.NewService(database, signer, auditMgr)
	_ = settingsSvc.SeedDefaults()

	if adminRole.IntegrityHash == "" {
		permIDs, _ := repo.FindPermissionIDsByRoleID(adminRole.ID)
		adminRole.IntegrityHash = user.SignRoleIntegrityHash(signer, adminRole, permIDs)
		_ = repo.UpdateRole(adminRole)
	}

	fmt.Printf("کاربر Admin ایجاد شد:\n  Username: %s\n  Password: %s\n  ID: %d\n", username, password, created.ID)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// اطمینان از import gorm
var _ = gorm.ErrRecordNotFound
