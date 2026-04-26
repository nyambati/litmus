package server

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nyambati/litmus/internal/config"
	"github.com/sirupsen/logrus"
)

var (
	staticFS fs.FS
)

type contextKey string

const LitmusConfigKey contextKey = "litmusConfig"

// SetStaticFS registers the embedded UI filesystem before the server starts.
func SetStaticFS(f fs.FS) {
	staticFS = f
}

// RunUIServer starts the Litmus UI backend.
func RunUIServer(port int, dev bool) error {
	if !dev {
		// use gin in production mode to serve static files
		gin.SetMode(gin.ReleaseMode)
	}

	litmusConfig, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("loading litmus config: %w", err)
	}

	router := gin.Default()

	// CORS Middleware for development
	router.Use(corsMiddleware())
	router.Use(litmusConfigMiddleware(litmusConfig))

	// API Endpoints
	api := router.Group("/api/v1")
	{
		api.GET("/config", configHandler)
		api.GET("/fragments", fragmentsHandler)
		api.GET("/tests", testsHandler)
		api.GET("/tests/grouped", groupedTestsHandler)
		api.POST("/tests/run", runTestsHandler)
		api.POST("/evaluate", evaluateHandler)
		api.GET("/label_values", suggestHandler)
		api.POST("/regressions/generate", generateRegressionsHandler)
		api.GET("/diff", diffHandler)
		api.GET("/health", healthHandler)
	}

	// Serve embedded UI in production mode

	router.Use(serveStatic)

	addr := fmt.Sprintf(":%d", port)
	url := fmt.Sprintf("http://localhost%s", addr)
	logrus.Printf("Litmus UI running at %s", url)

	return router.Run(addr)
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

func litmusConfigMiddleware(litmusConfig *config.LitmusConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(string(LitmusConfigKey), litmusConfig)
		c.Next()
	}
}
