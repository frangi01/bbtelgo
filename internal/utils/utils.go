package utils

import (
	"bytes"
	"encoding/json"

	"github.com/frangi01/bbtelgo/internal/config"
	"github.com/frangi01/bbtelgo/internal/db"
	"github.com/frangi01/bbtelgo/internal/i18n"
	"github.com/frangi01/bbtelgo/internal/logx"
	"github.com/go-telegram/bot/models"
)

func JSON(v any, pretty bool, escapeHTML bool) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(escapeHTML)
	if pretty {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(v); err != nil {
		return nil, err
	}

	b := bytes.TrimRight(buf.Bytes(), "\n")
	return b, nil
}

type HandlerDeps struct {
	Logger         	*logx.Logger
	Cfg            	config.Config
	RepositoryList 	*db.RepositoryList
	Cache          	*db.CacheClient
	I18n			*i18n.Bundle
}

func NewDeps(
	logger *logx.Logger,
	cfg config.Config,
	repositoryList *db.RepositoryList,
	cache *db.CacheClient,
	i18n *i18n.Bundle,
) *HandlerDeps {
	return &HandlerDeps{
		Logger:         logger,
		Cfg:            cfg,
		RepositoryList: repositoryList,
		Cache:          cache,
		I18n: i18n,
	}
}


func ChatIDFromUpdate(u *models.Update) int64 {
	if u.Message != nil {
		return u.Message.Chat.ID
	}
	if u.CallbackQuery != nil {
		return u.CallbackQuery.Message.Message.Chat.ID
	}
	return 0
}
