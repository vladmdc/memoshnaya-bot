package handlers

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

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
	case u.CallbackQuery != nil:
		h.log.Debug().
			Str("type", "group").
			Msg("new callback")
		if err := h.groupCallback(u.CallbackQuery); err != nil {
			h.log.Error().Err(err).Str("type", "group").Msg("failed to handle cb")
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
	case m.Video != nil:
		return h.groupVideo(m)
	case m.Animation != nil:
		return h.groupAnimation(m)
	default:
		return nil
	}
}

func (h *Handler) groupPhoto(m *tgbotapi.Message) error {
	fileId := m.Photo[0].FileID
	msg := tgbotapi.NewPhotoShare(m.Chat.ID, fileId)
	msg.ReplyMarkup = newReactionsKeyboard(0, 0)
	msg.DisableNotification = true
	msg.Caption = "from: " + from(m.From)
	if m.Caption != "" {
		msg.Caption = m.Caption + "\n" + msg.Caption
	}
	if m.ReplyToMessage != nil {
		msg.ReplyToMessageID = m.ReplyToMessage.MessageID
	}

	return h.newUserMediaPost(msg, m, fileId)
}

func (h *Handler) groupVideo(m *tgbotapi.Message) error {
	fileId := m.Video.FileID
	msg := tgbotapi.NewVideoShare(m.Chat.ID, fileId)
	msg.ReplyMarkup = newReactionsKeyboard(0, 0)
	msg.DisableNotification = true
	msg.Caption = "from: " + from(m.From)
	if m.Caption != "" {
		msg.Caption = fmt.Sprintf("%s\n%s", m.Caption, msg.Caption)
	}
	if m.ReplyToMessage != nil {
		msg.ReplyToMessageID = m.ReplyToMessage.MessageID
	}

	return h.newUserMediaPost(msg, m, fileId)
}

func (h *Handler) groupAnimation(m *tgbotapi.Message) error {
	fileId := m.Animation.FileID
	msg := tgbotapi.NewAnimationShare(m.Chat.ID, fileId)
	msg.ReplyMarkup = newReactionsKeyboard(0, 0)
	msg.DisableNotification = true
	msg.Caption = "from: " + from(m.From)
	if m.Caption != "" {
		msg.Caption = fmt.Sprintf("%s\n%s", m.Caption, msg.Caption)
	}
	if m.ReplyToMessage != nil {
		msg.ReplyToMessageID = m.ReplyToMessage.MessageID
	}

	return h.newUserMediaPost(msg, m, fileId)
}

func (h *Handler) newUserMediaPost(msg tgbotapi.Chattable, m *tgbotapi.Message, fileID string) error {
	sent, err := h.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("sending new post: %w", err)
	}

	err = h.st.AddPost(
		context.Background(),
		models.NewPost(&models.Post{
			FileID:    fileID,
			MessageID: sent.MessageID,
			ChatID:    m.Chat.ID,
			UserID:    m.From.ID,
		}),
	)
	if err != nil {
		return fmt.Errorf("saving post: %w", err)
	}

	return nil
}
