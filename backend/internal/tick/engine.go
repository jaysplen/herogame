package tick

import (
	"context"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/herogame/backend/internal/redisx"
	"github.com/herogame/backend/internal/store"
	"github.com/herogame/backend/internal/ws"
)

// Engine runs the authoritative 1 Hz game loop.
type Engine struct {
	store    *store.Store
	redis    *redisx.Client
	hub      *ws.Hub
	log      *slog.Logger
	arrivals *Arrivals
	economy  *Economy
	upkeep   *Upkeep
	cancel   context.CancelFunc
}

// NewEngine constructs a tick engine.
func NewEngine(st *store.Store, rdb *redisx.Client, hub *ws.Hub, log *slog.Logger) *Engine {
	return &Engine{
		store:    st,
		redis:    rdb,
		hub:      hub,
		log:      log,
		arrivals: NewArrivals(st, rdb, hub, log),
		economy:  NewEconomy(st, log),
		upkeep:   NewUpkeep(st, log),
	}
}

// Arrivals exposes the arrivals scheduler for move handlers (BETA-003).
func (e *Engine) Arrivals() *Arrivals {
	return e.arrivals
}

// Start rehydrates Redis and runs the tick loop until ctx is cancelled.
func (e *Engine) Start(ctx context.Context) error {
	if err := e.arrivals.Rehydrate(ctx); err != nil {
		return err
	}
	runCtx, cancel := context.WithCancel(ctx)
	e.cancel = cancel
	go e.runLoop(runCtx)
	e.log.Info("tick engine started")
	return nil
}

// Stop cancels the tick loop.
func (e *Engine) Stop() {
	if e.cancel != nil {
		e.cancel()
	}
}

func (e *Engine) runLoop(ctx context.Context) {
	for {
		e.runLoopOnce(ctx)
		if ctx.Err() != nil {
			return
		}
	}
}

func (e *Engine) runLoopOnce(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			e.log.Error("tick panic recovered",
				slog.Any("panic", r),
				slog.String("stack", string(debug.Stack())),
			)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.tick(ctx)
		}
	}
}

// TickOnce runs one tick cycle (arrivals, economy, upkeep). Exported for tests.
func (e *Engine) TickOnce(ctx context.Context) {
	e.tick(ctx)
}

func (e *Engine) tick(ctx context.Context) {
	now := time.Now().UTC()
	if err := e.arrivals.Sweep(ctx, now); err != nil {
		e.log.Error("arrivals sweep", slog.String("error", err.Error()))
	}
	if err := e.economy.Sweep(ctx); err != nil {
		e.log.Error("economy sweep", slog.String("error", err.Error()))
	}
	if err := e.upkeep.Sweep(ctx); err != nil {
		e.log.Error("upkeep sweep", slog.String("error", err.Error()))
	}
}
