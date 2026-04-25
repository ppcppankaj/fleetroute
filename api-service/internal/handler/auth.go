// Package handler contains all API request handlers.
package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	pkgauth "gpsgo/pkg/auth"
	pkgdb "gpsgo/pkg/db"
)

// AuthHandler handles login, token refresh, and logout.
type AuthHandler struct {
	pool    *pgxpool.Pool
	rdb     *redis.Client
	authMgr *pkgauth.Manager
	logger  *zap.Logger
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(pool *pgxpool.Pool, rdb *redis.Client, authMgr *pkgauth.Manager, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{pool: pool, rdb: rdb, authMgr: authMgr, logger: logger}
}

// loginRequest represents the login request body.
type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// tokenResponse is returned on successful auth.
type tokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         userInfo  `json:"user"`
}

type userInfo struct {
	ID       string       `json:"id"`
	TenantID string       `json:"tenant_id"`
	Email    string       `json:"email"`
	Role     pkgauth.Role `json:"role"`
}

// Login godoc
// @Summary      Authenticate user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body loginRequest true "Login credentials"
// @Success      200 {object} tokenResponse
// @Failure      400 {object} errorResponse
// @Failure      401 {object} errorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	// Look up user by email
	var (
		userID       string
		tenantID     string
		passwordHash string
		roleStr      string
	)
	row := h.pool.QueryRow(context.Background(),
		`SELECT id, tenant_id, password_hash, role FROM users WHERE email=$1 AND deleted_at IS NULL`,
		req.Email,
	)
	if err := row.Scan(&userID, &tenantID, &passwordHash, &roleStr); err != nil {
		respondError(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		respondError(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	role := pkgauth.Role(roleStr)
	access, err := h.authMgr.GenerateAccess(userID, tenantID, role)
	if err != nil {
		h.logger.Error("generate access token", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "token generation failed")
		return
	}
	refresh, err := h.authMgr.GenerateRefresh(userID, tenantID, role)
	if err != nil {
		h.logger.Error("generate refresh token", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "token generation failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tokenResponse{
			AccessToken:  access,
			RefreshToken: refresh,
			ExpiresAt:    time.Now().Add(15 * time.Minute),
			User:         userInfo{ID: userID, TenantID: tenantID, Email: req.Email, Role: role},
		},
	})
}

// Refresh godoc
// @Summary      Refresh access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	claims, err := h.authMgr.Validate(body.RefreshToken)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	// Check Redis denylist — token may have been revoked via logout.
	if h.isRevoked(c.Request.Context(), claims.RegisteredClaims.ID) {
		respondError(c, http.StatusUnauthorized, "refresh token has been revoked")
		return
	}

	access, err := h.authMgr.GenerateAccess(claims.UserID, claims.TenantID, claims.Role)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "token generation failed")
		return
	}
	refresh, err := h.authMgr.GenerateRefresh(claims.UserID, claims.TenantID, claims.Role)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "token generation failed")
		return
	}

	// Rotate: revoke the old refresh token immediately.
	h.revokeToken(c.Request.Context(), claims.RegisteredClaims.ID, time.Until(claims.ExpiresAt.Time))

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"access_token":  access,
			"refresh_token": refresh,
			"expires_at":    time.Now().Add(15 * time.Minute),
		},
	})
}

// Logout godoc
// @Summary      Revoke refresh token and logout
// @Tags         auth
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	// Best-effort parse — if no token is provided, still return 200.
	if err := c.ShouldBindJSON(&body); err == nil && body.RefreshToken != "" {
		claims, err := h.authMgr.Validate(body.RefreshToken)
		if err == nil && claims.RegisteredClaims.ID != "" {
			ttl := time.Until(claims.ExpiresAt.Time)
			if ttl > 0 {
				h.revokeToken(c.Request.Context(), claims.RegisteredClaims.ID, ttl)
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "logged_out"}})
}

// revokeToken adds the token jti to the Redis denylist with a TTL matching
// the token's remaining validity window.
func (h *AuthHandler) revokeToken(ctx context.Context, jti string, ttl time.Duration) {
	if jti == "" || ttl <= 0 {
		return
	}
	if err := h.rdb.Set(ctx, pkgdb.KeyRevoked(jti), "1", ttl).Err(); err != nil {
		h.logger.Warn("failed to add token to revocation denylist",
			zap.String("jti", jti),
			zap.Error(err),
		)
	}
}

// isRevoked returns true if the jti is present in the Redis denylist.
// Fails open on Redis error to avoid locking out users during Redis downtime.
func (h *AuthHandler) isRevoked(ctx context.Context, jti string) bool {
	if jti == "" {
		return false
	}
	exists, err := h.rdb.Exists(ctx, pkgdb.KeyRevoked(jti)).Result()
	if err != nil {
		h.logger.Warn("revocation denylist check failed — failing open",
			zap.String("jti", jti),
			zap.Error(err),
		)
		return false
	}
	return exists > 0
}

// ── Shared response helpers ────────────────────────────────────────────────────

type errorResponse struct {
	Error string `json:"error"`
}

func respondError(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
}

func respondOK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func respondCreated(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, gin.H{"data": data})
}

func respondPaginated(c *gin.Context, data any, meta any) {
	c.JSON(http.StatusOK, gin.H{"data": data, "meta": meta})
}
