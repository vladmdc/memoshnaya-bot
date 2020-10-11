package store

import (
	"cloud.google.com/go/firestore"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Store struct {
	c *firestore.Client
}

func New(c *firestore.Client) *Store {
	return &Store{c: c}
}

func (s *Store) UpsertUser(chat *tgbotapi.Chat, from *tgbotapi.User) error {
	s.c.Collection("users").Doc()
}