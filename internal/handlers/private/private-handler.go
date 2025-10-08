package private

import (
	"context"
	"strings"

	"github.com/frangi01/bbtelgo/internal/utils"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)




type HandleFunc func(ctx context.Context, b *bot.Bot, u *models.Update, _ []string, deps *utils.HandlerDeps)


var routes = map[string]HandleFunc{
	"/start": 	startHandler,
}

var callbackRoutes = map[string]HandleFunc{
	"button_1": button1Handler,
	"button_2": button2Handler,
	"button_3": button3Handler,
}



func HandlerMessage(ctx context.Context, b *bot.Bot, update *models.Update, handlerDeps *utils.HandlerDeps) {
	
	if update.Message.Text != "" {
		msg := strings.TrimSpace(update.Message.Text)

		// Extract command and args (ex: "/start foo bar")
		fields := strings.Fields(msg)
		cmd := fields[0]
		args := []string{}
		if len(fields) > 1 {
			args = fields[1:]
		}

		// Normalize: remove @botname if exist (ex: "/start@MyBot")
		if i := strings.IndexByte(cmd, '@'); i > 0 {
			cmd = cmd[:i]
		}

		// Dispatch
		if h, ok := routes[cmd]; ok {
			h(ctx, b, update, args, handlerDeps)
			return
		}
	}
	
	if update.Message.Photo != nil {
		photoHandler(ctx, b, update, nil, handlerDeps)
	}



	// enable for send msg when user send a command not in routes
	//defaultHandler(ctx, b, update, args)
}

func HandlerCallBackQuery(ctx context.Context, b *bot.Bot, update *models.Update, handlerDeps *utils.HandlerDeps) {
	cb 		:= update.CallbackQuery
	data 	:= cb.Data

	// Example format: "command:arg1:arg2"
	parts := strings.Split(data, ":")
	cmd := parts[0]
	args := []string{}
	if len(parts) > 1 {
		args = parts[1:]
	}

	if h, ok := callbackRoutes[cmd]; ok {
		h(ctx, b, update, args, handlerDeps)
	}

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: cb.ID,
	})
}


func defaultHandler(ctx context.Context, b *bot.Bot, u *models.Update, _ []string) {
	if u.Message == nil {
		return
	}

	b.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: u.Message.Chat.ID,
		Action: models.ChatActionTyping,
	})

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: u.Message.Chat.ID,
		Text:   "Command not found.",
	})
}
