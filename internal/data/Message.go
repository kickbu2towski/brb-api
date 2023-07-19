package data

import (
	"time"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Message struct {
	Type string `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

type MessageDM struct {
	ID string `json:"id"`
	Content string `json:"content"`
	ReceiverID string `json:"receiver_id"`
	SenderID string `json:"sender_id"`
	CreatedAt time.Time `json:"created_at"`
	IsDeleted bool `json:"is_deleted"`
	IsEdited bool `json:"is_edited"`
	ReplyToID string `json:"reply_to_id"`
}

type MessageModel struct {
	Pool *pgxpool.Pool
}

func (m *MessageModel) Insert(message *MessageDM) error {
	return nil
}

