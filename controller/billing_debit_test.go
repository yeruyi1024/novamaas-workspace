package controller

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestReviewedDebitRequiresAdminSessionAndServerEvidence(t *testing.T) {
	for _, endpoint := range []string{"preview", "collect"} {
		for _, tc := range []struct {
			role    int
			session string
			want    int
		}{{1, "session", 403}, {10, "", 403}, {10, "session", 400}, {100, "session", 400}} {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				c.Set("id", 7)
				c.Set("role", tc.role)
				c.Set("session_id", tc.session)
				c.Set("session_version", int64(1))
				c.Set("auth_version", int64(1))
			})
			r.POST("/preview", PreviewHistoricalDebits)
			r.POST("/collect", CollectHistoricalDebit)
			res := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/"+endpoint, strings.NewReader(`{"month":"2025-10","account":7}`))
			r.ServeHTTP(res, req)
			assert.Equal(t, tc.want, res.Code)
		}
	}
}
