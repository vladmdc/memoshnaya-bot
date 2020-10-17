package handlers

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	positiveReactionData = "p"
	negativeReactionData = "n"
)

func newReactionsKeyboard(positive, negative int) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("👍 %d", positive), positiveReactionData),
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("👎 %d", negative), negativeReactionData),
		))
}

func from(user *tgbotapi.User) string {
	if user.UserName != "" {
		return "@" + user.UserName
	}

	f := user.FirstName
	if user.LastName != "" {
		f = f + " " + user.LastName
	}
	return f
}
