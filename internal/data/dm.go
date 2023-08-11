package data

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DM struct {
	id int
}

type DMModel struct {
	Pool *pgxpool.Pool
}

func (m *DMModel) GetDMListForUser(ctx context.Context, userID int) ([]*BasicUserResp, error) {
	stmt := `
		SELECT u.id, u.username, u.avatar
		  FROM dm_participants AS dp1
		  JOIN dm_participants AS dp2 ON dp1.dm_id = dp2.dm_id AND dp1.participant_id != dp2.participant_id
		  JOIN follow_relations AS f1 ON dp1.participant_id = f1.following_id AND dp2.participant_id = f1.follower_id
		  JOIN follow_relations AS f2 ON dp1.participant_id = f2.follower_id AND dp2.participant_id = f2.following_id
		  JOIN users AS u ON u.id = dp2.participant_id
		  JOIN messages AS m ON m.dm_id = dp2.dm_id
		WHERE dp1.participant_id = $1
		GROUP BY dp2.dm_id, u.id
		HAVING COUNT(m.id) > 0;
  `

	users := make([]*BasicUserResp, 0)
	rows, err := m.Pool.Query(ctx, stmt, userID)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return users, nil
		default:
			return nil, err
		}
	}

	for rows.Next() {
		var user BasicUserResp
		err := rows.Scan(&user.ID, &user.Username, &user.Avatar)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (m *DMModel) GetDMForParticipants(ctx context.Context, participants []int) (int, error) {
	stmt := `SELECT dp1.dm_id
		FROM dm_participants dp1
		JOIN dm_participants dp2 ON dp1.dm_id = dp2.dm_id
		WHERE dp1.participant_id = $1
	  AND dp2.participant_id = $2;
	`

	var (
		dmID      int
		isInitial bool
	)

	err := m.Pool.QueryRow(ctx, stmt, participants[0], participants[1]).Scan(&dmID)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			isInitial = true
		default:
			return dmID, err
		}
	}

	if !isInitial {
		return dmID, nil
	}

	tx, err := m.Pool.Begin(ctx)
	if err != nil {
		return dmID, err
	}
	defer tx.Rollback(ctx)

	stmt = `INSERT INTO dms(id) VALUES(DEFAULT) RETURNING id`
	err = tx.QueryRow(ctx, stmt).Scan(&dmID)
	if err != nil {
		return dmID, err
	}

	stmt = `INSERT INTO dm_participants(dm_id, participant_id) VALUES($1, $2), ($1, $3);`
	_, err = tx.Exec(ctx, stmt, dmID, participants[0], participants[1])
	if err != nil {
		return dmID, err
	}

	tx.Commit(ctx)
	return dmID, nil
}
