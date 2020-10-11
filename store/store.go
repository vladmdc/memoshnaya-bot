package store

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"

	"github.com/vladmdc/memoshnaya-bot/models"
)

const (
	chatsColl = "chats"
	usersColl = "users"
)

type Store struct {
	c *firestore.Client
}

func New(c *firestore.Client) *Store {
	return &Store{c: c}
}

func (s *Store) UpsertUserToChat(ctx context.Context, chat *models.Chat, from *models.User) error {
	c := s.c.Collection(chatsColl).Doc(fmt.Sprint(chat.ID))
	err := s.c.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		if err := tx.Set(c, chat); err != nil {
			return fmt.Errorf("updating chat: %w", err)
		}

		u := c.Collection(usersColl).Doc(fmt.Sprint(from.ID))
		if err := tx.Set(u, from); err != nil {
			return fmt.Errorf("upserting user: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("upserting user to chat: %w", err)
	}

	return nil
}
