package cli

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/tesserabox/bentotask/internal/api"
)

func init() {
	serveCmd.Flags().Int("port", 7878, "Port to listen on")
	serveCmd.Flags().String("host", "127.0.0.1", "Host to bind to")
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the REST API server",
	Long: `Start the BentoTask REST API server.

The server exposes all task, habit, routine, and scheduling operations
via a JSON REST API at /api/v1/*.

Examples:
  bt serve                    Start on localhost:7878 (default)
  bt serve --port 9090        Custom port
  bt serve --host 0.0.0.0     Expose to network`,
	RunE: runServe,
}

func runServe(cmd *cobra.Command, _ []string) error {
	a, err := openApp(cmd)
	if err != nil {
		return err
	}

	port, _ := cmd.Flags().GetInt("port")
	host, _ := cmd.Flags().GetString("host")
	addr := fmt.Sprintf("%s:%d", host, port)

	srv := api.NewServer(a)

	// Graceful shutdown on SIGINT/SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		log.Printf("BentoTask API listening on http://%s", addr)
		if listenErr := srv.ListenAndServe(addr); listenErr != nil && listenErr != http.ErrServerClosed {
			errCh <- listenErr
		}
		close(errCh)
	}()

	select {
	case sig := <-sigCh:
		log.Printf("Received %v, shutting down...", sig)
	case err := <-errCh:
		if err != nil {
			_ = a.Close()
			return fmt.Errorf("server error: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}

	return a.Close()
}
