package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/mihari-bot/bot/internal/bot"
	"go.uber.org/zap"
)

func main() {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	loggerNotSugared, err := loggerConfig.Build()
	if err != nil {
		panic(err)
	}
	logger := loggerNotSugared.Sugar()

	baseDir, err := os.Getwd()
	if err != nil {
		logger.Fatalw("failed get cwd",
			err)
	}
	baseDir = filepath.Join(baseDir, "mihari")

	b := bot.New(logger, baseDir)

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := b.Init(rootCtx); err != nil {
		logger.Fatal("init bot failed", zap.Error(err))
	}

	sigCh := make(chan os.Signal, 1)

	var wg sync.WaitGroup
	wg.Go(func() {
		err := b.Start(rootCtx)
		if err != nil {
			logger.Errorw("run failed",
				err)
		}
		sigCh <- syscall.SIGTERM
		cancel()
	})

	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info("shutdown signal received")
	cancel()
	wg.Wait()

	logger.Info("service stopped")
}
