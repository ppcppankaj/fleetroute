package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	pkgauth "gpsgo/pkg/auth"
)

// VideoHandler manages video telematics (dashcam) endpoints.
// This is a stub for the M15 Video Telematics module.
type VideoHandler struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewVideoHandler(pool *pgxpool.Pool, logger *zap.Logger) *VideoHandler {
	return &VideoHandler{pool: pool, logger: logger}
}

// ListEvents returns video recording events for the tenant.
func (h *VideoHandler) ListEvents(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	vehicleID := c.Query("vehicle_id")

	h.logger.Info("List video events", zap.String("tenant_id", tenantID), zap.String("vehicle_id", vehicleID))

	// STUB REPSONSE
	// In a real implementation this would query `video_events` table
	events := []map[string]any{}
	
	respondOK(c, events)
}

// TriggerSnapshot requests a manual image snapshot from the device.
func (h *VideoHandler) TriggerSnapshot(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	deviceID := c.Param("deviceId")

	h.logger.Info("Triggering video snapshot", zap.String("tenant_id", tenantID), zap.String("device_id", deviceID))

	// STUB: send MQTT/TCP command to device to capture snapshot
	respondOK(c, gin.H{"status": "captured_requested"})
}

// GetLiveStreamCredentials returns WebRTC signaling details for a camera.
func (h *VideoHandler) GetLiveStreamCredentials(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	deviceID := c.Param("deviceId")

	h.logger.Info("Requesting livestream creds", zap.String("tenant", tenantID), zap.String("device", deviceID))

	// STUB
	respondError(c, http.StatusNotFound, "Device not connected to WebRTC relay")
}
