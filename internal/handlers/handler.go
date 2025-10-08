package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/frangi01/bbtelgo/internal/config"
	"github.com/frangi01/bbtelgo/internal/handlers/private"
	"github.com/frangi01/bbtelgo/internal/utils"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)



func Handler(handlerDeps *utils.HandlerDeps) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
        // rate limit
		if handlerDeps.Cache != nil && handlerDeps.Cfg.RedisCfg.RateLimitMessages > 0 {
			key := fmt.Sprintf("rl:user:%d:msg", utils.ChatIDFromUpdate(update))

			limit := handlerDeps.Cfg.RedisCfg.RateLimitMessages
			window := time.Duration(handlerDeps.Cfg.RedisCfg.RateLimitMs) * time.Millisecond

			if window <= 0 {
				window = time.Minute
			}

			allowed := true
			var err error

			switch handlerDeps.Cfg.RedisCfg.RateLimitType {
				case config.RLFixedWindow:
					allowed, _, _, err = handlerDeps.Cache.RateLimitFixedWindow(ctx, key, limit, window)
				case config.RLSlidingWindow:
					allowed, _, _, err = handlerDeps.Cache.RateLimitSlidingWindow(ctx, key, limit, window)
				default:
					allowed = true
			}

			if err != nil {
				handlerDeps.Logger.Errorf("rate-limit error: %v", err)
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
			handlerDeps.Logger.Errorf("marshal update: %v", err)
			return
		}
        handlerDeps.Logger.Debugf("update: %s", string(json))

		if update.Message != nil && update.Message.Chat.Type == "private" {
            private.HandlerMessage(ctx, b, update, handlerDeps)
        } else if update.CallbackQuery != nil && update.CallbackQuery.Message.Message.Chat.Type == "private" {
			private.HandlerCallBackQuery(ctx, b, update, handlerDeps)
		}

		// m := &entities.MessageEntity{
		// 	Message: *update.Message,
		// }
		// id, err = handlerDeps.repositoryList.MessageRepository.Create(ctx, m)
		// if err != nil {
		// 	handlerDeps.logger.Errorf("create message: %v", err)
		// }
		// handlerDeps.logger.Debugf("created message id: %v", id.Hex())
    }
}