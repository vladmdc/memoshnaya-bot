package store

import (
	"context"
	"fmt"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"github.com/vladmdc/memoshnaya-bot/models"
)

const (
	ratingDays      = 14
	minPostsForRate = 3
)

func (s *Store) CalcRates(ctx context.Context) error {
	chats, err := s.getChats(ctx)
	if err != nil {
		return fmt.Errorf("getting chats: %w", err)
	}

	for _, c := range chats {
		posts, err := s.getLastDaysPosts(ctx, c.ID, ratingDays)
		if err != nil {
			return fmt.Errorf("getting posts for chat (id: %d): %w", c.ID, err)
		}

		if len(posts) == 0 {
			continue
		}

		rate := calcRate(posts)
		if rate == nil {
			continue
		}

		if err = s.addRate(ctx, c.ID, rate); err != nil {
			return fmt.Errorf("creating new rate: %w", err)
		}
	}

	return nil
}

type userReactions struct {
	count     int
	reactions int
}

func calcRate(posts []*models.Post) *models.Rate {
	if len(posts) == 0 {
		return nil
	}

	reactionsMap := make(map[int]userReactions)
	for _, p := range posts {
		data := reactionsMap[p.UserID]
		data.count++
		data.reactions += len(p.Positives) - len(p.Negatives)
		reactionsMap[p.UserID] = data
	}

	var userRates []*models.UserRate
	for id, p := range reactionsMap {
		if p.count < minPostsForRate {
			continue
		}

		userRates = append(userRates, &models.UserRate{
			UserID: id,
			Rate:   float64(p.reactions) / float64(p.count),
		})
	}

	sort.Slice(userRates, func(i, j int) bool {
		return userRates[i].Rate > userRates[j].Rate
	})

	if len(userRates) > 10 {
		userRates = userRates[:10]
	}

	return &models.Rate{
		UserRates: userRates,
		Created:   time.Now(),
	}
}

func (s *Store) getLastDaysPosts(ctx context.Context, chatID int64, days int) ([]*models.Post, error) {
	location, _ := time.LoadLocation("Europe/Moscow")
	from := time.Now().In(location).AddDate(0, 0, -1*days)

	it := s.c.Collection(collChats).
		Doc(fmt.Sprint(chatID)).
		Collection(collPosts).
		Where("created", ">=", from).
		Documents(ctx)

	var posts []*models.Post
	for {
		doc, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("iterating: %w", err)
		}

		p := &models.Post{}
		if err := doc.DataTo(p); err != nil {
			return nil, fmt.Errorf("parsing post: %w", err)
		}

		posts = append(posts, p)
	}

	return posts, nil
}

func (s *Store) getChats(ctx context.Context) ([]*models.Chat, error) {
	iter := s.c.Collection(collChats).Documents(ctx)

	var chats []*models.Chat
	for {
		doc, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, fmt.Errorf("iterating: %w", err)
		}

		c := &models.Chat{}
		if err := doc.DataTo(c); err != nil {
			return nil, fmt.Errorf("parsing chat: %w", err)
		}

		chats = append(chats, c)
	}

	return chats, nil
}

func (s *Store) addRate(ctx context.Context, chatID int64, rate *models.Rate) error {
	_, _, err := s.c.Collection(collChats).Doc(fmt.Sprint(chatID)).Collection(collRates).Add(ctx, rate)
	if err != nil {
		return fmt.Errorf("adding rate to chat (id: %d): %w", chatID, err)
	}

	return nil
}

func (s *Store) GetLastRate(ctx context.Context, chatID int64) (*models.Rate, error) {
	it := s.c.Collection(collChats).Doc(fmt.Sprint(chatID)).
		Collection(collRates).OrderBy("created", firestore.Desc).
		Limit(1).Documents(ctx)

	doc, err := it.Next()
	if err != nil {
		if err == iterator.Done {
			return nil, nil
		}
		return nil, fmt.Errorf("iterating: %w", err)
	}

	r := &models.Rate{}
	if err = doc.DataTo(r); err != nil {
		return nil, fmt.Errorf("parsing rate: %w", err)
	}

	return r, nil
}
