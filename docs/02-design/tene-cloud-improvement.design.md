# Tene Cloud 개선 -- 상세 설계서

> **버전**: v1.0
> **일자**: 2026-04-08
> **기능**: tene-cloud-improvement
> **상위 문서**: [tene-cloud-improvement.plan.md](../01-plan/features/tene-cloud-improvement.plan.md)
> **방법론**: 10-관점 통합 설계 (DB/Backend/Auth/Frontend/CLI/Billing/QA/Infra/Integration/PM)

---

## Context Anchor

| Key | Value |
|-----|-------|
| **WHY** | Tene Cloud API가 in-memory 전용으로 프로덕션 불가. billing webhook panic, auth token 평문 저장, Dashboard 15분 로그아웃. 유료 전환 차단 |
| **WHO** | 솔로 개발자 (Pro $5/mo), 소규모 팀 (Team $10/user/mo), AI-first 개발자 |
| **RISK** | PG 마이그레이션 도중 스키마 불일치, LemonSqueezy webhook 테스트 난이도, go-keyring Linux headless 실패 |
| **SUCCESS** | E2E 플로우 무결 (가입-결제-push/pull-팀-Dashboard), Zero-Knowledge 유지, 유료 전환 가능 |
| **SCOPE** | 4 Sprint (19-23일): PG Repository + Auth/Security + Dashboard 연동 + 테스트/인프라 |

---

## 목차

