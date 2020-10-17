package handlers

import (
	"fmt"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/vladmdc/memoshnaya-bot/models"
)

func (h *Handler) sendBestMeme(p *models.Post, u *models.User) error {

	text := fmt.Sprintf("Лучший вчерашний пост\n%s топчик", fromUser(u))

	msg := tgbotapi.NewMessage(p.ChatID, text)
	msg.ReplyToMessageID = p.MessageID
	if _, err := h.bot.Send(msg); err != nil {
		return fmt.Errorf("sending msg: %w", err)
	}

	return nil
}

func fromUser(u *models.User) string {
	if u.UserName != "" {
		return "@" + u.UserName
	}

	f := u.FirstName
	if u.LastName != "" {
		f = f + " " + u.LastName
	}
	return f
}
