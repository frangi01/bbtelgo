package handlers

import (
	"context"

	"github.com/frangi01/bbtelgo/internal/config"
	"github.com/frangi01/bbtelgo/internal/db"
	"github.com/frangi01/bbtelgo/internal/entities"
	"github.com/frangi01/bbtelgo/internal/logx"
	"github.com/frangi01/bbtelgo/internal/utils"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func Handler(logger *logx.Logger, cfg config.Config, repositoryList *db.RepositoryList) bot.HandlerFunc {
	return func(context context.Context, bot *bot.Bot, update *models.Update) {
        json, err := utils.JSON(update, false, false)
		if err != nil {
			logger.Errorf("marshal update: %v", err)
			return
		}
        logger.Debugf("update: %s", string(json))

		u := &entities.UserEntity{
			User: *update.Message.From,
		}
		_, id, err := repositoryList.UserRepository.UpsertByTelegramID(context, u)
		if err != nil {
			logger.Errorf("upsert user: %v", err)
		}
		logger.Debugf("upsert user id: %v", id.Hex())

		m := &entities.MessageEntity{
			Message: *update.Message,
		}
		id, err = repositoryList.MessageRepository.Create(context, m)
		if err != nil {
			logger.Errorf("create message: %v", err)
		}
		logger.Debugf("created message id: %v", id.Hex())
    }
}