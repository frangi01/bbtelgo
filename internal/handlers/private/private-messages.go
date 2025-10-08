package private

import (
	"context"

	"github.com/frangi01/bbtelgo/internal/entities"
	"github.com/frangi01/bbtelgo/internal/utils"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func startHandler(ctx context.Context, b *bot.Bot, u *models.Update, _ []string, deps *utils.HandlerDeps) {
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
				{Text: "Button 1", CallbackData: "button_1"},
				{Text: "Button 2", CallbackData: "button_2:arg_1"},
			},
			{
				{Text: "Button 3", CallbackData: "button_3"},
			},
		},
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      u.Message.Chat.ID,
		Text:        "Welcome! Click a button",
		ReplyMarkup: kb,
	})
}

func photoHandler(ctx context.Context, b *bot.Bot, u *models.Update, _ []string, _ *utils.HandlerDeps) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      u.Message.Chat.ID,
		Text:        "Photo received!\nCaption: " + u.Message.Caption,
	})
}