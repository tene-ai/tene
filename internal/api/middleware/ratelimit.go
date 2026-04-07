package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/tomo-kay/tene/internal/api/response"
	"github.com/tomo-kay/tene/internal/auth"
	"golang.org/x/time/rate"
)

type userLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter implements per-user token bucket rate limiting.
type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*userLimiter
	freeRPM  int
	proRPM   int
}

// NewRateLimiter creates a rate limiter with plan-based limits.
func NewRateLimiter(freeRPM, proRPM int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*userLimiter),
		freeRPM:  freeRPM,
		proRPM:   proRPM,
	}
	go rl.cleanup()
	return rl
}

// Middleware returns an Echo middleware that applies rate limiting.
func (rl *RateLimiter) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key, rpm := rl.resolveKeyAndRPM(c)
			if !rl.allow(key, rpm) {
				c.Response().Header().Set("Retry-After", "60")
				return response.ErrMsg(c, http.StatusTooManyRequests, "RATE_LIMITED", "too many requests")
			}
			return next(c)
		}
	}
}

func (rl *RateLimiter) resolveKeyAndRPM(c echo.Context) (string, int) {
	claims, ok := c.Get(ContextKeyClaims).(*auth.Claims)
	if !ok {
		return "ip:" + c.RealIP(), rl.freeRPM
	}
	rpm := rl.freeRPM
	if claims.Plan == "pro" {
		rpm = rl.proRPM
	}
	return "user:" + claims.UserID, rpm
}

func (rl *RateLimiter) allow(key string, rpm int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	ul, ok := rl.limiters[key]
	if !ok {
		r := rate.Every(time.Minute / time.Duration(rpm))
		ul = &userLimiter{limiter: rate.NewLimiter(r, rpm/10+1)} // burst = 10% of RPM
		rl.limiters[key] = ul
	}
	ul.lastSeen = time.Now()
	return ul.limiter.Allow()
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(5 * time.Minute)
		rl.mu.Lock()
		for k, ul := range rl.limiters {
			if time.Since(ul.lastSeen) > 10*time.Minute {
				delete(rl.limiters, k)
			}
		}
		rl.mu.Unlock()
	}
}
