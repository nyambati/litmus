package server

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nyambati/litmus/internal/config"
)

var (
	staticFS fs.FS

	alertConfigPath string
)

type contextKey string

const LitmusConfigKey contextKey = "litmusConfig"

// SetStaticFS registers the embedded UI filesystem before the server starts.
func SetStaticFS(f fs.FS) {
	staticFS = f
}

// RunUIServer starts the Litmus UI backend.
func RunUIServer(port int, dev bool) error {
	litmusConfig, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("loading litmus config: %w", err)
	}
	alertConfigPath = litmusConfig.Config.File

	router := gin.Default()

	// CORS Middleware for development
	router.Use(corsMiddleware())
	router.Use(litmusConfigMiddleware(litmusConfig))

	// API Endpoints
	api := router.Group("/api/v1")
	{
		api.GET("/config", configHandler)
		api.GET("/tests", testsHandler)
		api.POST("/tests/run", runTestsHandler)
		api.POST("/evaluate", evaluateHandler)
		api.GET("/suggest", suggestHandler)
		api.GET("/regressions", regressionsHandler)
		api.POST("/regressions/run", regressionsRunHandler)
		api.GET("/diff", diffHandler)
		api.POST("/snapshot", snapshotHandler)
		api.GET("/health", healthHandler)
	}


	// Serve embedded UI in production mode
	if !dev && staticFS != nil {
		router.StaticFS("/assets", http.FS(staticFS))
		router.NoRoute(func(c *gin.Context) {
			if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.FileFromFS("index.html", http.FS(staticFS))
			}
		})
	}

	addr := fmt.Sprintf(":%d", port)
	url := fmt.Sprintf("http://localhost%s", addr)
	log.Printf("Litmus UI running at %s", url)

	if !dev {
		go func() {
			time.Sleep(150 * time.Millisecond)
			openBrowser(url)
		}()
	}

	return router.Run(addr)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("warning: failed to open browser: %v", err)
	}
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
