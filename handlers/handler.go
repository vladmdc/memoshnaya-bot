package handlers

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"

	"github.com/vladmdc/memoshnaya-bot/models"
)

type store interface {
	UpsertUserToChat(context.Context, *models.Chat, *models.User) error
	AddPost(context.Context, *models.Post) error
	UpsertUser(context.Context, *models.Chat, *models.User) error
	UpsertReaction(context.Context, int64, *models.Reaction) (int, int, error)
}

type Handler struct {
	st  store
	bot *tgbotapi.BotAPI
	log zerolog.Logger
}

func New(log zerolog.Logger, srv store, bot *tgbotapi.BotAPI) *Handler {
	return &Handler{st: srv, log: log, bot: bot}
}

const (
	undefined = iota
	private
	group
)

func (h *Handler) HandleUpdate(u tgbotapi.Update) {
	switch checkType(u) {
	case private:
		h.private(u)
	case group:
		h.group(u)
	}
}

func (h *Handler) sendDeletion(m *tgbotapi.Message) {
	deleteMsg := tgbotapi.NewDeleteMessage(m.Chat.ID, m.MessageID)
	resp, err := h.bot.Request(deleteMsg)
	if err != nil || !resp.Ok {
		h.log.Error().Err(err).Str("desc", resp.Description).Msg("sending deletion msg")
	}
}

func checkType(u tgbotapi.Update) int {
	switch {
	case u.Message != nil && u.Message.Chat.IsPrivate(),
		u.CallbackQuery != nil && u.CallbackQuery.Message.Chat.IsPrivate():
		return private
	case u.Message != nil && (u.Message.Chat.IsGroup() || u.Message.Chat.IsSuperGroup()),
		u.CallbackQuery != nil && (u.CallbackQuery.Message.Chat.IsGroup() || u.CallbackQuery.Message.Chat.IsSuperGroup()):
		return group
	default:
		return undefined
	}
}
