// Tene Cloud API Server entrypoint.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/tomo-kay/tene/internal/api"
)

func main() {
	cfg := api.Config{
		Port:               envOr("PORT", "8080"),
		JWTSecret:          envRequired("JWT_SECRET"),
		GitHubClientID:     envOr("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: envOr("GITHUB_CLIENT_SECRET", ""),
		CallbackBase:       envOr("CALLBACK_BASE", "http://127.0.0.1:8080"),
		DashboardURL:       envOr("DASHBOARD_URL", "https://app.tene.sh"),
		DatabaseURL:        buildDatabaseURL(),
		S3BucketName:       envOr("S3_BUCKET", ""),
		S3Region:           envOr("AWS_REGION", "ap-northeast-2"),
		S3Endpoint:         envOr("S3_ENDPOINT", ""),
		LemonAPIKey:        envOr("LEMON_API_KEY", ""),
		LemonWebhookSecret: envOr("LEMON_WEBHOOK_SECRET", ""),
		LemonStoreID:       envOr("LEMON_STORE_ID", ""),
		LemonProVariantID:  envOr("LEMON_VARIANT_PRO", ""),
		FreeRPM:            100,
		ProRPM:             1000,
	}

	// Run database migrations if DATABASE_URL is set
	if cfg.DatabaseURL != "" {
		// golang-migrate pgx/v5 driver requires "pgx5://" scheme
		migrateURL := strings.Replace(cfg.DatabaseURL, "postgres://", "pgx5://", 1)
		runMigrations(migrateURL)
	}

	e, cleanup, err := api.NewServer(cfg)
	if err != nil {
		log.Fatalf("server init: %v", err)
	}
	defer cleanup()

	// Graceful shutdown
	go func() {
		if err := e.Start(":" + cfg.Port); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
	log.Println("server exited cleanly")
}

// runMigrations applies pending database migrations from the migrations/ directory.
func runMigrations(databaseURL string) {
	m, err := migrate.New("file:///migrations", databaseURL)
	if err != nil {
		// Try relative path (local dev)
		m, err = migrate.New("file://migrations", databaseURL)
	}
	if err != nil {
		log.Printf("migration init: %v (skipping migrations)", err)
		return
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Printf("migration close source: %v", srcErr)
		}
		if dbErr != nil {
			log.Printf("migration close db: %v", dbErr)
		}
	}()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migration up: %v", err)
	}
	log.Println("database migrations applied")
}

// buildDatabaseURL returns DATABASE_URL if set, otherwise composes it from
// individual DB_* environment variables (DB_HOST, DB_PORT, DB_NAME, DB_USERNAME, DB_PASSWORD).
func buildDatabaseURL() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	host := os.Getenv("DB_HOST")
	if host == "" {
		return "" // no DB config at all → in-memory fallback
	}
	port := envOr("DB_PORT", "5432")
	name := envOr("DB_NAME", "tene")
	user := envOr("DB_USERNAME", "tene_admin")
	pass := os.Getenv("DB_PASSWORD")
	sslmode := envOr("DB_SSLMODE", "require")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, pass, host, port, name, sslmode)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envRequired(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required environment variable %s is not set", key)
	}
	return v
}
