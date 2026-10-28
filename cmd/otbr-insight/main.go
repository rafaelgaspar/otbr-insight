package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/otbr-insight/otbr-insight/internal/api"
	"github.com/otbr-insight/otbr-insight/internal/backup"
	"github.com/otbr-insight/otbr-insight/internal/config"
	"github.com/otbr-insight/otbr-insight/internal/mdns"
	"github.com/otbr-insight/otbr-insight/internal/names"
	"github.com/otbr-insight/otbr-insight/internal/otbr"
	"github.com/otbr-insight/otbr-insight/internal/otctl"
	"github.com/otbr-insight/otbr-insight/internal/service"
	"github.com/otbr-insight/otbr-insight/web"
)

func main() {
	cfg, err := config.Parse(os.Args[1:])
	if err != nil {
		// -h/--help is a request, not a failure: print usage on stdout and succeed.
		if errors.Is(err, flag.ErrHelp) {
			config.Usage(os.Stdout)
			return
		}
		fmt.Fprintf(os.Stderr, "otbr-insight: %v\n\n", err)
		config.Usage(os.Stderr)
		os.Exit(2)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel(cfg.LogLevel)}))
	slog.SetDefault(logger)

	client, err := otbr.NewClient(cfg.OTBRURL, &http.Client{Timeout: 4 * time.Second})
	if err == nil {
		client.SetLogger(logger)
	}
	if err != nil {
		logger.Error("create OTBR client", "error", err)
		os.Exit(2)
	}
	// Active scanning has no REST equivalent, so it comes from the OpenThread daemon
	// socket when otbr-insight runs on the border router. Absent socket, the otbr-web
	// endpoint is still tried, and absent that the feature reports itself unavailable.
	if strings.TrimSpace(cfg.OTBRSocket) != "" {
		// Meshdiag can run for minutes; everything else should not queue behind it
		// on the same process-wide mutex.
		meshSocket := otctl.New(cfg.OTBRSocket, cfg.OTBRSocketTimeout)
		interactive := otctl.New(cfg.OTBRSocket, 30*time.Second)
		client.SetNetworkScanner(interactive)
		client.SetEnergyScanner(interactive)
		client.SetMeshReader(meshSocket)
		client.SetStatusReader(interactive)
		client.SetPinger(interactive)
		client.SetHistoryReader(interactive)
		if meshSocket.Available() {
			logger.Info("OpenThread daemon socket present", "path", meshSocket.Path())
		} else {
			logger.Info("OpenThread daemon socket not present; nearby-network scanning will fall back to otbr-web", "path", meshSocket.Path())
		}
	}

	// Matter fabrics are discovered by browsing the LAN's mDNS. Any host can do
	// this; it is not tied to the border router.
	monitor := service.NewMonitor(client, cfg.PollInterval, logger,
		service.WithDiscoveryInterval(cfg.DiscoveryInterval),
		service.WithServiceBrowser(mdns.Browser{Wait: 3 * time.Second, Iface: cfg.MDNSInterface}))

	nameStore := names.New(cfg.DataDir)
	if err := nameStore.Load(); err != nil {
		logger.Warn("load device names", "error", err)
	}
	if nameStore.Enabled() {
		logger.Info("device naming enabled", "data_dir", cfg.DataDir)
	} else {
		logger.Info("device naming disabled: no data directory")
	}

	backupStore := backup.New(cfg.DataDir)
	if err := backupStore.Load(); err != nil {
		logger.Warn("load dataset backup", "error", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go monitor.Run(ctx)

	server := &http.Server{
		Addr: cfg.Listen, Handler: api.Handler(monitor, nameStore, client, backupStore, web.Handler(), logger),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second,
		WriteTimeout: 25 * time.Second, IdleTimeout: 60 * time.Second,
	}
	go func() {
		logger.Info("OTBR Insight started", "listen", cfg.Listen, "otbr_url", cfg.OTBRURL, "poll_interval", cfg.PollInterval.String(), "discovery_interval", cfg.DiscoveryInterval.String())
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown failed", "error", err)
	}
	logger.Info("OTBR Insight stopped")
}

func logLevel(value string) slog.Level {
	switch strings.ToLower(value) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
