package store

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"github.com/vladmdc/memoshnaya-bot/models"
)

const (
	collChats = "chats"
	collUsers = "users"
	collPosts = "posts"
	collRates = "rates"
)

type Store struct {
	c *firestore.Client
}

func New(c *firestore.Client) *Store {
	return &Store{c: c}
}

func (s *Store) UpsertUserToChat(ctx context.Context, chat *models.Chat, from *models.User) (*models.UserRate, error) {
	c := s.c.Collection(collChats).Doc(fmt.Sprint(chat.ID))
	var r *models.UserRate
	err := s.c.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		if err := tx.Set(c, chat); err != nil {
			return fmt.Errorf("updating chat: %w", err)
		}

		u := c.Collection(collUsers).Doc(fmt.Sprint(from.ID))
		if err := tx.Set(u, from); err != nil {
			return fmt.Errorf("upserting user: %w", err)
		}

		lastRate, err := s.GetLastRate(ctx, chat.ID)
		if err != nil {
			return fmt.Errorf("getting last rate: %w", err)
		}

		r = filterUserRate(from.ID, lastRate)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("upserting user to chat: %w", err)
	}

	return r, nil
}

func filterUserRate(userID int, rate *models.Rate) *models.UserRate {
	if rate == nil {
		return nil
	}

	for i, r := range rate.UserRates {
		if r.UserID == userID {
			r.Idx = i
			return r
		}
	}
	return nil
}

func (s *Store) UpsertUser(ctx context.Context, c *models.Chat, u *models.User) error {
	_, err := s.c.Collection(collChats).Doc(fmt.Sprint(c.ID)).Collection(collUsers).Doc(fmt.Sprint(u.ID)).Set(ctx, u)
	if err != nil {
		return fmt.Errorf("upserting user: %w", err)
	}

	return nil
}

func (s *Store) AddPost(ctx context.Context, post *models.Post) error {
	_, err := s.c.Collection(collChats).
		Doc(fmt.Sprint(post.ChatID)).
		Collection(collPosts).
		Doc(fmt.Sprint(post.MessageID)).
		Set(ctx, post)
	if err != nil {
		return fmt.Errorf("creating post: %w", err)
	}

	return nil
}

func (s *Store) UpsertReaction(ctx context.Context, chatID int64, r *models.Reaction) (pos, neg int, err error) {
	c := s.c.Collection(collChats).Doc(fmt.Sprint(chatID)).Collection(collPosts).Doc(fmt.Sprint(r.MessageID))
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

func (s *Store) GetYesterdayPosts(ctx context.Context) ([]models.PostUser, error) {
	location, _ := time.LoadLocation("Europe/Moscow")
	y, m, d := time.Now().In(location).Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, location)
	yesterday := today.Add(-24 * time.Hour)

	it := s.c.CollectionGroup(collPosts).
		Where("created", ">=", yesterday).
		Where("created", "<", today).
		Documents(ctx)
	var posts []models.Post
	for {
		doc, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("iterating: %w", err)
		}

		var p models.Post
		if err := doc.DataTo(&p); err != nil {
			return nil, fmt.Errorf("parsing post: %w", err)
		}

		posts = append(posts, p)
	}

	bestPosts := make(map[int64]models.Post)
	for _, post := range posts {
		p, ok := bestPosts[post.ChatID]
		if ok && len(post.Positives)-len(post.Negatives) <= len(p.Positives)-len(p.Negatives) {
			continue
		}

		bestPosts[post.ChatID] = post
	}

	postUsers := make([]models.PostUser, 0, len(bestPosts))
	for _, p := range bestPosts {
		pu := models.PostUser{
			Post: p,
		}
		dr, err := s.c.Collection(collChats).
			Doc(fmt.Sprint(p.ChatID)).
			Collection(collUsers).
			Doc(fmt.Sprint(p.UserID)).
			Get(ctx)
		if err != nil {
			continue
		}

		if err := dr.DataTo(&pu.User); err != nil {
			continue
		}

		postUsers = append(postUsers, pu)
	}

	return postUsers, nil
}
