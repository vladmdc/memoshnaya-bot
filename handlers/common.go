package handlers

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/vladmdc/memoshnaya-bot/models"
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

var ratePostfixes = []string{"🥇", "🥈", "🥉", "🏅"}

func from(ctx context.Context, user *tgbotapi.User) string {
	rate := ctx.Value(usrRate{}).(*models.UserRate)

	postfix := ""
	if rate != nil {
		if rate.Idx < 3 {
			postfix = " " + ratePostfixes[rate.Idx]
		}
		if rate.Idx >= 3 && rate.Idx < 10 {
			postfix = " " + ratePostfixes[3]
		}
	}

	if user.UserName != "" {
		return fmt.Sprintf("%s%s", "@"+user.UserName, postfix)
	}

	f := user.FirstName
	if user.LastName != "" {
		f = f + " " + user.LastName
	}
	return fmt.Sprintf("%s%s", f, postfix)
}
