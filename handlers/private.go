package handlers

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (h *Handler) private(u tgbotapi.Update) {
	switch {
	case u.Message != nil:
		h.log.Debug().
			Str("type", "private").
			Str("username", u.Message.From.UserName).
			Msg("new message")
		if err := h.privateMsg(u.Message); err != nil {
			h.log.Error().Err(err).Str("type", "private").Msg("failed to handle private msg")
		}
	}
}

func (h *Handler) privateMsg(m *tgbotapi.Message) error {
	go h.sendDeletion(m)
	msg := tgbotapi.NewMessage(m.Chat.ID, m.Text)
	if _, err := h.bot.Send(msg); err != nil {
		return fmt.Errorf("sending msg: %w", err)
	}
	return nil
}
