package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cortex/internal/app"
	"cortex/internal/config"
	"cortex/internal/utils"
)

func main() {
	logger := utils.GetLogger("cortex-backend")

	cfg, err := config.Load()
	if err != nil {
		logger.Error("APP", "/startup", 0, 0, "failed to load config: "+err.Error())
		log.Fatalf("failed to load config: %v", err)
	}

	application, err := app.NewApp(cfg, logger)
	if err != nil {
		logger.Error("APP", "/startup", 0, 0, "failed to initialize app: "+err.Error())
		log.Fatalf("failed to initialize app: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)

	go func() {
		errCh <- application.Run(ctx)
	}()

	sigCh := make(chan os.Signal, 1)

	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case sig := <-sigCh:
		log.Printf("received signal: %s", sig)

		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := application.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown error: %v", err)
		}

		select {
		case err := <-errCh:
			if err != nil {
				log.Printf("application stopped with error: %v", err)
				os.Exit(1)
			}
			log.Printf("application stopped gracefully")
		case <-shutdownCtx.Done():
			log.Printf("timed out waiting for application to stop")
			os.Exit(1)
		}

	case err := <-errCh:
		if err != nil {
			log.Printf("application stopped with error: %v", err)
			os.Exit(1)
		}
		log.Printf("application stopped gracefully")
	}
}
