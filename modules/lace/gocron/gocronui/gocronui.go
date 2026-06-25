// Package gocronui provides a thin wrapper around the gocron-ui dashboard
// library (github.com/go-co-op/gocron-ui) with internal-key authentication.
//
// # Overview
//
// gocron-ui exposes a web dashboard for monitoring and managing gocron jobs
// (real-time status, manual triggering, WebSocket live updates). This package
// adds an authentication layer so the dashboard is not publicly accessible.
//
// # Authentication flow
//
// Access is granted via the project's internal key, which can be provided in
// two ways:
//
//   - Query parameter: GET /?key=<internal-key>
//   - Request header:  X-Internal-Key: <internal-key>
//
// # Why cookies are used
//
// The gocron-ui dashboard is a single-page application that loads static
// assets (JS, CSS), makes REST API calls (/api/jobs, /api/config), and opens
// a WebSocket connection (/ws) — all as separate HTTP requests. These
// secondary requests are initiated by the browser automatically and cannot
// include a custom query parameter or header.
//
// To solve this, on the first successful key-authenticated request the
// middleware sets an HttpOnly session cookie. All subsequent requests from the
// same browser session (static assets, XHR/fetch API calls, WebSocket upgrade)
// automatically include this cookie and pass through without requiring the key
// again.
//
// The cookie value is an HMAC-SHA256 signature derived from the internal key,
// so it cannot be guessed or forged without knowing the key.
//
// # Usage
//
//	cronUICfg := appConfig.App.CronUI
//	if cronUICfg.Enable {
//	    httpSrv := gocronui.NewServer(scheduler, gocronui.Config{
//	        Port:        cronUICfg.Port,
//	        Title:       appConfig.Core.AppName + " Cron",
//	        InternalKey: appConfig.Core.InternalKey,
//	    })
//	    gocronui.ListenAndLog(httpSrv, cronUICfg.Port, func(err error) {
//	        logger.Error("gocron-ui server failed", zap.Error(err))
//	    })
//	    // call httpSrv.Shutdown(ctx) during graceful shutdown
//	}
//
// # Configuration (pkl)
//
// The enable/port settings live in the pkl config under app.cronUI:
//
//	app {
//	  cronUI {
//	    enable = true
//	    port = "8089"
//	  }
//	}
//
// Environment variables for production: CRON_UI_ENABLE, CRON_UI_PORT.
package gocronui

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/go-co-op/gocron-ui/server"
	"github.com/go-co-op/gocron/v2"
)

// Config holds the configuration for the gocron-ui dashboard server.
type Config struct {
	// Port is the TCP port the dashboard listens on (e.g. "8089").
	Port string

	// Title is displayed in the dashboard header (e.g. "thewildai Cron").
	Title string

	// InternalKey is the shared secret used for authentication.
	// Callers provide it via ?key= query param or X-Internal-Key header.
	InternalKey string
}

// NewServer creates an *http.Server serving the gocron-ui dashboard,
// protected by [InternalKeyAuth]. The returned server is ready to be
// started with ListenAndLog or manually with ListenAndServe.
func NewServer(scheduler gocron.Scheduler, cfg Config) *http.Server {
	srv := server.NewServer(scheduler, 0, server.WithTitle(cfg.Title))
	handler := InternalKeyAuth(cfg.InternalKey, srv.Router)
	return &http.Server{Addr: ":" + cfg.Port, Handler: handler}
}

// InternalKeyAuth wraps an http.Handler with internal-key authentication.
//
// Authentication is checked in the following order:
//  1. Session cookie — allows static assets, API, and WebSocket requests
//     to pass through after the initial authentication.
//  2. "key" query parameter — e.g. /?key=<secret>.
//  3. "X-Internal-Key" header — for programmatic access.
//
// On successful key authentication (methods 2 or 3), an HttpOnly session
// cookie is set so the browser can load all subsequent resources without
// needing the key on every request. See the package documentation for the
// full rationale on why cookies are necessary.
func InternalKeyAuth(key string, next http.Handler) http.Handler {
	cookieName := "cron_session"
	cookieValue := hmacSign(key, "cron-ui-session")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check cookie first (for static assets, API, WebSocket)
		if c, err := r.Cookie(cookieName); err == nil && c.Value == cookieValue {
			next.ServeHTTP(w, r)
			return
		}

		// Check key via query param or header
		if r.URL.Query().Get("key") == key || r.Header.Get("X-Internal-Key") == key {
			http.SetCookie(w, &http.Cookie{
				Name:     cookieName,
				Value:    cookieValue,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
			next.ServeHTTP(w, r)
			return
		}

		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

// ListenAndLog starts the HTTP server in a background goroutine and prints
// the dashboard URL to stdout. Errors (other than ErrServerClosed) are
// forwarded to the logError callback.
func ListenAndLog(httpSrv *http.Server, port string, logError func(error)) {
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logError(err)
		}
	}()
	fmt.Printf("Cron dashboard at http://localhost:%s\n", port)
}

// hmacSign produces a hex-encoded HMAC-SHA256 signature. Used to derive
// a cookie value from the internal key so the cookie cannot be forged.
func hmacSign(key, msg string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(msg))
	return hex.EncodeToString(mac.Sum(nil))
}
