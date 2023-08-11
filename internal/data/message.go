package data

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/guregu/null.v4"
)

type Message struct {
	ID        string      `json:"id"`
	Content   string      `json:"content"`
	DmID      int         `json:"dm_id"`
	UserID    int      `json:"user_id,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	IsDeleted bool        `json:"is_deleted"`
	IsEdited  bool        `json:"is_edited"`
	ReplyToID null.String `json:"reply_to_id"`
}

type MessageResp struct {
	Message
	User      BasicUserResp       `json:"user"`
	Reactions map[string][]int `json:"reactions"`
}

type MessageModel struct {
	Pool *pgxpool.Pool
}

func (m *MessageModel) GetMessages(ctx context.Context, dmID int) ([]*MessageResp, error) {
	stmt := `
	  SELECT 
		sq3.id, sq3.content, sq3.dm_id, sq3.created_at,
		sq3.is_deleted, sq3.is_edited, sq3.reply_to_id,
		sq3.user_id, sq3.username, sq3.avatar, sq3.reactions
		FROM (
			SELECT m.id, m.content, m.dm_id, m.created_at, 
			m.is_deleted, m.is_edited, m.reply_to_id,
			m.user_id, u.username, u.avatar, reactions
			FROM messages m
			JOIN users u ON m.user_id = u.id
			LEFT JOIN (
				SELECT message_id, json_object_agg(reaction, user_ids) AS reactions FROM
				(SELECT message_id, reaction, json_agg(r.user_id) AS user_ids FROM reactions  r
				GROUP BY reaction, message_id) AS sq1 GROUP BY message_id 
			) sq2 ON sq2.message_id = m.id 
			WHERE dm_id = $1
			ORDER BY created_at DESC
			LIMIT 50
		) AS sq3
		ORDER BY sq3.created_at ASC
  `

	messages := make([]*MessageResp, 0)
	rows, err := m.Pool.Query(ctx, stmt, dmID)
	if err != nil {
		return messages, err
	}

	for rows.Next() {
		var message MessageResp
		err := rows.Scan(
			&message.ID,
			&message.Content,
			&message.DmID,
			&message.CreatedAt,
			&message.IsDeleted,
			&message.IsEdited,
			&message.ReplyToID,
			&message.User.ID,
			&message.User.Username,
			&message.User.Avatar,
			&message.Reactions,
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

func (m *MessageModel) GetMessage(ctx context.Context, id string, userID int) (*MessageResp, error) {
	var withUserID int
	if userID == -1 {
		withUserID = 1
	}

	stmt := `
	  SELECT 
			m.id, 
			m.content, 
			m.dm_id,
			m.created_at, 
			m.is_deleted, 
			m.is_edited, 
			m.reply_to_id,
			u.id AS user_id, 
			u.username, 
			u.avatar, 
			sq2.reactions
		FROM messages m
		JOIN users u ON u.id = m.user_id
		LEFT JOIN (
			SELECT message_id, json_object_agg(reaction, user_ids) AS reactions FROM 
				(SELECT r.message_id, r.reaction, json_agg(r.user_id) AS user_ids FROM reactions r
				  GROUP BY reaction, message_id) sq1 GROUP BY message_id
		) sq2 ON m.id = message_id
		WHERE m.id = $1 AND (u.id = $2 OR 1 = $3)`

	var message MessageResp
	err := m.Pool.QueryRow(ctx, stmt, id, userID, withUserID).Scan(
		&message.ID,
		&message.Content,
		&message.DmID,
		&message.CreatedAt,
		&message.IsDeleted,
		&message.IsEdited,
		&message.ReplyToID,
		&message.User.ID,
		&message.User.Username,
		&message.User.Avatar,
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

func (m *MessageModel) UpdateMessage(ctx context.Context, id string, msg *MessageResp) error {
	args := []any{
		msg.Content,
		msg.IsDeleted,
		msg.IsEdited,
		id,
	}
	stmt := `
	  UPDATE messages
		SET content = $1, is_deleted = $2, is_edited = $3
		WHERE id = $4
	`
	_, err := m.Pool.Exec(ctx, stmt, args...)
	return err
}
