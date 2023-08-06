package data

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/guregu/null.v4"
)

type Message struct {
	ID        string              `json:"id"`
	Content   string              `json:"content"`
	DmID      int                 `json:"dm_id"`
	UserID    string              `json:"user_id,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
	IsDeleted bool                `json:"is_deleted"`
	IsEdited  bool                `json:"is_edited"`
	ReplyToID null.String         `json:"reply_to_id"`
	Reactions map[string][]string `json:"reactions"`
}

type MessagesResp struct {
	Message
	User BasicUserResp `json:"user"`
}

type MessageModel struct {
	Pool *pgxpool.Pool
}

func (m *MessageModel) GetMessages(ctx context.Context, dmID int) ([]*MessagesResp, error) {
	stmt := `
	  SELECT 
		sq.id, sq.content, sq.dm_id, sq.created_at,
		sq.is_deleted, sq.is_edited, sq.reply_to_id, sq.reactions,
		sq.user_id, sq.username, sq.avatar
		FROM (
			SELECT m.id, m.content, m.dm_id, m.created_at, 
			m.is_deleted, m.is_edited, m.reply_to_id, m.reactions,
			m.user_id, u.username, u.avatar
			FROM messages m
			JOIN users u ON m.user_id = u.id
			WHERE dm_id = $1
			ORDER BY created_at DESC
			LIMIT 50
		) AS sq
		ORDER BY sq.created_at ASC;
  `

	messages := make([]*MessagesResp, 0)
	rows, err := m.Pool.Query(ctx, stmt, dmID)
	if err != nil {
		return messages, err
	}

	for rows.Next() {
		var message MessagesResp

		err := rows.Scan(
			&message.ID,
			&message.Content,
			&message.DmID,
			&message.CreatedAt,
			&message.IsDeleted,
			&message.IsEdited,
			&message.ReplyToID,
			&message.Reactions,
			&message.User.ID,
			&message.User.Username,
			&message.User.Avatar,
		)
		if err != nil {
			return messages, err
		}

		messages = append(messages, &message)
	}

	err = rows.Err()
	if err != nil {
		return messages, err
	}

	return messages, nil
}

func (m *MessageModel) GetMessage(ctx context.Context, id string) (*Message, error) {
	stmt := `
	  SELECT 
			id, 
			content, 
			dm_id,
			user_id, 
			created_at, 
			is_deleted, 
			is_edited, 
			reply_to_id, 
			reactions
		FROM messages
		WHERE id = $1`

	var message Message
	err := m.Pool.QueryRow(ctx, stmt, id).Scan(
		&message.ID,
		&message.Content,
		&message.DmID,
		&message.UserID,
		&message.CreatedAt,
		&message.IsDeleted,
		&message.IsEdited,
		&message.ReplyToID,
		&message.Reactions,
	)

	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (m *MessageModel) InsertMessage(ctx context.Context, msg *Message) error {
	args := []any{
		msg.ID,
		msg.Content,
		msg.DmID,
		msg.UserID,
		msg.CreatedAt,
		msg.ReplyToID,
	}
	stmt := `INSERT INTO messages(id, content, dm_id, user_id, created_at, reply_to_id) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := m.Pool.Exec(ctx, stmt, args...)
	return err
}

func (m *MessageModel) UpdateMessage(ctx context.Context, id string, msg *Message) error {
	args := []any{
		msg.Content,
		msg.IsDeleted,
		msg.IsEdited,
		msg.Reactions,
		id,
	}
	stmt := `
	  UPDATE messages
		SET content = $1, is_deleted = $2, is_edited = $3, reactions = $4
		WHERE id = $5
	`
	_, err := m.Pool.Exec(ctx, stmt, args...)
	return err
}
