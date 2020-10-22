package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	firebase "firebase.google.com/go/v4"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"

	"github.com/vladmdc/memoshnaya-bot/handlers"
	"github.com/vladmdc/memoshnaya-bot/store"
)

var projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")

// client is a Firestore client, reused between function invocations.
var (
	bot *tgbotapi.BotAPI
	h   *handlers.Handler
	s   *store.Store
	l   zerolog.Logger
)

func init() {
	// Use the application default credentials.
	conf := &firebase.Config{ProjectID: projectID}

	// Use context.Background() because the app/client should persist across
	// invocations.
	ctx := context.Background()

	var hook SeverityHook
	l = zerolog.New(os.Stdout).Level(zerolog.DebugLevel).Hook(hook).With().Timestamp().Logger()

	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		l.Fatal().Err(err).Msg("firebase.NewApp")
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		l.Fatal().Err(err).Msg("app.Firestore")
	}

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		l.Fatal().Msg("token not set")
	}

	bot, err = tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		l.Fatal().Err(err).Msg("failed to init bot")
	}

	s = store.New(client)

	h = handlers.New(
		l.With().Str("component", "handler").Logger(),
		s,
		bot,
	)
}

func Handle(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()

	update := &tgbotapi.Update{}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		fmt.Fprint(w, "Hello World!")
		return
	}

	// todo: handle error
	h.HandleUpdate(*update)

	return
}

func YesterdayMeme(w http.ResponseWriter, r *http.Request) {
	h.YesterdayMemes()

	return
}
