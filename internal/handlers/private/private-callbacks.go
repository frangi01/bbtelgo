package private

import (
	"context"

	"github.com/frangi01/bbtelgo/internal/utils"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)



func button1Handler(ctx context.Context, b *bot.Bot, u *models.Update, _ []string, deps *utils.HandlerDeps) {

	b.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: u.CallbackQuery.Message.Message.Chat.ID,
		Action: models.ChatActionTyping,
	})

	
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      u.CallbackQuery.Message.Message.Chat.ID,
		Text:        "Button 1 clicked",
	})
}


func button2Handler(ctx context.Context, b *bot.Bot, u *models.Update, args []string, deps *utils.HandlerDeps) {

	b.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: u.CallbackQuery.Message.Message.Chat.ID,
		Action: models.ChatActionTyping,
	})

	
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      u.CallbackQuery.Message.Message.Chat.ID,
		Text:        "Button 2 clicked with args " + args[0],
	})
}

func button3Handler(ctx context.Context, b *bot.Bot, u *models.Update, args []string, deps *utils.HandlerDeps) {

	b.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: u.CallbackQuery.Message.Message.Chat.ID,
		Action: models.ChatActionTyping,
	})

	
	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		MessageID: u.CallbackQuery.Message.Message.ID,
		ChatID:      u.CallbackQuery.Message.Message.Chat.ID,
		Text:        "Button 3 clicked",
	})
}