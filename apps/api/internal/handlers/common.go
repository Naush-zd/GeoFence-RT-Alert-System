package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/geofencing-platform/apps/api/internal/realtime"
)

type Handler struct {
	DB  *pgxpool.Pool
	Hub *realtime.Hub
}

func New(db *pgxpool.Pool, hub *realtime.Hub) Handler {
	return Handler{DB: db, Hub: hub}
}
func (h Handler) Health(c *gin.Context) {
	started := time.Now()
	if err := h.DB.Ping(c.Request.Context()); err != nil {
		writeJSON(c, http.StatusServiceUnavailable, map[string]any{
			"status": "database_unavailable",
			"error":  "database is unavailable",
		}, started)
		return
	}
	writeJSON(c, http.StatusOK, map[string]any{"status": "ok"}, started)
}
func decodeJSON(c *gin.Context, destination any) error {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}
func writeError(c *gin.Context, status int, message string, started time.Time) {
	writeJSON(c, status, map[string]any{"error": message}, started)
}
func writeJSON(c *gin.Context, status int, payload map[string]any, started time.Time) {
	payload["time_ns"] = fmt.Sprintf("%d", time.Since(started).Nanoseconds())
	c.JSON(status, payload)
}
