package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v11"
	"github.com/elipatov/web-crawler/pkg/logger"
)

func main() {
	var mode string

	if len(os.Args) > 1 {
		mode = os.Args[1]
	}

	ctx := appContext()
	logger := logger.New("DEBUG")

	cfg, err := newConfig()
	if err != nil {
		logger.WithError(err).Error("failed to apply configuration")
	}

	logger.Infof("Mode: %s", mode)

	app, err := New(ctx, logger, cfg)
	if err != nil {
		logger.WithError(err).Error("failed to create app")
	}

	err = app.Run(ctx, mode, os.Args[1:]...)
	if err != nil {
		logger.WithError(err).Error("failed to run app")
	}
}

func newConfig() (config, error) {
	var conf config

	err := env.Parse(conf)
	if err != nil {
		return conf, fmt.Errorf("failed to read configuration: %w", err)
	}

	return conf, nil
}

func appContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		c := make(chan os.Signal, 1)

		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		cancel()
	}()

	return ctx
}
