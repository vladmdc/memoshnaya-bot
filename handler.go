package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	firebase "firebase.google.com/go/v4"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
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

	l = zerolog.New(os.Stdout).Level(zerolog.DebugLevel).With().Timestamp().Logger()

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
	update := &tgbotapi.Update{}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		fmt.Fprint(w, "Hello World!")
		return
	}

	// todo: handle error
	h.HandleUpdate(*update)

	return
}
