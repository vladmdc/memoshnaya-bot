package handlers

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/vladmdc/memoshnaya-bot/models"
)

var (
	positiveEmoji   = []string{"👌", "💪", "👍", "🤘", "🔥", "🏄"}
	negativeEmoji   = []string{"🤦‍♂️", "🤦‍♀️", "🥴", "💩", "👹"}
	positiveAnswers = []string{"Топчик", "Хорошо", "Кайф", "Огонь", "Пушка"}
	negativeAnswers = []string{"Не очень", "Э:)"}
)

func (h *Handler) groupCallback(q *tgbotapi.CallbackQuery) error {
	if q == nil {
		return fmt.Errorf("callback is nil")
	}

	cb := tgbotapi.NewCallback(q.ID, randomCallbackAnswerText(q.Data))
	resp, err := h.bot.Request(cb)
	if err != nil || !resp.Ok {
		return fmt.Errorf("sending callback: %w", err)
	}

	ctx := context.Background()

	t := models.Positive
	if q.Data == negativeReactionData {
		t = models.Negative
	}
	pos, neg, err := h.st.UpsertReaction(ctx, q.Message.Chat.ID, &models.Reaction{
		MessageID: q.Message.MessageID,
		UserID:    q.From.ID,
		Type:      t,
	})
	if err != nil {
		return fmt.Errorf("adding reaction: %w", err)
	}

	newKb := newReactionsKeyboard(pos, neg)
	_, err = h.bot.Send(tgbotapi.NewEditMessageReplyMarkup(q.Message.Chat.ID, q.Message.MessageID, newKb))
	if err != nil && !strings.Contains(err.Error(), "Bad Request: message is not modified") {
		return fmt.Errorf("could not update reply markup: %w", err)
	}

	return nil
}

func randomCallbackAnswerText(data string) string {
	rand.Seed(time.Now().Unix())

	if data == positiveReactionData {
		return positiveAnswers[rand.Intn(len(positiveAnswers))] + " " + positiveEmoji[rand.Intn(len(positiveEmoji))]
	}
	return negativeAnswers[rand.Intn(len(negativeAnswers))] + " " + negativeEmoji[rand.Intn(len(negativeEmoji))]
}
