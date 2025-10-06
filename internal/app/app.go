package app

import (
	"context"
	"net/http"
	"time"

	"github.com/frangi01/bbtelgo/internal/config"
	"github.com/frangi01/bbtelgo/internal/handlers"
	"github.com/frangi01/bbtelgo/internal/logx"

	tgbot "github.com/go-telegram/bot"
)

type App struct {
	logger 		*logx.Logger
	config 		config.Config
	bot	 		*tgbot.Bot
}

func New(logger *logx.Logger, config config.Config) (*App, error) {
	opts := []tgbot.Option{
		tgbot.WithDefaultHandler(
			handlers.Handler(logger, config),
		),
	}

	httpClient := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:    config.TransportMaxIdleConns,
			IdleConnTimeout: time.Duration(config.TransportIdleConnTimeout) * time.Second,
		},
	}
	opts = append(opts, tgbot.WithHTTPClient(time.Duration(config.Timeout) ,httpClient))

	if config.WebHookSecret != "" {
		opts = append(opts, tgbot.WithWebhookSecretToken(config.WebHookSecret))
	}

	botx, err := tgbot.New(config.Token, opts...)
	if err != nil {
		logger.Errorf("Error init bot %s", err)
	}

	return &App{logger: logger, config: config, bot: botx}, err
}

func (app *App) Run(context context.Context) {
	if app.config.ResetWebHook {
		deleteWebhookResult, err := app.bot.DeleteWebhook(
			context,
			&tgbot.DeleteWebhookParams{
				DropPendingUpdates: true,
			},
		)
		app.logger.Debugf("DeleteWebhook result: %s", deleteWebhookResult)
		if err != nil {
			app.logger.Debugf("DeleteWebhook error: %v", err)
		}
	}
	switch app.config.Mode {
		case "polling":
			app.bot.Start(context)
		case "webhook":
			if app.config.WebHookPublicUrl == "" {
				app.logger.Errorf("APP_WEBHOOK_PUBLIC_URL can't be empty")
			}

			setWebHookResult, err := app.bot.SetWebhook(context, &tgbot.SetWebhookParams{
				URL: app.config.WebHookPublicUrl,
				SecretToken: app.config.WebHookSecret,
			})

			app.logger.Debugf("SetWebHook result: %s", setWebHookResult)

			if err != nil {
				app.logger.Errorf("SetWebHook error: %v", err)
			}

			go app.bot.StartWebhook(context)
			
			addr := ":" +app.config.WebHookPort
			srv := &http.Server{
				Addr:    addr,
				Handler: app.bot.WebhookHandler(),
			}

			err = srv.ListenAndServeTLS(
				app.config.WebHookTLSCertFile,
				app.config.WebHookTLSKeyFile,
			)

			app.logger.Errorf("Listen and serve TLS error: %v", err)
		default:
			app.logger.Errorf("APP_MODE non valido: %s", app.config.Mode)
	}

}