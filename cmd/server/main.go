package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/tpdenta/afta-reception/internal/fund"
	"github.com/tpdenta/afta-reception/internal/logs"
	"github.com/tpdenta/afta-reception/internal/organization"
	"github.com/tpdenta/afta-reception/internal/platform/config"
	"github.com/tpdenta/afta-reception/internal/platform/db"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
	"github.com/tpdenta/afta-reception/internal/platform/migrate"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/encryption"
	"github.com/tpdenta/afta-reception/internal/platform/security/integrity"
	"github.com/tpdenta/afta-reception/internal/platform/security/loginguard"
	"github.com/tpdenta/afta-reception/internal/platform/security/logretention"
	"github.com/tpdenta/afta-reception/internal/platform/security/session"
	"github.com/tpdenta/afta-reception/internal/platform/security/settings"
	"github.com/tpdenta/afta-reception/internal/reception"
	"github.com/tpdenta/afta-reception/internal/tariff"
	"github.com/tpdenta/afta-reception/internal/user"
)

func main() {
	// بارگذاری .env در محیط توسعه
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("خطا در بارگذاری تنظیمات: %v", err)
	}

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("خطا در اتصال دیتابیس: %v", err)
	}
	migrate.Migrate(database, []interface{}{
		&user.User{},
		&organization.Organization{},
		&fund.Fund{},
		&tariff.Tariff{},
		&reception.Reception{},
		&settings.SecuritySetting{},
		&session.Session{},
		&audit.SecurityEvent{},
		&user.Role{},
		&user.Permission{},
		&user.RolePermission{},
		&loginguard.LoginAttempt{},
	})

	// لایه امنیتی
	signer := integrity.NewSigner(cfg.IntegrityHMACKey)
	encryptor, err := encryption.NewEncryptor(cfg.AESKey)
	if err != nil {
		log.Fatalf("خطا در راه‌اندازی رمزنگاری: %v", err)
	}
	userEncryptSvc := encryption.NewUserEncryptionService(encryptor)

	auditMgr := audit.NewManager(database, cfg.AuditHMACKey)
	settingsSvc := settings.NewService(database, signer, auditMgr)

	if err := settingsSvc.SeedDefaults(); err != nil {
		log.Printf("هشدار: seed تنظیمات: %v", err)
	}

	sessionSvc := session.NewService(database, settingsSvc, auditMgr)
	loginGuard := loginguard.NewGuard(database, settingsSvc)

	userSvc := user.NewService(database, signer, userEncryptSvc, settingsSvc, auditMgr, sessionSvc, loginGuard)
	if err := userSvc.FixRoleIntegrityHashes(); err != nil {
		log.Printf("هشدار: اصلاح هش نقش‌ها: %v", err)
	}

	// کارگر پس‌زمینه مدیریت حجم لاگ و یکپارچگی
	worker := logretention.NewWorker(database, settingsSvc, auditMgr, signer, loginGuard, userSvc, cfg.LogRetentionTicker)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go worker.Start(ctx)

	// راه‌اندازی HTTP
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// ترتیب میدلورها طبق سند
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.SecurityHeadersMiddleware())
	r.Use(middleware.RequestLoggerMiddleware())
	r.Use(middleware.CORSMiddleware(cfg.AllowedOrigins))
	r.Use(middleware.RateLimiterMiddleware(cfg.RateLimitPerMinute))

	api := r.Group("/api")
	api.Use(middleware.LoginRateLimiterMiddleware(loginGuard))
	api.Use(middleware.SessionMiddleware(sessionSvc, cfg.SecureCookies))
	api.Use(middleware.SetRoleMiddleware(userSvc.GetRoleName))
	api.Use(middleware.CSRFMiddleware(cfg.CSRFHMACKey, auditMgr))
	api.Use(middleware.SecurityCheckMiddleware(userSvc, sessionSvc, auditMgr))
	api.Use(middleware.AuthorizationMiddleware())

	// health check
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	userHandler := user.NewHandler(userSvc, sessionSvc, cfg.CSRFHMACKey, cfg.SecureCookies)
	user.RegisterRoutes(api, userHandler)

	receptionHandler := reception.NewHandler(reception.NewService(database, auditMgr))
	reception.RegisterRoutes(api, receptionHandler)

	orgHandler := organization.NewHandler(organization.NewService(database, auditMgr))
	organization.RegisterRoutes(api, orgHandler)

	fundHandler := fund.NewHandler(fund.NewService(database, auditMgr))
	fund.RegisterRoutes(api, fundHandler)

	tariffHandler := tariff.NewHandler(tariff.NewService(database, auditMgr))
	tariff.RegisterRoutes(api, tariffHandler)

	logsHandler := logs.NewHandler(auditMgr.Repository())
	logs.RegisterRoutes(api, logsHandler)

	srv := &http.Server{
		Addr:         "192.168.1.60:" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("سرور روی پورت %s در حال اجرا...", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("خطا در اجرای سرور: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("در حال خاموش شدن...")
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)
}
