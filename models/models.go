package models

import (
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Chat struct {
	ID         int64  `firestore:"id"`
	Type       string `firestore:"type,omitempty"`
	Title      string `firestore:"title,omitempty"`
	UserName   string `firestore:"user_name,omitempty"`
	FirstName  string `firestore:"first_name,omitempty"`
	LastName   string `firestore:"last_name,omitempty"`
	InviteLink string `firestore:"invite_link,omitempty"`
}

func NewChat(c *tgbotapi.Chat) *Chat {
	return &Chat{
		ID:         c.ID,
		Type:       c.Type,
		Title:      c.Title,
		UserName:   c.UserName,
		FirstName:  c.FirstName,
		LastName:   c.LastName,
		InviteLink: c.InviteLink,
	}
}

type User struct {
	ID        int    `firestore:"id"`
	FirstName string `firestore:"first_name,omitempty"`
	LastName  string `firestore:"last_name,omitempty"`
	UserName  string `firestore:"username,omitempty"`
	IsBot     bool   `firestore:"is_bot,omitempty"`
}

func NewUser(u *tgbotapi.User) *User {
	return &User{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		UserName:  u.UserName,
		IsBot:     u.IsBot,
	}
}

type Post struct {
	FileID    string    `firestore:"file_id,omitempty"`
	MessageID int       `firestore:"message_id,omitempty"`
	ChatID    int64     `firestore:"chat_id,omitempty"`
	UserID    int       `firestore:"user_id,omitempty"`
	Positives []int `firestore:"positives,omitempty"`
	Negatives []int `firestore:"negatives,omitempty"`
	Created   time.Time `firestore:"created,omitempty"`
}

func NewPost(p *Post) *Post {
	return &Post{
		FileID:    p.FileID,
		MessageID: p.MessageID,
		ChatID:    p.ChatID,
		UserID:    p.UserID,
		Created:   time.Now(),
	}
}

const (
	Positive = 1
	Negative = 2
)

type Reaction struct {
	MessageID int       `firestore:"message_id,omitempty"`
	UserID    int       `firestore:"user_id,omitempty"`
	Type      int       `firestore:"type,omitempty"`
}


type PostUser struct {
	Post
	User
}
