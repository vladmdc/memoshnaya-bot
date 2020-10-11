package handlers

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/vladmdc/memoshnaya-bot/models"
)

func (h *Handler) group(u tgbotapi.Update) {
	switch {
	case u.Message != nil:
		h.log.Debug().
			Str("type", "group").
			Str("username", u.Message.From.UserName).
			Msg("new message")
		if err := h.groupMsg(u.Message); err != nil {
			h.log.Error().Err(err).Str("type", "group").Msg("failed to handle msg")
		}
	}
}

func (h *Handler) groupMsg(m *tgbotapi.Message) error {
	go h.sendDeletion(m)

	err := h.st.UpsertUserToChat(
		context.Background(),
		models.NewChat(m.Chat),
		models.NewUser(m.From),
	)
	if err != nil {
		return fmt.Errorf("upserting user: %w", err)
	}

	switch {
	case m.Photo != nil:
		return h.groupPhoto(m)
	case m.Text != "":
		return h.groupText(m)
	default:
		return nil
	}
}

func (h *Handler) groupText(m *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(m.Chat.ID, m.Text)
	if _, err := h.bot.Send(msg); err != nil {
		return fmt.Errorf("sending msg: %w", err)
	}
	return nil
}

func (h *Handler) groupPhoto(m *tgbotapi.Message) error {
	return nil
}