1. [DB 아키텍처 설계](#1-db-아키텍처-설계)
2. [백엔드 API 서버 설계](#2-백엔드-api-서버-설계)
3. [인증/보안 설계](#3-인증보안-설계)
4. [프론트엔드 Dashboard 설계](#4-프론트엔드-dashboard-설계)
5. [CLI/DX 설계](#5-clidx-설계)
6. [결제 통합 설계](#6-결제-통합-설계)
7. [QA/테스트 전략](#7-qa테스트-전략)
8. [인프라/운영 설계](#8-인프라운영-설계)
9. [통합 아키텍처 설계](#9-통합-아키텍처-설계)
10. [사용자 여정/UX 설계](#10-사용자-여정ux-설계)
11. [구현 가이드](#11-구현-가이드)

---

## 1. DB 아키텍처 설계

**담당**: DB 아키텍트

### 1.1 현재 상태

모든 store가 in-memory (`map` + `sync.RWMutex`), ECS 재시작 시 전체 데이터 소실.

| Store | 파일 | 인터페이스 |
|-------|------|-----------|
| MemVaultStore | `internal/api/handler/vault_store_mem.go` | VaultStore (7 메서드) |
| MemTeamStore | `internal/api/handler/team.go:32-191` | TeamStore (9 메서드) |
| MemDeviceStore | `internal/api/handler/device.go:22-67` | DeviceStore (3 메서드) |
| MemAuditStore | `internal/api/handler/audit.go:18-40` | AuditStore (1 메서드) |
| MemWaitlistStore | `internal/api/handler/waitlist.go:21-37` | WaitlistStore (1 메서드) |
| AuthHandler.states | `internal/api/handler/auth.go:43` | in-memory map (비인터페이스) |
| AuthHandler.refresh | `internal/api/handler/auth.go:44` | in-memory map (비인터페이스) |

### 1.2 대상 아키텍처

```
internal/repository/
  postgres/
    postgres.go           -- pgxpool.Pool 래퍼, 헬스체크, 커넥션 관리
    user_repo.go          -- UserStore 인터페이스 구현
    vault_repo.go         -- VaultStore 인터페이스 구현
    team_repo.go          -- TeamStore 인터페이스 구현
    device_repo.go        -- DeviceStore 인터페이스 구현
    audit_repo.go         -- AuditStore 인터페이스 구현
    refresh_token_repo.go -- RefreshTokenStore 인터페이스 구현 (신규)
    waitlist_repo.go      -- WaitlistStore 인터페이스 구현
```

### 1.3 DB 드라이버 및 커넥션 풀

**드라이버**: `github.com/jackc/pgx/v5/pgxpool` (pure Go, PostgreSQL 전용)

```go
// internal/repository/postgres/postgres.go
package postgres

import (
    "context"
    "fmt"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps pgxpool.Pool with health check support.
type DB struct {
    Pool *pgxpool.Pool
}

// New creates a new PostgreSQL connection pool.
func New(ctx context.Context, databaseURL string) (*DB, error) {
    config, err := pgxpool.ParseConfig(databaseURL)
    if err != nil {
        return nil, fmt.Errorf("postgres: parse config: %w", err)
    }

    // ECS Fargate 0.25 vCPU / 512 MiB에 적합한 풀 설정
    config.MaxConns = 10
    config.MinConns = 2
    config.MaxConnLifetime = 30 * time.Minute
    config.MaxConnIdleTime = 5 * time.Minute
    config.HealthCheckPeriod = 1 * time.Minute

    pool, err := pgxpool.NewWithConfig(ctx, config)
    if err != nil {
        return nil, fmt.Errorf("postgres: connect: %w", err)
    }

    // 연결 확인
    if err := pool.Ping(ctx); err != nil {
        pool.Close()
        return nil, fmt.Errorf("postgres: ping: %w", err)
    }

    return &DB{Pool: pool}, nil
}

// Close gracefully closes the pool.
func (db *DB) Close() {
    db.Pool.Close()
}

// Ping checks database connectivity.
func (db *DB) Ping(ctx context.Context) error {
    return db.Pool.Ping(ctx)
}
```

**DATABASE_URL 형식**:
```
postgres://tene_admin:${DB_PASSWORD}@${DB_HOST}:5432/tene?sslmode=require
```

### 1.4 테이블-인터페이스 매핑

#### 1.4.1 UserRepo (신규 인터페이스)

현재 User 관련 store 인터페이스가 없음. billing.UserStore만 존재.
하나의 통합 UserStore 인터페이스를 정의하고 billing.UserStore를 포함.

```go
// internal/repository/postgres/user_repo.go
package postgres

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/agent-kay-it/tene/internal/domain"
)

// UserStore defines all user operations. Satisfies billing.UserStore.
type UserStore interface {
    UpsertUser(ctx context.Context, u *domain.User) error
    GetUserByID(ctx context.Context, id string) (*domain.User, error)
    GetUserByGitHubID(ctx context.Context, githubID int64) (*domain.User, error)
    GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
    UpdatePlan(ctx context.Context, email string, plan string, lemonCustomerID string) error
    GetLemonCustomerID(ctx context.Context, userID string) (string, error)
    UpdatePublicKey(ctx context.Context, userID string, publicKey []byte) error
    GetPublicKey(ctx context.Context, userID string) ([]byte, error)
}

// UserRepo implements UserStore with PostgreSQL.
type UserRepo struct {
    pool *pgxpool.Pool
}

// NewUserRepo creates a PostgreSQL user repository.
func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
    return &UserRepo{pool: pool}
}

// UpsertUser inserts or updates a user (ON CONFLICT github_id).
func (r *UserRepo) UpsertUser(ctx context.Context, u *domain.User) error {
    query := `
        INSERT INTO users (email, name, auth_provider, github_id, google_id, avatar_url, plan)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (github_id) WHERE github_id IS NOT NULL
        DO UPDATE SET
            email = EXCLUDED.email,
            name = EXCLUDED.name,
            avatar_url = EXCLUDED.avatar_url,
            updated_at = now()
        RETURNING id, created_at, updated_at`

    return r.pool.QueryRow(ctx, query,
        u.Email, u.Name, u.AuthProvider, u.GitHubID, u.GoogleID,
        u.AvatarURL, u.Plan,
    ).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

// GetUserByID retrieves a user by UUID.
func (r *UserRepo) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
    query := `
        SELECT id, email, name, auth_provider, github_id, google_id,
               avatar_url, plan, lemon_customer_id, x25519_public_key,
               created_at, updated_at
        FROM users WHERE id = $1`

    u := &domain.User{}
    err := r.pool.QueryRow(ctx, query, id).Scan(
        &u.ID, &u.Email, &u.Name, &u.AuthProvider, &u.GitHubID, &u.GoogleID,
        &u.AvatarURL, &u.Plan, &u.LemonCustomerID, &u.X25519PublicKey,
        &u.CreatedAt, &u.UpdatedAt,
    )
    if err == pgx.ErrNoRows {
        return nil, domain.ErrNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("user: get by id: %w", err)
    }
    return u, nil
}

// GetUserByGitHubID retrieves a user by GitHub ID.
func (r *UserRepo) GetUserByGitHubID(ctx context.Context, githubID int64) (*domain.User, error) {
    query := `
        SELECT id, email, name, auth_provider, github_id, google_id,
               avatar_url, plan, lemon_customer_id, x25519_public_key,
               created_at, updated_at
        FROM users WHERE github_id = $1`

    u := &domain.User{}
    err := r.pool.QueryRow(ctx, query, githubID).Scan(
        &u.ID, &u.Email, &u.Name, &u.AuthProvider, &u.GitHubID, &u.GoogleID,
        &u.AvatarURL, &u.Plan, &u.LemonCustomerID, &u.X25519PublicKey,
        &u.CreatedAt, &u.UpdatedAt,
    )
    if err == pgx.ErrNoRows {
        return nil, domain.ErrNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("user: get by github id: %w", err)
    }
    return u, nil
}

// UpdatePlan implements billing.UserStore.
func (r *UserRepo) UpdatePlan(ctx context.Context, email string, plan string, lemonCustomerID string) error {
    query := `UPDATE users SET plan = $1, lemon_customer_id = $2 WHERE email = $3`
    ct, err := r.pool.Exec(ctx, query, plan, lemonCustomerID, email)
    if err != nil {
        return fmt.Errorf("user: update plan: %w", err)
    }
    if ct.RowsAffected() == 0 {
        return domain.ErrNotFound
    }
    return nil
}

// GetLemonCustomerID implements billing.UserStore.
func (r *UserRepo) GetLemonCustomerID(ctx context.Context, userID string) (string, error) {
    var customerID *string
    err := r.pool.QueryRow(ctx,
        `SELECT lemon_customer_id FROM users WHERE id = $1`, userID,
    ).Scan(&customerID)
    if err == pgx.ErrNoRows {
        return "", domain.ErrNotFound
    }
    if err != nil {
        return "", fmt.Errorf("user: get lemon customer id: %w", err)
    }
    if customerID == nil {
        return "", nil
    }
    return *customerID, nil
}
```

**참고**: `domain.User.ID`는 현재 `string` 타입이지만 DB는 `UUID`. `UpsertUser`에서 RETURNING으로 DB가 생성한 UUID를 다시 할당. `GitHubCallback`의 `generateUserID("gh_"+id)` 패턴은 DB UUID로 교체.

#### 1.4.2 VaultRepo

기존 `handler.VaultStore` 인터페이스 (7 메서드)를 그대로 구현.

```go
// internal/repository/postgres/vault_repo.go
package postgres

import (
    "context"
    "fmt"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/agent-kay-it/tene/internal/domain"
)

// VaultRepo implements handler.VaultStore with PostgreSQL.
type VaultRepo struct {
    pool *pgxpool.Pool
}

func NewVaultRepo(pool *pgxpool.Pool) *VaultRepo {
    return &VaultRepo{pool: pool}
}

func (r *VaultRepo) CreateVault(v *domain.Vault) error {
    query := `
        INSERT INTO vaults (user_id, team_id, project_name, s3_key, vault_version, vault_hash)
        VALUES ($1, NULLIF($2, ''), $3, $4, 0, $5)
        RETURNING id, created_at, updated_at`

    var teamID *string
    if v.TeamID != "" {
        teamID = &v.TeamID
    }

    return r.pool.QueryRow(context.Background(), query,
        v.UserID, teamID, v.ProjectName, v.S3Key, v.VaultHash,
    ).Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt)
}

func (r *VaultRepo) GetVault(id, userID string) (*domain.Vault, error) {
    query := `
        SELECT id, user_id, COALESCE(team_id::text, ''), project_name, s3_key,
               vault_version, vault_hash, COALESCE(secret_count, 0),
               COALESCE(size, 0), created_at, updated_at, last_pushed_at
        FROM vaults WHERE id = $1 AND user_id = $2`

    v := &domain.Vault{}
    err := r.pool.QueryRow(context.Background(), query, id, userID).Scan(
        &v.ID, &v.UserID, &v.TeamID, &v.ProjectName, &v.S3Key,
        &v.Version, &v.VaultHash, &v.SecretCount,
        &v.Size, &v.CreatedAt, &v.UpdatedAt, &v.LastPushedAt,
    )
    if err == pgx.ErrNoRows {
        return nil, domain.ErrVaultNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("vault: get: %w", err)
    }
    return v, nil
}

func (r *VaultRepo) GetVaultByProject(userID, projectName string) (*domain.Vault, error) {
    query := `
        SELECT id, user_id, COALESCE(team_id::text, ''), project_name, s3_key,
               vault_version, vault_hash, COALESCE(secret_count, 0),
               COALESCE(size, 0), created_at, updated_at, last_pushed_at
        FROM vaults WHERE user_id = $1 AND project_name = $2`

    v := &domain.Vault{}
    err := r.pool.QueryRow(context.Background(), query, userID, projectName).Scan(
        &v.ID, &v.UserID, &v.TeamID, &v.ProjectName, &v.S3Key,
        &v.Version, &v.VaultHash, &v.SecretCount,
        &v.Size, &v.CreatedAt, &v.UpdatedAt, &v.LastPushedAt,
    )
    if err == pgx.ErrNoRows {
        return nil, domain.ErrVaultNotFound
    }
    return v, err
}

func (r *VaultRepo) ListVaults(userID string) ([]domain.Vault, error) {
    query := `
        SELECT id, user_id, COALESCE(team_id::text, ''), project_name, s3_key,
               vault_version, vault_hash, COALESCE(secret_count, 0),
               COALESCE(size, 0), created_at, updated_at, last_pushed_at
        FROM vaults WHERE user_id = $1
        ORDER BY updated_at DESC`

    rows, err := r.pool.Query(context.Background(), query, userID)
    if err != nil {
        return nil, fmt.Errorf("vault: list: %w", err)
    }
    defer rows.Close()

    var vaults []domain.Vault
    for rows.Next() {
        var v domain.Vault
        if err := rows.Scan(
            &v.ID, &v.UserID, &v.TeamID, &v.ProjectName, &v.S3Key,
            &v.Version, &v.VaultHash, &v.SecretCount,
            &v.Size, &v.CreatedAt, &v.UpdatedAt, &v.LastPushedAt,
        ); err != nil {
            return nil, err
        }
        vaults = append(vaults, v)
    }
    return vaults, rows.Err()
}

func (r *VaultRepo) UpdateVaultVersion(id string, currentVersion int64, hash []byte, size int64, secretCount int) (int64, error) {
    query := `
        UPDATE vaults
        SET vault_version = vault_version + 1,
            vault_hash = $1,
            size = $2,
            secret_count = CASE WHEN $3 > 0 THEN $3 ELSE secret_count END,
            last_pushed_at = now()
        WHERE id = $4 AND vault_version = $5
        RETURNING vault_version`

    var newVersion int64
    err := r.pool.QueryRow(context.Background(), query,
        hash, size, secretCount, id, currentVersion,
    ).Scan(&newVersion)
    if err == pgx.ErrNoRows {
        return 0, domain.ErrVersionConflict
    }
    if err != nil {
        return 0, fmt.Errorf("vault: update version: %w", err)
    }
    return newVersion, nil
}

func (r *VaultRepo) DeleteVault(id, userID string) error {
    ct, err := r.pool.Exec(context.Background(),
        `DELETE FROM vaults WHERE id = $1 AND user_id = $2`, id, userID)
    if err != nil {
        return fmt.Errorf("vault: delete: %w", err)
    }
    if ct.RowsAffected() == 0 {
        return domain.ErrVaultNotFound
    }
    return nil
}

func (r *VaultRepo) CreateAuditLog(log *domain.AuditLog) error {
    query := `
        INSERT INTO audit_logs (user_id, vault_id, action, detail, ip_address)
        VALUES ($1, NULLIF($2, '')::uuid, $3, $4, $5::inet)`

    _, err := r.pool.Exec(context.Background(), query,
        log.UserID, log.VaultID, log.Action, log.Detail, log.IPAddress)
    if err != nil {
        return fmt.Errorf("audit: create: %w", err)
    }
    return nil
}
```

#### 1.4.3 RefreshTokenRepo (신규)

현재 `AuthHandler`가 `map[string]refreshEntry`로 직접 관리. DB 이전을 위해 인터페이스 추출 필요.

```go
// internal/repository/postgres/refresh_token_repo.go
package postgres

import (
    "context"
    "fmt"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

// RefreshTokenStore defines refresh token operations.
type RefreshTokenStore interface {
    Store(ctx context.Context, userID string, tokenHash []byte, family string, expiresAt time.Time) error
    Find(ctx context.Context, tokenHash []byte) (*RefreshTokenEntry, error)
    Delete(ctx context.Context, tokenHash []byte) error
    RevokeFamily(ctx context.Context, family string) error
}

// RefreshTokenEntry represents a stored refresh token.
type RefreshTokenEntry struct {
    ID        string
    UserID    string
    TokenHash []byte
    Family    string
    ExpiresAt time.Time
    RevokedAt *time.Time
    CreatedAt time.Time
}

// RefreshTokenRepo implements RefreshTokenStore with PostgreSQL.
type RefreshTokenRepo struct {
    pool *pgxpool.Pool
}

func NewRefreshTokenRepo(pool *pgxpool.Pool) *RefreshTokenRepo {
    return &RefreshTokenRepo{pool: pool}
}

func (r *RefreshTokenRepo) Store(ctx context.Context, userID string, tokenHash []byte, family string, expiresAt time.Time) error {
    query := `
        INSERT INTO refresh_tokens (user_id, token_hash, family, expires_at)
        VALUES ($1, $2, $3, $4)`
    _, err := r.pool.Exec(ctx, query, userID, tokenHash, family, expiresAt)
    if err != nil {
        return fmt.Errorf("refresh_token: store: %w", err)
    }
    return nil
}

func (r *RefreshTokenRepo) Find(ctx context.Context, tokenHash []byte) (*RefreshTokenEntry, error) {
    query := `
        SELECT rt.id, rt.user_id, rt.token_hash, rt.family, rt.expires_at, rt.revoked_at, rt.created_at
        FROM refresh_tokens rt
        WHERE rt.token_hash = $1 AND rt.revoked_at IS NULL`

    e := &RefreshTokenEntry{}
    err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
        &e.ID, &e.UserID, &e.TokenHash, &e.Family,
        &e.ExpiresAt, &e.RevokedAt, &e.CreatedAt,
    )
    if err == pgx.ErrNoRows {
        return nil, nil // not found
    }
    if err != nil {
        return nil, fmt.Errorf("refresh_token: find: %w", err)
    }
    return e, nil
}

func (r *RefreshTokenRepo) Delete(ctx context.Context, tokenHash []byte) error {
    _, err := r.pool.Exec(ctx,
        `UPDATE refresh_tokens SET revoked_at = now() WHERE token_hash = $1`, tokenHash)
    return err
}

func (r *RefreshTokenRepo) RevokeFamily(ctx context.Context, family string) error {
    _, err := r.pool.Exec(ctx,
        `UPDATE refresh_tokens SET revoked_at = now() WHERE family = $1 AND revoked_at IS NULL`, family)
    return err
}
```

**마이그레이션 추가 필요**: `refresh_tokens` 테이블에 `family` 칼럼이 없음.

```sql
-- migrations/000009_add_refresh_token_family.up.sql
ALTER TABLE refresh_tokens ADD COLUMN IF NOT EXISTS family TEXT;
CREATE INDEX idx_refresh_tokens_family ON refresh_tokens (family) WHERE revoked_at IS NULL;
```

```sql
-- migrations/000009_add_refresh_token_family.down.sql
DROP INDEX IF EXISTS idx_refresh_tokens_family;
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS family;
```

#### 1.4.4 TeamRepo

```go
// internal/repository/postgres/team_repo.go
func (r *TeamRepo) CreateTeam(t *domain.Team) error {
    query := `
        INSERT INTO teams (name, slug, owner_id)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at`
    return r.pool.QueryRow(context.Background(), query,
        t.Name, t.Slug, t.OwnerID,
    ).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *TeamRepo) ListTeamsByUser(userID string) ([]domain.Team, error) {
    query := `
        SELECT t.id, t.name, t.slug, t.owner_id, t.created_at, t.updated_at
        FROM teams t
        WHERE t.owner_id = $1
        UNION
        SELECT t.id, t.name, t.slug, t.owner_id, t.created_at, t.updated_at
        FROM teams t
        JOIN team_members tm ON tm.team_id = t.id
        WHERE tm.user_id = $1
        ORDER BY created_at DESC`
    // ... rows.Scan
}

func (r *TeamRepo) AddMember(m *domain.TeamMember) error {
    query := `
        INSERT INTO team_members (team_id, user_id, role, env_permissions, wrapped_project_key)
        VALUES ($1, $2, $3, $4, $5)`
    envPermsJSON, _ := json.Marshal(m.EnvPermissions)
    _, err := r.pool.Exec(context.Background(), query,
        m.TeamID, m.UserID, m.Role, envPermsJSON, m.WrappedProjectKey)
    return err
}

func (r *TeamRepo) IsMember(teamID, userID string) bool {
    var exists bool
    _ = r.pool.QueryRow(context.Background(), `
        SELECT EXISTS(
            SELECT 1 FROM teams WHERE id = $1 AND owner_id = $2
            UNION ALL
            SELECT 1 FROM team_members WHERE team_id = $1 AND user_id = $2
        )`, teamID, userID).Scan(&exists)
    return exists
}

func (r *TeamRepo) IsAdmin(teamID, userID string) bool {
    var exists bool
    _ = r.pool.QueryRow(context.Background(), `
        SELECT EXISTS(
            SELECT 1 FROM teams WHERE id = $1 AND owner_id = $2
            UNION ALL
            SELECT 1 FROM team_members WHERE team_id = $1 AND user_id = $2 AND role = 'admin'
        )`, teamID, userID).Scan(&exists)
    return exists
}
```

#### 1.4.5 DeviceRepo, AuditRepo, WaitlistRepo

동일 패턴으로 간결 구현. 각 기존 인터페이스 메서드 1:1 매핑.

```go
// DeviceRepo
func (r *DeviceRepo) RegisterDevice(d *domain.Device) error {
    query := `INSERT INTO devices (user_id, device_name, x25519_public_key)
              VALUES ($1, $2, $3) RETURNING id, created_at, last_seen_at`
    return r.pool.QueryRow(ctx, query, d.UserID, d.DeviceName, d.X25519PublicKey).
        Scan(&d.ID, &d.CreatedAt, &d.LastSeenAt)
}

// AuditRepo
func (r *AuditRepo) ListAuditLogs(userID string, limit int) ([]domain.AuditLog, error) {
    query := `SELECT id, user_id, COALESCE(vault_id::text, ''), action, COALESCE(detail, ''),
                     COALESCE(host(ip_address), ''), created_at
              FROM audit_logs WHERE user_id = $1
              ORDER BY created_at DESC LIMIT $2`
    // ...
}

// WaitlistRepo
func (r *WaitlistRepo) AddToWaitlist(email, plan, source string) error {
    query := `INSERT INTO waitlist (email, plan, source) VALUES ($1, $2, $3)
              ON CONFLICT (email) DO NOTHING`
    _, err := r.pool.Exec(ctx, query, email, plan, source)
    return err
}
```

### 1.5 ID 타입 변환 전략

**문제**: 현재 `domain.User.ID`가 `string` 타입이며 `generateUserID("gh_" + githubID)` 형태. DB는 UUID.

**해결**: 
- `domain.User.ID`는 `string`을 유지 (UUID 문자열 저장)
- `UpsertUser`가 DB에서 UUID를 생성하여 RETURNING으로 반환
- `GitHubCallback`에서 `generateUserID()` 호출 제거 -> UpsertUser 결과 사용
- JWT claims의 `sub`는 UUID 문자열

### 1.6 트랜잭션 패턴

팀 생성 + owner 멤버 추가처럼 복수 write가 필요한 경우:

```go
func (r *TeamRepo) CreateTeamWithOwner(t *domain.Team) error {
    tx, err := r.pool.Begin(context.Background())
    if err != nil {
        return fmt.Errorf("team: begin tx: %w", err)
    }
    defer tx.Rollback(context.Background())

    // Insert team
    err = tx.QueryRow(context.Background(),
        `INSERT INTO teams (name, slug, owner_id) VALUES ($1, $2, $3)
         RETURNING id, created_at, updated_at`,
        t.Name, t.Slug, t.OwnerID,
    ).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
    if err != nil {
        return fmt.Errorf("team: create: %w", err)
    }

    // Auto-add owner as admin
    _, err = tx.Exec(context.Background(),
        `INSERT INTO team_members (team_id, user_id, role, env_permissions)
         VALUES ($1, $2, 'admin', '["*"]')`,
        t.ID, t.OwnerID,
    )
    if err != nil {
        return fmt.Errorf("team: add owner: %w", err)
    }

    return tx.Commit(context.Background())
}
```

---

## 2. 백엔드 API 서버 설계

**담당**: 백엔드 엔지니어

### 2.1 server.go 리팩터링

**현재**: `internal/api/server.go:40-190` -- NewServer가 모든 store를 직접 생성.

**변경**: Config에 DatabaseURL 추가, PG pool을 생성하고 모든 store에 주입.

```go
// internal/api/server.go — 변경 후
type Config struct {
    // ... 기존 필드 유지 ...
    DatabaseURL string // 신규: PostgreSQL connection string
}

func NewServer(cfg Config) (*echo.Echo, func(), error) {
    e := echo.New()
    e.HideBanner = true

    // Database
    var (
        vaultStore   handler.VaultStore
        teamStore    handler.TeamStore
        deviceStore  handler.DeviceStore
        auditStore   handler.AuditStore
        waitlistStore handler.WaitlistStore
        userStore    billing.UserStore
        refreshStore postgres.RefreshTokenStore
        cleanup      func() = func() {}
    )

    if cfg.DatabaseURL != "" {
        db, err := postgres.New(context.Background(), cfg.DatabaseURL)
        if err != nil {
            return nil, nil, fmt.Errorf("server: database: %w", err)
        }
        cleanup = db.Close

        vaultStore = postgres.NewVaultRepo(db.Pool)
        teamStore = postgres.NewTeamRepo(db.Pool)
        deviceStore = postgres.NewDeviceRepo(db.Pool)
        auditStore = postgres.NewAuditRepo(db.Pool)
        waitlistStore = postgres.NewWaitlistRepo(db.Pool)
        userStore = postgres.NewUserRepo(db.Pool)
        refreshStore = postgres.NewRefreshTokenRepo(db.Pool)
    } else {
        // 로컬 개발 폴백
        vaultStore = handler.NewMemVaultStore()
        teamStore = handler.NewMemTeamStore()
        deviceStore = handler.NewMemDeviceStore()
        auditStore = handler.NewMemAuditStore()
        waitlistStore = handler.NewMemWaitlistStore()
        // userStore remains nil (billing webhook disabled)
        // refreshStore remains nil (in-memory auth)
    }

    // Services
    jwtSvc := auth.NewJWTService(cfg.JWTSecret)
    oauthSvc := auth.NewOAuthService(cfg.GitHubClientID, cfg.GitHubClientSecret, cfg.CallbackBase)
    billingSvc := billing.NewService(billing.Config{...}, userStore) // nil -> PgUserStore

    // Handlers — AuthHandler now accepts UserStore + RefreshTokenStore
    authH := handler.NewAuthHandler(oauthSvc, jwtSvc, cfg.DashboardURL, userStore, refreshStore)
    // ... 나머지 동일

    return e, cleanup, nil
}
```

**cmd/server/main.go 변경**:

```go
func main() {
    cfg := api.Config{
        // ... 기존 ...
        DatabaseURL: envOr("DATABASE_URL", ""), // 비어있으면 in-memory 폴백
    }

    e, cleanup, err := api.NewServer(cfg)
    if err != nil {
        log.Fatalf("server init: %v", err)
    }
    defer cleanup()

    // ... 나머지 동일
}
```

### 2.2 신규 API Endpoints

#### GET /api/v1/vaults/:id (Vault 상세)

```go
// handler/vault.go 에 추가
func (h *VaultHandler) Get(c echo.Context) error {
    claims := middleware.GetClaims(c)
    if claims == nil {
        return response.Err(c, domain.ErrUnauthorized)
    }

    vaultID := c.Param("id")
    vault, err := h.store.GetVault(vaultID, claims.UserID)
    if err != nil {
        return response.Err(c, err)
    }

    return response.OK(c, http.StatusOK, vault)
}
```

**라우트 등록**: `authed.GET("/vaults/:id", vaultH.Get)` -- `/vaults/:id/push`, `/vaults/:id/pull` 앞에 등록 (Echo는 더 구체적인 패턴 우선)

#### GET /api/v1/teams/:id/members (팀 멤버 목록)

```go
// handler/team.go 에 추가
func (h *TeamHandler) ListMembers(c echo.Context) error {
    claims := middleware.GetClaims(c)
    if claims == nil {
        return response.Err(c, domain.ErrUnauthorized)
    }

    teamID := c.Param("id")
    if !h.store.IsMember(teamID, claims.UserID) {
        return response.Err(c, domain.ErrForbidden)
    }

    members, err := h.store.ListMembers(teamID)
    if err != nil {
        return response.Err(c, err)
    }
    return response.OK(c, http.StatusOK, members)
}
```

**라우트 등록**: `authed.GET("/teams/:id/members", teamH.ListMembers)`

#### POST /api/v1/auth/exchange (Auth Code Exchange)

3번 섹션에서 상세 설계.

### 2.3 NewServer 반환값 변경

**현재**: `func NewServer(cfg Config) *echo.Echo`
**변경**: `func NewServer(cfg Config) (*echo.Echo, func(), error)` -- cleanup 함수 + error 반환

이유: DB 연결 실패 시 에러 반환 필요, cleanup으로 pool.Close() 보장.

### 2.4 AuthHandler 리팩터링

**현재**: AuthHandler가 `states` map과 `refresh` map을 직접 관리.
**변경**:
- `states` map은 유지 (OAuth state는 단기 TTL, in-memory 적합)
- `refresh` map -> RefreshTokenStore 인터페이스 사용
- UserStore를 주입받아 `GitHubCallback`에서 user upsert 수행

```go
type AuthHandler struct {
    oauth        *auth.OAuthService
    jwt          *auth.JWTService
    dashboardURL string
    userStore    UserStore              // 신규: DB user operations
    refreshStore RefreshTokenStore      // 신규: DB refresh token operations
    mu           sync.RWMutex
    states       map[string]stateEntry  // in-memory 유지 (short-lived)
}
```

---

## 3. 인증/보안 설계

**담당**: 인증/보안 아키텍트

### 3.1 Auth Code Exchange (A-04 해결)

**현재 문제**: OAuth callback이 URL query parameter로 access_token/refresh_token 전달.
- `internal/api/handler/auth.go:177-188` -- `?access_token=...&refresh_token=...`
- 브라우저 히스토리, 서버 로그, Referrer 헤더에 노출

**해결 설계**:

#### 3.1.1 임시 코드 생성 + 교환 플로우

```
1. OAuth callback 완료
2. 서버가 access_token + refresh_token을 "auth code"에 매핑하여 저장
3. Dashboard로 redirect: ?code=TEMP_AUTH_CODE
4. Dashboard가 POST /api/v1/auth/exchange { code: "..." } 호출
5. 서버가 code를 검증하고 tokens 반환
6. code는 30초 TTL, 1회 사용 후 삭제
```

#### 3.1.2 서버 측 구현

```go
// internal/api/handler/auth.go 에 추가

type authCodeEntry struct {
    accessToken  string
    refreshToken string
    userID       string
    plan         string
    expiresAt    time.Time
}

// AuthHandler에 authCodes map 추가
type AuthHandler struct {
    // ... 기존 ...
    authCodes map[string]authCodeEntry // 신규: 임시 auth code 저장
}

// generateAuthCode creates a short-lived auth code.
func generateAuthCode() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return "ac_" + hex.EncodeToString(b), nil
}

// GitHubCallback — 변경 부분 (기존 177-188행 교체)
func (h *AuthHandler) GitHubCallback(c echo.Context) error {
    // ... 기존 코드 유지 (code exchange, user upsert, token generation) ...

    // Dashboard redirect: auth code 방식으로 변경
    if entry.redirectTo == "dashboard" && h.dashboardURL != "" {
        authCode, err := generateAuthCode()
        if err != nil {
            return response.Err(c, err)
        }

        h.mu.Lock()
        h.authCodes[authCode] = authCodeEntry{
            accessToken:  accessToken,
            refreshToken: refreshToken,
            userID:       user.ID,
            plan:         user.Plan,
            expiresAt:    time.Now().Add(30 * time.Second), // 30초 TTL
        }
        h.mu.Unlock()

        redirectURL := fmt.Sprintf("%s/auth/callback?code=%s", h.dashboardURL, authCode)
        return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
    }

    // CLI redirect: 기존 방식 유지 (localhost callback은 네트워크 노출 없음)
    if len(entry.redirectTo) > 4 && entry.redirectTo[:4] == "cli:" {
        // ... 기존 코드 유지 (CLI는 localhost로 토큰 전달, 보안 위험 없음)
    }

    // API-only: 기존 JSON 응답 유지
    // ...
}

// Exchange converts an auth code to token pair. POST /api/v1/auth/exchange
func (h *AuthHandler) Exchange(c echo.Context) error {
    var req struct {
        Code string `json:"code"`
    }
    if err := c.Bind(&req); err != nil || req.Code == "" {
        return response.ErrMsg(c, http.StatusBadRequest, "BAD_REQUEST", "code required")
    }

    h.mu.Lock()
    entry, ok := h.authCodes[req.Code]
    if ok {
        delete(h.authCodes, req.Code) // 1회 사용 후 삭제
    }
    h.mu.Unlock()

    if !ok || time.Now().After(entry.expiresAt) {
        return response.ErrMsg(c, http.StatusBadRequest, "INVALID_CODE", "invalid or expired auth code")
    }

    return response.OK(c, http.StatusOK, domain.TokenPair{
        AccessToken:  entry.accessToken,
        RefreshToken: entry.refreshToken,
        ExpiresIn:    int(auth.AccessTokenTTL.Seconds()),
    })
}
```

**라우트 등록**: `v1.POST("/auth/exchange", authH.Exchange)` (public, no JWT)

#### 3.1.3 Dashboard 측 변경

```typescript
// apps/dashboard/src/app/auth/callback/page.tsx — 변경
function CallbackHandler() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const login = useAuthStore((s) => s.login);

  useEffect(() => {
    const code = searchParams.get("code");

    if (code) {
      // Auth code를 토큰으로 교환
      api.exchangeAuthCode(code)
        .then(({ access_token, refresh_token, expires_in }) => {
          document.cookie = `tene_access_token=${access_token}; path=/; max-age=${expires_in}; SameSite=Lax`;
          login(access_token, refresh_token);
          router.replace("/");
        })
        .catch(() => {
          router.replace("/login?error=exchange_failed");
        });
    } else {
      router.replace("/login");
    }
  }, [searchParams, login, router]);

  // ... UI 동일
}
```

### 3.2 CLI Auth Token Keychain 이전 (S-01 해결)

**현재**: `internal/cli/login.go:150-215` -- `~/.tene/auth.json` 평문 저장.

**변경**: `go-keyring` 사용 (이미 `go.mod`에 의존성 존재).

```go
// internal/cli/login.go — token storage 교체

import "github.com/zalando/go-keyring"

const (
    keychainService = "tene-cli"
    keychainAccessKey  = "access_token"
    keychainRefreshKey = "refresh_token"
)

func loadAuthToken() (string, error) {
    token, err := keyring.Get(keychainService, keychainAccessKey)
    if err == keyring.ErrNotFound {
        // 마이그레이션: 기존 auth.json에서 읽기 시도
        return loadAuthField("access_token")
    }
    if err != nil {
        // keychain 사용 불가 (headless 등) -> file fallback
        return loadAuthField("access_token")
    }
    return token, nil
}

func saveAuthToken(token string) error {
    err := keyring.Set(keychainService, keychainAccessKey, token)
    if err != nil {
        // Fallback: file 저장 (--no-keychain 또는 headless 환경)
        return saveAuthField("access_token", token)
    }
    // 성공 시 기존 auth.json의 access_token 필드 정리
    _ = removeAuthField("access_token")
    return nil
}

func saveRefreshToken(token string) error {
    err := keyring.Set(keychainService, keychainRefreshKey, token)
    if err != nil {
        return saveAuthField("refresh_token", token)
    }
    _ = removeAuthField("refresh_token")
    return nil
}

func clearAuthTokens() error {
    _ = keyring.Delete(keychainService, keychainAccessKey)
    _ = keyring.Delete(keychainService, keychainRefreshKey)
    _ = clearAuthFile()
    return nil
}

// removeAuthField removes a single field from auth.json.
func removeAuthField(field string) error {
    data, _ := readAuthFile()
    if data == nil {
        return nil
    }
    delete(data, field)
    if len(data) == 0 {
        return os.Remove(authFilePath())
    }
    return writeAuthFile(data)
}
```

**자동 마이그레이션**: `loadAuthToken()`이 keychain에서 못 찾으면 기존 `auth.json`에서 읽고, 다음 `saveAuthToken()` 호출 시 keychain으로 이전.

**--no-keychain 플래그**: 기존 `root.go`의 global flag에 추가.

```go
// internal/cli/root.go
var flagNoKeychain bool

func init() {
    rootCmd.PersistentFlags().BoolVar(&flagNoKeychain, "no-keychain", false,
        "Store auth tokens in file instead of OS keychain")
}
```

`saveAuthToken`/`loadAuthToken`에서 `flagNoKeychain` 체크하여 file-only 모드 지원.

### 3.3 Refresh Token DB 이전

`AuthHandler`의 `refresh` map을 `RefreshTokenStore` 인터페이스로 교체. (1.4.3 참조)

`GitHubCallback`에서:
```go
// 기존: h.refresh[hashToken(refreshToken)] = refreshEntry{...}
// 변경:
if h.refreshStore != nil {
    tokenHashBytes, _ := hex.DecodeString(hashToken(refreshToken))
    err := h.refreshStore.Store(c.Request().Context(),
        user.ID, tokenHashBytes, family, time.Now().Add(auth.RefreshTokenTTL))
    // ...
}
```

`RefreshToken` handler에서:
```go
// 기존: entry, ok := h.refresh[tokenHash]
// 변경:
if h.refreshStore != nil {
    entry, err := h.refreshStore.Find(c.Request().Context(), tokenHashBytes)
    if entry == nil { /* reuse detection */ }
    // Delete used token
    _ = h.refreshStore.Delete(c.Request().Context(), tokenHashBytes)
    // ... issue new token pair ...
}
```

---

## 4. 프론트엔드 Dashboard 설계

**담당**: 프론트엔드 아키텍트

### 4.1 API Client 확장

**현재**: `apps/dashboard/src/lib/api.ts` -- 6개 메서드만 구현.

**목표**: 17/17 메서드 완성.

```typescript
// apps/dashboard/src/lib/api.ts — 전체 확장

class ApiClient {
  private accessToken: string | null = null;
  private refreshToken: string | null = null;

  setToken(token: string) { this.accessToken = token; }
  setRefreshToken(token: string) { this.refreshToken = token; }
  clearToken() { this.accessToken = null; this.refreshToken = null; }

  private async request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...(options.headers as Record<string, string>),
    };

    if (this.accessToken) {
      headers["Authorization"] = `Bearer ${this.accessToken}`;
    }

    let res = await fetch(`${API_BASE}${path}`, { ...options, headers });

    // 401 자동 갱신
    if (res.status === 401 && this.refreshToken) {
      const refreshed = await this.refreshTokens();
      if (refreshed) {
        headers["Authorization"] = `Bearer ${this.accessToken}`;
        res = await fetch(`${API_BASE}${path}`, { ...options, headers });
      }
    }

    const json: ApiResponse<T> = await res.json();
    if (!json.ok) {
      throw new ApiError(json.message || json.error || "API error", json.error, res.status);
    }
    return json.data;
  }

  private async refreshTokens(): Promise<boolean> {
    try {
      const res = await fetch(`${API_BASE}/api/v1/auth/refresh`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refresh_token: this.refreshToken }),
      });
      const json: ApiResponse<TokenPair> = await res.json();
      if (json.ok && json.data) {
        this.accessToken = json.data.access_token;
        this.refreshToken = json.data.refresh_token;
        // Cookie 갱신
        document.cookie = `tene_access_token=${json.data.access_token}; path=/; max-age=${json.data.expires_in}; SameSite=Lax`;
        // Zustand store 업데이트 (이벤트 방출)
        window.dispatchEvent(new CustomEvent("tene:token-refreshed", {
          detail: { accessToken: json.data.access_token, refreshToken: json.data.refresh_token },
        }));
        return true;
      }
    } catch {
      // refresh 실패 -> 로그아웃
    }
    window.dispatchEvent(new CustomEvent("tene:auth-expired"));
    return false;
  }

  // === Auth ===
  exchangeAuthCode(code: string) {
    return this.request<TokenPair>("/api/v1/auth/exchange", {
      method: "POST",
      body: JSON.stringify({ code }),
    });
  }

  getMe() {
    return this.request<UserInfo>("/api/v1/auth/me");
  }

  signout() {
    return this.request<{ message: string }>("/api/v1/auth/signout", {
      method: "POST",
      body: JSON.stringify({ refresh_token: this.refreshToken }),
    });
  }

  // === Vaults ===
  listVaults() {
    return this.request<Vault[]>("/api/v1/vaults");
  }

  getVault(id: string) {
    return this.request<Vault>(`/api/v1/vaults/${id}`);
  }

  deleteVault(id: string) {
    return this.request<{ message: string }>(`/api/v1/vaults/${id}`, { method: "DELETE" });
  }

  // === Teams ===
  listTeams() {
    return this.request<Team[]>("/api/v1/teams");
  }

  createTeam(name: string, slug: string) {
    return this.request<Team>("/api/v1/teams", {
      method: "POST",
      body: JSON.stringify({ name, slug }),
    });
  }

  listTeamMembers(teamId: string) {
    return this.request<TeamMember[]>(`/api/v1/teams/${teamId}/members`);
  }

  inviteTeamMember(teamId: string, userId: string, role: string) {
    return this.request<TeamMember>(`/api/v1/teams/${teamId}/invite`, {
      method: "POST",
      body: JSON.stringify({ user_id: userId, role }),
    });
  }

  removeTeamMember(teamId: string, userId: string) {
    return this.request<{ message: string }>(`/api/v1/teams/${teamId}/members/${userId}`, {
      method: "DELETE",
    });
  }

  updateMemberRole(teamId: string, userId: string, role: string) {
    return this.request<{ message: string }>(`/api/v1/teams/${teamId}/members/${userId}/role`, {
      method: "PATCH",
      body: JSON.stringify({ role }),
    });
  }

  // === Devices ===
  listDevices() {
    return this.request<Device[]>("/api/v1/devices");
  }

  deleteDevice(id: string) {
    return this.request<{ message: string }>(`/api/v1/devices/${id}`, { method: "DELETE" });
  }

  // === Audit ===
  listAuditLogs() {
    return this.request<AuditLog[]>("/api/v1/audit");
  }

  // === Billing ===
  getSubscription() {
    return this.request<Subscription>("/api/v1/billing/subscription");
  }

  createCheckout(email: string) {
    return this.request<{ checkout_url: string }>("/api/v1/billing/checkout", {
      method: "POST",
      body: JSON.stringify({ email }),
    });
  }

  getPortal() {
    return this.request<{ portal_url: string }>("/api/v1/billing/portal", {
      method: "POST",
    });
  }

  // === Waitlist ===
  joinWaitlist(email: string) {
    return this.request<{ message: string }>("/api/v1/waitlist", {
      method: "POST",
      body: JSON.stringify({ email, source: "dashboard" }),
    });
  }
}
```

**타입 정의 추가**:

```typescript
// apps/dashboard/src/lib/types.ts (신규)
export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface UserInfo {
  user_id: string;
  plan: string;
  email?: string;
  name?: string;
  avatar_url?: string;
}

export interface Vault {
  id: string;
  user_id: string;
  project_name: string;
  vault_version: number;
  vault_hash: string;
  secret_count: number;
  size: number;
  created_at: string;
  updated_at: string;
  last_pushed_at?: string;
}

export interface Team {
  id: string;
  name: string;
  slug: string;
  owner_id: string;
  created_at: string;
}

export interface TeamMember {
  team_id: string;
  user_id: string;
  role: string;
  env_permissions: string[];
  joined_at: string;
}

export interface Device {
  id: string;
  device_name: string;
  last_seen_at: string;
  created_at: string;
}

export interface AuditLog {
  id: string;
  action: string;
  vault_id?: string;
  detail?: string;
  ip_address?: string;
  created_at: string;
}

export interface Subscription {
  plan: string;
  status: string;
  expires_at?: string;
}

export class ApiError extends Error {
  constructor(
    message: string,
    public code?: string,
    public status?: number,
  ) {
    super(message);
    this.name = "ApiError";
  }
}
```

### 4.2 TanStack Query Hooks

```typescript
// apps/dashboard/src/hooks/use-vaults.ts (신규)
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useVaults() {
  return useQuery({
    queryKey: ["vaults"],
    queryFn: () => api.listVaults(),
  });
}

export function useVault(id: string) {
  return useQuery({
    queryKey: ["vaults", id],
    queryFn: () => api.getVault(id),
    enabled: !!id,
  });
}

export function useDeleteVault() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.deleteVault(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["vaults"] }),
  });
}
```

```typescript
// apps/dashboard/src/hooks/use-teams.ts (신규)
export function useTeams() {
  return useQuery({
    queryKey: ["teams"],
    queryFn: () => api.listTeams(),
  });
}

export function useTeamMembers(teamId: string) {
  return useQuery({
    queryKey: ["teams", teamId, "members"],
    queryFn: () => api.listTeamMembers(teamId),
    enabled: !!teamId,
  });
}

export function useRemoveTeamMember() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ teamId, userId }: { teamId: string; userId: string }) =>
      api.removeTeamMember(teamId, userId),
    onSuccess: (_, { teamId }) =>
      queryClient.invalidateQueries({ queryKey: ["teams", teamId, "members"] }),
  });
}
```

```typescript
// apps/dashboard/src/hooks/use-devices.ts (신규)
export function useDevices() {
  return useQuery({
    queryKey: ["devices"],
    queryFn: () => api.listDevices(),
  });
}

// apps/dashboard/src/hooks/use-audit.ts (신규)
export function useAuditLogs() {
  return useQuery({
    queryKey: ["audit"],
    queryFn: () => api.listAuditLogs(),
  });
}

// apps/dashboard/src/hooks/use-billing.ts (신규)
export function useSubscription() {
  return useQuery({
    queryKey: ["billing", "subscription"],
    queryFn: () => api.getSubscription(),
  });
}
```

### 4.3 Auth Store 확장 (Token Refresh)

```typescript
// apps/dashboard/src/lib/auth-store.ts — 변경
interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;  // 신규
  user: UserInfo | null;
  isAuthenticated: boolean;
  login: (accessToken: string, refreshToken: string) => void;
  logout: () => void;
  setUser: (user: UserInfo | null) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      accessToken: null,
      refreshToken: null,
      user: null,
      isAuthenticated: false,

      login: (accessToken, refreshToken) => {
        api.setToken(accessToken);
        api.setRefreshToken(refreshToken);
        set({ accessToken, refreshToken, isAuthenticated: true });

        // User profile fetch
        api.getMe().then((user) => set({ user })).catch(() => {});
      },

      logout: () => {
        const { refreshToken } = get();
        if (refreshToken) {
          api.signout().catch(() => {});
        }
        api.clearToken();
        document.cookie = "tene_access_token=; path=/; max-age=0";
        set({ accessToken: null, refreshToken: null, user: null, isAuthenticated: false });
      },

      setUser: (user) => set({ user }),
    }),
    {
      name: "tene-auth",
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        user: state.user,
        isAuthenticated: state.isAuthenticated,
      }),
      onRehydrateStorage: () => (state) => {
        if (state?.accessToken) {
          api.setToken(state.accessToken);
        }
        if (state?.refreshToken) {
          api.setRefreshToken(state.refreshToken);
        }
      },
    }
  )
);
```

**Token Refresh 이벤트 리스너** (`apps/dashboard/src/components/providers.tsx`에 추가):

```typescript
"use client";
import { useEffect } from "react";
import { useAuthStore } from "@/lib/auth-store";

export function AuthEventListener() {
  const login = useAuthStore((s) => s.login);
  const logout = useAuthStore((s) => s.logout);

  useEffect(() => {
    const handleRefreshed = (e: CustomEvent) => {
      login(e.detail.accessToken, e.detail.refreshToken);
    };
    const handleExpired = () => logout();

    window.addEventListener("tene:token-refreshed", handleRefreshed as EventListener);
    window.addEventListener("tene:auth-expired", handleExpired);
    return () => {
      window.removeEventListener("tene:token-refreshed", handleRefreshed as EventListener);
      window.removeEventListener("tene:auth-expired", handleExpired);
    };
  }, [login, logout]);

  return null;
}
```

### 4.4 페이지 라이브 연동

#### Overview 페이지 (라이브 데이터)

```typescript
// apps/dashboard/src/app/(dashboard)/page.tsx — 변경
"use client";

import { StatCard } from "@/components/stat-card";
import { useVaults } from "@/hooks/use-vaults";
import { useDevices } from "@/hooks/use-devices";
import { useAuditLogs } from "@/hooks/use-audit";

export default function OverviewPage() {
  const { data: vaults } = useVaults();
  const { data: devices } = useDevices();
  const { data: auditLogs } = useAuditLogs();

  const totalSecrets = vaults?.reduce((sum, v) => sum + v.secret_count, 0) ?? 0;
  const weekAgo = new Date(Date.now() - 7 * 24 * 60 * 60 * 1000);
  const recentEvents = auditLogs?.filter(l => new Date(l.created_at) > weekAgo).length ?? 0;

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold">Overview</h1>
        <p className="text-muted text-sm mt-1">Your Tene Cloud at a glance</p>
      </div>
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard label="Vaults" value={vaults?.length ?? 0} sub="synced projects" accent />
        <StatCard label="Secrets" value={totalSecrets} sub="total keys" />
        <StatCard label="Devices" value={devices?.length ?? 0} sub="registered" />
        <StatCard label="Events" value={recentEvents} sub="this week" />
      </div>
      {/* Recent Activity */}
      <div className="rounded-xl border border-border bg-surface p-6">
        <h2 className="text-sm font-semibold mb-4">Recent Activity</h2>
        {auditLogs && auditLogs.length > 0 ? (
          <div className="space-y-2">
            {auditLogs.slice(0, 5).map((log) => (
              <div key={log.id} className="flex justify-between text-sm py-1 border-b border-border last:border-0">
                <span className="font-mono text-accent">{log.action}</span>
                <span className="text-muted text-xs">
                  {new Date(log.created_at).toLocaleString()}
                </span>
              </div>
            ))}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-12 text-muted">
            <span className="font-mono text-3xl mb-3">---</span>
            <p className="text-sm">No activity yet</p>
          </div>
        )}
      </div>
    </div>
  );
}
```

#### Billing 페이지 (실제 결제 연결)

```typescript
// apps/dashboard/src/app/(dashboard)/billing/page.tsx — 변경
"use client";

import { useSubscription } from "@/hooks/use-billing";
import { useAuthStore } from "@/lib/auth-store";
import { api } from "@/lib/api";

export default function BillingPage() {
  const { data: subscription } = useSubscription();
  const user = useAuthStore((s) => s.user);
  const isPro = subscription?.plan === "pro";

  const handleUpgrade = async () => {
    if (!user?.email) return;
    try {
      const { checkout_url } = await api.createCheckout(user.email);
      window.location.href = checkout_url;
    } catch (err) {
      // error handling
    }
  };

  const handlePortal = async () => {
    try {
      const { portal_url } = await api.getPortal();
      window.open(portal_url, "_blank");
    } catch (err) {
      // error handling
    }
  };

  return (
    <div className="space-y-6">
      {/* ... 기존 UI 유지, 버튼만 연결 ... */}
      {!isPro ? (
        <button onClick={handleUpgrade} className="...">
          Upgrade to Pro
        </button>
      ) : (
        <button onClick={handlePortal} className="...">
          Manage Subscription
        </button>
      )}
    </div>
  );
}
```

---

## 5. CLI/DX 설계

**담당**: CLI/DX 엔지니어

### 5.1 CLI Plan Pre-flight Check (B-05)

Pro 기능 (push/pull/sync/team) 사용 전에 JWT claims에서 plan을 로컬 디코딩하여 사전 체크.

**현재 문제**: Free 유저가 `tene push` 실행 시 HTTP 402 에러만 표시.

```go
// internal/cli/helpers.go 에 추가

import (
    "encoding/base64"
    "encoding/json"
    "strings"
)

// checkProPlan decodes the JWT locally (no verification) to check the plan claim.
// Returns nil if pro, error with upgrade message if free.
func checkProPlan(token string) error {
    parts := strings.Split(token, ".")
    if len(parts) != 3 {
        return nil // can't decode, let server decide
    }

    // Decode payload (base64url, no padding)
    payload, err := base64.RawURLEncoding.DecodeString(parts[1])
    if err != nil {
        return nil // can't decode, let server decide
    }

    var claims struct {
        Plan string `json:"plan"`
    }
    if err := json.Unmarshal(payload, &claims); err != nil {
        return nil
    }

    if claims.Plan == "pro" {
        return nil
    }

    return fmt.Errorf(`this feature requires a Pro plan ($5/month).

  Your current plan: Free

  Upgrade options:
    tene billing upgrade     Open checkout in browser
    https://tene.sh/#pricing Visit pricing page

  Pro includes: cloud sync, team sharing, audit log, device management`)
}
```

**적용 위치**:

```go
// internal/cli/push.go — runPush 시작 부분에 추가
func runPush(cmd *cobra.Command, args []string) error {
    token, err := loadAuthToken()
    if err != nil || token == "" {
        return fmt.Errorf("not logged in. Run 'tene login' first")
    }

    // Plan pre-check (로컬 JWT 디코딩, 서버 호출 없음)
    if err := checkProPlan(token); err != nil {
        return err
    }
    // ... 기존 코드 ...
}

// internal/cli/pull.go — 동일 패턴 적용
// internal/cli/sync_cmd.go — 동일 패턴 적용
// internal/cli/team.go — 각 서브커맨드(create, invite, remove)에 적용
```

### 5.2 CLI 에러 메시지 개선

**현재 문제**: API 에러가 그대로 표시 (`PRO_PLAN_REQUIRED (HTTP 402)`)

```go
// internal/cli/helpers.go — apiErrMsg 개선
func apiErrMsg(code, message string, status int) string {
    switch code {
    case "PRO_PLAN_REQUIRED":
        return "Pro plan required. Run 'tene billing upgrade' to subscribe ($5/month)"
    case "VERSION_CONFLICT":
        return "Remote vault has newer changes. Run 'tene pull' first, then push again"
    case "VAULT_NOT_FOUND":
        return "Vault not found. Run 'tene push' to create your first cloud vault"
    case "UNAUTHORIZED":
        return "Session expired. Run 'tene login' to re-authenticate"
    default:
        if message != "" {
            return message
        }
        return fmt.Sprintf("API error (HTTP %d)", status)
    }
}
```

---

## 6. 결제 통합 설계

**담당**: 결제 전문가

### 6.1 Billing Webhook 수정 (A-02 해결)

**현재 문제**: `billing.NewService(..., nil)` -- UserStore가 nil이어서 webhook에서 panic.

**해결**: Section 2.1에서 server.go가 PgUserStore를 주입하도록 변경. 추가로 nil guard 보강:

```go
// internal/billing/billing.go — HandleWebhook 에 nil guard 추가
func (s *Service) HandleWebhook(ctx context.Context, payload []byte, signature string) error {
    // ... 기존 코드 ...

    switch event.Meta.EventName {
    case "subscription_created", "subscription_updated":
        if s.store == nil {
            return fmt.Errorf("billing: user store not configured")
        }
        if event.Data.Attributes.Status == "active" {
            return s.store.UpdatePlan(ctx, email, "pro", customerID)
        }
        // ...
    }
}
```

### 6.2 Plan 업데이트 → JWT 반영 경로

Webhook이 DB의 `users.plan`을 `pro`로 변경. 하지만 기존 JWT의 `plan` claim은 변경 불가 (서명됨).

**반영 경로**:
1. Webhook → DB `users.plan = 'pro'`
2. 다음 token refresh 시 DB에서 최신 plan 조회 → 새 JWT에 `plan: "pro"` 포함
3. Dashboard: 401 자동 갱신으로 새 JWT 자동 적용

**AuthHandler.RefreshToken 변경**:

```go
func (h *AuthHandler) RefreshToken(c echo.Context) error {
    // ... 기존 토큰 검증 ...

    // DB에서 최신 plan 조회 (webhook이 업데이트했을 수 있음)
    currentPlan := entry.plan
    if h.userStore != nil {
        user, err := h.userStore.GetUserByID(c.Request().Context(), entry.userID)
        if err == nil {
            currentPlan = user.Plan // 최신 plan 반영
        }
    }

    // Issue new token pair with latest plan
    newAccess, err := h.jwt.GenerateAccessToken(entry.userID, currentPlan, "", "user")
    // ...
}
```

### 6.3 Landing Pro CTA Intent 흐름

**Landing 변경** (`apps/web/src/data/pricing.ts`):

```typescript
// 기존: cta: { label: "Get started", action: "signup" }
// 변경:
cta: { label: "Get started", action: "signup" }
// signup action의 URL 변경 (pricing 컴포넌트에서):
// 기존: ${dashboardUrl}/login
// 변경: ${dashboardUrl}/login?intent=upgrade
```

**Dashboard Login 변경** (`apps/dashboard/src/app/(auth)/login/page.tsx`):

```typescript
export default function LoginPage() {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || "https://api.tene.sh";
  const searchParams = useSearchParams();
  const intent = searchParams.get("intent"); // "upgrade" or null

  // intent를 OAuth state에 포함시키기 위해 redirect param에 전달
  const authUrl = intent === "upgrade"
    ? `${apiUrl}/api/v1/auth/github/authorize?redirect=dashboard&intent=upgrade`
    : `${apiUrl}/api/v1/auth/github/authorize?redirect=dashboard`;

  return (
    // ... 기존 UI ...
    <a href={authUrl}>Continue with GitHub</a>
    // ...
  );
}
```

**Auth Callback 변경** (intent=upgrade인 경우):

```typescript
// apps/dashboard/src/app/auth/callback/page.tsx
// 코드 교환 성공 후:
const intent = searchParams.get("intent");
if (intent === "upgrade") {
  router.replace("/billing"); // 바로 billing으로 이동
} else {
  router.replace("/");
}
```

**서버 측**: `stateEntry`에 `intent` 필드 추가, callback redirect URL에 `&intent=upgrade` 포함.

---

## 7. QA/테스트 전략

**담당**: QA 전략가

### 7.1 Handler 테스트 설계

**테스트 파일 위치**: `internal/api/handler/*_test.go`

**테스트 프레임워크**: `httptest` + `testify` (table-driven)

#### 7.1.1 VaultHandler 테스트 예시

```go
// internal/api/handler/vault_test.go
package handler_test

import (
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/labstack/echo/v4"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/agent-kay-it/tene/internal/api/handler"
    "github.com/agent-kay-it/tene/internal/api/middleware"
    "github.com/agent-kay-it/tene/internal/auth"
)

func TestVaultHandler_List(t *testing.T) {
    tests := []struct {
        name       string
        setup      func(store *handler.MemVaultStore)
        claims     *auth.Claims
        wantStatus int
        wantCount  int
    }{
        {
            name:       "no vaults returns empty list",
            claims:     &auth.Claims{UserID: "user1", Plan: "pro"},
            wantStatus: http.StatusOK,
            wantCount:  0,
        },
        {
            name: "returns only user's vaults",
            setup: func(store *handler.MemVaultStore) {
                _ = store.CreateVault(&domain.Vault{
                    UserID: "user1", ProjectName: "proj1",
                    S3Key: "vaults/user1/proj1/vault.enc", VaultHash: make([]byte, 32),
                })
                _ = store.CreateVault(&domain.Vault{
                    UserID: "user2", ProjectName: "proj2",
                    S3Key: "vaults/user2/proj2/vault.enc", VaultHash: make([]byte, 32),
                })
            },
            claims:     &auth.Claims{UserID: "user1", Plan: "pro"},
            wantStatus: http.StatusOK,
            wantCount:  1,
        },
        {
            name:       "unauthorized without claims",
            claims:     nil,
            wantStatus: http.StatusUnauthorized,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            store := handler.NewMemVaultStore()
            if tt.setup != nil {
                tt.setup(store)
            }

            h := handler.NewVaultHandler(store, nil)
            e := echo.New()
            req := httptest.NewRequest(http.MethodGet, "/api/v1/vaults", nil)
            rec := httptest.NewRecorder()
            c := e.NewContext(req, rec)

            if tt.claims != nil {
                c.Set(middleware.ContextKeyClaims, tt.claims)
            }

            err := h.List(c)
            require.NoError(t, err)
            assert.Equal(t, tt.wantStatus, rec.Code)
        })
    }
}
```

### 7.2 테스트 대상 매트릭스

| Handler | 테스트 시나리오 수 | 우선순위 |
|---------|:------------------:|:--------:|
| AuthHandler | 12 (authorize, callback, refresh, exchange, me, signout) | P0 |
| VaultHandler | 10 (list, create, push, pull, delete, get) | P0 |
| BillingHandler | 8 (subscription, checkout, portal, webhook) | P1 |
| TeamHandler | 8 (create, list, invite, remove, update_role, list_members) | P1 |
| DeviceHandler | 4 (register, list, delete) | P2 |
| AuditHandler | 2 (list) | P2 |
| WaitlistHandler | 2 (register) | P3 |

### 7.3 Repository 테스트 (Integration)

`testcontainers-go`로 PostgreSQL 컨테이너 기동하여 실제 DB 테스트.

```go
// internal/repository/postgres/vault_repo_test.go
func TestVaultRepo_CRUD(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    ctx := context.Background()
    pool := setupTestDB(t) // testcontainers 로 PG 기동
    defer pool.Close()

    repo := postgres.NewVaultRepo(pool)

    // Create
    v := &domain.Vault{
        UserID:      "test-user-uuid",
        ProjectName: "test-project",
        S3Key:       "vaults/test/vault.enc",
        VaultHash:   make([]byte, 32),
    }
    err := repo.CreateVault(v)
    require.NoError(t, err)
    assert.NotEmpty(t, v.ID)

    // Get
    got, err := repo.GetVault(v.ID, v.UserID)
    require.NoError(t, err)
    assert.Equal(t, v.ProjectName, got.ProjectName)

    // ... List, Update, Delete
}
```

### 7.4 E2E 시나리오

| 시나리오 | 커버 범위 |
|----------|----------|
| 가입 → push → pull | OAuth → JWT → vault CRUD → S3 |
| Free → Pro 업그레이드 | Billing webhook → plan update → JWT refresh |
| 팀 생성 → 초대 → 제거 | Team CRUD → key rotation |
| Dashboard 로그인 → 페이지 탐색 | Auth code exchange → token refresh → API calls |

---

## 8. 인프라/운영 설계

**담당**: 인프라 엔지니어

### 8.1 Migration 자동화

**현재**: 수동으로 SQL 파일 실행.
**변경**: 서버 시작 시 자동 마이그레이션 (golang-migrate 사용).

```go
// cmd/server/main.go — 추가
import (
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func runMigrations(databaseURL string) error {
    m, err := migrate.New("file://migrations", databaseURL)
    if err != nil {
        return fmt.Errorf("migration: init: %w", err)
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("migration: up: %w", err)
    }

    log.Println("migrations applied successfully")
    return nil
}

func main() {
    // ... cfg 설정 ...

    // DB 마이그레이션 실행
    if dbURL := envOr("DATABASE_URL", ""); dbURL != "" {
        if err := runMigrations(dbURL); err != nil {
            log.Fatalf("migration failed: %v", err)
        }
    }

    e, cleanup, err := api.NewServer(cfg)
    // ...
}
```

**go.mod 추가**: `github.com/golang-migrate/migrate/v4`

**Dockerfile 변경**: migrations/ 디렉터리를 컨테이너에 복사.

```dockerfile
# Dockerfile.server 에 추가
COPY migrations/ /app/migrations/
```

### 8.2 Health Check 개선

**현재**: `HealthHandler.Readiness`가 항상 200 반환.
**변경**: DB ping 추가.

```go
// internal/api/handler/health.go — 변경
type HealthHandler struct {
    db interface{ Ping(ctx context.Context) error } // postgres.DB
}

func NewHealthHandler(db interface{ Ping(ctx context.Context) error }) *HealthHandler {
    return &HealthHandler{db: db}
}

func (h *HealthHandler) Readiness(c echo.Context) error {
    if h.db != nil {
        ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
        defer cancel()
        if err := h.db.Ping(ctx); err != nil {
            return c.JSON(http.StatusServiceUnavailable, map[string]string{
                "status": "not_ready",
                "error":  "database unreachable",
            })
        }
    }

    return c.JSON(http.StatusOK, map[string]string{"status": "ready"})
}
```

### 8.3 Graceful Shutdown 개선

**현재**: `cmd/server/main.go:29-48` -- 기본 graceful shutdown은 구현됨.
**추가**: DB pool close를 cleanup에 포함.

```go
func main() {
    // ...
    e, cleanup, err := api.NewServer(cfg)
    if err != nil {
        log.Fatalf("server init: %v", err)
    }

    go func() {
        if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
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

    cleanup() // DB pool close
    log.Println("server exited cleanly")
}
```

### 8.4 ECS Task Definition 변경

**기존 환경변수에 추가**:

```json
{
    "name": "DATABASE_URL",
    "valueFrom": "arn:aws:secretsmanager:ap-northeast-2:507221376909:secret:tene/prod/api-secrets:DATABASE_URL::"
}
```

**Terraform 변경** (`infra/terraform/modules/ecs/main.tf`):
- `secrets` 블록에 DATABASE_URL 추가
- health check path는 기존 `/health` 유지 (liveness)
- ECS service의 `health_check_grace_period_seconds` 를 60으로 설정 (migration 시간 고려)

---

## 9. 통합 아키텍처 설계

**담당**: 통합 아키텍트

### 9.1 데이터 흐름도 (변경 후)

```
=== 가입 플로우 ===
Browser -> Landing Pro CTA
  -> Dashboard /login?intent=upgrade
    -> GET /api/v1/auth/github/authorize?redirect=dashboard&intent=upgrade
      -> GitHub OAuth
        -> GET /api/v1/auth/github/callback
          -> PG: INSERT users (plan=free)
          -> 임시 auth code 생성
          -> redirect: Dashboard /auth/callback?code=ac_xxx&intent=upgrade
            -> POST /api/v1/auth/exchange { code }
              -> { access_token, refresh_token }
            -> intent=upgrade -> redirect /billing
              -> POST /api/v1/billing/checkout { email }
                -> LemonSqueezy checkout URL
                  -> 결제 완료
                    -> POST /api/v1/billing/webhook
                      -> PG: users.plan = 'pro'

=== Sync 플로우 ===
CLI: tene push
  -> loadAuthToken() (keychain or file)
  -> checkProPlan(token) (로컬 JWT 디코딩)
  -> POST /api/v1/vaults/:id/push (Bearer token)
    -> JWT middleware validates
    -> Plan check (claims.Plan == "pro")
    -> S3 upload
    -> PG: UPDATE vaults SET vault_version = vault_version + 1
    -> PG: INSERT audit_logs

=== Token Refresh ===
Dashboard API call -> 401
  -> POST /api/v1/auth/refresh { refresh_token }
    -> PG: Find refresh_token by hash
    -> PG: Get user (latest plan from DB)
    -> Generate new JWT with current plan
    -> PG: Delete old refresh_token, store new
    -> { new access_token, new refresh_token }
  -> Retry original request

=== Dashboard 페이지 ===
Overview -> GET /vaults + GET /devices + GET /audit (TanStack Query parallel)
Vaults   -> GET /vaults (list) + GET /vaults/:id (detail)
Team     -> GET /teams + GET /teams/:id/members
Billing  -> GET /billing/subscription + POST /billing/checkout
Devices  -> GET /devices
Audit    -> GET /audit
```

### 9.2 컴포넌트 의존성 그래프

```
cmd/server/main.go
  |
  +-- internal/api/server.go
  |     |
  |     +-- internal/repository/postgres/  (신규)
  |     |     +-- postgres.go (pgxpool.Pool)
  |     |     +-- user_repo.go
  |     |     +-- vault_repo.go
  |     |     +-- team_repo.go
  |     |     +-- device_repo.go
  |     |     +-- audit_repo.go
  |     |     +-- refresh_token_repo.go
  |     |     +-- waitlist_repo.go
  |     |
  |     +-- internal/api/handler/
  |     |     +-- auth.go (UserStore, RefreshTokenStore 주입)
  |     |     +-- vault.go (VaultStore 인터페이스)
  |     |     +-- team.go (TeamStore 인터페이스)
  |     |     +-- health.go (DB ping 주입)
  |     |     +-- billing.go (billing.Service -> UserStore)
  |     |
  |     +-- internal/auth/ (변경 없음)
  |     +-- internal/billing/ (nil guard 추가)
  |
  +-- migrations/ (golang-migrate 자동 실행)

cmd/tene/ (CLI)
  +-- internal/cli/
        +-- login.go (go-keyring 이전)
        +-- push.go (checkProPlan 추가)
        +-- pull.go (checkProPlan 추가)
        +-- team.go (checkProPlan 추가)

apps/dashboard/
  +-- src/lib/api.ts (17 메서드 + token refresh)
  +-- src/lib/auth-store.ts (refreshToken 저장)
  +-- src/lib/types.ts (신규)
  +-- src/hooks/ (신규: TanStack Query hooks)
  +-- src/app/auth/callback/page.tsx (auth code exchange)
  +-- src/app/(dashboard)/**/page.tsx (라이브 데이터)
```

### 9.3 인터페이스 일관성

모든 store 인터페이스는 `internal/api/handler/` 패키지에 정의됨. PostgreSQL 구현체는 `internal/repository/postgres/`에서 이 인터페이스를 구현.

**신규 인터페이스 (handler 패키지에 추가)**:

```go
// internal/api/handler/auth.go 에 추가
type UserStore interface {
    UpsertUser(ctx context.Context, u *domain.User) error
    GetUserByID(ctx context.Context, id string) (*domain.User, error)
    GetUserByGitHubID(ctx context.Context, githubID int64) (*domain.User, error)
    GetPublicKey(ctx context.Context, userID string) ([]byte, error)
}

type RefreshTokenStore interface {
    Store(ctx context.Context, userID string, tokenHash []byte, family string, expiresAt time.Time) error
    Find(ctx context.Context, tokenHash []byte) (*RefreshTokenEntry, error)
    Delete(ctx context.Context, tokenHash []byte) error
    RevokeFamily(ctx context.Context, family string) error
}
```

billing.UserStore 인터페이스 (UpdatePlan, GetLemonCustomerID)는 postgres.UserRepo가 함께 구현. server.go에서 `*postgres.UserRepo`를 billing.UserStore와 handler.UserStore 양쪽에 전달.

### 9.4 cross-component 데이터 일관성

| 데이터 | 소스 | 소비자 | 일관성 보장 |
|--------|------|--------|------------|
| user.plan | PG users 테이블 | JWT claims, Dashboard | Token refresh 시 DB에서 최신 plan 조회 |
| vault metadata | PG vaults 테이블 | CLI push/pull, Dashboard | Optimistic locking (vault_version) |
| refresh_token | PG refresh_tokens 테이블 | AuthHandler | Family tracking, revoke on reuse |
| team membership | PG team_members 테이블 | CLI team, Dashboard | ACID (트랜잭션) |

---

## 10. 사용자 여정/UX 설계

**담당**: 제품 매니저

### 10.1 Free 사용자 여정

```
1. tene.sh 방문 -> "Install now" -> CLI 설치
2. tene init -> 로컬 vault 생성 (Master Password + Recovery Key)
3. tene set API_KEY "sk-xxx" -> 시크릿 저장
4. tene run -- node server.js -> 시크릿 주입하여 실행
5. tene push -> "Pro plan required. Run 'tene billing upgrade' to subscribe ($5/month)"
   (친절한 업그레이드 안내 + 가격 + 기능 나열)
```

### 10.2 Free -> Pro 업그레이드 여정 (3가지 경로)

**경로 A: CLI에서 업그레이드**
```
1. tene billing upgrade
2. 브라우저 열림 -> LemonSqueezy checkout
3. 결제 완료 -> webhook -> plan=pro
4. tene push -> 성공 (다음 token refresh 시 plan=pro 반영)
```

**경로 B: Landing 페이지에서 업그레이드**
```
1. tene.sh -> Pro "Get started" 클릭
2. Dashboard /login?intent=upgrade -> GitHub OAuth
3. 로그인 성공 -> 자동 /billing 이동
4. "Upgrade to Pro" 클릭 -> LemonSqueezy checkout
5. 결제 완료 -> webhook -> plan=pro -> Dashboard 자동 반영
```

**경로 C: Dashboard에서 업그레이드**
```
1. Dashboard /billing -> "Upgrade to Pro" 클릭
2. LemonSqueezy checkout
3. 결제 완료 -> webhook -> plan=pro
4. 다음 API 호출 시 token refresh -> plan=pro JWT
```

### 10.3 Dashboard 정보 아키텍처

```
Dashboard
  |
  +-- Overview (/)
  |     - 4개 stat 카드 (Vaults, Secrets, Devices, Events)
  |     - Recent Activity (최근 5개 audit log)
  |     - Quick Start 가이드
  |
  +-- Vaults (/vaults)
  |     - 테이블: Project, Secrets, Version, Last Sync, Size
  |     - 빈 상태: "Run tene push" 안내
  |     +-- Vault Detail (/vaults/[id])
  |           - 메타데이터 상세 (push 이력, 버전)
  |
  +-- Team (/team)
  |     - 팀 없음: "tene team create" 안내
  |     - 팀 있음: 멤버 테이블 + Invite 모달
  |     - Key rotation 안내
  |
  +-- Devices (/devices)
  |     - 디바이스 카드 그리드
  |     - Last seen, 삭제 기능
  |
  +-- Audit (/audit)
  |     - 필터 (All/Push/Pull/Login/Delete)
  |     - 테이블: Time, Action, Target, IP
  |
  +-- Billing (/billing)
        - Free/Pro 비교 카드
        - 업그레이드/관리 버튼
        - Usage 바 (Vaults, Members, Storage)
```

### 10.4 에러 상태 UX

| 상황 | 표시 | 행동 |
|------|------|------|
| API 연결 실패 | "Unable to connect to Tene Cloud" 배너 | 5초 후 재시도 |
| 401 (토큰 만료) | 자동 refresh, 실패 시 로그인 redirect | 투명한 처리 |
| 402 (Pro 필요) | "Upgrade to Pro" CTA와 기능 설명 | billing 페이지로 안내 |
| 404 (Vault 없음) | "Run tene push" 안내 | 빈 상태 UI |
| 409 (버전 충돌) | "Pull first" 안내 | CLI 명령어 표시 |

---

## 11. 구현 가이드

### 11.1 Sprint별 구현 순서

#### Sprint 1: 기반 구축 (7-8일)

| 순서 | 작업 | 파일 | 의존성 |
|:----:|------|------|--------|
| 1 | `go get github.com/jackc/pgx/v5` | `go.mod` | 없음 |
| 2 | `go get github.com/golang-migrate/migrate/v4` | `go.mod` | 없음 |
| 3 | migration 000009 (refresh_tokens.family) | `migrations/000009_*.sql` | 없음 |
| 4 | `internal/repository/postgres/postgres.go` | 신규 | 1 |
| 5 | `internal/repository/postgres/user_repo.go` | 신규 | 4 |
| 6 | `internal/repository/postgres/vault_repo.go` | 신규 | 4 |
| 7 | `internal/repository/postgres/team_repo.go` | 신규 | 4 |
| 8 | `internal/repository/postgres/device_repo.go` | 신규 | 4 |
| 9 | `internal/repository/postgres/audit_repo.go` | 신규 | 4 |
| 10 | `internal/repository/postgres/refresh_token_repo.go` | 신규 | 3, 4 |
| 11 | `internal/repository/postgres/waitlist_repo.go` | 신규 | 4 |
| 12 | `internal/api/handler/auth.go` -- UserStore/RefreshTokenStore 인터페이스 추가, 주입 | 수정 | 5, 10 |
| 13 | `internal/api/server.go` -- NewServer 리팩터링 | 수정 | 5-12 |
| 14 | `internal/billing/billing.go` -- nil guard 추가 | 수정 | 없음 |
| 15 | `internal/api/handler/health.go` -- DB ping 추가 | 수정 | 4 |
| 16 | `cmd/server/main.go` -- DatabaseURL + migration + cleanup | 수정 | 2, 13 |
| 17 | `internal/cli/login.go` -- go-keyring 이전 | 수정 | 없음 |
| 18 | `apps/dashboard/src/lib/api.ts` -- token refresh interceptor | 수정 | 없음 |
| 19 | `apps/dashboard/src/lib/auth-store.ts` -- refreshToken 저장 | 수정 | 18 |

#### Sprint 2: 핵심 경험 (4-5일)

| 순서 | 작업 | 파일 | 의존성 |
|:----:|------|------|--------|
| 1 | Auth code exchange endpoint | `handler/auth.go`, `server.go` | Sprint 1 완료 |
| 2 | Dashboard callback auth code 방식 변경 | `auth/callback/page.tsx` | 1 |
| 3 | CLI checkProPlan 헬퍼 추가 | `internal/cli/helpers.go` | 없음 |
| 4 | push/pull/sync/team에 plan pre-check 적용 | `push.go`, `pull.go`, `sync_cmd.go`, `team.go` | 3 |
| 5 | Landing Pro CTA intent=upgrade | `apps/web/src/data/pricing.ts`, 관련 컴포넌트 | 없음 |
| 6 | Dashboard login?intent=upgrade 처리 | `login/page.tsx`, `auth/callback/page.tsx` | 2, 5 |
| 7 | Dashboard billing 버튼 실제 연결 | `billing/page.tsx` | Sprint 1 완료 |
| 8 | Dashboard overview 라이브 데이터 | `page.tsx` (overview) | Sprint 1 완료 |

#### Sprint 3: 기능 완성 (5-6일)

| 순서 | 작업 | 파일 | 의존성 |
|:----:|------|------|--------|
| 1 | `apps/dashboard/src/lib/types.ts` 타입 정의 | 신규 | 없음 |
| 2 | API client 11개 메서드 추가 | `api.ts` | 1 |
| 3 | TanStack Query hooks 생성 | `hooks/use-*.ts` | 2 |
| 4 | GET /vaults/:id endpoint 추가 | `handler/vault.go`, `server.go` | Sprint 1 완료 |
| 5 | GET /teams/:id/members endpoint 추가 | `handler/team.go`, `server.go` | Sprint 1 완료 |
| 6 | Vaults 페이지 라이브 연동 | `vaults/page.tsx`, `vaults/[id]/page.tsx` | 3, 4 |
| 7 | Team 페이지 라이브 연동 | `team/page.tsx` | 3, 5 |
| 8 | Devices 페이지 라이브 연동 | `devices/page.tsx` | 3 |
| 9 | Audit 페이지 라이브 연동 | `audit/page.tsx` | 3 |
| 10 | Team key rotation 구현 | `handler/team.go` | 5 |

#### Sprint 4: 마무리 (3-4일)

| 순서 | 작업 | 파일 | 의존성 |
|:----:|------|------|--------|
| 1 | VaultHandler 테스트 | `handler/vault_test.go` | Sprint 1-3 완료 |
| 2 | AuthHandler 테스트 | `handler/auth_test.go` | Sprint 1-2 완료 |
| 3 | BillingHandler 테스트 | `handler/billing_test.go` | Sprint 1 완료 |
| 4 | TeamHandler 테스트 | `handler/team_test.go` | Sprint 1 완료 |
| 5 | 문서 정렬 (CLAUDE.md HS256 명시 등) | `CLAUDE.md`, `.claude/rules/` | 없음 |

### 11.2 Module Map

| Module | 범위 | 예상 코드 |
|--------|------|----------|
| module-1 | PG Repository (7 repos + DB pool) | ~800 lines, 8 files |
| module-2 | Server.go 리팩터링 + handler 인터페이스 추가 | ~200 lines, 4 files |
| module-3 | Auth code exchange + token refresh | ~300 lines, 3 files (Go + TS) |
| module-4 | CLI keychain + plan pre-check | ~150 lines, 5 files |
| module-5 | Dashboard API client + hooks | ~400 lines, 8 files (TS) |
| module-6 | Dashboard 페이지 라이브 연동 | ~500 lines, 6 files (TSX) |
| module-7 | Billing 통합 (webhook fix + CTA + portal) | ~200 lines, 4 files |
| module-8 | 신규 endpoints (vault detail, team members) | ~100 lines, 3 files |
| module-9 | Infra (migration 자동화, health check, shutdown) | ~100 lines, 3 files |
| module-10 | Tests (handler + repo) | ~600 lines, 6 files |

### 11.3 Session Guide

**권장 세션 분할**:

| 세션 | 모듈 | 예상 시간 | 비고 |
|:----:|------|:--------:|------|
| 1 | module-1 (PG Repository) | 3-4h | 가장 큰 모듈, 집중 필요 |
| 2 | module-2 + module-9 (Server + Infra) | 2-3h | server.go + main.go 변경 |
| 3 | module-3 + module-4 (Auth + CLI) | 2-3h | 보안 관련 묶음 |
| 4 | module-7 + module-8 (Billing + Endpoints) | 2h | 백엔드 마무리 |
| 5 | module-5 + module-6 (Dashboard 전체) | 3-4h | 프론트엔드 일괄 |
| 6 | module-10 (Tests) | 3h | 마지막 세션 |

**사용법**: `/pdca do tene-cloud-improvement --scope module-1`

---

## 부록 A: 변경 파일 전체 목록

### 신규 파일 (18개)

| 파일 | 크기 |
|------|------|
| `internal/repository/postgres/postgres.go` | ~50 lines |
| `internal/repository/postgres/user_repo.go` | ~120 lines |
| `internal/repository/postgres/vault_repo.go` | ~150 lines |
| `internal/repository/postgres/team_repo.go` | ~150 lines |
| `internal/repository/postgres/device_repo.go` | ~60 lines |
| `internal/repository/postgres/audit_repo.go` | ~50 lines |
| `internal/repository/postgres/refresh_token_repo.go` | ~80 lines |
| `internal/repository/postgres/waitlist_repo.go` | ~30 lines |
| `migrations/000009_add_refresh_token_family.up.sql` | ~3 lines |
| `migrations/000009_add_refresh_token_family.down.sql` | ~2 lines |
| `apps/dashboard/src/lib/types.ts` | ~70 lines |
| `apps/dashboard/src/hooks/use-vaults.ts` | ~30 lines |
| `apps/dashboard/src/hooks/use-teams.ts` | ~40 lines |
| `apps/dashboard/src/hooks/use-devices.ts` | ~15 lines |
| `apps/dashboard/src/hooks/use-audit.ts` | ~15 lines |
| `apps/dashboard/src/hooks/use-billing.ts` | ~15 lines |
| `internal/api/handler/vault_test.go` | ~100 lines |
| `internal/api/handler/auth_test.go` | ~150 lines |

### 수정 파일 (17개)

| 파일 | 변경 범위 |
|------|----------|
| `go.mod` | pgx/v5, golang-migrate 추가 |
| `cmd/server/main.go` | DatabaseURL, migration, cleanup |
| `internal/api/server.go` | NewServer 시그니처 변경, PG 주입 |
| `internal/api/handler/auth.go` | UserStore/RefreshTokenStore 주입, auth code exchange, user upsert |
| `internal/api/handler/vault.go` | Get() 메서드 추가 |
| `internal/api/handler/team.go` | ListMembers() 메서드 추가 |
| `internal/api/handler/health.go` | DB ping 추가 |
| `internal/billing/billing.go` | nil guard 추가 |
| `internal/cli/login.go` | go-keyring 이전 |
| `internal/cli/helpers.go` | checkProPlan, apiErrMsg 개선 |
| `internal/cli/push.go` | plan pre-check 추가 |
| `internal/cli/pull.go` | plan pre-check 추가 |
| `internal/cli/sync_cmd.go` | plan pre-check 추가 |
| `internal/cli/team.go` | plan pre-check 추가 |
| `apps/dashboard/src/lib/api.ts` | 17 메서드 + token refresh |
| `apps/dashboard/src/lib/auth-store.ts` | refreshToken 저장 + getMe |
| `apps/dashboard/src/app/auth/callback/page.tsx` | auth code exchange 방식 |

### Dashboard 페이지 수정 (6개)

| 파일 | 변경 |
|------|------|
| `apps/dashboard/src/app/(dashboard)/page.tsx` | 라이브 데이터 (overview) |
| `apps/dashboard/src/app/(dashboard)/billing/page.tsx` | 실제 결제 연결 |
| `apps/dashboard/src/app/(dashboard)/vaults/page.tsx` | TanStack Query 연동 |
| `apps/dashboard/src/app/(dashboard)/team/page.tsx` | TanStack Query 연동 |
| `apps/dashboard/src/app/(dashboard)/devices/page.tsx` | TanStack Query 연동 |
| `apps/dashboard/src/app/(dashboard)/audit/page.tsx` | TanStack Query 연동 |

---

## 부록 B: SQL 쿼리 인덱스 활용

기존 마이그레이션에서 정의된 인덱스가 모든 repo 쿼리를 커버하는지 검증:

| 쿼리 패턴 | 사용 인덱스 | 상태 |
|-----------|------------|:----:|
| `users WHERE github_id = ?` | `idx_users_github_id` | OK |
| `users WHERE email = ?` | `idx_users_email` | OK |
| `vaults WHERE user_id = ? AND project_name = ?` | `idx_vaults_user_project` | OK |
| `vaults WHERE user_id = ?` | `idx_vaults_user_id` | OK |
| `refresh_tokens WHERE token_hash = ?` | `idx_refresh_tokens_hash` | OK |
| `refresh_tokens WHERE family = ?` | `idx_refresh_tokens_family` | 신규 (000009) |
| `teams WHERE slug = ?` | `idx_teams_slug` | OK |
| `teams WHERE owner_id = ?` | `idx_teams_owner_id` | OK |
| `team_members WHERE team_id = ? AND user_id = ?` | PK (team_id, user_id) | OK |
| `team_members WHERE user_id = ?` | `idx_team_members_user_id` | OK |
| `devices WHERE user_id = ?` | `idx_devices_user_id` | OK |
| `audit_logs WHERE user_id = ?` | `idx_audit_logs_user_id` | OK |
| `waitlist WHERE email = ?` | `idx_waitlist_email` | OK |

모든 쿼리가 인덱스를 활용하여 full table scan 없음.

---

## 부록 C: 보류 사항 (YAGNI)

Plan에서 보류로 결정된 항목은 이 설계서에 포함하지 않음:

| 항목 | 보류 사유 |
|------|----------|
| HS256 -> ES256 JWT 전환 | 단일 서버에서 불필요한 복잡성 |
| localStorage -> httpOnly cookie | Token refresh 구현 후 재평가 |
| CSRF 토큰 | Bearer 인증 사용 중 |
| Redis 세션/레이트리밋 | 단일 ECS에서 in-memory 충분 |
| WAF rules | 트래픽 증가 후 도입 |
| Playwright E2E | Go 테스트 먼저, 프론트엔드 E2E는 후순위 |
