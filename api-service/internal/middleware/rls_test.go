package middleware

import (
	"testing"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestRLSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	// basic instantiate test
	mw := RLS(nil, logger)
	if mw == nil {
		t.Fatal("expected middleware, got nil")
	}
}
