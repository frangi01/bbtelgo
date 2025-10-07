package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/frangi01/bbtelgo/internal/config"
	"github.com/frangi01/bbtelgo/internal/db"
	"github.com/frangi01/bbtelgo/internal/entities"
	"github.com/frangi01/bbtelgo/internal/logx"
	"github.com/frangi01/bbtelgo/internal/utils"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func Handler(logger *logx.Logger, cfg config.Config, repositoryList *db.RepositoryList, cache *db.CacheClient) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
        if cache != nil && cfg.RedisCfg.RateLimitMessages > 0 {
			key := fmt.Sprintf("rl:user:%d:msg", update.Message.From.ID)

			limit := cfg.RedisCfg.RateLimitMessages
			window := time.Duration(cfg.RedisCfg.RateLimitMs) * time.Millisecond

			if window <= 0 {
				window = time.Minute // default 60s
			}

			allowed := true
			var err error

			switch cfg.RedisCfg.RateLimitType {
				case config.RLFixedWindow:
					allowed, _, _, err = cache.RateLimitFixedWindow(ctx, key, limit, window)
				case config.RLSlidingWindow:
					allowed, _, _, err = cache.RateLimitSlidingWindow(ctx, key, limit, window)
				default:
					allowed = true
			}

			if err != nil {
				logger.Errorf("rate-limit error: %v", err)
				return
			}
			if !allowed {
				_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "You are banned.",
				})
				return
			}
		}
		
		
		
		json, err := utils.JSON(update, false, false)
		if err != nil {
			logger.Errorf("marshal update: %v", err)
			return
		}
        logger.Debugf("update: %s", string(json))

		u := &entities.UserEntity{
			User: *update.Message.From,
		}
		_, id, err := repositoryList.UserRepository.UpsertByTelegramID(ctx, u)
		if err != nil {
			logger.Errorf("upsert user: %v", err)
		}
		logger.Debugf("upsert user id: %v", id.Hex())

		m := &entities.MessageEntity{
			Message: *update.Message,
		}
		id, err = repositoryList.MessageRepository.Create(ctx, m)
		if err != nil {
			logger.Errorf("create message: %v", err)
		}
		logger.Debugf("created message id: %v", id.Hex())

		

    }
}