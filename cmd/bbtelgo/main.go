package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/frangi01/bbtelgo/internal/app"
	"github.com/frangi01/bbtelgo/internal/config"
	"github.com/frangi01/bbtelgo/internal/logx"
)

func main() {

	logger, err := logx.New("logs/bot.log", logx.Options{
		Level:      logx.Debug, // minimum level to print
		MaxSizeMB:  10,         // rotate after ~10MB (0 to disable)
		IncludeSrc: true,       // show file:line
		// TimeFormat: time.RFC3339, // or customize
	})
	if err != nil {
		panic(err)
	}
	defer logger.Close()
	
	config, err := config.Load(logger)
	if err != nil {
		logger.Errorf("config - env")
		return
	}

	context, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	app, err := app.New(logger, config)
	if err != nil {
		logger.Errorf("bot - new")
		return
	}

	app.Run(context)
}
