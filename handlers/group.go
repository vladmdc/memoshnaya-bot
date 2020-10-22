package handlers

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/vladmdc/memoshnaya-bot/models"
)

func (h *Handler) group(u tgbotapi.Update) {
	switch {
	case u.Message != nil:
		h.log.Debug().
			Str("type", "group").
			Str("username", u.Message.From.UserName).
			Int("message-id", u.Message.MessageID).
			Int64("chat-id", u.Message.Chat.ID).
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

	h.log.Debug().Msg("upserting user")
	err := h.st.UpsertUserToChat(
		context.Background(),
		models.NewChat(m.Chat),
		models.NewUser(m.From),
	)
	if err != nil {
		return fmt.Errorf("upserting user: %w", err)
	}

	h.log.Debug().
		Str("username", m.From.UserName).
		Int("message-id", m.MessageID).
		Int64("chat-id", m.Chat.ID).
		Msg("group message")

	switch {
	case m.Photo != nil:
		return h.groupPhoto(m)
	case m.Video != nil:
		return h.groupVideo(m)
	case m.Animation != nil:
		return h.groupAnimation(m)
	case m.Entities != nil:
		return h.groupEntity(m)
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

	h.log.Debug().
		Str("username", m.From.UserName).
		Int("message-id", m.MessageID).
		Int64("chat-id", m.Chat.ID).
		Msg("group photo")

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

	h.log.Debug().
		Str("username", m.From.UserName).
		Int("message-id", m.MessageID).
		Int64("chat-id", m.Chat.ID).
		Msg("group video")

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

	h.log.Debug().
		Str("username", m.From.UserName).
		Int("message-id", m.MessageID).
		Int64("chat-id", m.Chat.ID).
		Msg("group animation")

	return h.newUserMediaPost(msg, m, fileId)
}

func (h *Handler) newUserMediaPost(msg tgbotapi.Chattable, m *tgbotapi.Message, fileID string) error {
	h.log.Debug().Msg("sending new post")
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

func (h *Handler) groupEntity(m *tgbotapi.Message) error {
	if m.Entities == nil {
		return nil
	}

	e := m.Entities[0]
	if e.Type != "url" {
		h.log.Warn().Str("text", m.Text).Str("entity-type", e.Type).Msg("undefined entity")
		return nil
	}

	eURL := string([]rune(m.Text)[e.Offset : e.Offset+e.Length])
	urlName := "Ссылочка"
	switch {
	case strings.Contains(eURL, "youtube") || strings.Contains(eURL, "youtu.be"):
		urlName = "Видос"
	case strings.Contains(eURL, "coub"):
		urlName = "coub"
	case strings.Contains(eURL, "twitter.com"):
		urlName = "Твит"
	}

	linkText := fmt.Sprintf(
		"[%s](%s) _от_ %s",
		urlName,
		eURL,
		from(m.From),
	)
	caption := string([]rune(m.Text)[:e.Offset])
	if e.Offset == 0 {
		caption = string([]rune(m.Text)[e.Offset+e.Length:])
	}
	if caption != "" {
		linkText = fmt.Sprintf("%s\n%s", caption, linkText)
	}

	msg := tgbotapi.NewMessage(m.Chat.ID, linkText)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.DisableNotification = true
	msg.ReplyMarkup = newReactionsKeyboard(0, 0)
	if m.ReplyToMessage != nil {
		msg.ReplyToMessageID = m.ReplyToMessage.MessageID
	}
	sentMsg, _ := h.bot.Send(msg)

	err := h.st.AddPost(
		context.Background(),
		models.NewPost(&models.Post{
			MessageID: sentMsg.MessageID,
			ChatID:    m.Chat.ID,
			UserID:    m.From.ID,
		}),
	)
	if err != nil {
		return fmt.Errorf("saving post: %w", err)
	}

	h.log.Info().Msg("new post sent")

	return nil
}
