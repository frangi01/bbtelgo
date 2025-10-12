package app

import (
	"context"
	"net/http"
	"time"

	"github.com/frangi01/bbtelgo/internal/config"
	"github.com/frangi01/bbtelgo/internal/db"
	"github.com/frangi01/bbtelgo/internal/handlers"
	"github.com/frangi01/bbtelgo/internal/i18n"
	"github.com/frangi01/bbtelgo/internal/logx"
	"github.com/frangi01/bbtelgo/internal/utils"
	"go.mongodb.org/mongo-driver/mongo"

	tgbot "github.com/go-telegram/bot"
)

type App struct {
	logger 		*logx.Logger
	config 		config.Config
	bot	 		*tgbot.Bot
}

func New(logger *logx.Logger, cfg config.Config, dbclient *mongo.Client, repositoryList *db.RepositoryList, cache *db.CacheClient, i18nBundle *i18n.Bundle) (*App, error) {
	deps := utils.NewDeps(logger, cfg, repositoryList, cache, i18nBundle)
	h := handlers.Handler(deps)
	
	opts := []tgbot.Option{
		tgbot.WithDefaultHandler(h),
	}

	httpClient := &http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:    cfg.TransportMaxIdleConns,
			IdleConnTimeout: time.Duration(cfg.TransportIdleConnTimeout) * time.Second,
		},
	}
	opts = append(opts, tgbot.WithHTTPClient(time.Duration(cfg.Timeout) ,httpClient))

	if cfg.WebHookSecret != "" {
		opts = append(opts, tgbot.WithWebhookSecretToken(cfg.WebHookSecret))
	}

	botx, err := tgbot.New(cfg.Token, opts...)
	if err != nil {
		logger.Errorf("Error init bot %s", err)
	}

	return &App{logger: logger, config: cfg, bot: botx}, err
}

func (app *App) Run(context context.Context) {
	if app.config.ResetWebHook {
		deleteWebhookResult, err := app.bot.DeleteWebhook(
			context,
			&tgbot.DeleteWebhookParams{
				DropPendingUpdates: true,
			},
		)
		app.logger.Debugf("DeleteWebhook result: %v", deleteWebhookResult)
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

			app.logger.Debugf("SetWebHook result: %v", setWebHookResult)

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