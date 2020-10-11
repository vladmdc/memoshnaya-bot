package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rs/zerolog"
)

type store interface {
	UpsertUser(chat *tgbotapi.Chat, from *tgbotapi.User) error
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

func (h *Handler) HandleUpdate(u tgbotapi.Update){
	switch checkType(u) {
	case private:
		h.private(u)
	case group:
		h.group(u)
	}
}

func (h *Handler) sendDeletion(m *tgbotapi.Message) {
	deleteMsg := tgbotapi.NewDeleteMessage(m.Chat.ID, m.MessageID)
	if _, err := h.bot.Send(deleteMsg); err != nil {
		h.log.Error().Err(err).Msg("sending deletion msg")
	}
}

func checkType(u tgbotapi.Update) int {
	switch {
	case u.Message != nil && u.Message.Chat.IsPrivate(),
		u.CallbackQuery != nil && u.CallbackQuery.Message.Chat.IsPrivate():
		return private
	case u.Message != nil && (u.Message.Chat.IsGroup() || u.Message.Chat.IsSuperGroup()):
		return group
	default:
		return undefined
	}
}
