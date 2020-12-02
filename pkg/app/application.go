package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"url_reader/config"
	"url_reader/pkg/handlers"
	"url_reader/pkg/service"
)

func Run(cfg config.Config) {
	// create channel for listening shutdown signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	// create context for signal that shutdown is complete
	ctx, done := context.WithCancel(context.Background())

	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(cfg.Port))

	svc := service.NewUrlReader(config.Get())
	api := handlers.NewApi(
		handlers.WithConfig(config.Get()),
		handlers.WithReader(svc),
	)

	srv := &http.Server{
		Addr:    addr,
		Handler: handlers.NewRouter(api),
	}

	// run shutdown listener
	go func() {
		defer done()
		<-stop

		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
		defer shutdownCancel()
		fmt.Println("shutting down...")

		if err := srv.Shutdown(shutdownCtx); err != nil {
			fmt.Printf("can not gracefully shutdown. error: %v\n", err)
			return
		}
		fmt.Println("complete")
	}()

	// run http listener
	go func() {
		fmt.Printf("starting listener at '%s'", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("can not start listener: %v\n", err)
			os.Exit(-1)
		}
	}()

	// waiting shutdown is complete
	<-ctx.Done()
}
