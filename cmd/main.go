package main

import (
	"context"
	"os"

	"cloud.google.com/go/firestore"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/vladmdc/memoshnaya-bot/handlers"
)

type RequestBody struct {
	HttpMethod string `json:"httpMethod"`
	Body       string `json:"body"`
}

func main() {
	initLogger(true)
	ctx := context.Background()
	c := createClient(ctx)
	defer func() { _ = c.Close() }()

	_, _, err := c.Collection("users").Add(ctx, map[string]string{
		"msg": "Vlados",
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to insert")
	}

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal().Msg("token not set")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal().Err(err).Msg("create bot failed")
	}

	h := handlers.New(
		log.With().Str("component", "handler").Logger(),
		nil,
		bot,
	)

	log.Info().Str("bot", bot.Self.UserName).Msg("authorized on account")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		h.HandleUpdate(update)
	}
}

func initLogger(local bool) zerolog.Logger {
	log.Logger = zerolog.New(os.Stdout).Level(zerolog.DebugLevel).With().Timestamp().Logger()

	if local {
		log.Logger = log.Logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	return log.Logger
}

func createClient(ctx context.Context) *firestore.Client {
	// Sets your Google Cloud Platform project ID.
	projectID := "telegram-test-291719"

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create client")
	}
	// Close client when done with
	// defer client.Close()
	return client
}
