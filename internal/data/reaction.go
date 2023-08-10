package data

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Reaction struct {
	id           int
	Participants []string
}

type ReactionModel struct {
	Pool *pgxpool.Pool
}

func (m *ReactionModel) Insert(ctx context.Context, reaction, msgID, userID string) error {
	stmt := `INSERT INTO reactions(reaction, message_id, user_id) VALUES($1, $2, $3)`
	_, err := m.Pool.Exec(ctx, stmt, reaction, msgID, userID)
	return err
}

func (m *ReactionModel) Delete(ctx context.Context, reaction, msgID, userID string) error {
	stmt := `DELETE FROM reactions WHERE reaction = $1 AND message_id = $2 AND user_id = $3`
	_, err := m.Pool.Exec(ctx, stmt, reaction, msgID, userID)
	return err
}
