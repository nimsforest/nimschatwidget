package cli

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
	nimschatwidget "github.com/nimsforest/nimschatwidget"
	"github.com/nimsforest/nimsforest2/pkg/nim"
	"github.com/nimsforest/nimschatwidget/internal/config"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	var (
		addr       string
		natsURL    string
		webhookURL string
	)

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the chat widget server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configPath)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			// Flags override config
			if addr == "" {
				addr = cfg.Server.Addr
			}
			if addr == "" {
				addr = ":8096"
			}
			if natsURL == "" {
				natsURL = cfg.NATS.URL
			}
			if natsURL == "" {
				return fmt.Errorf("NATS URL required (--nats or config nats.url)")
			}
			if webhookURL == "" {
				webhookURL = cfg.Server.WebhookURL
			}
			if webhookURL == "" {
				return fmt.Errorf("webhook URL required (--webhook or config server.webhook_url)")
			}

			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			// Connect to NATS
			nc, err := nats.Connect(natsURL)
			if err != nil {
				return fmt.Errorf("connecting to NATS: %w", err)
			}
			defer nc.Close()

			wind := nim.NewWind(nc)
			log.Printf("[Wind] connected to %s", natsURL)

			// Create chat widget components
			source := nimschatwidget.NewSource(webhookURL, "nimschatwidget")
			songbird := nimschatwidget.NewSongbird(wind)
			if err := songbird.Start(); err != nil {
				return fmt.Errorf("starting songbird: %w", err)
			}

			// Mount handler
			mux := http.NewServeMux()
			mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"status":"ok"}`))
			})
			mux.Handle("/", nimschatwidget.Handler(source, songbird))

			go func() {
				<-ctx.Done()
				log.Println("shutting down...")
				os.Exit(0)
			}()

			log.Printf("nimschatwidget %s listening on %s (webhook: %s)", Version, addr, webhookURL)
			return http.ListenAndServe(addr, mux)
		},
	}

	cmd.Flags().StringVar(&addr, "addr", "", "listen address (default :8096)")
	cmd.Flags().StringVar(&natsURL, "nats", "", "NATS URL")
	cmd.Flags().StringVar(&webhookURL, "webhook", "", "forest webhook URL")

	return cmd
}
