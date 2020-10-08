package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")

// client is a Firestore client, reused between function invocations.
var client *firestore.Client

func init() {
	// Use the application default credentials.
	conf := &firebase.Config{ProjectID: projectID}

	// Use context.Background() because the app/client should persist across
	// invocations.
	ctx := context.Background()

	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatalf("firebase.NewApp: %v", err)
	}

	client, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalf("app.Firestore: %v", err)
	}
}

func Handle(w http.ResponseWriter, r *http.Request) {
	update := &tgbotapi.Update{}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		fmt.Fprint(w, "Hello World!")
		return
	}

	if update.Message == nil {
		fmt.Println("msg is nil", update)
		return
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		fmt.Println(err)
		return
	}

	_, _, err = client.Collection("users").Add(context.Background(), map[string]string{
		"msg": update.Message.Text,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

	if _, err := bot.Send(msg); err != nil {
		fmt.Println(err)
		return
	}

	return
}

