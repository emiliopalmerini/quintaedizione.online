package web

import (
	"net/http"
	"os"
	"strconv"

	"github.com/emiliopalmerini/quintaedizione.online/internal/application/services"
	"github.com/gin-gonic/gin"
)

type AnalyticsHandlers struct {
	analyticsService *services.AnalyticsService
}

func NewAnalyticsHandlers(analyticsService *services.AnalyticsService) *AnalyticsHandlers {
	return &AnalyticsHandlers{
		analyticsService: analyticsService,
	}
}

func (h *AnalyticsHandlers) RegisterRoutes(router *gin.Engine) {
	analytics := router.Group("/api/v1/analytics")
	analytics.Use(h.adminAuthMiddleware())
	{
		analytics.GET("/stats", h.handleGetStats)
		analytics.GET("/heatmap", h.handleGetHeatmap)
		analytics.GET("/search-stats", h.handleGetSearchStats)
	}
}

func (h *AnalyticsHandlers) adminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-Admin-Key")
		expectedKey := os.Getenv("ADMIN_API_KEY")

		if expectedKey == "" {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Admin API not configured"})
			c.Abort()
			return
		}

		if apiKey != expectedKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (h *AnalyticsHandlers) handleGetStats(c *gin.Context) {
	period := c.DefaultQuery("period", "30d")

	stats, err := h.analyticsService.GetAggregateStats(c.Request.Context(), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"period": period,
		"stats":  stats,
	})
}

func (h *AnalyticsHandlers) handleGetHeatmap(c *gin.Context) {
	period := c.DefaultQuery("period", "30d")
	limitStr := c.DefaultQuery("limit", "50")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 200 {
		limit = 50
	}

	heatmap, err := h.analyticsService.GetContentHeatmap(c.Request.Context(), period, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"period": period,
		"items":  heatmap,
		"count":  len(heatmap),
	})
}

func (h *AnalyticsHandlers) handleGetSearchStats(c *gin.Context) {
	period := c.DefaultQuery("period", "30d")
	limitStr := c.DefaultQuery("limit", "100")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 500 {
		limit = 100
	}

	searchStats, err := h.analyticsService.GetSearchStats(c.Request.Context(), period, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"period":       period,
		"top_searches": searchStats,
		"count":        len(searchStats),
	})
}
