package main

import (
	"context"
	"fmt"
	"log/slog"
	shttp "net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hoenirvili/axiogate/http"
	"github.com/hoenirvili/axiogate/http/handler"
	"github.com/hoenirvili/axiogate/http/request"
	"github.com/hoenirvili/axiogate/log"
	"github.com/hoenirvili/axiogate/provider/a"
	"github.com/hoenirvili/axiogate/provider/b"
	"github.com/hoenirvili/axiogate/shipment"
	"github.com/hoenirvili/axiogate/storage"
)

func db(ctx context.Context) (*pgxpool.Pool, error) {
	const url = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	dbpool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the db, %w", err)
	}
	if err := dbpool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("db ping failed, %w", err)
	}
	return dbpool, nil
}

var providers = map[string]shipment.Payloader{
	"a": a.Provider,
	"b": b.Provider,
}

func run() int {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGKILL,
	)

	logger := slog.New(slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	))

	db, err := db(ctx)
	if err != nil {
		logger.With(log.Error(err)).
			Error("Failed to connect to db")
		return 1
	}
	defer db.Close()

	svr := http.NewServer(
		http.WithLogger(logger),
		http.WithWhenToClose(ctx, stop),
	)

	st := storage.New(db, storage.WithLogger(logger))
	cli := request.NewClient(new(shttp.Client))
	service := shipment.New(cli, providers, st, shipment.WithLogger(logger))
	shipmentHandler := handler.NewShipment(service, handler.WithLogger(logger))
	svr.Routes(shipmentHandler)

	if err := http.Start(svr); err != nil {
		logger.With(log.Error(err)).Error("failed to start http server")
		return 1
	}

	return 0
}

func main() {
	os.Exit(run())
}
