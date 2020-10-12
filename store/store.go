package store

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"

	"github.com/vladmdc/memoshnaya-bot/models"
)

const (
	chatsColl     = "chats"
	usersColl     = "users"
	postsColl     = "posts"
	reactionsColl = "reactions"
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

func (s *Store) UpsertUser(ctx context.Context, c *models.Chat, u *models.User) error {
	_, err := s.c.Collection(chatsColl).Doc(fmt.Sprint(c.ID)).Collection(usersColl).Doc(fmt.Sprint(u.ID)).Set(ctx, u)
	if err != nil {
		return fmt.Errorf("upserting user: %w", err)
	}

	return nil
}

func (s *Store) AddPost(ctx context.Context, post *models.Post) error {
	_, err := s.c.Collection(chatsColl).
		Doc(fmt.Sprint(post.ChatID)).
		Collection(postsColl).
		Doc(fmt.Sprint(post.MessageID)).
		Set(ctx, post)
	if err != nil {
		return fmt.Errorf("creating post: %w", err)
	}

	return nil
}

func (s *Store) UpsertReaction(ctx context.Context, chatID int64, r *models.Reaction) (pos, neg int, err error) {
	c := s.c.Collection(chatsColl).Doc(fmt.Sprint(chatID)).Collection(postsColl).Doc(fmt.Sprint(r.MessageID))
	var positives, negatives []int
	err = s.c.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		dSnap, err := tx.Get(c)
		if err != nil {
			return fmt.Errorf("getting post: %w", err)
		}

		var post models.Post
		if err := dSnap.DataTo(&post); err != nil {
			return fmt.Errorf("parsing post: %w", err)
		}

		pos, neg = len(post.Positives), len(post.Negatives)

		if r.Type == models.Positive {
			for _, p := range post.Positives {
				if r.UserID == p {
					return nil
				}
				positives = append(positives, p)
			}

			for _, n := range post.Negatives {
				if r.UserID == n {
					continue
				}
				negatives = append(negatives, n)
			}

			positives = append(positives, r.UserID)
		}

		if r.Type == models.Negative {
			for _, n := range post.Negatives {
				if r.UserID == n {
					return nil
				}
				negatives = append(negatives, n)
			}

			for _, p := range post.Positives {
				if r.UserID == p {
					continue
				}
				positives = append(positives, p)
			}

			negatives = append(negatives, r.UserID)
		}

		err = tx.Set(c, map[string][]int{
			"positives": positives,
			"negatives": negatives,
		}, firestore.MergeAll)
		if err != nil {
			return fmt.Errorf("updating reacts: %w", err)
		}

		pos, neg = len(positives), len(negatives)

		return nil
	})

	if err != nil {
		return 0, 0, fmt.Errorf("updating reactions: %w", err)
	}

	return pos, neg, nil
}
