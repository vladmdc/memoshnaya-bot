package models

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

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
