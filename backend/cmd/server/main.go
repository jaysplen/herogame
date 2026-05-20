package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/herogame/backend/internal/httpsrv"
	"github.com/herogame/backend/internal/redisx"
	"github.com/herogame/backend/internal/store"
	"github.com/herogame/backend/internal/tick"
	"github.com/herogame/backend/internal/ws"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dsn := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	if dsn != "" {
		runMigrations := os.Getenv("RUN_MIGRATIONS") != "0"
		if runMigrations {
			logger.Info("running database migrations")
			if err := store.MigrateUp(dsn); err != nil {
				logger.Error("migrate failed", slog.String("error", err.Error()))
				os.Exit(1)
			}
			logger.Info("migrations complete")
		}
	}

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	var wsHandler http.Handler
	var tickEngine *tick.Engine

	if dsn != "" {
		st, err := store.New(ctx, dsn)
		if err != nil {
			logger.Error("store init failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
		defer st.Close()

		rdb, err := redisx.New(ctx, redisURL)
		if err != nil {
			logger.Error("redis init failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
		defer rdb.Close()

		gw := ws.NewGateway(st, logger)
		wsHandler = ws.Handler(gw)
		logger.Info("websocket gateway enabled", slog.String("path", "/ws"))

		tickEngine = tick.NewEngine(st, rdb, gw.Hub(), logger)
		if err := tickEngine.Start(ctx); err != nil {
			logger.Error("tick engine failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
		defer tickEngine.Stop()
	} else {
		logger.Warn("DATABASE_URL unset; /ws and tick engine disabled")
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      httpsrv.NewRouter(logger, wsHandler),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server listening", slog.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", slog.String("error", err.Error()))
	}
}
