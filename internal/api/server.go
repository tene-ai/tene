// Package api provides the Tene Cloud API server.
package api

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	echoMw "github.com/labstack/echo/v4/middleware"
	"github.com/tomo-kay/tene/internal/api/handler"
	mw "github.com/tomo-kay/tene/internal/api/middleware"
	"github.com/tomo-kay/tene/internal/api/response"
	"github.com/tomo-kay/tene/internal/api/storage"
	"github.com/tomo-kay/tene/internal/auth"
	"github.com/tomo-kay/tene/internal/billing"
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
	LemonAPIKey        string
	LemonWebhookSecret string
	LemonStoreID       string
	LemonProVariantID  string
	DashboardURL       string
}

// NewServer creates and configures the Echo server with all routes.
func NewServer(cfg Config) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// H-02: Configure trusted proxy IP extraction (ALB/CloudFront)
	e.IPExtractor = echo.ExtractIPFromXFFHeader(
		echo.TrustLoopback(true),
		echo.TrustLinkLocal(false),
		echo.TrustPrivateNet(false),
	)

	// Services
	jwtSvc := auth.NewJWTService(cfg.JWTSecret)
	oauthSvc := auth.NewOAuthService(cfg.GitHubClientID, cfg.GitHubClientSecret, cfg.CallbackBase)
	rateLimiter := mw.NewRateLimiter(cfg.FreeRPM, cfg.ProRPM)

	// Vault store (in-memory for dev; replace with PostgreSQL in prod)
	vaultStore := handler.NewMemVaultStore()

	// S3 client (nil-safe: push/pull will fail with clear error if not configured)
	var s3Client *storage.S3Client
	if cfg.S3BucketName != "" && cfg.S3Region != "" {
		var s3Err error
		s3Client, s3Err = storage.NewS3Client(context.Background(), cfg.S3BucketName, cfg.S3Region)
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
	}, nil) // UserStore wired when DB connected

	// Team, device, audit stores (in-memory for dev)
	teamStore := handler.NewMemTeamStore()
	deviceStore := handler.NewMemDeviceStore()
	auditStore := handler.NewMemAuditStore()

	// Waitlist store
	waitlistStore := handler.NewMemWaitlistStore()

	// Handlers
	healthH := &handler.HealthHandler{}
	authH := handler.NewAuthHandler(oauthSvc, jwtSvc)
	vaultH := handler.NewVaultHandler(vaultStore, s3Client)
	billingH := handler.NewBillingHandler(billingSvc)
	waitlistH := handler.NewWaitlistHandler(waitlistStore)
	teamH := handler.NewTeamHandler(teamStore)
	deviceH := handler.NewDeviceHandler(deviceStore)
	auditH := handler.NewAuditHandler(auditStore)

	// Global middleware (order matters)
	e.Use(echoMw.RequestID())
	e.Use(echoMw.Logger())
	e.Use(echoMw.Recover())
	e.Use(mw.SecurityHeaders())
	e.Use(echoMw.BodyLimit("2M")) // Default body limit for most routes
	e.Use(echoMw.CORSWithConfig(echoMw.CORSConfig{
		AllowOrigins: []string{"https://tene.sh", "https://app.tene.sh", "http://localhost:3000", "http://localhost:3001"},
		AllowMethods: []string{"GET", "POST", "OPTIONS"}, // M-03: restrict to used methods
		AllowHeaders: []string{"Authorization", "Content-Type", "X-Request-ID"},
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

	return e
}
