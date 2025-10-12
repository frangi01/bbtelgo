package private

import (
	"context"

	"github.com/frangi01/bbtelgo/internal/entities"
	"github.com/frangi01/bbtelgo/internal/utils"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func startHandler(ctx context.Context, b *bot.Bot, u *models.Update, _ []string, deps *utils.HandlerDeps, lang string) {
	if u.Message == nil {
		return
	}
	
	user := &entities.UserEntity{
		User: *u.Message.From,
	}
	_, id, err := deps.RepositoryList.UserRepository.UpsertByTelegramID(ctx, user)
	if err != nil {
		deps.Logger.Errorf("upsert user: %v", err)
	}
	deps.Logger.Debugf("upsert user id: %v", id.Hex())
	

	b.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: u.Message.Chat.ID,
		Action: models.ChatActionTyping,
	})

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: deps.I18n.T(lang, "button.1", nil), CallbackData: "button_1"},
				{Text: deps.I18n.T(lang, "button.2", nil), CallbackData: "button_2:arg_1"},
			},
			{
				{Text: deps.I18n.T(lang, "button.3", nil), CallbackData: "button_3"},
			},
		},
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      	u.Message.Chat.ID,
		Text:       	deps.I18n.T(lang, "button.1", nil),
		ReplyMarkup: 	kb,
	})
}

func photoHandler(ctx context.Context, b *bot.Bot, u *models.Update, _ []string, deps *utils.HandlerDeps, lang string) {
	lang = "de"
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      u.Message.Chat.ID,
		Text:        deps.I18n.T(
			lang, "photo.received", map[string]any{
			"caption": u.Message.Caption,
		}),
	})
}