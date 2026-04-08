// Package api provides the Tene Cloud API server.
package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	echoMw "github.com/labstack/echo/v4/middleware"
	"github.com/tomo-kay/tene/internal/api/handler"
	mw "github.com/tomo-kay/tene/internal/api/middleware"
	"github.com/tomo-kay/tene/internal/api/response"
	"github.com/tomo-kay/tene/internal/api/storage"
	"github.com/tomo-kay/tene/internal/auth"
	"github.com/tomo-kay/tene/internal/billing"
	"github.com/tomo-kay/tene/internal/repository/postgres"
)

// Config holds API server configuration.
type Config struct {
	Port               string
	JWTSecret          string
	GitHubClientID     string
	GitHubClientSecret string
	CallbackBase       string // e.g. "https://api.tene.sh"
	FreeRPM            int
	ProRPM             int
	S3BucketName       string
	S3Region           string
	S3Endpoint         string // S3-compatible endpoint (e.g. MinIO); empty = AWS S3
	LemonAPIKey        string
	LemonWebhookSecret string
	LemonStoreID       string
	LemonProVariantID  string
	DashboardURL       string
	DatabaseURL        string // PostgreSQL connection string; empty = in-memory fallback
}

// NewServer creates and configures the Echo server with all routes.
// Returns the Echo instance, a cleanup function, and any initialization error.
// The cleanup function must be called on shutdown to release resources (e.g. DB pool).
func NewServer(cfg Config) (*echo.Echo, func(), error) {
	e := echo.New()
	e.HideBanner = true

	// H-02: Configure trusted proxy IP extraction (ALB/CloudFront)
	e.IPExtractor = echo.ExtractIPFromXFFHeader(
		echo.TrustLoopback(true),
		echo.TrustLinkLocal(false),
		echo.TrustPrivateNet(false),
	)

	// Database: PostgreSQL or in-memory fallback
	var (
		vaultStore    handler.VaultStore
		teamStore     handler.TeamStore
		deviceStore   handler.DeviceStore
		auditStore    handler.AuditStore
		waitlistStore handler.WaitlistStore
		userStore     billing.UserStore
		authUserStore handler.AuthUserStore // for enriched /auth/me responses
		dbPinger      handler.DBPinger      // for readiness health check
		cleanup       = func() {}
	)

	if cfg.DatabaseURL != "" {
		db, err := postgres.New(context.Background(), cfg.DatabaseURL)
		if err != nil {
			return nil, nil, fmt.Errorf("server: database: %w", err)
		}
		cleanup = db.Close
		dbPinger = db.Pool
		slog.Info("database.connected", "driver", "postgresql")

		userRepo := postgres.NewUserRepo(db.Pool)
		vaultStore = postgres.NewVaultRepo(db.Pool)
		teamStore = postgres.NewTeamRepo(db.Pool)
		deviceStore = postgres.NewDeviceRepo(db.Pool)
		auditStore = postgres.NewAuditRepo(db.Pool)
		waitlistStore = postgres.NewWaitlistRepo(db.Pool)
		userStore = userRepo
		authUserStore = userRepo
	} else {
		slog.Warn("database.fallback", "mode", "in-memory", "reason", "DATABASE_URL not set")
		vaultStore = handler.NewMemVaultStore()
		teamStore = handler.NewMemTeamStore()
		deviceStore = handler.NewMemDeviceStore()
		auditStore = handler.NewMemAuditStore()
		waitlistStore = handler.NewMemWaitlistStore()
		// userStore remains nil (billing webhook disabled without DB)
	}

	// Services
	jwtSvc := auth.NewJWTService(cfg.JWTSecret)
	oauthSvc := auth.NewOAuthService(cfg.GitHubClientID, cfg.GitHubClientSecret, cfg.CallbackBase)
	rateLimiter := mw.NewRateLimiter(cfg.FreeRPM, cfg.ProRPM)

	// S3 client (nil-safe: push/pull will fail with clear error if not configured)
	var s3Client *storage.S3Client
	if cfg.S3BucketName != "" && cfg.S3Region != "" {
		var s3Err error
		s3Client, s3Err = storage.NewS3Client(context.Background(), cfg.S3BucketName, cfg.S3Region, cfg.S3Endpoint)
		if s3Err != nil {
			e.Logger.Warnf("S3 client init failed (push/pull disabled): %v", s3Err)
		}
	}

	// Billing service
	billingSvc := billing.NewService(billing.Config{
		APIKey:        cfg.LemonAPIKey,
		WebhookSecret: cfg.LemonWebhookSecret,
		StoreID:       cfg.LemonStoreID,
		ProVariantID:  cfg.LemonProVariantID,
		DashboardURL:  cfg.DashboardURL,
	}, userStore)

	// Handlers
	healthH := &handler.HealthHandler{DB: dbPinger}
	authH := handler.NewAuthHandler(oauthSvc, jwtSvc, cfg.DashboardURL)
	vaultH := handler.NewVaultHandler(vaultStore, s3Client)
	billingH := handler.NewBillingHandler(billingSvc)
	waitlistH := handler.NewWaitlistHandler(waitlistStore)
	teamH := handler.NewTeamHandler(teamStore)
	deviceH := handler.NewDeviceHandler(deviceStore)
	auditH := handler.NewAuditHandler(auditStore)

	// Wire auth handler user store for enriched /auth/me responses
	if authUserStore != nil {
		authH.SetUserStore(authUserStore)
	}

	// Global middleware (order matters)
	e.Use(echoMw.RequestID())
	e.Use(echoMw.RequestLoggerWithConfig(echoMw.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		LogMethod:   true,
		LogLatency:  true,
		LogRemoteIP: true,
		LogError:    true,
		LogValuesFunc: func(c echo.Context, v echoMw.RequestLoggerValues) error {
			slog.Info("request",
				"method", v.Method,
				"uri", v.URI,
				"status", v.Status,
				"latency", v.Latency.String(),
				"remote_ip", v.RemoteIP,
				"error", v.Error,
			)
			return nil
		},
	}))
	e.Use(echoMw.Recover())
	e.Use(mw.SecurityHeaders())
	e.Use(echoMw.BodyLimit("2M")) // Default body limit for most routes
	// M-03: CORS origins from environment (no localhost in prod)
	corsOrigins := []string{"https://tene.sh", "https://app.tene.sh"}
	if extra := os.Getenv("CORS_EXTRA_ORIGINS"); extra != "" {
		corsOrigins = append(corsOrigins, strings.Split(extra, ",")...)
	}
	// Auto-add localhost origins when CALLBACK_BASE points to localhost (local dev)
	if strings.Contains(cfg.CallbackBase, "localhost") || strings.Contains(cfg.CallbackBase, "127.0.0.1") {
		corsOrigins = append(corsOrigins, "http://localhost:3000", "http://localhost:3001")
	}
	e.Use(echoMw.CORSWithConfig(echoMw.CORSConfig{
		AllowOrigins: corsOrigins,
		AllowMethods: []string{"GET", "POST", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders: []string{"Authorization", "Content-Type", "X-Request-ID", "If-Match"},
	}))
	e.Use(rateLimiter.Middleware())

	// Health (no auth)
	e.GET("/health", healthH.Liveness)
	e.GET("/health/ready", healthH.Readiness)

	// API v1
	v1 := e.Group("/api/v1")

	// Auth (public)
	v1.GET("/auth/github/authorize", authH.GitHubAuthorize)
	v1.GET("/auth/github/callback", authH.GitHubCallback)
	v1.POST("/auth/refresh", authH.RefreshToken)
	v1.POST("/auth/exchange", authH.Exchange)

	// Authenticated routes
	authed := v1.Group("", mw.JWTAuth(jwtSvc))
	authed.GET("/auth/me", authH.Me)
	authed.POST("/auth/signout", authH.Signout)

	// User public key (for team key exchange)
	authed.GET("/users/:id/public-key", func(c echo.Context) error {
		// Returns the X25519 public key for a user (needed for ECDH team invite)
		// TODO: wire to user repository when DB connected
		return response.ErrMsg(c, http.StatusNotFound, "NOT_FOUND", "user not found")
	})

	// Vault routes (W3)
	authed.GET("/vaults", vaultH.List)
	authed.POST("/vaults", vaultH.Create)
	authed.GET("/vaults/:id", vaultH.Get)
	authed.POST("/vaults/:id/push", vaultH.Push, echoMw.BodyLimit("50M"))
	authed.GET("/vaults/:id/pull", vaultH.Pull)
	authed.DELETE("/vaults/:id", vaultH.Delete)

	// Billing routes (W4)
	authed.GET("/billing/subscription", billingH.GetSubscription)
	authed.POST("/billing/checkout", billingH.CreateCheckout)
	authed.POST("/billing/portal", billingH.CreatePortal)

	// Billing webhook (public, HMAC-verified)
	v1.POST("/billing/webhook", billingH.Webhook)

	// Team routes (W5)
	authed.POST("/teams", teamH.Create)
	authed.GET("/teams", teamH.List)
	authed.POST("/teams/:id/invite", teamH.Invite)
	authed.GET("/teams/:id/members", teamH.ListMembers)
	authed.DELETE("/teams/:id/members/:uid", teamH.RemoveMember)
	authed.PATCH("/teams/:id/members/:uid/role", teamH.UpdateRole)

	// Device routes (W5)
	authed.POST("/devices", deviceH.Register)
	authed.GET("/devices", deviceH.List)
	authed.DELETE("/devices/:id", deviceH.Delete)

	// Audit routes (W5)
	authed.GET("/audit", auditH.List)

	// Waitlist (public)
	v1.POST("/waitlist", waitlistH.Register)

	return e, cleanup, nil
}
