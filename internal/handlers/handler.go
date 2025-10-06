package handlers

import (
	"context"

	"github.com/frangi01/bbtelgo/internal/config"
	"github.com/frangi01/bbtelgo/internal/logx"
	"github.com/frangi01/bbtelgo/internal/utils"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func Handler(logger *logx.Logger, cfg config.Config) bot.HandlerFunc {
	return func(context context.Context, bot *bot.Bot, update *models.Update) {
        json, err := utils.JSON(update, false, false)
		if err != nil {
			logger.Errorf("marshal update: %v", err)
			return
		}
        logger.Debugf("update: %s", string(json))
    }
}