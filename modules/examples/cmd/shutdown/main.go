package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ubgo/shutdown"
)

func main() {
	router := gin.Default()
	router.GET("/slow", func(c *gin.Context) {
		time.Sleep(5 * time.Second)
		c.String(http.StatusOK, "Welcome Gin Server")
	})

	srv := &http.Server{
		Addr:    ":9001",
		Handler: router.Handler(),
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	mgr := shutdown.New()
	_ = mgr.Register("http-server", func(ctx context.Context) error {
		log.Println("Shutting down server...")
		return srv.Shutdown(ctx)
	}, shutdown.WithPhase(shutdown.PhaseStopAccepting))

	if err := mgr.Listen(context.Background()); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
